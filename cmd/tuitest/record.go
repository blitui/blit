package main

// record.go implements the `tuitest record <name> -- <command>` subcommand.
//
// It runs the target command inside a pseudo-tty, intercepts every keypress,
// window resize, and tick event, then serialises the capture as a
// testdata/sessions/<name>.tuisess file.
//
// Record UI (status bar rendered to stderr):
//
//	[REC] keystrokes: 42  length: 00:01:23  (ctrl+p = pause/resume, ctrl+c = stop+save)
//
// The status bar is redrawn on every event using a \r carriage-return so it
// stays on a single line without scrolling the terminal.

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

// recordSession holds the mutable state accumulated during a recording.
type recordSession struct {
	Name      string
	Command   []string
	StartedAt time.Time
	Steps     []recordStep
	keystroke int
	paused    bool
}

// recordStep mirrors tuitest.SessionStep but lives in cmd/tuitest to avoid
// import cycles (the CLI has no direct dependency on the library package).
type recordStep struct {
	Kind   string `json:"kind"`
	Key    string `json:"key,omitempty"`
	Text   string `json:"text,omitempty"`
	Cols   int    `json:"cols,omitempty"`
	Lines  int    `json:"lines,omitempty"`
	Screen string `json:"screen,omitempty"`
}

// tuiSessFile is the on-disk format written by the record command.
// Version 2 adds Name, RecordedAt, Command metadata.
type tuiSessFile struct {
	Version    int          `json:"version"`
	Cols       int          `json:"cols"`
	Lines      int          `json:"lines"`
	Name       string       `json:"name,omitempty"`
	RecordedAt string       `json:"recorded_at,omitempty"`
	Command    []string     `json:"command,omitempty"`
	Steps      []recordStep `json:"steps"`
}

// runRecord is the entry point for `tuitest record`.
func runRecord(args []string) int {
	fs := flag.NewFlagSet("record", flag.ContinueOnError)
	outDir := fs.String("dir", "testdata/sessions", "directory to write .tuisess files")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: tuitest record [flags] <name> -- <command> [args...]")
		fs.PrintDefaults()
	}

	// Split on "--" to separate tuitest flags from the target command.
	splitIdx := -1
	for i, a := range args {
		if a == "--" {
			splitIdx = i
			break
		}
	}
	if splitIdx < 0 {
		fmt.Fprintln(os.Stderr, "tuitest record: missing '--' separator before command")
		fs.Usage()
		return 1
	}

	tuitestArgs := args[:splitIdx]
	cmdArgs := args[splitIdx+1:]

	if err := fs.Parse(tuitestArgs); err != nil {
		return 1
	}
	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "tuitest record: <name> required")
		fs.Usage()
		return 1
	}
	if len(cmdArgs) == 0 {
		fmt.Fprintln(os.Stderr, "tuitest record: command required after '--'")
		fs.Usage()
		return 1
	}

	name := fs.Arg(0)
	sess := &recordSession{
		Name:      name,
		Command:   cmdArgs,
		StartedAt: time.Now(),
	}

	return runRecordSession(sess, *outDir)
}

// runRecordSession drives the recording loop.
// It launches the target command, captures stdin keystrokes, and on exit
// (Ctrl+C or natural termination) saves the .tuisess file.
func runRecordSession(sess *recordSession, outDir string) int {
	// Determine terminal size.
	cols, lines := termSize()

	fmt.Fprintf(os.Stderr, "[tuitest] recording %q → %s/%s.tuisess\n", sess.Name, outDir, sess.Name)
	fmt.Fprintln(os.Stderr, "[tuitest] ctrl+p = pause/resume  ctrl+c = stop+save")

	// Set up signal handling for clean exit.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Channel for raw key bytes read from stdin.
	keyCh := make(chan []byte, 64)

	// Put stdin into raw mode so we can intercept individual keystrokes.
	oldState, err := makeRaw(int(os.Stdin.Fd()))
	if err != nil {
		fmt.Fprintf(os.Stderr, "[tuitest] warning: could not set raw mode: %v\n", err)
	}

	// Start reading keystrokes in a goroutine.
	go readKeys(keyCh)

	// Launch the target command, forwarding its stdio.
	cmd := exec.Command(sess.Command[0], sess.Command[1:]...) //nolint:gosec
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if startErr := cmd.Start(); startErr != nil {
		if oldState != nil {
			restoreTerminal(int(os.Stdin.Fd()), oldState)
		}
		fmt.Fprintf(os.Stderr, "[tuitest] failed to start command: %v\n", startErr)
		return 1
	}

	// cmdDone fires when the child process exits.
	cmdDone := make(chan error, 1)
	go func() { cmdDone <- cmd.Wait() }()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	drawStatusBar(sess, cols)

loop:
	for {
		select {
		case raw := <-keyCh:
			key := parseRawKey(raw)

			// ctrl+p → pause/resume
			if key == "ctrl+p" {
				sess.paused = !sess.paused
				drawStatusBar(sess, cols)
				continue
			}
			// ctrl+c → stop and save
			if key == "ctrl+c" {
				_ = cmd.Process.Signal(syscall.SIGINT)
				break loop
			}

			if !sess.paused {
				sess.keystroke++
				sess.Steps = append(sess.Steps, recordStep{Kind: "key", Key: key})
			}
			drawStatusBar(sess, cols)

		case <-ticker.C:
			if !sess.paused {
				sess.Steps = append(sess.Steps, recordStep{Kind: "tick"})
			}
			drawStatusBar(sess, cols)

		case <-sigCh:
			_ = cmd.Process.Signal(syscall.SIGTERM)
			break loop

		case <-cmdDone:
			break loop
		}
	}

	// Restore terminal before writing output.
	if oldState != nil {
		restoreTerminal(int(os.Stdin.Fd()), oldState)
	}
	fmt.Fprintln(os.Stderr) // newline after status bar

	return saveRecordedSession(sess, outDir, cols, lines)
}

// saveRecordedSession writes the .tuisess file atomically (temp + rename).
func saveRecordedSession(sess *recordSession, outDir string, cols, lines int) int {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "[tuitest] mkdir %s: %v\n", outDir, err)
		return 1
	}

	out := tuiSessFile{
		Version:    2,
		Cols:       cols,
		Lines:      lines,
		Name:       sess.Name,
		RecordedAt: sess.StartedAt.UTC().Format(time.RFC3339),
		Command:    sess.Command,
		Steps:      sess.Steps,
	}

	data, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "[tuitest] marshal: %v\n", err)
		return 1
	}

	dest := filepath.Join(outDir, sess.Name+".tuisess")
	tmp := dest + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "[tuitest] write tmp: %v\n", err)
		return 1
	}
	if err := os.Rename(tmp, dest); err != nil {
		fmt.Fprintf(os.Stderr, "[tuitest] rename: %v\n", err)
		return 1
	}

	elapsed := time.Since(sess.StartedAt).Round(time.Second)
	fmt.Fprintf(os.Stderr, "[tuitest] saved %s (%d keystrokes, %s)\n", dest, sess.keystroke, elapsed)
	return 0
}

// drawStatusBar renders a single-line status bar to stderr using \r.
func drawStatusBar(sess *recordSession, cols int) {
	elapsed := time.Since(sess.StartedAt).Round(time.Second)
	h := int(elapsed.Hours())
	m := int(elapsed.Minutes()) % 60
	s := int(elapsed.Seconds()) % 60
	timeStr := fmt.Sprintf("%02d:%02d:%02d", h, m, s)

	pauseStr := ""
	if sess.paused {
		pauseStr = "  [PAUSED]"
	}

	bar := fmt.Sprintf("[REC] keystrokes: %d  length: %s%s  (ctrl+p=pause/resume ctrl+c=stop+save)",
		sess.keystroke, timeStr, pauseStr)

	// Truncate to terminal width to avoid line wrapping.
	if cols > 0 && len(bar) > cols {
		bar = bar[:cols]
	}
	// Pad with spaces to erase previous longer line.
	pad := strings.Repeat(" ", max(0, cols-len(bar)))
	fmt.Fprintf(os.Stderr, "\r%s%s", bar, pad)
}

// parseRawKey converts a raw byte sequence from the terminal into a key name
// compatible with tuitest's keyMap. This is a best-effort mapping for the
// most common keys; unknown sequences are hex-encoded.
func parseRawKey(raw []byte) string {
	if len(raw) == 0 {
		return ""
	}
	if len(raw) == 1 {
		b := raw[0]
		switch {
		case b == 0x01:
			return "ctrl+a"
		case b == 0x02:
			return "ctrl+b"
		case b == 0x03:
			return "ctrl+c"
		case b == 0x04:
			return "ctrl+d"
		case b == 0x05:
			return "ctrl+e"
		case b == 0x06:
			return "ctrl+f"
		case b == 0x07:
			return "ctrl+g"
		case b == 0x08:
			return "backspace"
		case b == 0x09:
			return "tab"
		case b == 0x0a || b == 0x0d:
			return "enter"
		case b == 0x0b:
			return "ctrl+k"
		case b == 0x0c:
			return "ctrl+l"
		case b == 0x0e:
			return "ctrl+n"
		case b == 0x0f:
			return "ctrl+o"
		case b == 0x10:
			return "ctrl+p"
		case b == 0x11:
			return "ctrl+q"
		case b == 0x12:
			return "ctrl+r"
		case b == 0x13:
			return "ctrl+s"
		case b == 0x14:
			return "ctrl+t"
		case b == 0x15:
			return "ctrl+u"
		case b == 0x16:
			return "ctrl+v"
		case b == 0x17:
			return "ctrl+w"
		case b == 0x18:
			return "ctrl+x"
		case b == 0x19:
			return "ctrl+y"
		case b == 0x1a:
			return "ctrl+z"
		case b == 0x1b:
			return "esc"
		case b == 0x7f:
			return "backspace"
		case b == 0x20:
			return "space"
		case b >= 0x21 && b <= 0x7e:
			return string(rune(b))
		}
	}
	// Escape sequences for arrow keys and common special keys.
	s := string(raw)
	switch s {
	case "\x1b[A":
		return "up"
	case "\x1b[B":
		return "down"
	case "\x1b[C":
		return "right"
	case "\x1b[D":
		return "left"
	case "\x1b[H", "\x1b[1~":
		return "home"
	case "\x1b[F", "\x1b[4~":
		return "end"
	case "\x1b[5~":
		return "pgup"
	case "\x1b[6~":
		return "pgdown"
	case "\x1b[2~":
		return "insert"
	case "\x1b[3~":
		return "delete"
	case "\x1bOP":
		return "f1"
	case "\x1bOQ":
		return "f2"
	case "\x1bOR":
		return "f3"
	case "\x1bOS":
		return "f4"
	case "\x1b[15~":
		return "f5"
	case "\x1b[17~":
		return "f6"
	case "\x1b[18~":
		return "f7"
	case "\x1b[19~":
		return "f8"
	case "\x1b[20~":
		return "f9"
	case "\x1b[21~":
		return "f10"
	case "\x1b[23~":
		return "f11"
	case "\x1b[24~":
		return "f12"
	}
	// Unknown: represent as hex so the session is still valid JSON.
	var sb strings.Builder
	for _, b := range raw {
		fmt.Fprintf(&sb, "%02x", b)
	}
	return "0x" + sb.String()
}

// max returns the larger of a and b.
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
