package internal

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

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

		fmt.Fprint(os.Stdout, "\rRe-enter key: ")
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

		var buf []rune
		for {
			var b [1]byte
			n, er := os.Stdin.Read(b[:])
			if er != nil || n == 0 {
				break
			}
			ch := rune(b[0])
			if ch == '\r' || ch == '\n' {
				fmt.Fprintln(os.Stdout)
				break
			}
			if ch == 0x7f || ch == '\b' { // backspace/delete
				if len(buf) > 0 {
					buf = buf[:len(buf)-1]
					// Erase last '*'
					fmt.Fprint(os.Stdout, "\b \b")
				}
				continue
			}
			// Ignore non-printable control characters
			if ch < 0x20 || ch == 0x7f {
				continue
			}
			buf = append(buf, ch)
			fmt.Fprint(os.Stdout, "*")
		}
		return string(buf), nil
	}

	k1, err := readMasked("Enter key: ")
	if err != nil {
		return "", err
	}
	k2, err := readMasked("Re-enter key: ")
	if err != nil {
		return "", err
	}
	if k1 != k2 {
		return "", fmt.Errorf("keys do not match")
	}
	return k1, nil
}
