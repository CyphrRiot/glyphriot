package internal

import (
	"fmt"
	"strings"
)

// EncodeWordsVerified encodes the provided BIP-39 words to glyph tokens and then
// immediately verifies a full round-trip by decoding those glyphs back to words
// using the same effective key. The comparison is order-sensitive and exact.
// If verification fails for any reason, an error is returned and no glyphs are produced.
//
// Parameters:
//   - words:      input seed words (12 or 24 recommended). They will be normalized (lowercased, trimmed).
//   - index:      map of lowercased word -> index in wordsList
//   - wordsList:  canonical BIP-39 list (length must be Total, e.g., 2048)
//   - key:        the user key/passphrase (can be empty for identity order)
//   - policy:     key policy for KDF/enforcement (Argon2id defaults recommended)
//
// Returns:
//   - glyphTokens: encoded glyphs (without separators)
//   - error:       non-nil if encode or verify fails
func EncodeWordsVerified(words []string, index map[string]int, wordsList []string, key string, policy KeyPolicy) ([]string, error) {
	if len(wordsList) != Total {
		return nil, fmt.Errorf("wordsList length must be %d, got %d", Total, len(wordsList))
	}
	normalized := normalizeWords(words)
	if len(normalized) == 0 {
		return nil, fmt.Errorf("no words provided")
	}

	// Determine enforcement target (12 -> 128 bits; 24 -> 256 bits; default 128).
	minBits := MinBitsForContext(len(normalized))

	// Compute effective key material according to policy (KDF/enforcement).
	effKey := ""
	if strings.TrimSpace(key) != "" {
		seed32, err := MustEffectiveKeyMaterial(key, minBits, policy)
		if err != nil {
			return nil, err
		}
		// Use the derived 32-byte seed as the key to the permutation engine.
		effKey = string(seed32[:])
	}

	// Encode words -> glyphs
	glyphs, err := EncodeWords(normalized, index, wordsList, effKey)
	if err != nil {
		return nil, fmt.Errorf("encode failed: %w", err)
	}

	// Decode glyphs -> words and verify order-sensitive equality
	decoded, err := DecodeGlyphTokens(glyphs, wordsList, effKey)
	if err != nil {
		return nil, fmt.Errorf("decode failed: %w", err)
	}
	if len(decoded) != len(normalized) {
		return nil, fmt.Errorf("round-trip mismatch: decoded length %d != %d", len(decoded), len(normalized))
	}
	for i := range normalized {
		if decoded[i] != normalized[i] {
			return nil, fmt.Errorf("round-trip mismatch at position %d: have %q, want %q", i, decoded[i], normalized[i])
		}
	}

	return glyphs, nil
}

// VerifyWordsRoundTrip checks that encoding and then decoding the provided words
// (under the given key/policy) yields the exact same word sequence. It returns
// nil on success or a detailed error on any failure.
//
// This is equivalent to calling EncodeWordsVerified and discarding the glyphs.
func VerifyWordsRoundTrip(words []string, index map[string]int, wordsList []string, key string, policy KeyPolicy) error {
	_, err := EncodeWordsVerified(words, index, wordsList, key, policy)
	return err
}

// normalizeWords lowercases and trims each word, skipping empties.
func normalizeWords(words []string) []string {
	out := make([]string, 0, len(words))
	for _, w := range words {
		lw := strings.ToLower(strings.TrimSpace(w))
		if lw != "" {
			out = append(out, lw)
		}
	}
	return out
}
