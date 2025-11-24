<div align="center">

![Version](https://img.shields.io/badge/version-v1.0-blue?labelColor=0052cc)
![Code](https://img.shields.io/badge/code-Go-00ADD8?logo=go&logoColor=white&labelColor=0f172a)
![Human Coded](https://img.shields.io/badge/human-coded-1e3a8a?labelColor=111827&color=1e3a8a)
![CypherRiot](https://img.shields.io/badge/CypherRiot-18181B?logo=github&logoColor=white&labelColor=0f172a)

</div>

# GlyphRiot – Glyph Seed Words

Standardized, wallet‑style seed word glyphs.

- Fixed glyph digits (exactly 7): △ □ ○ × • ◇ ☆
- Fixed length: exactly 4 glyphs per word (unique mapping; decoding returns the exact word)
- Optional keyed salt: SHA‑256(key) → PRNG → deterministic permutation of the 2048 BIP‑39 English words
- Behavior mirrors wallet UX (Ledger/Trezor/MetaMask):
    - Encoding produces 4 glyphs per word
    - Decoding a 4‑glyph code returns the exact word (1→1)

This is exactly as accurate and reliable as typing the first 4 letters during wallet recovery, with the same user experience but faster input.

## How it works

- Word list: BIP‑39 English (default), or a custom 2048‑line file.
- Salt (key): If provided, a deterministic permutation P of the 2048 words is generated from SHA‑256(key). No key → identity order.
- Code space: 7^4 = 2401 unique codes. Each word position n (0..2047) maps to code b = n.
- Encoding (word → 4 glyphs):
    1. Find the position n (0..2047) of the word in the permuted order P.
    2. Convert n to base‑7 with exactly 4 digits (pad left with 0s).
    3. Map digits 0..6 to △ □ ○ × • ◇ ☆ and emit the 4 glyphs.
- Decoding (4 glyphs → word):
    1. Convert 4 glyphs back to a base‑7 code b (0..2400).
    2. If b ≥ 2048 → invalid code. Otherwise the word is P[b]. Exact 1→1.

Input rules:

- Only the seven glyphs △, □, ○, ×, •, ◇, ☆ are valid (× also accepts x/X when typing)
- Anything else is rejected

## Install

Download for Linux (latest binary): https://raw.githubusercontent.com/CyphrRiot/glyphriot/main/bin/glyphriot

Requires Go 1.25+.

```bash
make          # build ./bin/glyphriot
make install  # install to ~/.local/bin/glyphriot
make clean    # remove build artifact
```

## Usage

Help:

```bash
./bin/glyphriot --help
```

Words → glyphs (always 4 glyphs per word):

```bash
./bin/glyphriot brave coconut drift zebra
./bin/glyphriot --key 'my secret' brave coconut drift zebra
./bin/glyphriot --prompt --mask brave coconut drift zebra   # secure key prompt
```

Glyphs → words (exact):

```bash
./bin/glyphriot '△□○×' '□□○×'
List: bip39-en
△□○× → <word>
□□○× → <word>
```

Full table (for the active list, salted if key set):

```bash
./bin/glyphriot --all | less
```

Options summary:

- --list bip39-en|auto (default: bip39-en)
- --list-file PATH (custom 2048‑line list; overrides --list)
- --key KEY | --prompt [--mask] (salt to reorder words)
- --all (print index, word, 4‑glyph code; paged on TTY by default; use --pager=false to disable)
- --glyph-sep SEP (insert a visible separator between glyphs in output; decoding strips it)
- --phrase-only (when decoding glyphs, print only the recovered phrase)

## Decoding

- Each 4‑glyph token decodes to exactly one word. No candidate lists.
- If you use a key (salt), use the same key for encoding and decoding to recover the same mapping.

## Security notes

- Each 4‑glyph code maps to exactly one word in the selected list and key context.
- The optional key (salt) permutes the entire list. Use the same key for encoding and decoding to recover the same mapping. Different keys produce unrelated mappings.

## Examples

```bash
# No key
$ ./bin/glyphriot brave coconut drift zebra
•○△•  □□×•  ○○•△  ••○×

# With key
$ ./bin/glyphriot --key hunter2 brave coconut drift zebra
□○△×  ○•×□  △○•×  ○•××

# Decoding (example glyphs)
$ ./bin/glyphriot '△□○×' '□□○×'
List: bip39-en
△□○× → <word>
□□○× → <word>
```

## License

MIT
