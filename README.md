<div align="center">

![Version](https://img.shields.io/badge/version-v1.0-blue?labelColor=0052cc)
![Code](https://img.shields.io/badge/code-Go-00ADD8?logo=go&logoColor=white&labelColor=0f172a)
![Human Coded](https://img.shields.io/badge/human-coded-1e3a8a?labelColor=111827&color=1e3a8a)
![CypherRiot](https://img.shields.io/badge/CypherRiot-18181B?logo=github&logoColor=white&labelColor=0f172a)

</div>

# GlyphRiot – Glyph Seed Words

Standardized, wallet‑style seed word glyphs.

- Fixed glyph digits (exactly 7): △ □ ○ × • ◇ ▽ (☆ also accepted)
  Note: ▽ replaces ☆ for improved readability; old glyphs using ☆ still decode correctly.
- Fixed length: exactly 4 glyphs per word (unique mapping; decoding returns the exact word)
- Optional keyed salt: SHA‑256(key) → PRNG → deterministic permutation of the 2048 BIP‑39 English words
- Behavior mirrors wallet UX (Ledger/Trezor/MetaMask):
    - Encoding produces 4 glyphs per word
    - Decoding a 4‑glyph code returns the exact word (1→1)

This is exactly as accurate and reliable as typing the first 4 letters during wallet recovery, with the same user experience but faster input.

## How it works

- What this tool does (in one sentence):
    - It gives every seed word its own 4‑glyph “nickname,” and those 4 glyphs always decode back to the exact word.

- Why it works (the simple math):
    - We use 7 easy‑to‑read symbols and always show 4 of them → 7^4 = 2401 possible codes.
    - There are 2048 BIP‑39 words, so each word can have its own unique code with room to spare.

- With or without a key (salt):
    - Optional key reshuffles the word list in a predictable way (wallet‑style). Use the same key to get the same codes. No key → standard order.

- Encode (words → glyphs), in plain English:

*   1.  Look up the word’s place in the (possibly reshuffled) 2048‑word list.
*   2.  Turn that position into four base‑7 digits.
*   3.  Replace digits 0..6 with these symbols: △ □ ○ × • ◇ ▽ (☆ also accepted).

- Decode (glyphs → word), in plain English:
    1. Turn the four symbols back into digits 0..6.
    2. Convert those digits into a number.
    3. Use that number to pick the word from the same list/key.
    4. Result: one code → exactly one word. No guessing.

- Guarantees:
    - Exact round‑trip: the 4 glyphs you see always decode to the original word.
    - Works offline; deterministic; no external services required.

Input rules:

- Only the seven glyphs △, □, ○, ×, •, ◇, ▽ are valid (× also accepts x/X; ☆ also accepted)
- Anything else is rejected

## Install

Download for Linux (latest binary): https://raw.githubusercontent.com/CyphrRiot/glyphriot/main/bin/glyphriot
Windows (prebuilt or build yourself): https://raw.githubusercontent.com/CyphrRiot/glyphriot/main/bin/glyphriot-windows-amd64.exe (or run: make build-windows-amd64)

Quick download and run:

```bash
curl -L -o glyphriot https://raw.githubusercontent.com/CyphrRiot/glyphriot/main/bin/glyphriot
chmod +x glyphriot
./glyphriot --help
```

From source (Go 1.25+):

```bash
make                   # build ./bin/glyphriot (native platform)
make install           # install to ~/.local/bin/glyphriot
make clean             # remove build artifact

# Cross-compile targets (artifacts in ./bin)
make build-linux-amd64     # build ./bin/glyphriot-linux-amd64
make build-linux-arm64     # build ./bin/glyphriot-linux-arm64
make build-windows-amd64   # build ./bin/glyphriot-windows-amd64.exe
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

Self-test (with and without a key):

```bash
# No key
./bin/glyphriot --self-test

# With key (deterministic shuffle; use the same key to get the same codes)
./bin/glyphriot --key "test-key" --self-test
```

Key handling (prompt or flag):

```bash
# Prompt for key (masked by default); overrides --key if both are present
./bin/glyphriot --prompt brave coconut drift zebra

# Prompt and decode glyphs
./bin/glyphriot --prompt '△□○×' '□□○×'
```

Options summary:

- --list bip39-en|auto (default: bip39-en)
- --list-file PATH (custom 2048‑line list; overrides --list)
- --key KEY | --prompt [--mask] (salt to reorder words)
- --all (print index, word, 4‑glyph code; paged on TTY by default; use --pager=false to disable)
- --glyph-sep SEP (insert a visible separator between glyphs in output; decoding strips it)
- --phrase-only (when decoding glyphs, print only the recovered phrase)

## Custom list validation

If you use --list-file, your file must meet these rules:

- UTF‑8 encoded (a UTF‑8 BOM is accepted and stripped automatically)
- Exactly 2048 non‑empty lines (one word per line)
- Whitespace is trimmed; words are treated case‑insensitively (lowercased)
- Duplicate words are rejected with a clear error
- Newlines are normalized (CRLF and CR are handled)

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
