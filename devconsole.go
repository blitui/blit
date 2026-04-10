package blit

import (
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// devConsoleToggleMsg is sent by ctrl+\ to toggle the dev console.
type devConsoleToggleMsg struct{}

// DevConsoleToggleCmd returns a tea.Cmd that toggles the dev console.
func DevConsoleToggleCmd() tea.Cmd {
	return func() tea.Msg { return devConsoleToggleMsg{} }
}

// devConsole is the overlay that renders the developer console.
// It is toggled by ctrl+\ or by setting TUIKIT_DEVCONSOLE=1.
//
// The console shows:
//   - FPS / frame time
//   - Component tree (focus state)
//   - Active signals and their current values
//   - Recent keypresses (ring buffer of last 20)
//   - Recent log messages from ctx.Logger
//   - Theme name + primary color swatches
//
// It renders as a full-screen overlay (implements Overlay) and is mounted
// via the slot system at SlotOverlay with the highest z-order.
//
// Zero cost when disabled: when TUIKIT_DEVCONSOLE=0 and the console has
// never been toggled, all frame-recording hooks are no-ops and the console
// never enters the overlay stack.
type devConsole struct {
	active  bool
	focused bool
	width   int
	height  int
	theme   Theme

	// position and size within the screen
	x, y int
	w, h int

	// FPS tracking — ring buffer of last 60 frame timestamps
	frameTimes [60]time.Time
	frameHead  int
	frameCount int

	// recent keypresses — ring buffer of last 20
	keyBuf  [20]string
	keyHead int
	keyFull bool

	// snapshot of app state for rendering (set each frame by the app)
	snapshot devConsoleSnapshot
}

// devConsoleSnapshot captures a point-in-time view of app state so the
// console can render without holding locks.
type devConsoleSnapshot struct {
	focusIdx       int
	focusName      string
	componentNames []string
	componentFocus []bool
	signals        []signalInfo
	logLines       []string
	themeName      string
	theme          Theme
}

// signalInfo is a rendered representation of one signal's current value.
type signalInfo struct {
	label string
	value string
}

// newDevConsole creates a devConsole. autoEnable checks TUIKIT_DEVCONSOLE env.
func newDevConsole() *devConsole {
	dc := &devConsole{}
	if os.Getenv("TUIKIT_DEVCONSOLE") == "1" {
		dc.active = true
	}
	return dc
}

// recordFrame pushes the current timestamp into the FPS ring buffer.
// This is called by the app on every View() invocation when the console exists.
func (dc *devConsole) recordFrame(t time.Time) {
	dc.frameTimes[dc.frameHead] = t
	dc.frameHead = (dc.frameHead + 1) % len(dc.frameTimes)
	if dc.frameCount < len(dc.frameTimes) {
		dc.frameCount++
	}
}

// recordKey pushes a keypress string into the ring buffer.
func (dc *devConsole) recordKey(key string) {
	dc.keyBuf[dc.keyHead] = key
	dc.keyHead = (dc.keyHead + 1) % len(dc.keyBuf)
	if !dc.keyFull && dc.keyHead == 0 {
		dc.keyFull = true
	}
}

// fps returns the approximate frames-per-second over the last N frames.
func (dc *devConsole) fps() float64 {
	if dc.frameCount < 2 {
		return 0
	}
	count := dc.frameCount
	if count > len(dc.frameTimes) {
		count = len(dc.frameTimes)
	}
	// newest is at frameHead-1, oldest is at frameHead (wrapping)
	newest := dc.frameTimes[(dc.frameHead-1+len(dc.frameTimes))%len(dc.frameTimes)]
	oldest := dc.frameTimes[(dc.frameHead-count+len(dc.frameTimes))%len(dc.frameTimes)]
	dur := newest.Sub(oldest)
	if dur <= 0 {
		return 0
	}
	return float64(count-1) / dur.Seconds()
}

// frameTimeMs returns the last frame duration in milliseconds.
func (dc *devConsole) frameTimeMs() float64 {
	if dc.frameCount < 2 {
		return 0
	}
	i1 := (dc.frameHead - 1 + len(dc.frameTimes)) % len(dc.frameTimes)
	i2 := (dc.frameHead - 2 + len(dc.frameTimes)) % len(dc.frameTimes)
	return float64(dc.frameTimes[i1].Sub(dc.frameTimes[i2]).Microseconds()) / 1000.0
}

// recentKeys returns up to 20 recent keypresses in chronological order.
func (dc *devConsole) recentKeys() []string {
	var keys []string
	total := len(dc.keyBuf)
	if !dc.keyFull {
		total = dc.keyHead
	}
	if total == 0 {
		return nil
	}
	start := (dc.keyHead - total + len(dc.keyBuf)) % len(dc.keyBuf)
	for i := 0; i < total; i++ {
		keys = append(keys, dc.keyBuf[(start+i)%len(dc.keyBuf)])
	}
	return keys
}

// --- Component interface ---

// Init implements Component.
func (dc *devConsole) Init() tea.Cmd { return nil }

// Update implements Component.
func (dc *devConsole) Update(msg tea.Msg, ctx Context) (Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "ctrl+\\":
			dc.active = false
			return dc, nil
		// Resize: alt+arrows
		case "alt+up":
			if dc.h > 5 {
				dc.h--
			}
		case "alt+down":
			if dc.y+dc.h < dc.height {
				dc.h++
			}
		case "alt+left":
			if dc.w > 20 {
				dc.w--
			}
		case "alt+right":
			if dc.x+dc.w < dc.width {
				dc.w++
			}
		// Move: shift+arrows
		case "shift+up":
			if dc.y > 0 {
				dc.y--
			}
		case "shift+down":
			if dc.y+dc.h < dc.height {
				dc.y++
			}
		case "shift+left":
			if dc.x > 0 {
				dc.x--
			}
		case "shift+right":
			if dc.x+dc.w < dc.width {
				dc.x++
			}
		}
	case tea.MouseMsg:
		// drag support: mouse button 1 drag moves the console
		if msg.Action == tea.MouseActionMotion && msg.Button == tea.MouseButtonLeft {
			// Move top-left corner to mouse position, clamped
			nx := msg.X
			ny := msg.Y
			if nx < 0 {
				nx = 0
			}
			if ny < 0 {
				ny = 0
			}
			if nx+dc.w > dc.width {
				nx = dc.width - dc.w
			}
			if ny+dc.h > dc.height {
				ny = dc.height - dc.h
			}
			dc.x = nx
			dc.y = ny
		}
	}
	return dc, nil
}

// View implements Component. For a FloatingOverlay the app calls FloatView
// instead; View is still required by the interface and returns the raw panel.
func (dc *devConsole) View() string {
	if !dc.active {
		return ""
	}
	return dc.renderPanel()
}

// FloatView implements FloatingOverlay. It overlays the dev console panel on
// top of the background content by replacing the appropriate lines.
func (dc *devConsole) FloatView(background string) string {
	if !dc.active {
		return background
	}
	panel := dc.renderPanel()
	if panel == "" {
		return background
	}

	bgLines := strings.Split(background, "\n")
	panelLines := strings.Split(panel, "\n")

	// Ensure background has enough lines
	for len(bgLines) < dc.y+len(panelLines) {
		bgLines = append(bgLines, "")
	}

	for i, pLine := range panelLines {
		row := dc.y + i
		if row >= len(bgLines) {
			break
		}
		bgLine := bgLines[row]
		bgRunes := []rune(bgLine)
		pRunes := []rune(pLine)

		// Expand bg line if shorter than dc.x
		for len(bgRunes) < dc.x {
			bgRunes = append(bgRunes, ' ')
		}

		// Replace characters at [dc.x : dc.x+len(pRunes)]
		end := dc.x + len(pRunes)
		if end > len(bgRunes) {
			// Extend with spaces then overwrite
			for len(bgRunes) < end {
				bgRunes = append(bgRunes, ' ')
			}
		}
		copy(bgRunes[dc.x:], pRunes)
		bgLines[row] = string(bgRunes)
	}

	return strings.Join(bgLines, "\n")
}

// renderPanel renders the console box without positioning (used by FloatView).
func (dc *devConsole) renderPanel() string {
	w := dc.w
	h := dc.h
	if w < 20 {
		w = 20
	}
	if h < 5 {
		h = 5
	}

	t := dc.theme
	border := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(t.Accent)).
		Foreground(lipgloss.Color(t.Text)).
		Background(lipgloss.Color("#1a1a2e")).
		Width(w - 2).
		Height(h - 2)

	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(t.Accent)).
		Render("blit dev console")

	fps := dc.fps()
	ft := dc.frameTimeMs()
	perfLine := fmt.Sprintf("FPS: %.1f  frame: %.2fms", fps, ft)

	var treeLines []string
	snap := dc.snapshot
	for i, name := range snap.componentNames {
		focused := i < len(snap.componentFocus) && snap.componentFocus[i]
		marker := "  "
		if focused {
			marker = "* "
		}
		treeLines = append(treeLines, marker+name)
	}
	treeSection := "Components:\n" + strings.Join(treeLines, "\n")

	var sigLines []string
	for _, s := range snap.signals {
		sigLines = append(sigLines, fmt.Sprintf("  %s = %s", s.label, s.value))
	}
	sigSection := "Signals:"
	if len(sigLines) > 0 {
		sigSection += "\n" + strings.Join(sigLines, "\n")
	} else {
		sigSection += " (none)"
	}

	keys := dc.recentKeys()
	keySection := "Keys: " + strings.Join(keys, " ")

	var logSection string
	if len(snap.logLines) > 0 {
		logSection = "Logs:\n" + strings.Join(snap.logLines, "\n")
	}

	accentSwatch := lipgloss.NewStyle().Background(lipgloss.Color(t.Accent)).Render("  ")
	textSwatch := lipgloss.NewStyle().Background(lipgloss.Color(t.Text)).Render("  ")
	mutedSwatch := lipgloss.NewStyle().Background(lipgloss.Color(t.Muted)).Render("  ")
	posSwatch := lipgloss.NewStyle().Background(lipgloss.Color(t.Positive)).Render("  ")
	negSwatch := lipgloss.NewStyle().Background(lipgloss.Color(t.Negative)).Render("  ")
	themeLine := fmt.Sprintf("Theme: %s  Accent%s Text%s Muted%s Pos%s Neg%s",
		snap.themeName, accentSwatch, textSwatch, mutedSwatch, posSwatch, negSwatch)

	innerW := w - 4
	if innerW < 1 {
		innerW = 1
	}
	truncate := func(s string) string {
		var out []string
		for _, line := range strings.Split(s, "\n") {
			runes := []rune(line)
			if len(runes) > innerW {
				runes = runes[:innerW]
			}
			out = append(out, string(runes))
		}
		return strings.Join(out, "\n")
	}

	sections := []string{
		header,
		truncate(perfLine),
		truncate(treeSection),
		truncate(sigSection),
		truncate(keySection),
	}
	if logSection != "" {
		sections = append(sections, truncate(logSection))
	}
	sections = append(sections, truncate(themeLine))

	innerH := h - 2
	if innerH < 1 {
		innerH = 1
	}
	body := strings.Join(sections, "\n")
	bodyLines := strings.Split(body, "\n")
	if len(bodyLines) > innerH {
		bodyLines = bodyLines[:innerH]
	}
	body = strings.Join(bodyLines, "\n")

	return border.Render(body)
}

// KeyBindings implements Component.
func (dc *devConsole) KeyBindings() []KeyBind {
	return []KeyBind{
		{Key: "esc", Label: "Close dev console", Group: "DEV"},
		{Key: "alt+arrows", Label: "Resize", Group: "DEV"},
		{Key: "shift+arrows", Label: "Move", Group: "DEV"},
	}
}

// SetSize implements Component.
func (dc *devConsole) SetSize(width, height int) {
	dc.width = width
	dc.height = height
	// Default console size: 60% width, 60% height, centered
	if dc.w == 0 {
		dc.w = width * 6 / 10
		if dc.w < 40 {
			dc.w = 40
		}
		dc.h = height * 6 / 10
		if dc.h < 12 {
			dc.h = 12
		}
		dc.x = (width - dc.w) / 2
		dc.y = (height - dc.h) / 2
	}
}

// Focused implements Component.
func (dc *devConsole) Focused() bool { return dc.focused }

// SetFocused implements Component.
func (dc *devConsole) SetFocused(f bool) { dc.focused = f }

// SetTheme implements Themed.
func (dc *devConsole) SetTheme(t Theme) { dc.theme = t }

// --- Overlay interface ---

// IsActive implements Overlay.
func (dc *devConsole) IsActive() bool { return dc.active }

// Close implements Overlay.
func (dc *devConsole) Close() { dc.active = false }

// SetActive implements Activatable.
func (dc *devConsole) SetActive(v bool) { dc.active = v }

// --- app integration helpers ---

// WithDevConsole enables the dev console overlay on an App. The console is
// also auto-enabled when TUIKIT_DEVCONSOLE=1 is set in the environment.
// When this option is not provided and the env var is not set, the console
// costs nothing at runtime.
func WithDevConsole() Option {
	return func(a *appModel) {
		if a.devConsole == nil {
			a.devConsole = newDevConsole()
		}
		a.devConsole.active = true
	}
}
