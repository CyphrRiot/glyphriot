package internal

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"golang.org/x/crypto/argon2"
)

// KeyPolicy defines how we validate and derive the effective key material.
//   - If KDF == "argon2id" (default), we use Argon2id to slow down brute force
//     and enforce practical minimum lengths (ergonomic).
//   - If KDF == "none", we enforce pure minimum entropy by format (BIP‑39/hex/base64)
//     and reject everything else unless AllowWeak == true.
type KeyPolicy struct {
	KDF         string // "argon2id" (default) or "none"
	KDFMemMB    uint32 // memory in MB (e.g., 512)
	KDFTime     uint32 // iterations (e.g., 3)
	KDFParallel uint8  // parallelism (e.g., 1)
	AllowWeak   bool   // allow weak keys (bypass enforcement)
}

// DefaultKeyPolicy returns a recommended default that is ergonomic and strong.
// These parameters make each guess expensive and greatly slow down brute force.
// Tweak to your environment if necessary (e.g., reduced mem on low-RAM hosts).
func DefaultKeyPolicy() KeyPolicy {
	return KeyPolicy{
		KDF:         "argon2id",
		KDFMemMB:    512,
		KDFTime:     3,
		KDFParallel: 1,
		AllowWeak:   false,
	}
}

// MinBitsForContext returns the minimum key entropy required based on seed/glyph length.
// We default to 128-bit unless the caller signals a 24-word context (256-bit).
func MinBitsForContext(tokenOrWordCount int) int {
	if tokenOrWordCount >= 24 {
		return 256
	}
	return 128
}

// EffectiveKeyMaterial derives a 32-byte seed from the provided key string using the given policy.
//   - If policy.KDF == "argon2id": we run Argon2id with configured parameters and a fixed domain salt.
//     The Argon2id output is then hashed with SHA-256 to get a canonical 32-byte seed.
//   - If policy.KDF == "none": we directly SHA-256 the key string to 32 bytes.
func EffectiveKeyMaterial(key string, policy KeyPolicy) ([32]byte, error) {
	var seed32 [32]byte

	switch strings.ToLower(strings.TrimSpace(policy.KDF)) {
	case "", "argon2id":
		// Use a domain-separated salt so the same passphrase doesn't collide across tools.
		salt := []byte("GlyphRiot/v1/argon2id/domain-sep")
		mem := policy.KDFMemMB
		if mem == 0 {
			mem = 512
		}
		time := policy.KDFTime
		if time == 0 {
			time = 3
		}
		par := policy.KDFParallel
		if par == 0 {
			par = 1
		}

		derived := argon2.IDKey([]byte(key), salt, time, mem*1024, par, 32)
		// Hash once more to canonicalize
		seed32 = sha256.Sum256(derived)
		return seed32, nil

	case "none":
		seed32 = sha256.Sum256([]byte(key))
		return seed32, nil

	default:
		return seed32, fmt.Errorf("unknown KDF %q (supported: argon2id, none)", policy.KDF)
	}
}

// ValidateKeyStrength enforces minimum key strength based on policy and required bits.
// - With kdf=argon2id (default): enforce minimum passphrase length (ergonomic hardness).
//   - For minBits=128 → require >=16 characters
//   - For minBits=256 → require >=20 characters
//     These are practical values given a strong KDF; feel free to tune upward.
//
// - With kdf=none: enforce pure entropy by format only (no guessing the "quality" of ASCII).
//   - Accept a 12- or 24-word BIP‑39 phrase as the key (128/256 bits).
//   - Accept hex with length >= minBits/4.
//   - Accept base64 whose decoded length >= minBits/8.
//   - Everything else is rejected unless policy.AllowWeak is true.
func ValidateKeyStrength(key string, minBits int, policy KeyPolicy) error {
	key = strings.TrimSpace(key)
	if key == "" {
		if policy.AllowWeak {
			return nil
		}
		return fmt.Errorf("key is empty; provide a strong key or use --allow-weak-key")
	}

	switch strings.ToLower(strings.TrimSpace(policy.KDF)) {
	case "", "argon2id":
		// Practical hardness with KDF: enforce minimal passphrase length thresholds.
		var minLen int
		if minBits >= 256 {
			minLen = 20
		} else {
			minLen = 16
		}
		// Count runes (not bytes) to avoid trivially bypassing with multi-byte empty-like inputs.
		if utf8.RuneCountInString(key) < minLen {
			if policy.AllowWeak {
				return nil
			}
			return fmt.Errorf("key too short: need %d+ characters for this context with Argon2id enabled (or supply a 24-word BIP‑39 phrase, 64+ hex chars, or 32-byte base64; or use --allow-weak-key)", minLen)
		}
		return nil

	case "none":
		// Pure entropy enforcement by accepted format
		if ok := satisfiesPureEntropyFormats(key, minBits); ok {
			return nil
		}
		if policy.AllowWeak {
			return nil
		}
		return fmt.Errorf("key does not meet %d-bit minimum. Use a 12/24-word BIP‑39 phrase, %d+ hex chars, or base64 of %d+ bytes; or use --allow-weak-key",
			minBits, minBits/4, minBits/8)

	default:
		if policy.AllowWeak {
			return nil
		}
		return fmt.Errorf("unknown KDF %q (supported: argon2id, none)", policy.KDF)
	}
}

// satisfiesPureEntropyFormats returns true if key (a string) meets the pure entropy
// thresholds for minBits via one of the recognized high-entropy formats:
// - BIP‑39 English phrase with 12 or 24 words (we check membership; checksum optional).
// - Hex string with length >= minBits/4.
// - Base64 string where decoded length >= minBits/8 bytes.
func satisfiesPureEntropyFormats(key string, minBits int) bool {
	if bits, ok := bip39BitsIfValid(key); ok {
		return bits >= minBits
	}
	if isStrongHex(key, minBits) {
		return true
	}
	if isStrongBase64(key, minBits) {
		return true
	}
	return false
}

// bip39BitsIfValid checks if the key is a 12- or 24-word BIP‑39 English phrase.
// We validate that all words are in the canonical list; checksum is not required here
// because for key-space purposes we only need the entropy tier (128/256).
func bip39BitsIfValid(key string) (int, bool) {
	words := normalizedWords(key)
	if len(words) != 12 && len(words) != 24 {
		return 0, false
	}
	// Build index for quick lookup once per call; small map, negligible cost
	bipIndex := make(map[string]struct{}, len(WordsBIP39EN))
	for _, w := range WordsBIP39EN {
		bipIndex[strings.ToLower(strings.TrimSpace(w))] = struct{}{}
	}
	for _, w := range words {
		if _, ok := bipIndex[w]; !ok {
			return 0, false
		}
	}
	if len(words) == 12 {
		return 128, true
	}
	return 256, true
}

func normalizedWords(s string) []string {
	fields := strings.Fields(strings.ToLower(strings.TrimSpace(s)))
	out := make([]string, 0, len(fields))
	for _, f := range fields {
		if f != "" {
			out = append(out, f)
		}
	}
	return out
}

func isStrongHex(s string, minBits int) bool {
	if len(s) == 0 || len(s)%2 != 0 { // even-length hex only
		return false
	}
	// Hex must decode
	_, err := hex.DecodeString(s)
	if err != nil {
		return false
	}
	// Each hex char = 4 bits
	return len(s)*4 >= minBits
}

func isStrongBase64(s string, minBits int) bool {
	// Try standard base64
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		// Try URL-safe base64 without padding
		data, err = base64.RawURLEncoding.DecodeString(s)
		if err != nil {
			// Try raw base64 (no padding)
			data, err = base64.RawStdEncoding.DecodeString(s)
			if err != nil {
				return false
			}
		}
	}
	// Each byte = 8 bits
	return len(data)*8 >= minBits
}

// EnforceOrError validates key strength according to policy and required bits.
// It returns nil if the key is acceptable under the current policy (or AllowWeak).
// Otherwise it returns an actionable error message.
func EnforceOrError(key string, minBits int, policy KeyPolicy) error {
	if err := ValidateKeyStrength(key, minBits, policy); err != nil {
		return err
	}
	return nil
}

// MustEffectiveKeyMaterial is a convenience that enforces strength (unless AllowWeak)
// and derives the 32-byte seed using the policy KDF. If enforcement fails, returns error.
func MustEffectiveKeyMaterial(key string, minBits int, policy KeyPolicy) ([32]byte, error) {
	var zero [32]byte
	if err := EnforceOrError(key, minBits, policy); err != nil {
		return zero, err
	}
	return EffectiveKeyMaterial(key, policy)
}

// ErrWeakKey is a sentinel error for weak keys (you may wrap or replace messages).
var ErrWeakKey = errors.New("weak key")
