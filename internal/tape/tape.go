// Package tape converts a btest.Session into a VHS .tape script.
//
// VHS (github.com/charmbracelet/vhs) is a tool that renders terminal
// sessions as GIFs from a declarative tape script. This package bridges
// the blit .tuisess format to the VHS tape format so recorded sessions
// can be turned into shareable GIFs.
package tape

import (
	"fmt"
	"strings"

	"github.com/blitui/blit/btest"
)

// defaultSleepMs is the inter-step pause inserted between input directives
// when no explicit timing information is available in the session.
const defaultSleepMs = 100

// Generate converts a btest.Session into a VHS tape script string.
//
// The resulting script contains:
//   - A preamble that sets the terminal size from the session metadata.
//   - A Type, Key, Sleep, or Hide directive for every input step.
//   - Screen steps are skipped (they are blit assertions, not VHS directives).
func Generate(sess *btest.Session) string {
	var sb strings.Builder

	// Preamble: terminal dimensions.
	fmt.Fprintf(&sb, "Set Width %d\n", sess.Cols*8)    // VHS uses pixels; approximate 8px per col
	fmt.Fprintf(&sb, "Set Height %d\n", sess.Lines*16) // approximate 16px per line
	sb.WriteString("Set FontSize 14\n")
	sb.WriteString("\n")

	for _, step := range sess.Steps {
		switch step.Kind {
		case "type":
			// Escape double-quotes inside the text for VHS.
			escaped := strings.ReplaceAll(step.Text, `"`, `\"`)
			fmt.Fprintf(&sb, "Type \"%s\"\n", escaped)
			fmt.Fprintf(&sb, "Sleep %dms\n", defaultSleepMs)

		case "key":
			directive := keyToVHS(step.Key)
			if directive != "" {
				fmt.Fprintf(&sb, "%s\n", directive)
				fmt.Fprintf(&sb, "Sleep %dms\n", defaultSleepMs)
			}

		case "resize":
			// VHS does not support mid-script resize; emit a comment.
			fmt.Fprintf(&sb, "# Resize %dx%d\n", step.Cols, step.Lines)

		case "tick":
			// A tick represents a timer event; map to a short sleep.
			fmt.Fprintf(&sb, "Sleep %dms\n", defaultSleepMs)

		case "screen":
			// Screen steps are blit assertion snapshots — not emitted.
		}
	}

	return sb.String()
}

// keyToVHS maps a blit key name to the equivalent VHS directive.
// Returns an empty string for keys that have no VHS representation.
func keyToVHS(key string) string {
	switch key {
	case "enter":
		return "Enter"
	case "space":
		return "Space"
	case "tab":
		return "Tab"
	case "backspace":
		return "Backspace"
	case "esc":
		return "Escape"
	case "up":
		return "Up"
	case "down":
		return "Down"
	case "left":
		return "Left"
	case "right":
		return "Right"
	case "home":
		return "Home"
	case "end":
		return "End"
	case "pgup":
		return "PageUp"
	case "pgdown":
		return "PageDown"
	case "delete":
		return "Delete"
	case "insert":
		return "Insert"
	case "f1":
		return "F1"
	case "f2":
		return "F2"
	case "f3":
		return "F3"
	case "f4":
		return "F4"
	case "f5":
		return "F5"
	case "f6":
		return "F6"
	case "f7":
		return "F7"
	case "f8":
		return "F8"
	case "f9":
		return "F9"
	case "f10":
		return "F10"
	case "f11":
		return "F11"
	case "f12":
		return "F12"
	}

	// ctrl+<letter> → Ctrl+<Letter>
	if strings.HasPrefix(key, "ctrl+") {
		letter := strings.TrimPrefix(key, "ctrl+")
		if len(letter) == 1 {
			return "Ctrl+" + strings.ToUpper(letter)
		}
	}

	// Single printable character → Type directive handled upstream,
	// but if a single char arrives as a "key" step, emit a Type.
	if len(key) == 1 {
		escaped := strings.ReplaceAll(key, `"`, `\"`)
		return fmt.Sprintf(`Type "%s"`, escaped)
	}

	// Unknown / hex sequences — skip silently.
	return ""
}
