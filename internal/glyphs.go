package internal

// Package glyphs encapsulates the 4×7 unique glyph mapping scheme used by GlyphRiot.
//
// Rationale
// - We need exactly one 4-glyph code per BIP-39 word (2048 total).
// - With a base of 7 symbols and fixed length 4, the code space is 7^4 = 2401 ≥ 2048.
// - Therefore, we can map the permuted position n∈[0,2047] directly to a unique code
//   (identity mapping): code = n. Decode is the exact inverse.
//
// This package exposes:
// - Constants: Base, Len, Total
// - Glyph set: Digits (ordered), Decode (rune → digit)
// - Conversions:
//     ToDigits(code)    → 4 base-7 digits (MSB-first)
//     FromDigits(d)     → code (validates range and digits)
//     PosToCode(pos)    → code (identity; validates range)
//     CodeToPos(code)   → pos  (identity; validates range)
// - Render/Parse helpers for glyph strings:
//     Render(code)      → 4 runes string
//     Parse(s)          → code from a 4-rune glyph string
//
// Notes
// - All functions validate inputs and return (value, ok) where appropriate.
// - The caller is responsible for applying the permutation P derived from the salt/key.
//   I.e., encode: pos = P^-1[wordIndex]; code = PosToCode(pos)
//         decode: wordIndex = P[CodeToPos(code)]
//
// Glyph set (0..6):
//   0: △   1: □   2: ○   3: ×   4: •   5: ◇   6: ☆
// Convenience: 'x' or 'X' are accepted as aliases for × when parsing.

const (
	// Base is the numeric base for glyph digits (7 unique symbols).
	Base = 7
	// Len is the fixed number of glyphs per word.
	Len = 4
	// Total is the total number of entries in the list (BIP-39 English).
	Total = 2048
)

// Digits defines the ordered set of 7 glyphs (0..6) used by the scheme.
var Digits = []rune{'△', '□', '○', '×', '•', '◇', '☆'}

// Decode maps each glyph rune (and x/X convenience for ×) back to its digit (0..6).
var Decode = map[rune]int{
	'△': 0,
	'□': 1,
	'○': 2,
	'×': 3,
	'•': 4,
	'◇': 5,
	'☆': 6,
	// Convenience aliases for ×:
	'x': 3,
	'X': 3,
}

// ToDigits converts a code/index (0..Total-1) into 4 base-7 digits (MSB-first).
// Returns (digits, true) if code is within range; otherwise returns (zero, false).
func ToDigits(code int) ([Len]int, bool) {
	var out [Len]int
	if code < 0 || code >= Total {
		return out, false
	}
	for i := Len - 1; i >= 0; i-- {
		out[i] = code % Base
		code /= Base
	}
	return out, true
}

// FromDigits converts 4 base-7 digits back to a code/index, validating digit range
// and final code range [0, Total).
func FromDigits(d []int) (int, bool) {
	if len(d) != Len {
		return 0, false
	}
	code := 0
	for i := 0; i < Len; i++ {
		if d[i] < 0 || d[i] >= Base {
			return 0, false
		}
		code = code*Base + d[i]
	}
	if code < 0 || code >= Total {
		return 0, false
	}
	return code, true
}

// PosToCode maps a word position in the active permutation (0..Total-1) to its code/index.
// With the 4×7 unique scheme, this is the identity function. Returns (code, true) if pos is valid.
func PosToCode(pos int) (int, bool) {
	if pos < 0 || pos >= Total {
		return 0, false
	}
	return pos, true
}

// CodeToPos returns the word position for a code/index (identity mapping).
// Returns (pos, true) if code is within [0, Total); otherwise (0, false).
func CodeToPos(code int) (int, bool) {
	if code < 0 || code >= Total {
		return 0, false
	}
	return code, true
}

// Render converts a valid code into its 4-glyph string representation.
// Returns ("", false) if the code is out of range.
func Render(code int) (string, bool) {
	d, ok := ToDigits(code)
	if !ok {
		return "", false
	}
	b := make([]rune, 0, Len)
	for i := 0; i < Len; i++ {
		b = append(b, Digits[d[i]])
	}
	return string(b), true
}

// Parse converts a 4-rune glyph string into the corresponding code.
// Returns (code, true) when valid; otherwise (0, false).
func Parse(s string) (int, bool) {
	runes := []rune(s)
	if len(runes) != Len {
		return 0, false
	}
	d := make([]int, Len)
	for i, r := range runes {
		val, ok := Decode[r]
		if !ok {
			return 0, false
		}
		d[i] = val
	}
	return FromDigits(d)
}
