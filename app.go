package tuikit

import (
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Option configures an App.
type Option func(*appModel)

// WithTheme sets the theme for all components.
func WithTheme(t Theme) Option {
	return func(a *appModel) { a.theme = t }
}

// WithComponent registers a named component with the App.
func WithComponent(name string, c Component) Option {
	return func(a *appModel) {
		a.components = append(a.components, namedComponent{name: name, component: c})
	}
}

// WithLayout sets a dual-pane layout.
func WithLayout(l *DualPane) Option {
	return func(a *appModel) { a.dualPane = l }
}

// WithStatusBar adds a status bar at the bottom of the screen.
func WithStatusBar(left, right func() string) Option {
	return func(a *appModel) {
		a.statusBar = NewStatusBar(StatusBarOpts{Left: left, Right: right})
	}
}

// WithHelp enables the auto-generated help overlay (toggle with '?').
func WithHelp() Option {
	return func(a *appModel) { a.helpEnabled = true }
}

// WithOverlay registers a named overlay with an explicit trigger key.
// Press triggerKey to open the overlay; Esc closes it.
func WithOverlay(name string, triggerKey string, o Overlay) Option {
	return func(a *appModel) {
		a.namedOverlays = append(a.namedOverlays, namedOverlay{
			name:       name,
			triggerKey: triggerKey,
			overlay:    o,
		})
	}
}

// WithKeyBind adds a global keybinding with an optional handler.
// If Handler is set, it's called when the key is pressed. The binding
// also appears in the auto-generated help screen.
func WithKeyBind(kb KeyBind) Option {
	return func(a *appModel) { a.globalBindings = append(a.globalBindings, kb) }
}

// WithMouseSupport enables mouse event tracking.
func WithMouseSupport() Option {
	return func(a *appModel) { a.mouseSupport = true }
}

// WithTickInterval enables periodic tick messages at the given interval.
// Components receive TickMsg in their Update method.
func WithTickInterval(d time.Duration) Option {
	return func(a *appModel) { a.tickInterval = d }
}

// TickMsg is sent to all components at the configured tick interval.
// Use it for animations, countdowns, polling, and flash effects.
type TickMsg struct {
	Time time.Time
}

type namedComponent struct {
	name      string
	component Component
}

type namedOverlay struct {
	name       string
	triggerKey string
	overlay    Overlay
}

// appModel is the top-level Bubble Tea model that coordinates everything.
type appModel struct {
	theme          Theme
	components     []namedComponent
	dualPane       *DualPane
	statusBar      *StatusBar
	helpEnabled    bool
	help           *Help
	namedOverlays  []namedOverlay
	overlays       *overlayStack
	registry       *registry
	globalBindings []KeyBind
	mouseSupport   bool
	tickInterval   time.Duration
	focusIdx       int
	width          int
	height         int
	notifyMsg      string
	notifyExpiry   time.Time
}

// newAppModel creates an appModel for testing (does not start tea.Program).
func newAppModel(opts ...Option) *appModel {
	a := &appModel{
		theme:    DefaultTheme(),
		overlays: newOverlayStack(),
		registry: newRegistry(),
	}
	for _, opt := range opts {
		opt(a)
	}
	a.setup()
	return a
}

// applyTheme sets the theme on any value that implements the Themed interface.
func applyTheme(v interface{}, t Theme) {
	if themed, ok := v.(Themed); ok {
		themed.SetTheme(t)
	}
}

func (a *appModel) setup() {
	// Set theme on all components
	for _, nc := range a.components {
		applyTheme(nc.component, a.theme)
	}

	// Set theme on dual pane components
	if a.dualPane != nil {
		applyTheme(a.dualPane.Main, a.theme)
		if a.dualPane.Side != nil {
			applyTheme(a.dualPane.Side, a.theme)
		}
	}

	// Status bar
	if a.statusBar != nil {
		a.statusBar.SetTheme(a.theme)
	}

	// Help overlay
	if a.helpEnabled {
		a.help = NewHelp()
		a.help.SetTheme(a.theme)
	}

	// Set theme on named overlays
	for _, no := range a.namedOverlays {
		applyTheme(no.overlay, a.theme)
	}

	// Build keybinding registry
	for _, nc := range a.components {
		a.registry.addBindings(nc.name, nc.component.KeyBindings())
	}
	if a.dualPane != nil {
		if a.dualPane.Main != nil {
			a.registry.addBindings("main", a.dualPane.Main.KeyBindings())
		}
		if a.dualPane.Side != nil {
			a.registry.addBindings("side", a.dualPane.Side.KeyBindings())
		}
	}

	// Built-in global bindings
	globals := []KeyBind{
		{Key: "q", Label: "Quit", Group: "OTHER"},
		{Key: "tab/←/→", Label: "Switch focus", Group: "OTHER"},
	}
	if a.helpEnabled {
		globals = append(globals, KeyBind{Key: "?", Label: "Help", Group: "OTHER"})
	}
	if a.dualPane != nil && a.dualPane.ToggleKey != "" {
		globals = append(globals, KeyBind{
			Key:   a.dualPane.ToggleKey,
			Label: "Toggle sidebar",
			Group: "OTHER",
		})
	}

	// Register overlay trigger keys
	for _, no := range a.namedOverlays {
		if no.triggerKey != "" {
			globals = append(globals, KeyBind{
				Key:   no.triggerKey,
				Label: no.name,
				Group: "OTHER",
			})
		}
	}

	// User-defined global bindings
	globals = append(globals, a.globalBindings...)
	a.registry.addBindings("global", globals)

	// Give help the registry
	if a.help != nil {
		a.help.setRegistry(a.registry)
	}

	// Set initial focus
	a.setFocus(0)
}

func (a *appModel) focusableComponents() []Component {
	if a.dualPane != nil {
		result := []Component{a.dualPane.Main}
		if a.dualPane.Side != nil {
			result = append(result, a.dualPane.Side)
		}
		return result
	}
	var result []Component
	for _, nc := range a.components {
		result = append(result, nc.component)
	}
	return result
}

func (a *appModel) setFocus(idx int) {
	focusable := a.focusableComponents()
	if len(focusable) == 0 {
		return
	}
	for i, c := range focusable {
		c.SetFocused(i == idx)
	}
	a.focusIdx = idx
}

func (a *appModel) cycleFocus() {
	focusable := a.focusableComponents()
	if len(focusable) <= 1 {
		return
	}
	next := (a.focusIdx + 1) % len(focusable)
	a.setFocus(next)
}

func (a *appModel) tickCmd() tea.Cmd {
	if a.tickInterval <= 0 {
		return nil
	}
	return tea.Tick(a.tickInterval, func(t time.Time) tea.Msg {
		return TickMsg{Time: t}
	})
}

func (a *appModel) Init() tea.Cmd {
	var cmds []tea.Cmd
	for _, nc := range a.components {
		if cmd := nc.component.Init(); cmd != nil {
			cmds = append(cmds, cmd)
		}
	}
	if a.dualPane != nil {
		if cmd := a.dualPane.Main.Init(); cmd != nil {
			cmds = append(cmds, cmd)
		}
		if a.dualPane.Side != nil {
			if cmd := a.dualPane.Side.Init(); cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
	}
	if cmd := a.tickCmd(); cmd != nil {
		cmds = append(cmds, cmd)
	}
	return tea.Batch(cmds...)
}

func (a *appModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.resize()
		return a, nil

	case tea.KeyMsg:
		return a.handleKey(msg)

	case tea.MouseMsg:
		return a.handleMouse(msg)

	case TickMsg:
		return a.handleTick(msg)

	case NotifyMsg:
		a.notifyMsg = msg.Text
		dur := msg.Duration
		if dur <= 0 {
			dur = 2 * time.Second
		}
		a.notifyExpiry = time.Now().Add(dur)
		a.resize()
		return a, nil
	}

	// Forward unknown messages to all components (for custom app messages)
	return a, a.broadcastMsg(msg)
}

// broadcastMsg sends a message to all registered components.
func (a *appModel) broadcastMsg(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd
	for _, nc := range a.components {
		if _, cmd := nc.component.Update(msg); cmd != nil {
			cmds = append(cmds, cmd)
		}
	}
	if a.dualPane != nil {
		if _, cmd := a.dualPane.Main.Update(msg); cmd != nil {
			cmds = append(cmds, cmd)
		}
		if a.dualPane.Side != nil {
			if _, cmd := a.dualPane.Side.Update(msg); cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
	}
	return tea.Batch(cmds...)
}

func (a *appModel) handleTick(msg TickMsg) (tea.Model, tea.Cmd) {
	if a.notifyMsg != "" && time.Now().After(a.notifyExpiry) {
		a.notifyMsg = ""
		a.resize()
	}

	var cmds []tea.Cmd
	cmds = append(cmds, a.broadcastMsg(msg))
	if cmd := a.tickCmd(); cmd != nil {
		cmds = append(cmds, cmd)
	}
	return a, tea.Batch(cmds...)
}

func (a *appModel) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	focusable := a.focusableComponents()
	if a.focusIdx < len(focusable) {
		_, cmd := focusable[a.focusIdx].Update(msg)
		return a, cmd
	}
	return a, nil
}

func (a *appModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	// 1. Active overlay gets first crack
	if overlay := a.overlays.active(); overlay != nil {
		if key == "esc" {
			a.overlays.pop()
			return a, nil
		}
		_, cmd := overlay.Update(msg)
		if isConsumed(cmd) {
			return a, nil
		}
		return a, cmd
	}

	// 2. Built-in global bindings
	switch key {
	case "q", "ctrl+c":
		return a, tea.Quit
	case "?":
		if a.help != nil {
			a.help.active = true
			a.help.SetSize(a.width, a.height)
			a.overlays.push(a.help)
			return a, nil
		}
	case "tab", "left", "right":
		if len(a.focusableComponents()) > 1 {
			a.cycleFocus()
			return a, nil
		}
	}

	// 3. Dual pane toggle
	if a.dualPane != nil && a.dualPane.ToggleKey != "" && key == a.dualPane.ToggleKey {
		a.dualPane.Toggle()
		a.resize()
		return a, nil
	}

	// 4. Named overlay triggers (each overlay has its own trigger key)
	for _, no := range a.namedOverlays {
		if no.triggerKey == key {
			a.openOverlay(no.overlay)
			return a, nil
		}
	}

	// 5. User-defined global keybindings with handlers
	for _, kb := range a.globalBindings {
		if kb.Key == key && kb.Handler != nil {
			kb.Handler()
			return a, nil
		}
	}

	// 6. Focused component
	focusable := a.focusableComponents()
	if a.focusIdx < len(focusable) {
		_, cmd := focusable[a.focusIdx].Update(msg)
		return a, cmd
	}

	return a, nil
}

// Activatable is implemented by overlays that can be explicitly activated.
type Activatable interface {
	SetActive(bool)
}

// openOverlay activates and pushes an overlay onto the stack.
func (a *appModel) openOverlay(o Overlay) {
	o.SetSize(a.width, a.height)
	// Activate the overlay
	if act, ok := o.(Activatable); ok {
		act.SetActive(true)
	} else {
		// Fall back to known built-in types
		switch ov := o.(type) {
		case *ConfigEditor:
			ov.active = true
		case *Help:
			ov.active = true
		}
	}
	a.overlays.push(o)
}

func (a *appModel) resize() {
	contentHeight := a.height
	if a.statusBar != nil {
		contentHeight--
		a.statusBar.SetSize(a.width, 1)
	}

	if a.notifyMsg != "" {
		contentHeight--
	}

	compHeight := contentHeight
	if a.showBadges() {
		compHeight--
	}

	if a.dualPane != nil {
		main, side, sideVisible := a.dualPane.compute(a.width, compHeight)
		a.dualPane.Main.SetSize(main.width, main.height)
		if a.dualPane.Side != nil && sideVisible {
			a.dualPane.Side.SetSize(side.width, side.height)
		}
	} else {
		for _, nc := range a.components {
			nc.component.SetSize(a.width, compHeight)
		}
	}

	if a.help != nil {
		a.help.SetSize(a.width, a.height)
	}
	for _, no := range a.namedOverlays {
		no.overlay.SetSize(a.width, a.height)
	}
}

func (a *appModel) showBadges() bool {
	return len(a.focusableComponents()) > 1
}

func (a *appModel) renderBadge(name string, focused bool) string {
	if focused {
		return lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(a.theme.TextInverse)).
			Background(lipgloss.Color(a.theme.Accent)).
			Padding(0, 1).
			Render(name)
	}
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(a.theme.Muted)).
		Padding(0, 1).
		Render(name)
}

func (a *appModel) View() string {
	// Active overlay takes over the screen
	if overlay := a.overlays.active(); overlay != nil {
		return overlay.View()
	}

	badges := a.showBadges()

	var content string
	if a.dualPane != nil {
		contentHeight := a.height
		if a.statusBar != nil {
			contentHeight--
		}
		compHeight := contentHeight
		if badges {
			compHeight--
		}
		_, _, sideVisible := a.dualPane.compute(a.width, compHeight)

		mainName := a.dualPane.MainName
		if mainName == "" {
			mainName = "Main"
		}
		mainView := a.dualPane.Main.View()
		if badges {
			badge := a.renderBadge(mainName, a.focusIdx == 0)
			mainView = lipgloss.JoinVertical(lipgloss.Left, badge, mainView)
		}

		if sideVisible && a.dualPane.Side != nil {
			sideName := a.dualPane.SideName
			if sideName == "" {
				sideName = "Side"
			}
			sideView := a.dualPane.Side.View()
			if badges {
				badge := a.renderBadge(sideName, a.focusIdx == 1)
				sideView = lipgloss.JoinVertical(lipgloss.Left, badge, sideView)
			}
			content = a.joinPanes(mainView, sideView, contentHeight)
		} else {
			content = mainView
		}
	} else {
		var views []string
		for i, nc := range a.components {
			v := nc.component.View()
			if badges {
				badge := a.renderBadge(nc.name, i == a.focusIdx)
				v = lipgloss.JoinVertical(lipgloss.Left, badge, v)
			}
			views = append(views, v)
		}
		content = lipgloss.JoinVertical(lipgloss.Left, views...)
	}

	var bottom []string
	if a.notifyMsg != "" {
		notiStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(a.theme.TextInverse)).
			Background(lipgloss.Color(a.theme.Flash)).
			Width(a.width).
			Padding(0, 1)
		bottom = append(bottom, notiStyle.Render(a.notifyMsg))
	}
	if a.statusBar != nil {
		bottom = append(bottom, a.statusBar.View())
	}
	if len(bottom) > 0 {
		return lipgloss.JoinVertical(lipgloss.Left, content, strings.Join(bottom, "\n"))
	}
	return content
}

// joinPanes renders main + separator + side with a full-height border.
func (a *appModel) joinPanes(mainView, sideView string, height int) string {
	sepStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(a.theme.Border))

	// Build a full-height separator
	sep := sepStyle.Render(strings.Repeat("│\n", height-1) + "│")

	if a.dualPane.SideRight {
		return lipgloss.JoinHorizontal(lipgloss.Top, mainView, sep, sideView)
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, sideView, sep, mainView)
}

// App is the main entry point for a tuikit application.
type App struct {
	model   *appModel
	opts    []tea.ProgramOption
	program *tea.Program
}

// NewApp creates a new App with the given options.
func NewApp(opts ...Option) *App {
	model := newAppModel(opts...)
	progOpts := []tea.ProgramOption{tea.WithAltScreen()}
	if model.mouseSupport {
		progOpts = append(progOpts, tea.WithMouseCellMotion())
	}
	return &App{model: model, opts: progOpts}
}

// Run starts the TUI application. It blocks until the user quits.
func (a *App) Run() error {
	a.program = tea.NewProgram(a.model, a.opts...)
	_, err := a.program.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	}
	return err
}

// AddKeyBind adds a global keybinding after construction.
// The binding appears in the help screen and its Handler is called on key press.
func (a *App) AddKeyBind(kb KeyBind) {
	a.model.globalBindings = append(a.model.globalBindings, kb)
	a.model.registry.addBindings("global", []KeyBind{kb})
}

// Notify displays a timed notification. If duration is 0, defaults to 2 seconds.
func (a *App) Notify(msg string, duration time.Duration) {
	if a.program != nil {
		a.program.Send(NotifyMsg{Text: msg, Duration: duration})
	}
}

// Send sends a message to the running App from outside the Bubble Tea event loop.
// Use this to push data from background goroutines (WebSocket streams, API polling, etc.).
// The message will be delivered to all components via Update.
func (a *App) Send(msg tea.Msg) {
	if a.program != nil {
		a.program.Send(msg)
	}
}
