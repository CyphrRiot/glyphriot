package internal

import (
	"fmt"
	"strings"
)

// EncodeWords encodes each input word into a 4-glyph token using the 4×7 unique scheme.
//
// Parameters:
//   - input: the list of words to encode
//   - index: map of lowercased word -> index in wordsList (0..Total-1)
//   - wordsList: the canonical word list (length must be Total, e.g., 2048 for BIP‑39 EN)
//   - key: optional salt for deterministic permutation
//
// Behavior:
//  1. Derive permutation P from key.
//  2. For each input word: normalize (trim/lower), find its index in wordsList via index.
//  3. Get position pos = P^-1[index] (0..Total-1).
//  4. Convert pos to a code (identity in 4×7), then to 4 base‑7 digits.
//  5. Map digits to glyphs (Digits) and emit the 4‑glyph token.
//
// Returns:
//   - slice of encoded 4‑glyph tokens (same order as input)
//   - error if any word is not present in index or invalid parameters were provided
func EncodeWords(input []string, index map[string]int, wordsList []string, key string) ([]string, error) {
	if len(wordsList) != Total {
		return nil, fmt.Errorf("wordsList length must be %d, got %d", Total, len(wordsList))
	}

	_, inv := Derive(len(wordsList), key)

	out := make([]string, 0, len(input))
	for _, w := range input {
		lw := strings.ToLower(strings.TrimSpace(w))
		if lw == "" {
			continue
		}
		idx, ok := index[lw]
		if !ok {
			return nil, fmt.Errorf("%q not in active list", w)
		}
		pos := inv[idx] // 0..Total-1
		code, ok := PosToCode(pos)
		if !ok {
			return nil, fmt.Errorf("internal error: invalid position %d for %q", pos, w)
		}
		d, ok := ToDigits(code) // 4 base-7 digits
		if !ok {
			return nil, fmt.Errorf("internal error: invalid code %d for %q", code, w)
		}
		var sb strings.Builder
		for i := 0; i < Len; i++ {
			sb.WriteRune(Digits[d[i]])
		}
		out = append(out, sb.String())
	}
	return out, nil
}

// DecodeGlyphToken decodes a single 4‑glyph token back to its exact word using the 4×7 unique scheme.
//
// Parameters:
//   - tok: the 4‑glyph token to decode (must be exactly Len runes long)
//   - wordsList: the canonical word list (length must be Total, e.g., 2048 for BIP‑39 EN)
//   - key: optional salt for deterministic permutation
//
// Behavior:
//  1. Derive permutation P from key.
//  2. Map glyphs to base‑7 digits via Decode; validate length.
//  3. Convert digits to code (validate range).
//  4. wordIndex = P[code]; return wordsList[wordIndex].
//
// Returns:
//   - the decoded word
//   - error if the token is invalid or parameters are inconsistent
func DecodeGlyphToken(tok string, wordsList []string, key string) (string, error) {
	if len(wordsList) != Total {
		return "", fmt.Errorf("wordsList length must be %d, got %d", Total, len(wordsList))
	}

	p, _ := Derive(len(wordsList), key)

	runes := []rune(strings.TrimSpace(tok))
	if len(runes) != Len {
		return "", fmt.Errorf("glyph %q must be exactly %d symbols", tok, Len)
	}
	d := make([]int, Len)
	for i, r := range runes {
		val, ok := Decode[r]
		if !ok {
			return "", fmt.Errorf("invalid glyph rune %q in %q", r, tok)
		}
		d[i] = val
	}
	code, ok := FromDigits(d)
	if !ok {
		return "", fmt.Errorf("invalid glyph code %q", tok)
	}
	idx := p[code]
	if idx < 0 || idx >= len(wordsList) {
		return "", fmt.Errorf("invalid glyph code %q", tok)
	}
	return wordsList[idx], nil
}
