# GlyphRiot: Plan and Status

Last updated: 2025-11-23

## Where we are (Glyph Seed System v1.0)

Implemented and validated updated v1.0 with exact decoding:

- Fixed glyph set (digits 0..6):
    - 0: △, 1: □, 2: ○, 3: × (accept x/X), 4: •, 5: ◇, 6: ☆
- Fixed length: exactly 4 glyphs per word.
- Optional key/salt: SHA‑256(key) → deterministic seed → Fisher–Yates permutation P over 2048 BIP‑39 words. No key → identity.
- Encoding (word → 4 glyphs; 4×7 unique mapping):
    - Find word’s position n in P (i.e., inverse(P) for original list index).
    - Since 7^4 = 2401 ≥ 2048, use b = n directly (no buckets/packing).
    - Convert b to 4 base‑7 digits → map 0..6 → △ □ ○ × • ◇ ☆
- Decoding (4 glyphs → exact word):
    - Convert 4 glyphs → base‑7 code b (0..2400).
    - If b ≥ 2048 → invalid code; else the word is P[b] (unique 1→1).
- Input validation: accepts only △ □ ○ × • ◇ ☆; x/X treated as alias for ×; all else rejected.
- CLI/UX:
    - --all: prints salted table (index, word, 4‑glyph code); paged on TTY by default; --pager=false disables.
    - --list bip39-en|auto; --list-file PATH (expects exactly 2048 lines). Deprecated aliases removed from help.
    - --key KEY | --prompt [--mask]; masked prompt defaults to true; TTY required; signal‑safe terminal restore.
    - --self-test: built‑in harness runs randomized 12‑word sets and paginates results on TTY.
    - --glyph-sep SEP: insert a visible separator between glyphs on output; decoding strips it.
    - --phrase-only: decoding glyph tokens prints only the exact phrase (single line).
    - --no-color: disable ANSI colors; help is concise and fits one screen with Tokyo‑Night style colors by default.
- Build/Docs:
    - Makefile builds ./bin/glyphriot via `go build .`; install to ~/.local/bin; clean removes artifacts.
    - go.mod go 1.25; requires golang.org/x/term and x/sys.
    - README updated to 4×7 exact scheme with minimal badges and new flags.
- Internal structure (cleanup complete):
    - internal/glyphs.go — 4×7 scheme (Base/Len/Total, Digits/Decode, conversions)
    - internal/perm.go — Derive/Inv permutation helpers
    - internal/codec.go — EncodeWords/DecodeGlyphToken
    - internal/ui.go — styling and glyph separator helpers
    - main.go — flags, I/O, orchestration only

## Overview rules (product + dev process)

Product/spec rules

- Fixed 7‑symbol alphabet only (△ □ ○ × • ◇ ☆); no alternates, no confusables.
- No lossy compression; exactly 4 glyphs per word.
- Deterministic key‑salted permutation (wallet‑style behavior).
- Decode is exact 1→1 per token under the 4×7 scheme.
- Strict input validation: only allowed glyphs; exact 4‑glyph tokens.

Developer workflow rules

- Build with `make` only; verify after each change.
- Incremental changes: one focused change per commit/iteration; avoid sweeping edits.
- Robust UX: masked prompts by default; pager for long output; concise, one‑screen help with colors by default.
- Safety: signal‑safe terminal restore; TTY enforcement for interactive prompts.
- Quality: clean, readable code; deterministic behavior; tests/self‑tests before release.

## Wordlist status and plan

Current state

- english.txt is embedded (BIP‑39 EN). A legacy words.txt previously duplicated content.

Goal

- Maintain a single canonical BIP‑39 EN list embedded in the binary. If an alternate list is ever added, keep it distinct with documented provenance and validation.

Action items

- Completed: english.txt is the sole embedded list; deprecated aliases removed from help/docs.
- Future (if alternate list is ever added):
    - Document provenance; validate exactly 2048 UTF‑8 lines.
    - Keep lists unambiguous; update flags, README, and tests accordingly.

## Audit & Release Plan

Focus: full-scale code audit, determinism verification, static analysis, and release readiness.

Planned work

- Code audit:
    - Review modules under internal/ for clarity, error handling, and testability.
    - Verify glyph validation, separator stripping, and edge cases in DecodeGlyphToken.
    - Confirm terminal interactions (prompt, pager) restore TTY state robustly.
- Determinism:
    - Same key/list yields identical outputs across runs/hosts.
    - Stable --all output for a given key.
- Static analysis:
    - go vet; staticcheck (if available).
- Docs/UX:
    - Final pass over README/help; ensure options and examples are correct.
- Release:
    - make builds clean; tag and prepare release notes.

Risk/compat notes

- 4×7 is a breaking change vs prior 5^4 scheme; old 5‑glyph codes will not decode under v1.0 4×7. Decision: accept the break for exact decoding guarantees and document accordingly.

## Next steps (checklist)

- [ ] Full code audit (internal/\*): clarity, error handling, and tests where warranted
- [ ] Determinism checks: same key/list stable across runs/hosts; stable --all output
- [ ] Static analysis: go vet; staticcheck (if available)
- [ ] Docs/UX final pass: README/help options and examples verified
- [ ] Release prep: make build clean; tag and prepare release notes

## Reference commands (do not auto-run)

```
# Build
make

# Self-test (no key; paginated on TTY)
./bin/glyphriot --self-test

# All entries with pager
./bin/glyphriot --all

# All entries without pager
./bin/glyphriot --all --pager=false

# Decode 12 tokens: phrase only (one line)
./bin/glyphriot --phrase-only '<12 tokens>'
```

## Notes

- The lists are embedded via //go:embed. Deduplicate to a single canonical list (english.txt) until a truly distinct alternate exists.
- Finalized 4×7 exact scheme eliminates candidate lists and relies on the expanded glyph set for unique mapping.
