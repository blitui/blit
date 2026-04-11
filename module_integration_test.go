package blit

import (
	"path/filepath"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// --- integration test helpers ---

// orderModule records Init and Update call order into a shared slice.
type orderModule struct {
	name      string
	initOrder *[]string
	updOrder  *[]string
}

func (m *orderModule) Name() string { return m.name }
func (m *orderModule) Init() tea.Cmd {
	*m.initOrder = append(*m.initOrder, m.name)
	return nil
}
func (m *orderModule) Update(msg tea.Msg, ctx Context) (Module, tea.Cmd) {
	*m.updOrder = append(*m.updOrder, m.name)
	return m, nil
}

// contextCapture captures the Context passed to Update for later inspection.
type contextCapture struct {
	name    string
	lastCtx *Context
}

func (m *contextCapture) Name() string  { return m.name }
func (m *contextCapture) Init() tea.Cmd { return nil }
func (m *contextCapture) Update(msg tea.Msg, ctx Context) (Module, tea.Cmd) {
	*m.lastCtx = ctx
	return m, nil
}

// replacingModule returns a brand-new Module from Update to test replacement.
type replacingModule struct {
	name      string
	iteration int
}

func (m *replacingModule) Name() string  { return m.name }
func (m *replacingModule) Init() tea.Cmd { return nil }
func (m *replacingModule) Update(msg tea.Msg, ctx Context) (Module, tea.Cmd) {
	return &replacingModule{name: m.name, iteration: m.iteration + 1}, nil
}

// providerModule implements ModuleWithProviders.
type providerModule struct {
	name  string
	provs []DebugProvider
}

func (m *providerModule) Name() string  { return m.name }
func (m *providerModule) Init() tea.Cmd { return nil }
func (m *providerModule) Update(msg tea.Msg, ctx Context) (Module, tea.Cmd) {
	return m, nil
}
func (m *providerModule) Providers() []DebugProvider { return m.provs }

// keybindModule implements ModuleWithKeybinds.
type keybindModule struct {
	name  string
	binds []KeyBind
}

func (m *keybindModule) Name() string  { return m.name }
func (m *keybindModule) Init() tea.Cmd { return nil }
func (m *keybindModule) Update(msg tea.Msg, ctx Context) (Module, tea.Cmd) {
	return m, nil
}
func (m *keybindModule) Keybinds() []KeyBind { return m.binds }

// --- integration tests ---

func TestMultipleModulesInitOrder(t *testing.T) {
	var initOrder, updOrder []string
	m1 := &orderModule{name: "alpha", initOrder: &initOrder, updOrder: &updOrder}
	m2 := &orderModule{name: "beta", initOrder: &initOrder, updOrder: &updOrder}
	m3 := &orderModule{name: "gamma", initOrder: &initOrder, updOrder: &updOrder}

	a := newAppModel(
		WithModule(m1),
		WithModule(m2),
		WithModule(m3),
	)
	a.Init()

	if len(initOrder) != 3 {
		t.Fatalf("expected 3 init calls, got %d", len(initOrder))
	}
	want := []string{"alpha", "beta", "gamma"}
	for i, name := range want {
		if initOrder[i] != name {
			t.Errorf("init order[%d]: want %q, got %q", i, name, initOrder[i])
		}
	}
}

func TestMultipleModulesUpdateOrder(t *testing.T) {
	var initOrder, updOrder []string
	m1 := &orderModule{name: "alpha", initOrder: &initOrder, updOrder: &updOrder}
	m2 := &orderModule{name: "beta", initOrder: &initOrder, updOrder: &updOrder}
	m3 := &orderModule{name: "gamma", initOrder: &initOrder, updOrder: &updOrder}

	a := newAppModel(
		WithModule(m1),
		WithModule(m2),
		WithModule(m3),
	)
	a.width = 80
	a.height = 24

	type customMsg struct{}
	a.Update(customMsg{})

	if len(updOrder) != 3 {
		t.Fatalf("expected 3 update calls, got %d", len(updOrder))
	}
	want := []string{"alpha", "beta", "gamma"}
	for i, name := range want {
		if updOrder[i] != name {
			t.Errorf("update order[%d]: want %q, got %q", i, name, updOrder[i])
		}
	}
}

func TestModuleReplacementAfterUpdate(t *testing.T) {
	m := &replacingModule{name: "replace", iteration: 0}
	a := newAppModel(WithModule(m))
	a.width = 80
	a.height = 24

	type pingMsg struct{}
	a.Update(pingMsg{})

	got, ok := a.modules[0].(*replacingModule)
	if !ok {
		t.Fatal("expected module to be *replacingModule")
	}
	if got.iteration != 1 {
		t.Errorf("expected iteration 1 after one Update, got %d", got.iteration)
	}

	// Second update should produce iteration 2
	a.Update(pingMsg{})
	got = a.modules[0].(*replacingModule)
	if got.iteration != 2 {
		t.Errorf("expected iteration 2 after two Updates, got %d", got.iteration)
	}
}

func TestContextFieldsPopulated(t *testing.T) {
	var captured Context
	m := &contextCapture{name: "ctx-test", lastCtx: &captured}

	a := newAppModel(
		WithTheme(DefaultTheme()),
		WithModule(m),
	)
	a.width = 120
	a.height = 40

	type customMsg struct{}
	a.Update(customMsg{})

	// Verify Theme is populated
	if captured.Theme.Accent == "" {
		t.Error("expected Theme.Accent to be populated")
	}
	if captured.Theme.Text == "" {
		t.Error("expected Theme.Text to be populated")
	}

	// Verify Size is populated
	if captured.Size.Width != 120 {
		t.Errorf("expected Size.Width=120, got %d", captured.Size.Width)
	}
	if captured.Size.Height != 40 {
		t.Errorf("expected Size.Height=40, got %d", captured.Size.Height)
	}

	// Verify Hotkeys registry exists
	if captured.Hotkeys == nil {
		t.Error("expected Hotkeys registry to be non-nil")
	}
}

func TestDevConsoleAndCustomModuleTogether(t *testing.T) {
	prov := &mockDebugProvider{name: "custom-metrics"}
	mod := &providerModule{
		name:  "metrics",
		provs: []DebugProvider{prov},
	}

	a := newAppModel(
		WithDevConsole(),
		WithModule(mod),
	)

	if a.devConsole == nil {
		t.Fatal("expected devConsole to exist")
	}

	// DevConsole should have 6 built-in + 1 custom provider
	if len(a.devConsole.providers) != 7 {
		t.Fatalf("expected 7 providers, got %d", len(a.devConsole.providers))
	}

	// The custom provider should be last
	last := a.devConsole.providers[len(a.devConsole.providers)-1]
	if last.Name() != "custom-metrics" {
		t.Errorf("expected last provider %q, got %q", "custom-metrics", last.Name())
	}
}

func TestPollerDebugProviderRenders(t *testing.T) {
	p := NewPollerWithOpts(PollerOpts{
		Name:     "test-api",
		Interval: 5 * time.Second,
		Fetch:    func() (tea.Msg, error) { return nil, nil },
	})

	dp := p.DebugProvider()
	if dp == nil {
		t.Fatal("expected DebugProvider to be non-nil for enhanced poller")
	}

	if dp.Name() != "test-api Poller" {
		t.Errorf("expected provider name %q, got %q", "test-api Poller", dp.Name())
	}

	// View should render stats
	view := dp.View(60, 10, DefaultTheme())
	if view == "" {
		t.Error("expected non-empty View output")
	}

	// Data should return structured data
	data := dp.Data()
	if data["name"] != "test-api" {
		t.Errorf("expected data[name]=%q, got %v", "test-api", data["name"])
	}
}

func TestConfigEditorGeneration(t *testing.T) {
	type testConfig struct {
		Name  string `blit:"label=Name,group=General,default=test-app"`
		Port  int    `blit:"label=Port,group=Network,default=8080,min=1,max=65535"`
		Debug bool   `blit:"label=Debug,group=General,default=false"`
	}

	tmpDir := t.TempDir()
	cfg, err := LoadConfig[testConfig]("test-app", WithConfigPath(filepath.Join(tmpDir, "config.yaml")))
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	editor := cfg.Editor()
	if editor == nil {
		t.Fatal("expected Editor to be non-nil")
	}

	// Should have 3 fields (Name, Port, Debug)
	if len(editor.fields) != 3 {
		t.Fatalf("expected 3 fields, got %d", len(editor.fields))
	}

	// Verify field labels
	labels := make(map[string]bool)
	for _, f := range editor.fields {
		labels[f.Label] = true
	}
	for _, want := range []string{"Name", "Port", "Debug"} {
		if !labels[want] {
			t.Errorf("expected field with label %q", want)
		}
	}

	// Test Get/Set roundtrip on the Name field
	for _, f := range editor.fields {
		if f.Label == "Name" && f.Set != nil {
			if err := f.Set("new-name"); err != nil {
				t.Fatalf("Set failed: %v", err)
			}
			got := f.Get()
			if got != "new-name" {
				t.Errorf("expected Get()=%q after Set, got %q", "new-name", got)
			}
		}
	}
}

func TestThreeModulesTogether(t *testing.T) {
	var initOrder, updOrder []string

	mod1 := &orderModule{name: "mod-a", initOrder: &initOrder, updOrder: &updOrder}
	mod2 := &orderModule{name: "mod-b", initOrder: &initOrder, updOrder: &updOrder}

	prov := &mockDebugProvider{name: "extra-tab"}
	mod3 := &providerModule{name: "mod-c", provs: []DebugProvider{prov}}

	a := newAppModel(
		WithDevConsole(),
		WithModule(mod1),
		WithModule(mod2),
		WithModule(mod3),
	)
	a.width = 80
	a.height = 24
	a.Init()

	// All three should have initialized in order
	if len(initOrder) != 2 {
		// mod3 is a providerModule, not an orderModule, so only mod1 and mod2 track init
		t.Fatalf("expected 2 tracked inits, got %d", len(initOrder))
	}

	// Send a TickMsg — modules should receive it
	a.Update(TickMsg{Time: time.Now()})

	if len(updOrder) != 2 {
		t.Fatalf("expected 2 tracked update calls from TickMsg, got %d", len(updOrder))
	}

	// DevConsole should have built-in(6) + mod3's provider(1) = 7
	if len(a.devConsole.providers) != 7 {
		t.Fatalf("expected 7 providers, got %d", len(a.devConsole.providers))
	}
}

func TestModuleWithProvidersInDevConsole(t *testing.T) {
	p1 := &mockDebugProvider{name: "p1"}
	p2 := &mockDebugProvider{name: "p2"}
	mod := &providerModule{
		name:  "multi-prov",
		provs: []DebugProvider{p1, p2},
	}

	a := newAppModel(WithModule(mod))

	if a.devConsole == nil {
		t.Fatal("expected devConsole to be auto-created for module with providers")
	}

	// 6 built-in + 2 from module
	if len(a.devConsole.providers) != 8 {
		t.Fatalf("expected 8 providers, got %d", len(a.devConsole.providers))
	}

	// Verify they appear at the end in order
	provs := a.devConsole.providers
	if provs[6].Name() != "p1" {
		t.Errorf("expected provider[6]=%q, got %q", "p1", provs[6].Name())
	}
	if provs[7].Name() != "p2" {
		t.Errorf("expected provider[7]=%q, got %q", "p2", provs[7].Name())
	}
}

func TestEmptyAppWithModulesOnly(t *testing.T) {
	var initOrder, updOrder []string
	m := &orderModule{name: "sole", initOrder: &initOrder, updOrder: &updOrder}

	a := newAppModel(WithModule(m))
	a.width = 80
	a.height = 24
	a.Init()

	if len(initOrder) != 1 || initOrder[0] != "sole" {
		t.Fatalf("expected init of 'sole', got %v", initOrder)
	}

	type customMsg struct{}
	a.Update(customMsg{})

	if len(updOrder) != 1 || updOrder[0] != "sole" {
		t.Fatalf("expected update of 'sole', got %v", updOrder)
	}

	// View should not panic with no components
	_ = a.View()
}

func TestModulesReceiveTickMsg(t *testing.T) {
	received := false
	var receivedMsg tea.Msg

	m := &mockModule{name: "tick-test"}
	a := newAppModel(WithModule(m))
	a.width = 80
	a.height = 24

	tick := TickMsg{Time: time.Now()}
	a.Update(tick)

	for _, msg := range m.updates {
		if tm, ok := msg.(TickMsg); ok {
			received = true
			receivedMsg = tm
		}
	}

	if !received {
		t.Fatal("expected module to receive TickMsg")
	}
	_ = receivedMsg
}

func TestKeybindModuleRegistration(t *testing.T) {
	m := &keybindModule{
		name: "shortcuts",
		binds: []KeyBind{
			{Key: "ctrl+k", Label: "Search", Group: "NAV"},
			{Key: "ctrl+j", Label: "Jump", Group: "NAV"},
		},
	}

	a := newAppModel(WithModule(m))

	binds, ok := a.registry.sources["shortcuts"]
	if !ok {
		t.Fatal("expected registry source 'shortcuts'")
	}
	if len(binds) != 2 {
		t.Fatalf("expected 2 keybinds, got %d", len(binds))
	}

	keys := make(map[string]bool)
	for _, b := range binds {
		keys[b.Key] = true
	}
	if !keys["ctrl+k"] || !keys["ctrl+j"] {
		t.Errorf("expected ctrl+k and ctrl+j in binds, got %v", keys)
	}
}

func TestMultiModuleApp(t *testing.T) {
	// Simulate a realistic multi-module app with:
	// 1. A poller-like module with debug provider
	// 2. A config-like module with keybinds
	// 3. DevConsole to show everything

	pollerProv := &mockDebugProvider{name: "API Poller"}
	pollerMod := &providerModule{
		name:  "poller",
		provs: []DebugProvider{pollerProv},
	}

	configMod := &keybindModule{
		name: "config",
		binds: []KeyBind{
			{Key: "ctrl+,", Label: "Open config", Group: "APP"},
		},
	}

	var captured Context
	ctxMod := &contextCapture{name: "ctx-spy", lastCtx: &captured}

	a := newAppModel(
		WithDevConsole(),
		WithTheme(DefaultTheme()),
		WithModule(pollerMod),
		WithModule(configMod),
		WithModule(ctxMod),
	)
	a.width = 100
	a.height = 50
	a.Init()

	// Verify DevConsole has built-in(6) + poller provider(1) = 7
	if len(a.devConsole.providers) != 7 {
		t.Fatalf("expected 7 providers, got %d", len(a.devConsole.providers))
	}

	// Verify config keybinds registered
	binds, ok := a.registry.sources["config"]
	if !ok {
		t.Fatal("expected registry source 'config'")
	}
	found := false
	for _, b := range binds {
		if b.Key == "ctrl+," {
			found = true
		}
	}
	if !found {
		t.Error("expected ctrl+, keybind from config module")
	}

	// Send a custom message, verify context spy captured it
	type appEvent struct{ value int }
	a.Update(appEvent{value: 42})

	if captured.Size.Width != 100 || captured.Size.Height != 50 {
		t.Errorf("expected ctx size 100x50, got %dx%d", captured.Size.Width, captured.Size.Height)
	}
	if captured.Theme.Accent == "" {
		t.Error("expected Theme to be populated in context")
	}
	if captured.Hotkeys == nil {
		t.Error("expected Hotkeys registry in context")
	}

	// Send a TickMsg, verify modules still receive it
	a.Update(TickMsg{Time: time.Now()})

	// Context spy should have updated again
	if captured.Size.Width != 100 {
		t.Error("expected context to be populated on TickMsg dispatch")
	}
}
