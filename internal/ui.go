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

	// Brand accent colors for optional QR tinting (foreground 24-bit ANSI)
	BTCOrange     = "\x1b[38;2;247;147;26m" // Bitcoin orange  #F7931A
	BTCDark       = "\x1b[38;2;168;95;0m"   // Bitcoin dark orange (watermark)
	XMROrange     = "\x1b[38;2;255;110;0m"  // Monero orange   #FF6E00
	ZECGold       = "\x1b[38;2;236;178;68m" // Zcash gold      #ECB244
	ZECDark       = "\x1b[38;2;156;122;31m" // Zcash dark gold (watermark)
	XMROrangeDark = "\x1b[38;2;176;74;0m"   // Monero dark orange (watermark)

	// Subtle watermark brand tones (very light BG on white; near‑black FG on black)
	// These add a faint hue without harming QR contrast.
	BTCWatermarkBg = "\x1b[48;2;255;245;200m" // warm pale yellow (#FFF5C8) for white halves
	BTCWatermarkFg = "\x1b[38;2;10;10;0m"     // near‑black with slight yellow bias (#0A0A00) for black halves

	ZECWatermarkBg = "\x1b[48;2;250;235;190m" // soft pale gold (#FAEBBE) for white halves
	ZECWatermarkFg = "\x1b[38;2;12;10;2m"     // near‑black gold‑tint (#0C0A02) for black halves

	XMRWatermarkBg = "\x1b[48;2;255;240;210m" // soft pale orange (#FFF0D2) for white halves
	XMRWatermarkFg = "\x1b[38;2;12;8;0m"      // near‑black orange‑tint (#0C0800) for black halves

	// Brand background colors for QR overlays (background 24-bit ANSI)
	BTCOrangeBg     = "\x1b[48;2;247;147;26m" // Bitcoin orange background
	BTCDarkBg       = "\x1b[48;2;168;95;0m"   // Bitcoin dark orange background (watermark)
	XMROrangeBg     = "\x1b[48;2;255;110;0m"  // Monero orange background
	ZECGoldBg       = "\x1b[48;2;236;178;68m" // Zcash gold background
	ZECDarkBg       = "\x1b[48;2;156;122;31m" // Zcash dark gold background (watermark)
	XMROrangeDarkBg = "\x1b[48;2;176;74;0m"   // Monero dark orange background (watermark)
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

// Banner returns the styled CLI header.
func Banner(version string) string {
	return Style("GlyphRiot — Glyph Seed System - "+version, Bold, Purple)
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
