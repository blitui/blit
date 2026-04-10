//go:build windows

package main

import (
	"errors"
	"os"
)

// windowsTermState is a placeholder; Windows raw-mode is not implemented.
type windowsTermState struct{}

// makeRaw returns an error on Windows because raw terminal mode is not
// implemented. The record command requires a Unix pseudo-tty.
func makeRaw(fd int) (*windowsTermState, error) {
	return nil, errors.New("blit record requires a Unix terminal (raw mode is not supported on Windows)")
}

// restoreTerminal is a no-op on Windows.
func restoreTerminal(fd int, state *windowsTermState) {}

// termSize returns the terminal dimensions on Windows, defaulting to 80x24.
func termSize() (cols, lines int) {
	return 80, 24
}

// readKeys reads raw bytes from stdin and sends them on ch until done is closed.
func readKeys(ch chan<- []byte, done <-chan struct{}) {
	buf := make([]byte, 32)
	for {
		select {
		case <-done:
			return
		default:
		}
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
