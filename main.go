// GlyphRiot — SeedGlyph generator (standardized wallet-style mapping)
//
// Single, standardized scheme:
// - Fixed 7 glyph digits (0..6): 0:△ 1:□ 2:○ 3:× 4:• 5:◇ 6:☆
// - Fixed length: 4 glyphs per word
// - Key/salt: SHA-256(key) → PRNG → deterministic permutation P over 2048
//
// Encoding (word -> glyphs):
// - Find word index i in active list, get pos = P^-1[i]
// - Since 7^4=2401 >=2048, bucket index b = pos
// - Convert b to base-7 (4 digits) → map digits to glyphs
//
// Decoding (glyphs -> candidates):
// - Convert 4 glyphs to base-7 bucket index b
// - If b >=2048, invalid
// - Else, candidate = P[b] → exact word
// - This provides 100% exact, unique round-trip without guessing or checksum reliance.
//
// Notes:
// - We keep list selection (--list, --list-file) and key input (--key/--prompt)
// - Input decoding accepts the exact 7 glyphs and x/X for ×

package main

import (
	_ "embed"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"path/filepath"

	"strings"
	"syscall"
	"time"

	"glyphriot/internal"

	"golang.org/x/term"
)

//go:embed english.txt
var englishTxt string

// WordList encapsulates a seed word list and fast lookup index
type WordList struct {
	Name  string
	Words []string
	Index map[string]int // lowercased word -> index
}

func buildWordList(name, txt string) WordList {
	lines := strings.Split(strings.TrimSpace(txt), "\n")
	idx := make(map[string]int, len(lines))
	for i, w := range lines {
		lw := strings.ToLower(strings.TrimSpace(w))
		if lw == "" {
			continue
		}
		idx[lw] = i
	}
	return WordList{Name: name, Words: lines, Index: idx}
}

func loadListFile(path string) (WordList, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return WordList{}, err
	}
	lines := strings.Split(strings.ReplaceAll(string(b), "\r\n", "\n"), "\n")
	// trim trailing empty
	if len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}
	idx := make(map[string]int, len(lines))
	for i, w := range lines {
		lw := strings.ToLower(strings.TrimSpace(w))
		lines[i] = lw
		if lw == "" {
			continue
		}
		idx[lw] = i
	}
	if len(lines) != 2048 {
		fmt.Fprintf(os.Stderr, "warning: --list-file expected 2048 lines, got %d\n", len(lines))
	}
	return WordList{Name: "custom", Words: lines, Index: idx}, nil
}

var wlBip39 = buildWordList("bip39-en", englishTxt)

func usage() {
	prog := filepath.Base(os.Args[0])

	// Headline
	fmt.Println(internal.Style("GlyphRiot — Glyph Seed System v1.0 (standardized)", internal.Bold, internal.Purple))
	fmt.Println()

	// Usage
	fmt.Println(internal.Style("Usage:", internal.Bold, internal.Blue))
	fmt.Printf("  %s %s\n", prog, internal.Style("[options] [word ...] | [glyphs]", internal.Cyan))
	fmt.Println()

	// Flags
	fmt.Println(internal.Style("Flags:", internal.Bold, internal.Blue))
	fmt.Println(internal.Style("  --all  --list  --list-file  --key|--prompt  --pager  --glyph-sep  --phrase-only  --no-color", internal.Cyan))
	fmt.Println()

	// Glyphs and rules
	fmt.Println(internal.Style("Glyphs (fixed):", internal.Bold, internal.Blue), "△ □ ○ × • ◇ ☆", internal.Style("(always 4 glyphs per word)", internal.Gray))
	fmt.Println(internal.Style("Words → glyphs; glyphs → exact word. Accepts x/X as ×.", internal.Gray))
	fmt.Println()

	// Examples (12-word, valid BIP-39)
	fmt.Println(internal.Style("Examples (12-word mnemonic):", internal.Bold, internal.Blue))
	ex := "letter advice cage absurd amount doctor acoustic avoid letter advice cage above"
	fmt.Printf("  %s %s\n", prog, ex)

	// Decoding (concise)
	fmt.Printf("  %s '<12 tokens>'\n", prog)

	// Table
	fmt.Printf("  %s --all\n", prog)
}

// Fixed glyph digit set and decoding (from internal/glyphs)
var glyphDigits = internal.Digits
var glyphDecode = internal.Decode

// Bucket/base constants (from internal/glyphs)
const (
	bucketBase = internal.Base
	bucketLen  = internal.Len
	wordsTotal = internal.Total
)

func posToBucket(pos int) int {
	code, _ := internal.PosToCode(pos)
	return code
}

func bucketStartSize(b int) (start, size int) {
	if _, ok := internal.CodeToPos(b); !ok {
		return 0, 0
	}
	return b, 1
}

func bucketToDigits(b int) [bucketLen]int {
	d, _ := internal.ToDigits(b)
	return d
}

func digitsToBucket(d []int) (int, bool) {
	return internal.FromDigits(d)
}

func promptForKey(mask bool) (string, error) {
	fd := int(syscall.Stdin)
	if !term.IsTerminal(fd) {
		return "", fmt.Errorf("--prompt requires a TTY (interactive terminal)")
	}
	fmt.Fprint(os.Stderr, "Enter key: ")
	if !mask {
		b, err := term.ReadPassword(fd)
		fmt.Fprintln(os.Stderr)
		if err != nil {
			return "", err
		}
		return string(b), nil
	}
	// masked: raw-mode with '*' echo and signal-safe restore
	oldState, err := term.GetState(fd)
	if err != nil {
		return "", err
	}
	restore := func() { _ = term.Restore(fd, oldState) }
	done := make(chan struct{})
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, syscall.SIGTERM)
	go func() {
		select {
		case <-sigc:
			restore()
			os.Exit(130)
		case <-done:
		}
	}()
	state, err := term.MakeRaw(fd)
	if err != nil {
		signal.Stop(sigc)
		close(done)
		return "", err
	}
	defer func() { restore(); signal.Stop(sigc); close(done) }()
	_ = state
	var buf []rune
	for {
		var b [1]byte
		n, er := os.Stdin.Read(b[:])
		if er != nil || n == 0 {
			break
		}
		ch := rune(b[0])
		if ch == '\r' || ch == '\n' {
			fmt.Fprintln(os.Stderr)
			break
		}
		if ch == 0x7f || ch == '\b' { // backspace
			if len(buf) > 0 {
				buf = buf[:len(buf)-1]
				fmt.Fprint(os.Stderr, "\b \b")
			}
			continue
		}
		buf = append(buf, ch)
		fmt.Fprint(os.Stderr, "*")
	}
	return string(buf), nil
}

func runSelfTest(active WordList, keyStr string, glyphSep string, paginate bool, height int) int {
	p, _ := internal.Derive(len(active.Words), keyStr)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	sets := []int{12, 12, 12, 12}
	failed := 0

	printed := 0
	header := func() {
		fmt.Println(internal.Style("Self-test (12-word sets)", internal.Bold, internal.Blue))
		printed++
	}

	if paginate {
		header()
	}

	for si, sz := range sets {
		poseen := make(map[int]bool)
		seq := make([]int, 0, sz)
		for len(seq) < sz {
			p := r.Intn(len(active.Words))
			if poseen[p] {
				continue
			}
			poseen[p] = true
			seq = append(seq, p)
		}
		words := make([]string, sz)
		for i := 0; i < sz; i++ {
			words[i] = active.Words[p[seq[i]]]
		}
		glyphs, err := internal.EncodeWords(words, active.Index, active.Words, keyStr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "self-test encode error: %v\n", err)
			failed++
			continue
		}
		// Format glyphs with separator
		for i := range glyphs {
			glyphs[i] = internal.InsertSep(glyphs[i], glyphSep)
		}
		// Verify each token still contains its word
		okAll := true
		for i := 0; i < sz; i++ {
			tokNorm := internal.StripSepAndSpaces(glyphs[i], glyphSep)
			runes := []rune(tokNorm)
			d := make([]int, bucketLen)
			for j, rch := range runes {
				d[j] = glyphDecode[rch]
			}
			b, _ := digitsToBucket(d)
			start, size := bucketStartSize(b)
			found := false
			for k := 0; k < size; k++ {
				idx := p[start+k]
				if idx >= 0 && idx < len(active.Words) && active.Words[idx] == words[i] {
					found = true
					break
				}
			}
			if !found {
				okAll = false
				break
			}
		}

		// Print block for this set
		title := fmt.Sprintf("Set %d:", si+1)
		fmt.Println(internal.Style(title, internal.Bold, internal.Purple))
		printed++
		fmt.Printf("  Words:  %s\n", strings.Join(words, " "))
		printed++
		fmt.Printf("  Glyphs: %s\n", strings.Join(glyphs, "  "))
		printed++

		printed++

		label := fmt.Sprintf("Result: %s", map[bool]string{true: "PASSED"}[okAll])
		if !okAll {
			label = "Result: FAILED"
			failed++
		}
		fmt.Println(internal.Style("  "+label, internal.Bold))
		printed++

		// Pagination
		if paginate && printed >= height-1 {
			fmt.Fprint(os.Stderr, "-- more -- (Enter to continue, q to quit) ")
			var buf [1]byte
			_, er := os.Stdin.Read(buf[:])
			fmt.Fprintln(os.Stderr)
			if er == nil && (buf[0] == 'q' || buf[0] == 'Q') {
				break
			}
			printed = 0
			header()
		}
	}

	// Summary
	fmt.Printf("%s %d, %s %d\n",
		internal.Style("Total sets:", internal.Bold), len(sets),
		internal.Style("Failed:", internal.Bold), failed)

	return failed
}

func main() {
	all := flag.Bool("all", false, "Generate full table for the selected word list")
	table := flag.Bool("table", false, "Tabular output for provided words/phrase")
	sep := flag.String("sep", "  ", "Separator between glyphs for phrase output")
	list := flag.String("list", "bip39-en", "Word list: bip39-en (default), auto")
	listFile := flag.String("list-file", "", "Load a custom 2048-word list from file (overrides --list)")
	key := flag.String("key", "", "User key to reorder word mapping")
	prompt := flag.Bool("prompt", false, "Securely prompt for key (no echo); overrides --key")
	mask := flag.Bool("mask", true, "With --prompt, show * while typing (use --mask=false to disable)")
	pager := flag.Bool("pager", true, "Paginate --all output when writing to a TTY (press Enter per page); --pager=false to disable")
	selfTest := flag.Bool("self-test", false, "Run built-in test harness (4×12-word phrases)")
	phraseOnly := flag.Bool("phrase-only", false, "Print only the recovered phrase when decoding glyphs")

	glyphSep := flag.String("glyph-sep", "", "Insert this separator between glyphs when printing; decoding strips it")
	noColor := flag.Bool("no-color", false, "Disable colored output (TTY-safe)")

	flag.Parse()

	// Color enablement: default on for TTY unless --no-color
	internal.SetColorEnabled(!*noColor && term.IsTerminal(int(syscall.Stdout)))

	// Determine active word list
	var active WordList
	if strings.TrimSpace(*listFile) != "" {
		wl, err := loadListFile(*listFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: failed to load --list-file: %v\n", err)
			os.Exit(2)
		}
		active = wl
	} else {
		switch strings.ToLower(strings.TrimSpace(*list)) {
		case "", "bip39-en", "auto":
			active = wlBip39
		default:
			fmt.Fprintf(os.Stderr, "warning: unknown --list=%q; defaulting to bip39-en\n", *list)
			active = wlBip39
		}
	}

	// Resolve key
	keyStr := *key
	if *prompt {
		ks, err := promptForKey(*mask)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(2)
		}
		keyStr = ks
	}

	// Self-test
	if *selfTest {
		// Paginate self-test output similar to --all
		outIsTTY := term.IsTerminal(int(syscall.Stdout))
		paginate := *pager && outIsTTY
		_, height, _ := term.GetSize(int(syscall.Stdout))
		if height <= 0 {
			height = 24
		}
		failed := runSelfTest(active, keyStr, *glyphSep, paginate, height)
		if failed > 0 {
			os.Exit(1)
		}
		return
	}

	// table flag retained for compatibility; not used in standardized scheme
	_ = table
	if *all {
		p, _ := internal.Derive(len(active.Words), keyStr)
		outIsTTY := term.IsTerminal(int(syscall.Stdout))
		paginate := *pager && outIsTTY
		_, height, _ := term.GetSize(int(syscall.Stdout))
		if height <= 0 {
			height = 24
		}
		header := func() int {
			fmt.Printf("List: %s\n", active.Name)
			c := 1
			if strings.TrimSpace(keyStr) != "" {
				fmt.Printf("Key: set\n")
				c++
			}
			fmt.Printf("%-4s %-12s %s\n", "Idx", "Word", "Glyph")
			c++
			fmt.Printf("%s\n", strings.Repeat("─", 30))
			c++
			return c
		}
		printed := header()
		inv2 := internal.Inv(p)
		for i, word := range active.Words {
			pos := inv2[i]
			b := pos // unique mapping in base-7
			d := bucketToDigits(b)
			var sb strings.Builder
			for j := 0; j < bucketLen; j++ {
				sb.WriteRune(glyphDigits[d[j]])
			}
			fmt.Printf("%4d %-12s %s\n", i, word, internal.InsertSep(sb.String(), *glyphSep))
			printed++
			if paginate && printed >= height-1 {
				fmt.Fprint(os.Stderr, "-- more -- (Enter to continue, q to quit) ")
				var buf [1]byte
				_, er := os.Stdin.Read(buf[:])
				fmt.Fprintln(os.Stderr)
				if er == nil && (buf[0] == 'q' || buf[0] == 'Q') {
					return
				}
				printed = header()
			}
		}
		return
	}

	tokens := flag.Args()
	if len(tokens) == 0 {
		usage()
		os.Exit(0)
	}

	// Normalize glyph tokens and detect glyph input
	normTokens := make([]string, len(tokens))
	for i, t := range tokens {
		normTokens[i] = internal.StripSepAndSpaces(strings.TrimSpace(t), *glyphSep)
	}
	isGlyph := true
	for _, t := range normTokens {
		r := []rune(t)
		if len(r) != bucketLen {
			isGlyph = false
			break
		}
		for _, ch := range r {
			if _, ok := glyphDecode[ch]; !ok {
				isGlyph = false
				break
			}
		}
		if !isGlyph {
			break
		}
	}

	if isGlyph {

		decoded := make([]string, 0, len(normTokens))

		// Decode all tokens first
		for _, tok := range normTokens {
			word, err := internal.DecodeGlyphToken(strings.TrimSpace(tok), active.Words, keyStr)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(2)
			}
			decoded = append(decoded, word)
		}

		// If phrase-only, print just the phrase and exit
		if *phraseOnly {
			fmt.Println(strings.Join(decoded, " "))
			return
		}

		// Otherwise, print per-token lines, plus the full phrase
		fmt.Printf("List: %s\n", active.Name)
		if strings.TrimSpace(keyStr) != "" {
			fmt.Printf("Key: set\n")
		}
		for i, tok := range normTokens {
			fmt.Printf("%s → %s\n", internal.InsertSep(tok, *glyphSep), decoded[i])
		}
		fmt.Println("Phrase:", strings.Join(decoded, " "))

		return
	}

	// Otherwise treat as words → glyphs
	glyphs, err := internal.EncodeWords(tokens, active.Index, active.Words, keyStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(2)
	}
	for i := range glyphs {
		glyphs[i] = internal.InsertSep(glyphs[i], *glyphSep)
	}
	if len(glyphs) == 24 {
		fmt.Println(strings.Join(glyphs[:12], *sep))
		fmt.Println(strings.Join(glyphs[12:], *sep))
	} else {
		fmt.Println(strings.Join(glyphs, *sep))
	}
}

// Helpers
