package blit

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// --- test doubles ---

// mockModule is a minimal Module for testing.
type mockModule struct {
	name       string
	initCalled bool
	initCmd    tea.Cmd
	updates    []tea.Msg
}

func (m *mockModule) Name() string { return m.name }
func (m *mockModule) Init() tea.Cmd {
	m.initCalled = true
	return m.initCmd
}
func (m *mockModule) Update(msg tea.Msg, ctx Context) (Module, tea.Cmd) {
	m.updates = append(m.updates, msg)
	return m, nil
}

// mockModuleWithKeybinds implements ModuleWithKeybinds.
type mockModuleWithKeybinds struct {
	mockModule
	binds []KeyBind
}

func (m *mockModuleWithKeybinds) Keybinds() []KeyBind { return m.binds }

// mockModuleWithProviders implements ModuleWithProviders.
type mockModuleWithProviders struct {
	mockModule
	provs []DebugProvider
}

func (m *mockModuleWithProviders) Providers() []DebugProvider { return m.provs }

// mockDebugProvider is a minimal DebugProvider for testing.
type mockDebugProvider struct {
	name string
}

func (p *mockDebugProvider) Name() string                         { return p.name }
func (p *mockDebugProvider) View(w, h int, theme Theme) string    { return p.name }

// mockDebugDataProvider implements DebugDataProvider.
type mockDebugDataProvider struct {
	mockDebugProvider
	data map[string]any
}

func (p *mockDebugDataProvider) Data() map[string]any { return p.data }

// --- compile-time interface assertions ---

func TestModuleInterface(t *testing.T) {
	// These assignments verify that mock types satisfy the interfaces at compile time.
	var _ Module = (*mockModule)(nil)
	var _ ModuleWithKeybinds = (*mockModuleWithKeybinds)(nil)
	var _ ModuleWithProviders = (*mockModuleWithProviders)(nil)
	var _ DebugProvider = (*mockDebugProvider)(nil)
	var _ DebugDataProvider = (*mockDebugDataProvider)(nil)
}

// --- functional tests ---

func TestWithModule(t *testing.T) {
	m := &mockModule{name: "test"}
	a := newAppModel(WithModule(m))

	if len(a.modules) != 1 {
		t.Fatalf("expected 1 module, got %d", len(a.modules))
	}
	if a.modules[0].Name() != "test" {
		t.Errorf("expected module name %q, got %q", "test", a.modules[0].Name())
	}

	// Init should have been called during newAppModel -> Init is called by tea.Program,
	// but setup() is called. Let's call Init() explicitly to verify.
	a.Init()
	if !m.initCalled {
		t.Error("expected Init to be called on module")
	}
}

func TestModuleUpdateReceivesMessages(t *testing.T) {
	m := &mockModule{name: "test"}
	c := &stubComponent{name: "comp"}
	a := newAppModel(
		WithModule(m),
		WithComponent("comp", c),
	)
	a.width = 80
	a.height = 24

	// Send a custom message through Update (not a key/mouse/tick — those have special handling)
	type customMsg struct{}
	a.Update(customMsg{})

	if len(m.updates) == 0 {
		t.Fatal("expected module to receive at least one Update call")
	}

	found := false
	for _, msg := range m.updates {
		if _, ok := msg.(customMsg); ok {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected module to receive customMsg")
	}
}

func TestModuleWithProvidersRegistration(t *testing.T) {
	prov := &mockDebugProvider{name: "metrics"}
	m := &mockModuleWithProviders{
		mockModule: mockModule{name: "mod"},
		provs:      []DebugProvider{prov},
	}
	a := newAppModel(WithModule(m))

	if a.devConsole == nil {
		t.Fatal("expected devConsole to be created for module with providers")
	}
	if len(a.devConsole.providers) != 1 {
		t.Fatalf("expected 1 provider, got %d", len(a.devConsole.providers))
	}
	if a.devConsole.providers[0].Name() != "metrics" {
		t.Errorf("expected provider name %q, got %q", "metrics", a.devConsole.providers[0].Name())
	}
}

func TestModuleWithKeybindsRegistration(t *testing.T) {
	m := &mockModuleWithKeybinds{
		mockModule: mockModule{name: "nav"},
		binds: []KeyBind{
			{Key: "ctrl+p", Label: "Command palette", Group: "NAV"},
		},
	}
	a := newAppModel(WithModule(m))

	// Verify the keybinding was registered under the module's name
	binds, ok := a.registry.sources["nav"]
	if !ok {
		t.Fatal("expected registry to contain source \"nav\"")
	}
	found := false
	for _, b := range binds {
		if b.Key == "ctrl+p" && b.Label == "Command palette" {
			found = true
		}
	}
	if !found {
		t.Error("expected module keybinding ctrl+p to be registered")
	}
}

func TestModuleLifecycleOrder(t *testing.T) {
	var order []string
	makeModule := func(name string) *orderTrackingModule {
		return &orderTrackingModule{name: name, order: &order}
	}

	m1 := makeModule("first")
	m2 := makeModule("second")
	m3 := makeModule("third")

	a := newAppModel(
		WithModule(m1),
		WithModule(m2),
		WithModule(m3),
	)
	a.Init()

	if len(order) != 3 {
		t.Fatalf("expected 3 init calls, got %d", len(order))
	}
	expected := []string{"first", "second", "third"}
	for i, name := range expected {
		if order[i] != name {
			t.Errorf("init order[%d]: expected %q, got %q", i, name, order[i])
		}
	}
}

func TestWithDebugProvider(t *testing.T) {
	prov := &mockDebugProvider{name: "standalone"}
	a := newAppModel(WithDebugProvider(prov))

	if a.devConsole == nil {
		t.Fatal("expected devConsole to be created for standalone provider")
	}
	if len(a.devConsole.providers) != 1 {
		t.Fatalf("expected 1 provider, got %d", len(a.devConsole.providers))
	}
	if a.devConsole.providers[0].Name() != "standalone" {
		t.Errorf("expected provider name %q, got %q", "standalone", a.devConsole.providers[0].Name())
	}
}

func TestModuleUpdateStoresUpdatedModule(t *testing.T) {
	m := &counterModule{name: "counter"}
	a := newAppModel(WithModule(m))
	a.width = 80
	a.height = 24

	type pingMsg struct{}
	a.Update(pingMsg{})

	// The module should have been replaced with the updated version
	cm, ok := a.modules[0].(*counterModule)
	if !ok {
		t.Fatal("expected module to be *counterModule")
	}
	if cm.count != 1 {
		t.Errorf("expected count 1 after one Update, got %d", cm.count)
	}
}

// --- helper types ---

// orderTrackingModule records the order of Init calls.
type orderTrackingModule struct {
	name  string
	order *[]string
}

func (m *orderTrackingModule) Name() string { return m.name }
func (m *orderTrackingModule) Init() tea.Cmd {
	*m.order = append(*m.order, m.name)
	return nil
}
func (m *orderTrackingModule) Update(msg tea.Msg, ctx Context) (Module, tea.Cmd) {
	return m, nil
}

// counterModule increments a count on each Update to verify module replacement.
type counterModule struct {
	name  string
	count int
}

func (m *counterModule) Name() string { return m.name }
func (m *counterModule) Init() tea.Cmd { return nil }
func (m *counterModule) Update(msg tea.Msg, ctx Context) (Module, tea.Cmd) {
	m.count++
	return m, nil
}
