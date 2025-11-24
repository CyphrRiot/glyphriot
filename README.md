<div align="center">

![Version](https://img.shields.io/badge/version-v1.2-blue?labelColor=0052cc)
![Code](https://img.shields.io/badge/code-Go-00ADD8?logo=go&logoColor=white&labelColor=0f172a)
![Human Coded](https://img.shields.io/badge/human-coded-1e3a8a?labelColor=111827&color=1e3a8a)
![CypherRiot](https://img.shields.io/badge/CypherRiot-18181B?logo=github&logoColor=white&labelColor=0f172a)

</div>

# GlyphRiot – Glyph'd Seed Words

### The ultimate seed phrase obfuscation system for self-custody warriors!

Standardized, wallet‑style seed word glyphs.

- Fixed glyph digits (exactly 7): △ □ ○ ✕ ● ◇ ▽ (aliases: × and • accepted; ☆ also accepted)
  Note: ▽ replaces ☆ for improved readability; old glyphs using ☆ still decode correctly.
- Fixed length: exactly 4 glyphs per word (unique mapping; decoding returns the exact word)
- Optional keyed salt: SHA‑256(key) → PRNG → deterministic permutation of the 2048 BIP‑39 English words
- Behavior mirrors wallet UX (Ledger/Trezor/MetaMask):
    - Encoding produces 4 glyphs per word
    - Decoding a 4‑glyph code returns the exact word (1→1)

This is exactly as accurate and reliable as typing the first 4 letters during wallet recovery, with the same user experience but faster input.

## Examples

Tip: Use --prompt to protect the mapping with a key (salt); Argon2id is enabled by default. With --prompt, enter your key first, then enter your seed words or glyph tokens interactively (not on the command line).

```bash
# Recommended: interactive with --prompt (enter key twice to confirm)
glyphriot --prompt
Enter key: ******************
Re-enter key: ******************
Seed/Glyph: ◇▽✕▽  ○△△✕  □◇✕□  ✕△✕▽  ●△●✕  ✕○◇△

Phrase: violin era grab thunder rescue case

# Words → glyphs (no key)
glyphriot letter advice cage absurd amount doctor
> ○▽▽▽  △△●●  △◇□◇  △△□□  △□○□  □✕✕✕

# Words → glyphs (with a key)
glyphriot --key "my secret is real" letter advice cage absurd amount doctor
> ◇●△●  ◇◇▽●  □□●●  ●△△◇  ◇◇✕●  □▽□◇

# Glyphs → words (with a key)
glyphriot --key "my secret is real" ◇●△●  ◇◇▽●  □□●●  ●△△◇  ◇◇✕●  □▽□◇

Glyph:
◇●△●  ◇◇▽●  □□●●  ●△△◇  ◇◇✕●  □▽□◇
Phrase: letter advice cage absurd amount doctor
```

## Inspiration

- Navajo Code Talkers: Human ingenuity using language and context to protect critical communications under extreme pressure during WWII.
- Enigma and Allied cryptanalysis: The cat‑and‑mouse evolution of practical cryptography and operational security.
- Historical ciphers: From substitution and transposition ciphers to rotor machines and one‑time pads—the lineage of obfuscation and key‑based secrecy.

GlyphRiot is not a cipher in the classical sense; it’s an exact, keyed mapping of BIP‑39 words to a compact “glyph alphabet.” It borrows the timeless principle that the secret (key) is what unlocks meaning—while the visible artifact (glyph) can be safely shared or stored.

## Why this matters (offline, salted, storable)

- Protect the words with your key (salt): Use --key or --prompt to apply a deterministic, SHA‑256‑based permutation; same key, same mapping anywhere.
- Offline conversion: Convert your seed phrase to a glyph representation entirely offline—no servers, no telemetry.
- Safe to store the glyph: You can save the glyph in cloud storage (Dropbox, Drive, email yourself). It’s the key that unlocks the proper order and yields the exact phrase.
- Restore forever: As long as you remember the one key you used, you can restore the exact phrase from the glyph—on any machine, offline.

Recommended workflow

- Encode (offline):
    - glyphriot --prompt <words...> (enter key twice; keep a simple, human‑memorable passphrase)
    - Save the resulting glyph somewhere convenient (txt file, notes, cloud).
- Decode (offline):
    - glyphriot --prompt <glyph tokens> (enter the same key; prints the exact phrase)

Threat model

- If the glyph is leaked without your key, the original phrase remains protected.
- If the key is leaked, the glyph must be treated like the seed phrase.
- If you choose no key, the glyph is equivalent to the original words (intentional—for “no‑key” workflows).

## Why it’s uncrackable — with a key

- Keyed permutation, not “just a cipher”: GlyphRiot uses your key to derive a deterministic, cryptographically strong permutation of the 2048 BIP‑39 words (via a SHA‑256 counter‑mode DRBG with unbiased Fisher–Yates). The glyph encodes the position within that permutation. Without the key, the mapping is a uniform shuffle of 2048 slots.

- 2^256 key space: The permutation is derived from SHA‑256(key). An attacker who only sees glyphs and doesn’t know your key faces a 2^256 search space—brute forcing is infeasible.

- No structural leakage: Each glyph token is just a base‑7 index (encoded with △ □ ○ ✕ ● ◇ ▽). Without the permutation P (your key), the index does not reveal the original word or its prefix; the permutation uniformly distributes words across indices.

- Offline and deterministic: Nothing is sent anywhere. There’s no server‑side oracle. Given the same key and list, you get the same mapping on any machine—so you can safely store the glyph and rely on the key alone for recovery.

Threat model

- If the glyph is exposed and the key remains secret: the original phrase is protected.
- If the key is exposed: treat the glyph like the seed phrase.
- If you choose no key: the glyph is equivalent to the original words (intentional for “no‑key” workflows).

Security tips

- Always use a key (salt): Use `--prompt` to enter it twice and confirm.
- Choose a strong passphrase: GlyphRiot hashes your key with SHA‑256 (fast and deterministic). Prefer a long passphrase (e.g., 4–6 random words) for high entropy.
- Keep the key in your head, store the glyph anywhere: The glyph can live in your cloud drive, notes app, or email—because it’s the key that unlocks the correct order.

## Key strength and KDF (default: Argon2id)

- Default: GlyphRiot derives the permutation seed from your key using Argon2id, then hashes with SHA‑256 to a 32‑byte seed. This makes brute‑forcing short passphrases prohibitively costly.
- Enforcement:
    - 12‑word context → require ≥16 characters (when Argon2id is enabled)
    - 24‑word context → require ≥20 characters
    - If you set --kdf=none, we enforce pure entropy by format only:
        - BIP‑39 12/24 words, or
        - Hex (≥32/64 chars), or
        - Base64 (decoded ≥16/32 bytes)
    - Use --allow-weak-key to bypass (not recommended)
- KDF flags (tune to your hardware):
    - --kdf argon2id|none (default: argon2id)
    - --kdf-mem-mb (default: 512)
    - --kdf-time (iterations; default: 3)
    - --kdf-parallel (default: 1)

## How it works

- What this tool does (in one sentence):
    - It gives every seed word its own 4‑glyph “nickname,” and those 4 glyphs always decode back to the exact word.

- Why it works (the simple math):
    - We use 7 easy‑to‑read symbols and always show 4 of them → 7^4 = 2401 possible codes.
    - There are 2048 BIP‑39 words, so each word can have its own unique code with room to spare.

- With or without a key (salt):
    - Optional key reshuffles the word list in a predictable way (wallet‑style).
      Use the same key to get the same codes. No key → standard order.

- Encode (words → glyphs), in plain English:

*   1.  Look up the word’s place in the (possibly reshuffled) 2048‑word list.
*   2.  Turn that position into four base‑7 digits.
*   3.  Replace digits 0..6 with these symbols: △ □ ○ ✕ ● ◇ ▽ (aliases: × and • accepted; ☆ also accepted).

- Decode (glyphs → word), in plain English:
    1. Turn the four symbols back into digits 0..6.
    2. Convert those digits into a number.
    3. Use that number to pick the word from the same list/key.
    4. Result: one code → exactly one word. No guessing.

- Guarantees:
    - Exact round‑trip: the 4 glyphs you see always decode to the original word.
    - Works offline; deterministic; no external services required.

Input rules:

- Only the seven glyphs △, □, ○, ✕, ●, ◇, ▽ are valid (aliases: x/X and × accepted as ✕; • accepted as ●; ☆ also accepted)
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
./bin/glyphriot --prompt                                   # secure key prompt
Enter key: *********
Re-enter key: *********
Seed or Glyph: brave coconut drift zebra
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
# Prompt for key (masked by default); overrides --key if both are present.
# Recommended: enter your seed words or glyph tokens interactively (avoids shell history).
./bin/glyphriot --prompt
Enter key: *********
Re-enter key: *********
Seed or Glyph: brave coconut drift zebra

# Prompt and decode glyphs (interactive entry)
./bin/glyphriot --prompt
Enter key: *********
Re-enter key: *********
Seed or Glyph: △□○✕  □□○✕
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

## License

MIT
