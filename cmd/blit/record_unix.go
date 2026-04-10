//go:build !windows

package main

import (
	"golang.org/x/sys/unix"
	"os"
)

// makeRaw puts fd into raw mode and returns the previous terminal state.
// Returns nil state (no error) when fd is not a terminal.
func makeRaw(fd int) (*unix.Termios, error) {
	if !isTerminal(fd) {
		return nil, nil
	}
	old, err := unix.IoctlGetTermios(fd, ioctlReadTermios)
	if err != nil {
		return nil, err
	}
	raw := *old
	raw.Iflag &^= unix.IGNBRK | unix.BRKINT | unix.PARMRK | unix.ISTRIP | unix.INLCR | unix.IGNCR | unix.ICRNL | unix.IXON
	raw.Oflag &^= unix.OPOST
	raw.Lflag &^= unix.ECHO | unix.ECHONL | unix.ICANON | unix.ISIG | unix.IEXTEN
	raw.Cflag &^= unix.CSIZE | unix.PARENB
	raw.Cflag |= unix.CS8
	raw.Cc[unix.VMIN] = 1
	raw.Cc[unix.VTIME] = 0
	if err := unix.IoctlSetTermios(fd, ioctlWriteTermios, &raw); err != nil {
		return nil, err
	}
	return old, nil
}

// restoreTerminal restores the terminal to its previous state.
func restoreTerminal(fd int, state *unix.Termios) {
	if state == nil {
		return
	}
	_ = unix.IoctlSetTermios(fd, ioctlWriteTermios, state)
}

// isTerminal reports whether fd is a tty.
func isTerminal(fd int) bool {
	_, err := unix.IoctlGetTermios(fd, ioctlReadTermios)
	return err == nil
}

// termSize returns the current terminal dimensions.
func termSize() (cols, lines int) {
	ws, err := unix.IoctlGetWinsize(int(os.Stdout.Fd()), unix.TIOCGWINSZ)
	if err != nil {
		return 80, 24
	}
	return int(ws.Col), int(ws.Row)
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
