package internal

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"
)

// RunSelfTest generates randomized self-test sets of size specified in `sets`,
// prints each set (words and glyphs), verifies exact round-trip decoding, and
// returns the number of failed sets.
//
// Parameters:
// - wordsList: the canonical word list (length must be Total, e.g., 2048)
// - index:     map of lowercased word -> index in wordsList
// - keyStr:    key/salt for permutation ("" = identity order)
// - glyphSep:  optional separator to insert between glyphs for readability
// - paginate:  whether to paginate output when the terminal is a TTY
// - height:    terminal height in rows for pagination logic
// - sets:      slice of test sizes (e.g., []int{12, 24})
// - title:     heading to print once at the top (empty to skip)
func RunSelfTest(wordsList []string, index map[string]int, keyStr string, glyphSep string, paginate bool, height int, sets []int, title string) int {
	perm, _ := Derive(len(wordsList), keyStr)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	failed := 0

	printed := 0
	header := func() {
		if title != "" {
			fmt.Println(Style(title, Bold, Blue))
			printed++
		}
	}

	if paginate {
		header()
	}

	for si, sz := range sets {
		// Assemble a randomized, non-repeating set of positions in [0..n-1]
		seen := make(map[int]bool, sz)
		seq := make([]int, 0, sz)
		for len(seq) < sz {
			p := r.Intn(len(wordsList))
			if seen[p] {
				continue
			}
			seen[p] = true
			seq = append(seq, p)
		}

		// Build the test words via permutation (pos -> word index)
		words := make([]string, sz)
		for i := 0; i < sz; i++ {
			words[i] = wordsList[perm[seq[i]]]
		}

		// Encode to glyphs
		glyphs, err := EncodeWords(words, index, wordsList, keyStr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "self-test encode error: %v\n", err)
			failed++
			continue
		}
		// Insert separator for readability (if provided)
		for i := range glyphs {
			glyphs[i] = InsertSep(glyphs[i], glyphSep)
		}

		// Verify exact round-trip using batch decode
		okAll := true
		decoded, derr := DecodeGlyphTokens(glyphs, wordsList, keyStr)
		if derr != nil {
			okAll = false
		} else {
			if len(decoded) != len(words) {
				okAll = false
			} else {
				for i := range words {
					if decoded[i] != words[i] {
						okAll = false
						break
					}
				}
			}
		}

		// Print this set
		// Only print "Set N:" when multiple sets are requested
		if len(sets) > 1 {
			title := fmt.Sprintf("Set %d:", si+1)
			fmt.Println(Style(title, Bold, Purple))
			printed++
		}

		// Words block (split across two lines if 24)
		if len(words) == 24 {
			fmt.Printf("  Words:  %s\n", strings.Join(words[:12], " "))
			printed++
			fmt.Printf("          %s\n", strings.Join(words[12:], " "))
			printed++
		} else {
			fmt.Printf("  Words:  %s\n", strings.Join(words, " "))
			printed++
		}

		// Glyphs block (split similarly for 24)
		if len(glyphs) == 24 {
			fmt.Printf("  Glyphs: %s\n", strings.Join(glyphs[:12], "  "))
			fmt.Printf("          %s\n", strings.Join(glyphs[12:], "  "))
			printed += 2
		} else {
			fmt.Printf("  Glyphs: %s\n", strings.Join(glyphs, "  "))
			printed++
		}

		// Result line
		var result string
		if okAll {
			result = "PASSED"
		} else {
			result = "FAILED"
			failed++
		}
		label := fmt.Sprintf("Result: %s — Verified: %d passes", result, sz)
		fmt.Println(Style("  "+label, Bold))
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

	// Summary (only when multiple sets)
	if len(sets) > 1 {
		fmt.Printf("%s %d, %s %d\n",
			Style("Total sets:", Bold), len(sets),
			Style("Failed:", Bold), failed)
	}

	return failed
}
