package blit

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// devConsoleToggleMsg is sent by ctrl+\ to toggle the dev console.
type devConsoleToggleMsg struct{}

// DevConsoleToggleCmd returns a tea.Cmd that toggles the dev console.
func DevConsoleToggleCmd() tea.Cmd {
	return func() tea.Msg { return devConsoleToggleMsg{} }
}

// devConsole is the overlay that renders the developer console.
// It is toggled by ctrl+\ or by setting BLIT_DEVCONSOLE=1.
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
// Zero cost when disabled: when BLIT_DEVCONSOLE=0 and the console has
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

	// providers holds registered DebugProviders for debug sections.
	providers []DebugProvider

	// activeTab is the index of the currently selected provider tab.
	activeTab int
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

// newDevConsole creates a devConsole with built-in providers.
// autoEnable checks BLIT_DEVCONSOLE env.
func newDevConsole() *devConsole {
	dc := &devConsole{}
	if os.Getenv("BLIT_DEVCONSOLE") == "1" {
		dc.active = true
	}
	dc.providers = []DebugProvider{
		&frameStatsProvider{dc: dc},
		&componentTreeProvider{dc: dc},
		&signalMonitorProvider{dc: dc},
		&keyLogProvider{dc: dc},
		&logViewerProvider{dc: dc},
		&themeInspectorProvider{dc: dc},
	}
	return dc
}

// Name returns the module name for the dev console.
func (dc *devConsole) Name() string { return "devConsole" }

// Providers returns the registered debug providers.
func (dc *devConsole) Providers() []DebugProvider {
	return dc.providers
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
//
//nolint:gocyclo // debug console dispatches per key and pane
func (dc *devConsole) Update(msg tea.Msg, ctx Context) (Component, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "ctrl+\\":
			dc.active = false
			return dc, nil
		// Tab navigation
		case "left":
			if len(dc.providers) > 0 {
				dc.activeTab = (dc.activeTab - 1 + len(dc.providers)) % len(dc.providers)
			}
		case "right":
			if len(dc.providers) > 0 {
				dc.activeTab = (dc.activeTab + 1) % len(dc.providers)
			}
		case "ctrl+e":
			_ = dc.exportJSON()
			return dc, nil
		case "1", "2", "3", "4", "5", "6", "7", "8", "9":
			idx := int(msg.String()[0]-'0') - 1
			if idx >= 0 && idx < len(dc.providers) {
				dc.activeTab = idx
			}
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

		// Use ANSI-aware truncation so escape sequences aren't mangled
		left := ansi.Truncate(bgLine, dc.x, "")
		// Pad to the target column if the visible width is short
		if w := ansi.StringWidth(left); w < dc.x {
			left += strings.Repeat(" ", dc.x-w)
		}
		bgLines[row] = left + pLine
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

	// Build tab bar
	tabBar := dc.renderTabBar(w-4, t)

	// Render active provider content
	innerW := w - 4
	if innerW < 1 {
		innerW = 1
	}
	innerH := h - 5 // border(2) + header(1) + tab bar(1) + padding
	if innerH < 1 {
		innerH = 1
	}

	var content string
	if dc.activeTab >= 0 && dc.activeTab < len(dc.providers) {
		content = dc.providers[dc.activeTab].View(innerW, innerH, t)
	}

	// Truncate content lines to fit
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

	content = truncate(content)

	// Clamp content to available height
	contentLines := strings.Split(content, "\n")
	if len(contentLines) > innerH {
		contentLines = contentLines[:innerH]
	}
	content = strings.Join(contentLines, "\n")

	body := strings.Join([]string{header, tabBar, content}, "\n")

	return border.Render(body)
}

// renderTabBar renders the tab bar showing all provider names.
func (dc *devConsole) renderTabBar(width int, t Theme) string {
	if len(dc.providers) == 0 {
		return ""
	}

	var tabs []string
	for i, p := range dc.providers {
		label := p.Name()
		if i == dc.activeTab {
			tab := lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#1a1a2e")).
				Background(lipgloss.Color(t.Accent)).
				Render(" " + label + " ")
			tabs = append(tabs, tab)
		} else {
			tab := lipgloss.NewStyle().
				Foreground(lipgloss.Color(t.Muted)).
				Render(" " + label + " ")
			tabs = append(tabs, tab)
		}
	}

	bar := strings.Join(tabs, "")
	// Truncate if wider than available space
	runes := []rune(bar)
	if len(runes) > width {
		runes = runes[:width]
	}
	return string(runes)
}

// KeyBindings implements Component.
func (dc *devConsole) KeyBindings() []KeyBind {
	return []KeyBind{
		{Key: "ctrl+e", Label: "Export JSON", Group: "DEV"},
		{Key: "esc", Label: "Close dev console", Group: "DEV"},
		{Key: "left/right", Label: "Switch tab", Group: "DEV"},
		{Key: "1-9", Label: "Jump to tab", Group: "DEV"},
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
// also auto-enabled when BLIT_DEVCONSOLE=1 is set in the environment.
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

// initProviders calls Init on all DebugProviderWithLifecycle providers.
func (dc *devConsole) initProviders() {
	for _, p := range dc.providers {
		if lp, ok := p.(DebugProviderWithLifecycle); ok {
			_ = lp.Init()
		}
	}
}

// destroyProviders calls Destroy on all DebugProviderWithLifecycle providers.
func (dc *devConsole) destroyProviders() {
	for _, p := range dc.providers {
		if lp, ok := p.(DebugProviderWithLifecycle); ok {
			_ = lp.Destroy()
		}
	}
}

// exportJSON writes all DebugDataProvider Data() to a JSON file in the
// current directory. Returns the path written or an error.
func (dc *devConsole) exportJSON() error {
	export := make(map[string]any, len(dc.providers))
	for _, p := range dc.providers {
		if dp, ok := p.(DebugDataProvider); ok {
			export[p.Name()] = dp.Data()
		}
	}

	data, err := json.MarshalIndent(export, "", "  ")
	if err != nil {
		return fmt.Errorf("devconsole export: %w", err)
	}

	name := fmt.Sprintf("blit-devconsole-%s.json", time.Now().Format("20060102-150405"))
	path := filepath.Join(os.TempDir(), name)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("devconsole export write: %w", err)
	}
	return nil
}

// --- Built-in DebugProvider implementations ---

// frameStatsProvider renders FPS and frame time.
type frameStatsProvider struct{ dc *devConsole }

func (p *frameStatsProvider) Name() string { return "Frame Stats" }
func (p *frameStatsProvider) View(w, h int, theme Theme) string {
	fps := p.dc.fps()
	ft := p.dc.frameTimeMs()
	return fmt.Sprintf("FPS: %.1f  frame: %.2fms", fps, ft)
}
func (p *frameStatsProvider) Data() map[string]any {
	return map[string]any{
		"fps":         p.dc.fps(),
		"frameTimeMs": p.dc.frameTimeMs(),
		"frameCount":  p.dc.frameCount,
	}
}

// componentTreeProvider renders the component tree with focus markers.
type componentTreeProvider struct{ dc *devConsole }

func (p *componentTreeProvider) Name() string { return "Components" }
func (p *componentTreeProvider) View(w, h int, theme Theme) string {
	snap := p.dc.snapshot
	var lines []string
	for i, name := range snap.componentNames {
		focused := i < len(snap.componentFocus) && snap.componentFocus[i]
		marker := "  "
		if focused {
			marker = "* "
		}
		lines = append(lines, marker+name)
	}
	if len(lines) == 0 {
		return "Components: (none)"
	}
	return "Components:\n" + strings.Join(lines, "\n")
}

// signalMonitorProvider renders signal values.
type signalMonitorProvider struct{ dc *devConsole }

func (p *signalMonitorProvider) Name() string { return "Signals" }
func (p *signalMonitorProvider) View(w, h int, theme Theme) string {
	snap := p.dc.snapshot
	if len(snap.signals) == 0 {
		return "Signals: (none)"
	}
	var lines []string
	for _, s := range snap.signals {
		lines = append(lines, fmt.Sprintf("  %s = %s", s.label, s.value))
	}
	return "Signals:\n" + strings.Join(lines, "\n")
}

// keyLogProvider renders recent keypresses.
type keyLogProvider struct{ dc *devConsole }

func (p *keyLogProvider) Name() string { return "Keys" }
func (p *keyLogProvider) View(w, h int, theme Theme) string {
	keys := p.dc.recentKeys()
	if len(keys) == 0 {
		return "Keys: (none)"
	}
	return "Keys: " + strings.Join(keys, " ")
}

// logViewerProvider renders recent log lines.
type logViewerProvider struct{ dc *devConsole }

func (p *logViewerProvider) Name() string { return "Logs" }
func (p *logViewerProvider) View(w, h int, theme Theme) string {
	snap := p.dc.snapshot
	if len(snap.logLines) == 0 {
		return "Logs: (none)"
	}
	return "Logs:\n" + strings.Join(snap.logLines, "\n")
}

// themeInspectorProvider renders theme name and color swatches.
type themeInspectorProvider struct{ dc *devConsole }

func (p *themeInspectorProvider) Name() string { return "Theme" }
func (p *themeInspectorProvider) View(w, h int, theme Theme) string {
	t := theme
	snap := p.dc.snapshot
	accentSwatch := lipgloss.NewStyle().Background(lipgloss.Color(t.Accent)).Render("  ")
	textSwatch := lipgloss.NewStyle().Background(lipgloss.Color(t.Text)).Render("  ")
	mutedSwatch := lipgloss.NewStyle().Background(lipgloss.Color(t.Muted)).Render("  ")
	posSwatch := lipgloss.NewStyle().Background(lipgloss.Color(t.Positive)).Render("  ")
	negSwatch := lipgloss.NewStyle().Background(lipgloss.Color(t.Negative)).Render("  ")
	return fmt.Sprintf("Theme: %s  Accent%s Text%s Muted%s Pos%s Neg%s",
		snap.themeName, accentSwatch, textSwatch, mutedSwatch, posSwatch, negSwatch)
}
