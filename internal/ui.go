package internal

import (
	"strings"
	"unicode"
)

// Package internal: UI helpers (exported)
//
// This file provides small, self-contained UI helpers for:
// - ANSI styling (Tokyo Night–inspired colors)
// - Glyph token formatting (insert/remove separators and whitespace)
//
// Color usage
// - Enable or disable color globally via SetColorEnabled(true/false).
// - Wrap text with Style("text", Bold, Blue) to apply codes when enabled.
// - When disabled, Style returns the input unchanged.
//
// Glyph formatting
// - InsertSep inserts a visible separator between runes for readability.
// - StripSepAndSpaces removes the configured separator and all Unicode spaces,
//   normalizing tokens prior to glyph validation/decoding.

// --- ANSI color/style (Tokyo Night–inspired) ---

// Default: colors enabled. Override via SetColorEnabled.
var colorEnabled = true

// ANSI escape codes (exported)
const (
	Reset  = "\x1b[0m"
	Bold   = "\x1b[1m"
	Blue   = "\x1b[38;2;122;162;247m" // Tokyo Night blue
	Cyan   = "\x1b[38;2;42;195;222m"  // Tokyo Night cyan
	Purple = "\x1b[38;2;187;154;247m" // Tokyo Night purple
	Gray   = "\x1b[38;2;136;146;176m" // Dimmed foreground
	Red    = "\x1b[38;2;247;118;142m" // Tokyo Night red
)

// SetColorEnabled toggles ANSI styling on or off.
func SetColorEnabled(on bool) {
	colorEnabled = on
}

// ColorEnabled reports whether ANSI styling is currently enabled.
func ColorEnabled() bool {
	return colorEnabled
}

// Style wraps s with the provided ANSI codes when color is enabled.
// When disabled, returns s unchanged.
//
// Example:
//
//	Style("Hello", Bold, Blue)
func Style(s string, codes ...string) string {
	if !colorEnabled {
		return s
	}
	var b strings.Builder
	for _, c := range codes {
		b.WriteString(c)
	}
	b.WriteString(s)
	b.WriteString(Reset)
	return b.String()
}

// --- Glyph token formatting helpers ---

// StripSepAndSpaces removes all Unicode whitespace and an optional configured
// separator from the provided string. This normalizes glyph tokens so that
// decoding accepts visually separated input (e.g., "△ □ ○ ×" or "△·□·○·×").
//
// If sep is empty, only whitespace is removed.
func StripSepAndSpaces(s, sep string) string {
	if sep != "" {
		s = strings.ReplaceAll(s, sep, "")
	}
	var buf []rune
	for _, r := range []rune(s) {
		if !unicode.IsSpace(r) {
			buf = append(buf, r)
		}
	}
	return string(buf)
}

// InsertSep inserts the configured separator between glyph runes when rendering
// output to improve readability in fonts where glyphs appear tightly spaced.
//
// If sep is empty, s is returned unchanged.
func InsertSep(s, sep string) string {
	if sep == "" {
		return s
	}
	r := []rune(s)
	if len(r) == 0 {
		return s
	}
	var b strings.Builder
	for i, ch := range r {
		if i > 0 {
			b.WriteString(sep)
		}
		b.WriteRune(ch)
	}
	return b.String()
}
