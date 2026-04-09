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
//
// Slot mapping: the first call populates SlotMain; subsequent calls stack
// additional components that share focus cycling with Main.
func WithComponent(name string, c Component) Option {
	return func(a *appModel) {
		a.components = append(a.components, namedComponent{name: name, component: c})
		if a.slots != nil && !a.slots.has(SlotMain) {
			a.slots.set(SlotMain, c)
		}
	}
}

// WithLayout sets a dual-pane layout.
//
// Slot mapping: DualPane.Main -> SlotMain, DualPane.Side -> SlotSidebar.
func WithLayout(l *DualPane) Option {
	return func(a *appModel) {
		a.dualPane = l
		if a.slots != nil && l != nil {
			if l.Main != nil {
				a.slots.set(SlotMain, l.Main)
			}
			if l.Side != nil {
				a.slots.set(SlotSidebar, l.Side)
			}
		}
	}
}

// WithStatusBar adds a status bar at the bottom of the screen.
//
// left and right are legacy `func() string` closures. For reactive status
// bar content driven by signals, use WithStatusBarSignal.
//
// Slot mapping: the resulting StatusBar is bound to SlotFooter.
func WithStatusBar(left, right func() string) Option {
	return func(a *appModel) {
		a.statusBar = NewStatusBar(StatusBarOpts{Left: left, Right: right})
		if a.slots != nil {
			a.slots.set(SlotFooter, a.statusBar)
		}
	}
}

// WithStatusBarSignal adds a status bar whose left and right content are
// driven by *Signal[string]. Either may be nil. Prefer this over
// WithStatusBar when the values change asynchronously (background polling,
// websocket streams, etc.) — updates collapse into one notification per
// frame thanks to Signal's dirty-bit coalescing.
func WithStatusBarSignal(left, right *Signal[string]) Option {
	return func(a *appModel) {
		a.statusBar = NewStatusBar(StatusBarOpts{Left: left, Right: right})
		if a.slots != nil {
			a.slots.set(SlotFooter, a.statusBar)
		}
		if left != nil {
			a.trackSignal(left)
		}
		if right != nil {
			a.trackSignal(right)
		}
	}
}

// WithHelp enables the auto-generated help overlay (toggle with '?').
func WithHelp() Option {
	return func(a *appModel) { a.helpEnabled = true }
}

// WithOverlay registers a named overlay with an explicit trigger key.
// Press triggerKey to open the overlay; Esc closes it.
//
// Slot mapping: the overlay is pushed onto SlotOverlay.
func WithOverlay(name string, triggerKey string, o Overlay) Option {
	return func(a *appModel) {
		a.namedOverlays = append(a.namedOverlays, namedOverlay{
			name:       name,
			triggerKey: triggerKey,
			overlay:    o,
		})
		if a.slots != nil {
			a.slots.push(slotEntry{
				name:        SlotOverlay,
				component:   o,
				overlayKey:  triggerKey,
				overlayName: name,
			})
		}
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

// WithFocusCycleKey changes the key used to cycle focus between panes.
// Default is "tab". Set to "" to disable focus cycling entirely.
func WithFocusCycleKey(key string) Option {
	return func(a *appModel) { a.focusCycleKey = key }
}

// WithAnimations enables or disables the internal animation tick bus (~60fps).
// Setting TUIKIT_NO_ANIM=1 in the environment overrides this to false.
func WithAnimations(enabled bool) Option {
	return func(a *appModel) {
		if animDisabled {
			enabled = false
		}
		a.animationsEnabled = enabled
	}
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
	focusCycleKey  string
	focusIdx       int
	width          int
	height         int
	notifyMsg      string
	notifyExpiry   time.Time
	updateConfig  *UpdateConfig
	pendingNotify     string
	toasts            *toastManager
	animationsEnabled bool
	signalBus         *signalBus
	signals           []AnySignal
	hotReloadPath     string
	hotReload         *ThemeHotReload
	slots             *slotRegistry
}

// trackSignal registers a signal with the app's bus so its Set calls fire
// subscribers on the UI goroutine. Idempotent — attaching the same signal
// twice is harmless.
func (a *appModel) trackSignal(s AnySignal) {
	if s == nil {
		return
	}
	s.attach(a.signalBus)
	a.signals = append(a.signals, s)
}

// newAppModel creates an appModel for testing (does not start tea.Program).
func newAppModel(opts ...Option) *appModel {
	a := &appModel{
		theme:         DefaultTheme(),
		overlays:      newOverlayStack(),
		registry:      newRegistry(),
		focusCycleKey: "tab",
		toasts:        newToastManager(ToastManagerOpts{}),
		signalBus:     newSignalBus(),
		slots:         newSlotRegistry(),
	}
	for _, opt := range opts {
		opt(a)
	}
	a.materialiseSlots()
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
			_, _, vis := a.dualPane.compute(a.width, 0)
			if vis {
				result = append(result, a.dualPane.Side)
			}
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

func (a *appModel) animTickCmd() tea.Cmd {
	if animDisabled || !a.animationsEnabled {
		return nil
	}
	const animFrameInterval = 16 * time.Millisecond
	return tea.Tick(animFrameInterval, func(t time.Time) tea.Msg {
		return animTickMsg{time: t}
	})
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
	if cmd := a.animTickCmd(); cmd != nil {
		cmds = append(cmds, cmd)
	}
	if a.pendingNotify != "" {
		cmds = append(cmds, NotifyCmd(a.pendingNotify, 5*time.Second))
		a.pendingNotify = ""
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
		if a.toasts != nil {
			a.toasts.theme = a.theme
			a.toasts.add(ToastMsg{Severity: SeverityInfo, Title: msg.Text, Duration: dur})
		}
		a.resize()
		return a, nil

	case ToastMsg:
		if a.toasts != nil {
			a.toasts.theme = a.theme
			a.toasts.add(msg)
		}
		return a, nil

	case dismissTopToastMsg:
		if a.toasts != nil {
			a.toasts.dismissTop()
		}
		return a, nil

	case dismissToastMsg:
		if a.toasts != nil {
			a.toasts.dismissAt(msg.index)
		}
		return a, nil
	case SetThemeMsg:
		a.theme = msg.Theme
		a.setup()
		a.resize()
		return a, nil

	case ThemeHotReloadMsg:
		a.theme = msg.Theme
		a.setup()
		a.resize()
		return a, nil

	case ThemeHotReloadErrMsg:
		if a.toasts != nil {
			a.toasts.theme = a.theme
			a.toasts.add(ToastMsg{
				Severity: SeverityError,
				Title:    "Theme reload failed",
				Body:     msg.Err.Error(),
				Duration: 5 * time.Second,
			})
		}
		return a, nil

	case signalFlushMsg:
		// Drain pending signal subscribers on the UI goroutine. A5: this
		// is how signal-driven re-renders surface — Bubble Tea repaints
		// after any Update, so View() calls that read Signal.Get() pick
		// up the new value on the next frame without any extra plumbing.
		if a.signalBus != nil {
			a.signalBus.drain()
		}
		return a, nil
	}

	// Forward unknown messages to all components (for custom app messages)
	cmd := a.broadcastMsg(msg)
	a.checkOverlayActivation()
	return a, cmd
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
	if a.toasts != nil {
		a.toasts.tick(msg.Time)
	}

	var cmds []tea.Cmd
	cmds = append(cmds, a.broadcastMsg(msg))
	if cmd := a.tickCmd(); cmd != nil {
		cmds = append(cmds, cmd)
	}

	a.checkOverlayActivation()

	return a, tea.Batch(cmds...)
}

func (a *appModel) handleAnimTick(msg animTickMsg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	cmds = append(cmds, a.broadcastMsg(msg))
	if cmd := a.animTickCmd(); cmd != nil {
		cmds = append(cmds, cmd)
	}
	return a, tea.Batch(cmds...)
}

func (a *appModel) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	// Route mouse events to the correct pane based on X position
	if a.dualPane != nil && a.dualPane.Side != nil {
		main, _, sideVisible := a.dualPane.compute(a.width, 0)
		if sideVisible {
			var mainStart, mainEnd int
			if a.dualPane.SideRight {
				mainStart = 0
				mainEnd = main.width
			} else {
				sideW := a.width - main.width - 3 // 3 for separator
				mainStart = sideW + 3
				mainEnd = a.width
			}

			target := 0 // main
			if msg.X < mainStart || msg.X >= mainEnd {
				target = 1 // side
			}

			// Switch focus to clicked pane
			if target != a.focusIdx {
				a.setFocus(target)
			}

			// Adjust X coordinate relative to the target pane
			if target == 0 {
				msg.X -= mainStart
			} else {
				if a.dualPane.SideRight {
					msg.X -= mainEnd + 3 // skip past main + separator
				}
			}
		}
	}

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
			a.resize()
			return a, nil
		}
		_, cmd := overlay.Update(msg)
		// If the overlay closed itself (e.g., CommandBar after executing), pop it
		if !overlay.IsActive() {
			a.overlays.stack = a.overlays.stack[:len(a.overlays.stack)-1]
			a.resize()
		}
		if isConsumed(cmd) {
			return a, nil
		}
		return a, cmd
	}

	// 2. Check if focused component captures input (e.g., search mode)
	focusable := a.focusableComponents()
	inputCaptured := false
	if a.focusIdx < len(focusable) {
		if ic, ok := focusable[a.focusIdx].(InputCapture); ok {
			inputCaptured = ic.CapturesInput()
		}
	}

	// When input is captured, only ctrl+c passes through to globals
	if inputCaptured {
		if key == "ctrl+c" {
			return a, tea.Quit
		}
		if a.focusIdx < len(focusable) {
			_, cmd := focusable[a.focusIdx].Update(msg)
			return a, cmd
		}
	}

	// 3. Built-in global bindings
	switch key {
	case "q", "ctrl+c":
		return a, tea.Quit
	case "esc":
		if a.toasts != nil && a.toasts.hasActive() {
			a.toasts.dismissTop()
			return a, nil
		}
	case "?":
		if a.help != nil {
			a.help.active = true
			a.help.SetSize(a.width, a.height)
			a.overlays.push(a.help)
			return a, nil
		}
	case "left", "right":
		if len(focusable) > 1 {
			a.cycleFocus()
			return a, nil
		}
	default:
		if a.focusCycleKey != "" && key == a.focusCycleKey && len(focusable) > 1 {
			a.cycleFocus()
			return a, nil
		}
	}

	// 4. Dual pane toggle
	if a.dualPane != nil && a.dualPane.ToggleKey != "" && key == a.dualPane.ToggleKey {
		a.dualPane.Toggle()
		a.resize()
		return a, nil
	}

	// 5. Named overlay triggers (each overlay has its own trigger key)
	for _, no := range a.namedOverlays {
		if no.triggerKey == key {
			a.openOverlay(no.overlay)
			return a, nil
		}
	}

	// 6. User-defined global keybindings with handlers
	for _, kb := range a.globalBindings {
		if kb.Key == key {
			if kb.HandlerCmd != nil {
				cmd := kb.HandlerCmd()
				return a, cmd
			}
			if kb.Handler != nil {
				kb.Handler()
				return a, nil
			}
		}
	}

	// 7. Focused component
	if a.focusIdx < len(focusable) {
		_, cmd := focusable[a.focusIdx].Update(msg)
		return a, cmd
	}

	return a, nil
}

// InputCapture is implemented by components that can enter text-input mode
// (e.g., search, filter). When CapturesInput returns true, the App skips
// global keybindings and sends keys directly to the component.
type InputCapture interface {
	CapturesInput() bool
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

// checkOverlayActivation pushes any named overlay that became active
// outside of the normal trigger-key flow (e.g., via component calling Show()).
func (a *appModel) checkOverlayActivation() {
	for _, no := range a.namedOverlays {
		if no.overlay.IsActive() && a.overlays.active() != no.overlay {
			// Check it's not already somewhere in the stack
			found := false
			for _, stacked := range a.overlays.stack {
				if stacked == no.overlay {
					found = true
					break
				}
			}
			if !found {
				no.overlay.SetSize(a.width, a.height)
				a.overlays.push(no.overlay)
			}
		}
	}
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

	// Inline overlay takes a line at the bottom
	if overlay := a.overlays.active(); overlay != nil {
		if _, ok := overlay.(InlineOverlay); ok {
			contentHeight--
		}
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
	// Active overlay takes over the screen (unless inline)
	if overlay := a.overlays.active(); overlay != nil {
		if _, ok := overlay.(InlineOverlay); !ok {
			return overlay.View()
		}
	}

	badges := a.showBadges()

	var content string
	if a.dualPane != nil {
		contentHeight := a.height
		if a.statusBar != nil {
			contentHeight--
		}
		if a.notifyMsg != "" {
			contentHeight--
		}
		if overlay := a.overlays.active(); overlay != nil {
			if _, ok := overlay.(InlineOverlay); ok {
				contentHeight--
			}
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
	if overlay := a.overlays.active(); overlay != nil {
		if _, ok := overlay.(InlineOverlay); ok {
			bottom = append(bottom, overlay.View())
		}
	}
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

	// Build a full-height separator with spacing on both sides
	sepLine := " │ "
	sep := sepStyle.Render(strings.Repeat(sepLine+"\n", height-1) + sepLine)

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
	// Auto-update check (before TUI starts)
	if a.model.updateConfig != nil {
		a.runUpdateCheck()
	}

	a.program = tea.NewProgram(a.model, a.opts...)
	if a.model.signalBus != nil {
		prog := a.program
		a.model.signalBus.setSender(func(msg tea.Msg) { prog.Send(msg) })
	}
	if a.model.hotReloadPath != "" {
		prog := a.program
		hr, hrErr := NewThemeHotReload(a.model.hotReloadPath, func(msg interface{}) {
			prog.Send(msg)
		})
		if hrErr != nil {
			fmt.Fprintf(os.Stderr, "theme hot-reload: %v\n", hrErr)
		} else {
			a.model.hotReload = hr
			defer hr.Stop()
		}
	}
	_, err := a.program.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	}
	return err
}

// runUpdateCheck performs the update check based on configured mode.
func (a *App) runUpdateCheck() {
	// Clean up leftover .old binary from a previous self-update
	CleanupOldBinary()

	cfg := a.model.updateConfig
	result, err := CheckForUpdate(*cfg)
	if err != nil || !result.Available {
		return
	}

	// Detect install method
	exePath, exeErr := os.Executable()
	if exeErr == nil {
		result.InstallMethod = DetectInstallMethod(exePath)
	}

	upgradeHint := a.upgradeHint(result)

	// Required releases (min-version marker) always take the forced gate,
	// regardless of the configured mode.
	mode := cfg.Mode
	if result.Required {
		mode = UpdateForced
	}

	switch mode {
	case UpdateBlocking:
		fmt.Printf("Update available: %s → %s\n", result.CurrentVersion, result.LatestVersion)
		if result.InstallMethod != InstallManual {
			// Package manager users get the upgrade command, no self-replace
			fmt.Println(upgradeHint)
			return
		}
		fmt.Print("Update now? [y/n]: ")
		var answer string
		fmt.Scanln(&answer)
		if strings.ToLower(strings.TrimSpace(answer)) == "y" {
			if err := SelfUpdate(*cfg); err != nil {
				fmt.Fprintf(os.Stderr, "Update failed: %v\n", err)
				return
			}
			fmt.Printf("Updated to %s. Please restart.\n", result.LatestVersion)
			os.Exit(0)
		}

	case UpdateNotify:
		// Queue a notification for after the TUI starts
		a.model.pendingNotify = fmt.Sprintf("Update available: %s → %s  %s", result.CurrentVersion, result.LatestVersion, upgradeHint)

	case UpdateSilent:
		// Fire-and-forget background self-update. Users see no prompts;
		// the replacement takes effect on next launch. Errors are
		// routed through the OnUpdateError hook if configured.
		go func(c UpdateConfig) {
			if err := SelfUpdate(c); err != nil && c.OnUpdateError != nil {
				c.OnUpdateError(err)
			}
		}(*cfg)

	case UpdateForced:
		// Full-screen blocking gate. If the user accepts, we run
		// SelfUpdate inline and exit; otherwise the app exits without
		// launching its main UI.
		gate := NewForcedUpdateScreen(result, *cfg)
		p := tea.NewProgram(gate, tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Update gate failed: %v\n", err)
			os.Exit(1)
		}
		switch gate.Choice {
		case ForcedChoiceUpdate:
			if result.InstallMethod != InstallManual {
				fmt.Println(upgradeHint)
				os.Exit(0)
			}
			if err := SelfUpdate(*cfg); err != nil {
				fmt.Fprintf(os.Stderr, "Update failed: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Updated to %s. Please restart.\n", result.LatestVersion)
			os.Exit(0)
		default:
			os.Exit(0)
		}

	case UpdateDryRun:
		if err := SelfUpdate(*cfg); err != nil {
			fmt.Fprintf(os.Stderr, "Dry-run update error: %v\n", err)
		}
	}
}

func (a *App) upgradeHint(result *UpdateResult) string {
	switch result.InstallMethod {
	case InstallHomebrew:
		return fmt.Sprintf("Run: brew upgrade %s", a.model.updateConfig.BinaryName)
	case InstallScoop:
		return fmt.Sprintf("Run: scoop update %s", a.model.updateConfig.BinaryName)
	default:
		return result.ReleaseURL
	}
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
// Model returns the underlying tea.Model for use with tuitest or other
// testing frameworks that need a tea.Model directly.
func (a *App) Model() tea.Model {
	return a.model
}

// Use this to push data from background goroutines (WebSocket streams, API polling, etc.).
// The message will be delivered to all components via Update.
func (a *App) Send(msg tea.Msg) {
	if a.program != nil {
		a.program.Send(msg)
	}
}
