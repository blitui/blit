package main

// replay.go implements the `tuitest replay <name>` subcommand.
//
// It reads a .tuisess file, drives the recorded steps through a virtual
// terminal, and renders the replay progress UI to stderr.
//
// Replay UI (status bar rendered to stderr):
//
//	[REPLAY] step 42/120 ████████░░░░░░░░ 35%  speed: 1.0x  (ctrl+c = abort)

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// runReplay is the entry point for `tuitest replay`.
func runReplay(args []string) int {
	fs := flag.NewFlagSet("replay", flag.ContinueOnError)
	inDir := fs.String("dir", "testdata/sessions", "directory containing .tuisess files")
	speed := fs.Float64("speed", 1.0, "playback speed multiplier (e.g. 2x, 0.5x)")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: tuitest replay [flags] <name>")
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		return 1
	}
	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "tuitest replay: <name> required")
		fs.Usage()
		return 1
	}

	name := fs.Arg(0)
	path := filepath.Join(*inDir, name+".tuisess")

	data, err := os.ReadFile(path)
	if err != nil {
		// Try without adding the extension in case name already has it.
		if filepath.Ext(name) == "" {
			fmt.Fprintf(os.Stderr, "[tuitest] cannot read %s: %v\n", path, err)
			return 1
		}
		path = filepath.Join(*inDir, name)
		data, err = os.ReadFile(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[tuitest] cannot read %s: %v\n", path, err)
			return 1
		}
	}

	var sess replaySessFile
	if err := json.Unmarshal(data, &sess); err != nil {
		fmt.Fprintf(os.Stderr, "[tuitest] parse %s: %v\n", path, err)
		return 1
	}
	if sess.Version != 1 && sess.Version != 2 {
		fmt.Fprintf(os.Stderr, "[tuitest] unsupported session version %d\n", sess.Version)
		return 1
	}

	return runReplaySession(&sess, path, *speed)
}

// replaySessFile is the on-disk representation read by the replay command.
type replaySessFile struct {
	Version    int          `json:"version"`
	Cols       int          `json:"cols"`
	Lines      int          `json:"lines"`
	Name       string       `json:"name,omitempty"`
	RecordedAt string       `json:"recorded_at,omitempty"`
	Command    []string     `json:"command,omitempty"`
	Steps      []recordStep `json:"steps"`
}

// runReplaySession drives the VT through recorded steps and renders the UI.
func runReplaySession(sess *replaySessFile, path string, speed float64) int {
	cols, _ := termSize()

	total := len(sess.Steps)
	if total == 0 {
		fmt.Fprintf(os.Stderr, "[tuitest] session %s has no steps\n", path)
		return 0
	}

	sessionName := sess.Name
	if sessionName == "" {
		sessionName = strings.TrimSuffix(filepath.Base(path), ".tuisess")
	}

	fmt.Fprintf(os.Stderr, "[tuitest] replaying %q (%d steps, %.1fx speed)\n",
		sessionName, total, speed)
	if len(sess.Command) > 0 {
		fmt.Fprintf(os.Stderr, "[tuitest] original command: %s\n", strings.Join(sess.Command, " "))
	}

	// Base interval between steps: 100ms at 1x speed (ticks are 1s).
	baseDelay := time.Duration(float64(100*time.Millisecond) / speed)

	// vt is our in-memory virtual terminal that accumulates the replayed view.
	vt := newVirtualTerminal(sess.Cols, sess.Lines)

	start := time.Now()
	for i, step := range sess.Steps {
		drawReplayBar(i+1, total, speed, cols)

		switch step.Kind {
		case "key":
			vt.applyKey(step.Key)
		case "type":
			for _, ch := range step.Text {
				vt.applyKey(string(ch))
			}
		case "resize":
			vt.resize(step.Cols, step.Lines)
		case "screen":
			// Screen assertion steps: in replay mode we just display progress.
		case "tick":
			// Tick step: advance simulated time.
		}

		time.Sleep(baseDelay)
	}

	drawReplayBar(total, total, speed, cols)
	fmt.Fprintln(os.Stderr) // newline after progress bar

	elapsed := time.Since(start).Round(time.Millisecond)
	fmt.Fprintf(os.Stderr, "[tuitest] replay complete in %s\n", elapsed)

	// Print the final virtual terminal state.
	fmt.Print(vt.render())

	return 0
}

// drawReplayBar renders a single-line progress bar to stderr using \r.
func drawReplayBar(step, total int, speed float64, cols int) {
	pct := 0
	if total > 0 {
		pct = (step * 100) / total
	}

	// Build a progress bar of width 16.
	const barWidth = 16
	filled := (step * barWidth) / max(total, 1)
	bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)

	line := fmt.Sprintf("[REPLAY] step %d/%d %s %3d%%  speed: %.1fx  (ctrl+c=abort)",
		step, total, bar, pct, speed)

	if cols > 0 && len(line) > cols {
		line = line[:cols]
	}
	pad := strings.Repeat(" ", max(0, cols-len(line)))
	fmt.Fprintf(os.Stderr, "\r%s%s", line, pad)
}

// virtualTerminal is a minimal in-memory VT that accumulates replayed content.
// It is intentionally simple: it just tracks the "current screen" as a line
// buffer so the replay command can render a final snapshot without requiring
// the full tuitest package (which is testing-only).
type virtualTerminal struct {
	cols  int
	lines int
	buf   []string // one entry per line
	cur   int      // current line index
}

func newVirtualTerminal(cols, lines int) *virtualTerminal {
	buf := make([]string, lines)
	return &virtualTerminal{cols: cols, lines: lines, buf: buf}
}

func (vt *virtualTerminal) applyKey(key string) {
	switch key {
	case "enter":
		vt.cur++
		if vt.cur >= vt.lines {
			// Scroll: drop the first line, append a blank.
			vt.buf = append(vt.buf[1:], "")
			vt.cur = vt.lines - 1
		}
	case "backspace":
		if vt.cur < len(vt.buf) && len(vt.buf[vt.cur]) > 0 {
			vt.buf[vt.cur] = vt.buf[vt.cur][:len(vt.buf[vt.cur])-1]
		}
	default:
		if len(key) == 1 && vt.cur < len(vt.buf) {
			line := vt.buf[vt.cur]
			if len(line) < vt.cols {
				vt.buf[vt.cur] = line + key
			}
		}
	}
}

func (vt *virtualTerminal) resize(cols, lines int) {
	vt.cols = cols
	// Grow or shrink the line buffer.
	for len(vt.buf) < lines {
		vt.buf = append(vt.buf, "")
	}
	if len(vt.buf) > lines {
		vt.buf = vt.buf[:lines]
	}
	vt.lines = lines
	if vt.cur >= vt.lines {
		vt.cur = vt.lines - 1
	}
}

func (vt *virtualTerminal) render() string {
	var sb strings.Builder
	for _, line := range vt.buf {
		sb.WriteString(line)
		sb.WriteByte('\n')
	}
	return sb.String()
}

// replaySpeedParsed parses a speed string like "2x", "0.5x", or "2" into a float.
// It is exported for use in tests.
func replaySpeedParsed(s string) (float64, error) {
	s = strings.TrimSuffix(s, "x")
	var v float64
	_, err := fmt.Sscanf(s, "%f", &v)
	if err != nil || v <= 0 {
		return 1.0, fmt.Errorf("invalid speed %q", s)
	}
	return v, nil
}
