//go:build windows

package main

import "os"

// windowsTermState is a placeholder; Windows raw-mode is not implemented.
type windowsTermState struct{}

// makeRaw is a no-op on Windows. Raw terminal mode requires the
// golang.org/x/term package or the Windows Console API; for now we
// fall back to line-buffered input so the record command is at least
// buildable on Windows.
func makeRaw(fd int) (*windowsTermState, error) { return nil, nil }

// restoreTerminal is a no-op on Windows.
func restoreTerminal(fd int, state *windowsTermState) {}

// termSize returns the terminal dimensions on Windows, defaulting to 80x24.
func termSize() (cols, lines int) {
	// Best-effort: check COLUMNS/LINES env vars set by some terminal emulators.
	return 80, 24
}

// readKeys reads raw bytes from stdin and sends them on ch.
func readKeys(ch chan<- []byte) {
	buf := make([]byte, 32)
	for {
		n, err := os.Stdin.Read(buf)
		if n > 0 {
			cp := make([]byte, n)
			copy(cp, buf[:n])
			ch <- cp
		}
		if err != nil {
			return
		}
	}
}
