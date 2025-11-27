package internal

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"unicode/utf8"

	"golang.org/x/term"
)

// PromptForKey securely prompts for a key twice and verifies they match.
// If mask is true, input is read in raw mode with '*' echo; otherwise it uses
// the terminal's hidden input (no echo) via ReadPassword.
// Errors are concise and never echo the key content.
func PromptForKey(mask bool) (string, error) {
	fd := int(syscall.Stdin)
	if !term.IsTerminal(fd) {
		return "", fmt.Errorf("prompt requires an interactive terminal")
	}

	// If masking is not requested, use the terminal's hidden password mode (no echo).
	if !mask {
		fmt.Fprint(os.Stdout, "\rEnter key: ")
		k1b, err := term.ReadPassword(fd)
		fmt.Fprintln(os.Stdout)
		if err != nil {
			return "", fmt.Errorf("failed to read key")
		}

		fmt.Fprint(os.Stdout, "\rRe-enter:  ")
		k2b, err := term.ReadPassword(fd)
		fmt.Fprintln(os.Stdout)
		if err != nil {
			return "", fmt.Errorf("failed to read key")
		}

		k1 := string(k1b)
		k2 := string(k2b)
		if k1 != k2 {
			return "", fmt.Errorf("keys do not match")
		}
		return k1, nil
	}

	// Masked input using raw mode with '*' echo and signal-safe restore.
	readMasked := func(prompt string) (string, error) {
		fmt.Fprint(os.Stdout, "\r"+prompt)

		oldState, err := term.GetState(fd)
		if err != nil {
			return "", fmt.Errorf("terminal not ready")
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
			return "", fmt.Errorf("terminal not ready")
		}
		defer func() { restore(); signal.Stop(sigc); close(done) }()
		_ = state

		// UTF-8 aware masked input: accumulate bytes until a full rune is read.
		var bytesBuf []byte
		var runeSizes []int // stack of rune byte lengths for backspace
		var partial []byte  // accumulate partial multi-byte sequences

		for {
			var b [1]byte
			n, er := os.Stdin.Read(b[:])
			if er != nil || n == 0 {
				break
			}
			by := b[0]

			// Enter ends input
			if by == '\r' || by == '\n' {
				fmt.Fprintln(os.Stdout)
				break
			}

			// Handle backspace/delete: remove last full rune (and erase one '*')
			if by == 0x7f || by == '\b' {
				if len(partial) > 0 {
					// Drop any partial sequence being built
					partial = partial[:0]
				} else if len(runeSizes) > 0 {
					sz := runeSizes[len(runeSizes)-1]
					runeSizes = runeSizes[:len(runeSizes)-1]
					if sz > 0 && sz <= len(bytesBuf) {
						bytesBuf = bytesBuf[:len(bytesBuf)-sz]
						fmt.Fprint(os.Stdout, "\b \b")
					}
				}
				continue
			}

			// Ignore other control characters
			if by < 0x20 {
				continue
			}

			// Build UTF-8 sequence
			partial = append(partial, by)
			if utf8.FullRune(partial) {
				r, size := utf8.DecodeRune(partial)
				if r == utf8.RuneError && size == 1 {
					// Invalid byte, skip
					partial = partial[:0]
					continue
				}
				// Commit full rune
				bytesBuf = append(bytesBuf, partial...)
				runeSizes = append(runeSizes, size)
				partial = partial[:0]
				fmt.Fprint(os.Stdout, "*")
			}
		}
		return string(bytesBuf), nil
	}

	k1, err := readMasked("Enter key: ")
	if err != nil {
		return "", err
	}
	k2, err := readMasked("Re-enter:  ")
	if err != nil {
		return "", err
	}
	if k1 != k2 {
		return "", fmt.Errorf("keys do not match")
	}
	return k1, nil
}
