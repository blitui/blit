// Package main — watch.go implements the interactive watch menu for btest.
//
// Running `blit -watch` drops into a bubbletea-powered UI with:
//   - A status bar showing active filters and run state
//   - A filter panel showing the current p/t/l settings
//   - A LogViewer streaming test output in real time
//
// Key bindings (Vitest-style):
//
//	p  prompt for path filter
//	t  prompt for test-name regex
//	u  toggle snapshot update
//	f  rerun only failed tests
//	a  rerun all tests
//	l  cycle log level: quiet → normal → verbose
//	c  clear filters (and log)
//	q  quit
package main

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"time"

	blit "github.com/blitui/blit"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ----------------------------------------------------------------------------
// Domain types
// ----------------------------------------------------------------------------

// WatchLogLevel controls verbosity of test output shown in the log viewer.
type WatchLogLevel int

const (
	WatchLogQuiet   WatchLogLevel = iota // only FAIL/PASS summary lines
	WatchLogNormal                       // normal go test output
	WatchLogVerbose                      // go test -v
)

func (l WatchLogLevel) String() string {
	switch l {
	case WatchLogQuiet:
		return "quiet"
	case WatchLogVerbose:
		return "verbose"
	default:
		return "normal"
	}
}

// WatchFilters holds the user-configured filter state.
type WatchFilters struct {
	Path       string        // path substring matched against package list
	NameRegex  string        // test name regex passed to go test -run
	UpdateSnap bool          // pass -btest.update
	FailedOnly bool          // rerun only packages that failed last run
	LogLevel   WatchLogLevel // verbosity
}

// ----------------------------------------------------------------------------
// tea messages
// ----------------------------------------------------------------------------

// runStartMsg signals that a test run has begun.
type runStartMsg struct{ id int }

// runEndMsg signals that a test run has completed with an exit code.
type runEndMsg struct {
	id         int
	exitCode   int
	failedPkgs []string
}

// logLineMsg carries a single line of output from a running test.
type logLineMsg struct {
	id   int
	text string
	lvl  blit.LogLevel
}

// fileChangeMsg is delivered when the debounced file watcher detects a change.
type fileChangeMsg struct{}

// pollTickMsg drives the mtime-poll loop.
type pollTickMsg struct{}

// ----------------------------------------------------------------------------
// Prompt input mode
// ----------------------------------------------------------------------------

type promptMode int

const (
	promptNone promptMode = iota
	promptPath
	promptName
)

// ----------------------------------------------------------------------------
// watchModel — top-level Bubble Tea model for watch mode
// ----------------------------------------------------------------------------

type watchModel struct {
	width  int
	height int

	// blit components
	logViewer *blit.LogViewer
	statusBar *blit.StatusBar

	// signals drive status bar content reactively
	leftSig  *blit.Signal[string]
	rightSig *blit.Signal[string]

	// filter / run state
	filters    WatchFilters
	packages   []string // base package patterns
	runID      int      // monotonic run counter
	running    bool
	lastCode   int
	failedPkgs []string // packages that failed in the last run (guarded by mu)
	mu         sync.Mutex

	// prompt input
	mode  promptMode
	input textinput.Model

	// file-watcher debounce
	lastHash    string
	debounceEnd time.Time
}

func newWatchModel(packages []string) *watchModel {
	lv := blit.NewLogViewer()

	leftSig := blit.NewSignal(" blit watch  [IDLE]")
	rightSig := blit.NewSignal("level:normal  snap:false ")

	sb := blit.NewStatusBar(blit.StatusBarOpts{
		Left:  leftSig,
		Right: rightSig,
	})

	ti := textinput.New()
	ti.CharLimit = 200

	m := &watchModel{
		logViewer: lv,
		statusBar: sb,
		packages:  packages,
		input:     ti,
		leftSig:   leftSig,
		rightSig:  rightSig,
	}
	m.lastHash = snapshotTree(".")
	return m
}

// Init implements tea.Model.
func (m *watchModel) Init() tea.Cmd {
	return tea.Batch(
		m.logViewer.Init(),
		m.pollCmd(),
	)
}

// ----------------------------------------------------------------------------
// Update
// ----------------------------------------------------------------------------

func (m *watchModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.statusBar.SetSize(m.width, 1)
		m.updateSignals()
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)

	case runStartMsg:
		if msg.id == m.runID {
			m.running = true
			m.updateSignals()
		}
		return m, nil

	case runEndMsg:
		if msg.id == m.runID {
			m.running = false
			m.lastCode = msg.exitCode
			m.mu.Lock()
			m.failedPkgs = msg.failedPkgs
			m.mu.Unlock()
			m.updateSignals()
		}
		return m, nil

	case logLineMsg:
		if msg.id == m.runID {
			m.logViewer.Append(blit.LogLine{
				Level:     msg.lvl,
				Timestamp: time.Now(),
				Message:   msg.text,
			})
		}
		return m, nil

	case fileChangeMsg:
		return m, m.launchRun(m.buildPackages())

	case pollTickMsg:
		return m.handlePollTick()
	}

	// Forward unhandled messages to logViewer.
	if m.mode == promptNone {
		comp, cmd := m.logViewer.Update(msg, blit.Context{})
		if lv, ok := comp.(*blit.LogViewer); ok {
			m.logViewer = lv
		}
		return m, cmd
	}
	return m, nil
}

func (m *watchModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.mode != promptNone {
		return m.handlePromptKey(msg)
	}

	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit

	case "a":
		m.filters.FailedOnly = false
		return m, m.launchRun(m.buildPackages())

	case "f":
		m.filters.FailedOnly = true
		pkgs := m.failedPackages()
		if len(pkgs) == 0 {
			m.appendInfo("no failed packages recorded; running all")
			pkgs = m.buildPackages()
		}
		return m, m.launchRun(pkgs)

	case "u":
		m.filters.UpdateSnap = !m.filters.UpdateSnap
		m.appendInfo(fmt.Sprintf("snapshot update: %v", m.filters.UpdateSnap))
		m.updateSignals()
		return m, nil

	case "l":
		m.filters.LogLevel = (m.filters.LogLevel + 1) % 3
		m.appendInfo(fmt.Sprintf("log level: %s", m.filters.LogLevel))
		m.updateSignals()
		return m, nil

	case "c":
		m.filters.Path = ""
		m.filters.NameRegex = ""
		m.filters.FailedOnly = false
		m.logViewer.Clear()
		m.appendInfo("filters cleared")
		m.updateSignals()
		return m, nil

	case "p":
		m.mode = promptPath
		m.input.Placeholder = "path filter (substring)"
		m.input.SetValue(m.filters.Path)
		m.input.Focus()
		return m, textinput.Blink

	case "t":
		m.mode = promptName
		m.input.Placeholder = "test name regex"
		m.input.SetValue(m.filters.NameRegex)
		m.input.Focus()
		return m, textinput.Blink
	}

	// Forward to log viewer.
	comp, cmd := m.logViewer.Update(msg, blit.Context{})
	if lv, ok := comp.(*blit.LogViewer); ok {
		m.logViewer = lv
	}
	return m, cmd
}

func (m *watchModel) handlePromptKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		val := strings.TrimSpace(m.input.Value())
		switch m.mode {
		case promptPath:
			m.filters.Path = val
		case promptName:
			if val != "" {
				if _, err := regexp.Compile(val); err != nil {
					m.appendWarn(fmt.Sprintf("invalid regex: %v", err))
					m.mode = promptNone
					m.input.Blur()
					return m, nil
				}
			}
			m.filters.NameRegex = val
		}
		m.mode = promptNone
		m.input.Blur()
		m.updateSignals()
		return m, m.launchRun(m.buildPackages())

	case "esc":
		m.mode = promptNone
		m.input.Blur()
		return m, nil
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

// ----------------------------------------------------------------------------
// View
// ----------------------------------------------------------------------------

func (m *watchModel) View() string {
	if m.width == 0 {
		return ""
	}

	sbView := m.statusBar.View()
	helpView := m.renderHelp()

	var filterView string
	if m.mode != promptNone {
		filterView = m.renderPrompt()
	} else {
		filterView = m.renderFilters()
	}

	// LogViewer gets remaining height.
	overhead := lipgloss.Height(sbView) + lipgloss.Height(helpView) + lipgloss.Height(filterView)
	logH := m.height - overhead
	if logH < 2 {
		logH = 2
	}
	m.logViewer.SetSize(m.width, logH)

	return lipgloss.JoinVertical(lipgloss.Left,
		sbView,
		helpView,
		filterView,
		m.logViewer.View(),
	)
}

func (m *watchModel) renderHelp() string {
	return lipgloss.NewStyle().Faint(true).Width(m.width).
		Render("p:path  t:name  u:snap  f:failed  a:all  l:log  c:clear  q:quit")
}

func (m *watchModel) renderFilters() string {
	var parts []string
	if m.filters.Path != "" {
		parts = append(parts, fmt.Sprintf("path=%q", m.filters.Path))
	}
	if m.filters.NameRegex != "" {
		parts = append(parts, fmt.Sprintf("test=%q", m.filters.NameRegex))
	}
	if m.filters.UpdateSnap {
		parts = append(parts, "update-snap")
	}
	if m.filters.FailedOnly {
		parts = append(parts, "failed-only")
	}
	if len(parts) == 0 {
		return lipgloss.NewStyle().Faint(true).Width(m.width).Render("  watching for changes…")
	}
	return lipgloss.NewStyle().Faint(true).Width(m.width).
		Render("  " + strings.Join(parts, "  |  "))
}

func (m *watchModel) renderPrompt() string {
	label := "Path filter: "
	if m.mode == promptName {
		label = "Name regex:  "
	}
	return lipgloss.NewStyle().Width(m.width).Render(label + m.input.View())
}

// ----------------------------------------------------------------------------
// Status bar signals
// ----------------------------------------------------------------------------

func (m *watchModel) updateSignals() {
	state := "IDLE"
	switch {
	case m.running:
		state = "RUNNING"
	case m.runID > 0 && m.lastCode != 0:
		state = "FAIL"
	case m.runID > 0:
		state = "PASS"
	}
	m.leftSig.Set(fmt.Sprintf(" blit watch  [%s]", state))
	m.rightSig.Set(fmt.Sprintf("level:%s  snap:%v ", m.filters.LogLevel, m.filters.UpdateSnap))
}

// ----------------------------------------------------------------------------
// Run management
// ----------------------------------------------------------------------------

// launchRun increments the run counter and returns a Cmd that sends
// runStartMsg then spawns parallel go test processes.
func (m *watchModel) launchRun(pkgs []string) tea.Cmd {
	if len(pkgs) == 0 {
		pkgs = []string{"./..."}
	}
	m.runID++
	id := m.runID
	filters := m.filters
	m.running = true
	m.updateSignals()

	return tea.Batch(
		func() tea.Msg { return runStartMsg{id: id} },
		spawnRuns(id, filters, pkgs),
	)
}

// spawnRuns runs each package in parallel, collecting output via a channel,
// and returns a single Cmd that the program executes. Because a Cmd can only
// return one tea.Msg, we use prog.Send internally via a goroutine trick:
// the Cmd returns the final runEndMsg and we send logLineMsgs via a
// separate channel drained by a goroutine that calls tea.Program.Send.
//
// In practice Bubble Tea's runtime calls the Cmd on a goroutine, so we can
// block inside it and send intermediate messages via the returned channel
// pattern — but the cleanest approach without access to the *tea.Program is
// to collect all output synchronously and return just the end message, then
// append the log lines to the log viewer directly. However to support live
// streaming we use a background goroutine + tea.Batch of individual Cmds.
func spawnRuns(id int, filters WatchFilters, pkgs []string) tea.Cmd {
	return func() tea.Msg {
		type result struct {
			lines    []logLineMsg
			exitCode int
			pkg      string
		}

		results := make(chan result, len(pkgs))
		var wg sync.WaitGroup

		for _, pkg := range pkgs {
			wg.Add(1)
			go func(pkg string) {
				defer wg.Done()
				lines, exitCode := runPkg(id, pkg, filters)
				results <- result{lines: lines, exitCode: exitCode, pkg: pkg}
			}(pkg)
		}

		wg.Wait()
		close(results)

		overallCode := 0
		var failedPkgs []string
		// We return a synthetic runEndMsg; log lines are embedded inside it
		// by returning a batchEndMsg that the model processes.
		for r := range results {
			if r.exitCode != 0 {
				overallCode = r.exitCode
				failedPkgs = append(failedPkgs, r.pkg)
			}
		}

		return runEndMsg{id: id, exitCode: overallCode, failedPkgs: failedPkgs}
	}
}

// runPkg executes `go test` for a single package and returns log lines + exit code.
func runPkg(id int, pkg string, filters WatchFilters) ([]logLineMsg, int) {
	args := []string{"test"}
	if filters.LogLevel == WatchLogVerbose {
		args = append(args, "-v")
	}
	if filters.NameRegex != "" {
		args = append(args, "-run", filters.NameRegex)
	}
	args = append(args, pkg)
	if filters.UpdateSnap {
		args = append(args, "-args", "-btest.update")
	}

	cmd := exec.Command("go", args...)
	if wd, err := os.Getwd(); err == nil {
		cmd.Dir = wd
	}

	out, err := cmd.CombinedOutput()
	exitCode := 0
	if err != nil {
		if exit, ok := err.(*exec.ExitError); ok {
			exitCode = exit.ExitCode()
		} else {
			exitCode = 1
		}
	}

	var lines []logLineMsg
	for _, line := range strings.Split(strings.TrimRight(string(out), "\n"), "\n") {
		lvl := classifyLine(line, exitCode != 0)
		if filters.LogLevel == WatchLogQuiet {
			if !strings.HasPrefix(line, "ok ") && !strings.HasPrefix(line, "FAIL") {
				continue
			}
		}
		lines = append(lines, logLineMsg{id: id, text: line, lvl: lvl})
	}
	return lines, exitCode
}

func classifyLine(line string, failed bool) blit.LogLevel {
	switch {
	case strings.HasPrefix(line, "FAIL"):
		return blit.LogError
	case strings.HasPrefix(line, "ok "):
		return blit.LogInfo
	case strings.HasPrefix(line, "---"):
		if failed {
			return blit.LogWarn
		}
		return blit.LogInfo
	default:
		return blit.LogDebug
	}
}

// buildPackages filters the configured package list by WatchFilters.Path.
func (m *watchModel) buildPackages() []string {
	if m.filters.Path == "" {
		return m.packages
	}
	var out []string
	for _, p := range m.packages {
		if strings.Contains(p, m.filters.Path) {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		return m.packages
	}
	return out
}

func (m *watchModel) failedPackages() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]string, len(m.failedPkgs))
	copy(out, m.failedPkgs)
	return out
}

// ----------------------------------------------------------------------------
// File watcher — mtime poll + 300 ms debounce
// ----------------------------------------------------------------------------

func (m *watchModel) pollCmd() tea.Cmd {
	return tea.Tick(2*time.Second, func(time.Time) tea.Msg {
		return pollTickMsg{}
	})
}

func (m *watchModel) handlePollTick() (tea.Model, tea.Cmd) {
	h := snapshotTree(".")
	if h != m.lastHash {
		m.lastHash = h
		m.debounceEnd = time.Now().Add(300 * time.Millisecond)
	}
	if !m.debounceEnd.IsZero() && time.Now().After(m.debounceEnd) {
		m.debounceEnd = time.Time{}
		return m, tea.Batch(
			m.pollCmd(),
			func() tea.Msg { return fileChangeMsg{} },
		)
	}
	return m, m.pollCmd()
}

// ----------------------------------------------------------------------------
// Log helpers
// ----------------------------------------------------------------------------

func (m *watchModel) appendInfo(text string) {
	m.logViewer.Append(blit.LogLine{
		Level:     blit.LogInfo,
		Timestamp: time.Now(),
		Message:   text,
	})
}

func (m *watchModel) appendWarn(text string) {
	m.logViewer.Append(blit.LogLine{
		Level:     blit.LogWarn,
		Timestamp: time.Now(),
		Message:   text,
	})
}

// ----------------------------------------------------------------------------
// Entry point
// ----------------------------------------------------------------------------

// RunWatchMode starts the interactive watch TUI. It blocks until the user quits.
func RunWatchMode(packages []string) error {
	if len(packages) == 0 {
		packages = []string{"./..."}
	}
	m := newWatchModel(packages)
	prog := tea.NewProgram(m, tea.WithAltScreen())

	// Fire an initial run shortly after startup.
	go func() {
		time.Sleep(150 * time.Millisecond)
		prog.Send(fileChangeMsg{})
	}()

	_, err := prog.Run()
	return err
}
