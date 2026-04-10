package blit

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// TestDevConsole_ImplementsComponent verifies devConsole satisfies Component.
func TestDevConsole_ImplementsComponent(t *testing.T) {
	var _ Component = (*devConsole)(nil)
}

// TestDevConsole_HasModuleMethods verifies devConsole has Name() and Providers()
// methods consistent with the module system. Note: devConsole cannot implement
// both Component and Module simultaneously because their Update methods have
// different return types (Component vs Module).
func TestDevConsole_HasModuleMethods(t *testing.T) {
	dc := newDevConsole()
	if dc.Name() != "devConsole" {
		t.Errorf("expected Name() = %q, got %q", "devConsole", dc.Name())
	}
	provs := dc.Providers()
	if len(provs) != 6 {
		t.Errorf("expected 6 providers from Providers(), got %d", len(provs))
	}
}

// TestDevConsole_BuiltinProviders verifies that newDevConsole registers 6
// built-in providers with the expected names.
func TestDevConsole_BuiltinProviders(t *testing.T) {
	dc := newDevConsole()

	expectedNames := []string{
		"Frame Stats",
		"Components",
		"Signals",
		"Keys",
		"Logs",
		"Theme",
	}

	if len(dc.providers) != len(expectedNames) {
		t.Fatalf("expected %d built-in providers, got %d", len(expectedNames), len(dc.providers))
	}

	for i, want := range expectedNames {
		got := dc.providers[i].Name()
		if got != want {
			t.Errorf("provider[%d]: expected name %q, got %q", i, want, got)
		}
	}
}

// TestDevConsole_CustomProvider verifies that WithDebugProvider appends a
// custom provider after the built-in ones.
func TestDevConsole_CustomProvider(t *testing.T) {
	custom := &mockDebugProvider{name: "custom-metrics"}
	a := newAppModel(
		WithDevConsole(),
		WithDebugProvider(custom),
	)

	if a.devConsole == nil {
		t.Fatal("expected devConsole to exist")
	}

	// 6 built-in + 1 custom
	if len(a.devConsole.providers) != 7 {
		t.Fatalf("expected 7 providers, got %d", len(a.devConsole.providers))
	}

	last := a.devConsole.providers[6]
	if last.Name() != "custom-metrics" {
		t.Errorf("expected custom provider name %q, got %q", "custom-metrics", last.Name())
	}
}

// TestDevConsole_TabNavigation verifies left/right arrow keys change the active tab.
func TestDevConsole_TabNavigation(t *testing.T) {
	dc := newDevConsole()
	dc.active = true
	dc.SetTheme(DefaultTheme())
	dc.SetSize(120, 40)
	ctx := Context{Theme: DefaultTheme(), Size: Size{Width: 120, Height: 40}}

	if dc.activeTab != 0 {
		t.Fatalf("expected initial activeTab 0, got %d", dc.activeTab)
	}

	tests := []struct {
		name    string
		key     tea.KeyMsg
		wantTab int
	}{
		{
			name:    "right advances tab",
			key:     tea.KeyMsg{Type: tea.KeyRight},
			wantTab: 1,
		},
		{
			name:    "right again advances tab",
			key:     tea.KeyMsg{Type: tea.KeyRight},
			wantTab: 2,
		},
		{
			name:    "left goes back",
			key:     tea.KeyMsg{Type: tea.KeyLeft},
			wantTab: 1,
		},
		{
			name:    "left wraps to last tab",
			key:     tea.KeyMsg{Type: tea.KeyLeft},
			wantTab: 0,
		},
		{
			name:    "left from 0 wraps to last",
			key:     tea.KeyMsg{Type: tea.KeyLeft},
			wantTab: 5,
		},
		{
			name:    "right from last wraps to 0",
			key:     tea.KeyMsg{Type: tea.KeyRight},
			wantTab: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dc.Update(tt.key, ctx)
			if dc.activeTab != tt.wantTab {
				t.Errorf("expected activeTab %d, got %d", tt.wantTab, dc.activeTab)
			}
		})
	}
}

// TestDevConsole_TabJumpByNumber verifies number keys 1-9 jump to tabs.
func TestDevConsole_TabJumpByNumber(t *testing.T) {
	dc := newDevConsole()
	dc.active = true
	dc.SetTheme(DefaultTheme())
	dc.SetSize(120, 40)
	ctx := Context{Theme: DefaultTheme(), Size: Size{Width: 120, Height: 40}}

	tests := []struct {
		key     string
		wantTab int
	}{
		{"3", 2},
		{"1", 0},
		{"6", 5},
		{"9", 5}, // only 6 providers, so 9 should not change from current
	}

	for _, tt := range tests {
		t.Run("key_"+tt.key, func(t *testing.T) {
			dc.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)}, ctx)
			if dc.activeTab != tt.wantTab {
				t.Errorf("key %q: expected activeTab %d, got %d", tt.key, tt.wantTab, dc.activeTab)
			}
		})
	}
}

// TestDevConsole_ProviderView verifies each built-in provider returns a non-empty View.
func TestDevConsole_ProviderView(t *testing.T) {
	dc := newDevConsole()
	dc.active = true
	dc.SetTheme(DefaultTheme())
	dc.SetSize(120, 40)

	// Populate some snapshot data so providers have content
	dc.snapshot = devConsoleSnapshot{
		componentNames: []string{"main", "sidebar"},
		componentFocus: []bool{true, false},
		signals:        []signalInfo{{label: "count", value: "42"}},
		logLines:       []string{"log line 1"},
		themeName:      "default",
		theme:          DefaultTheme(),
	}
	dc.recordKey("a")
	dc.recordKey("b")

	theme := DefaultTheme()
	for _, p := range dc.providers {
		t.Run(p.Name(), func(t *testing.T) {
			view := p.View(60, 20, theme)
			if view == "" {
				t.Errorf("provider %q returned empty View", p.Name())
			}
		})
	}
}

// TestDevConsole_FrameStatsData verifies the frameStatsProvider implements DebugDataProvider.
func TestDevConsole_FrameStatsData(t *testing.T) {
	dc := newDevConsole()
	p := dc.providers[0]

	dp, ok := p.(DebugDataProvider)
	if !ok {
		t.Fatal("frameStatsProvider should implement DebugDataProvider")
	}

	data := dp.Data()
	if _, ok := data["fps"]; !ok {
		t.Error("expected 'fps' key in Data()")
	}
	if _, ok := data["frameTimeMs"]; !ok {
		t.Error("expected 'frameTimeMs' key in Data()")
	}
	if _, ok := data["frameCount"]; !ok {
		t.Error("expected 'frameCount' key in Data()")
	}
}

// TestDevConsole_RenderPanelWithTabs verifies the panel output includes the
// tab bar with provider names.
func TestDevConsole_RenderPanelWithTabs(t *testing.T) {
	dc := newDevConsole()
	dc.active = true
	dc.SetTheme(DefaultTheme())
	dc.SetSize(120, 40)

	view := dc.View()
	if view == "" {
		t.Fatal("expected non-empty View when active")
	}

	// The tab bar should contain provider names
	for _, name := range []string{"Frame Stats", "Components", "Signals", "Keys", "Logs", "Theme"} {
		if !dcContainsText(view, name) {
			t.Errorf("expected tab bar to contain %q", name)
		}
	}
}

// dcContainsText checks if rendered output contains the given plain text.
func dcContainsText(rendered, text string) bool {
	for i := 0; i <= len(rendered)-len(text); i++ {
		if rendered[i:i+len(text)] == text {
			return true
		}
	}
	return false
}

// --- DebugProviderWithLifecycle tests ---

type mockLifecycleProvider struct {
	name      string
	inited    bool
	destroyed bool
}

func (m *mockLifecycleProvider) Name() string                      { return m.name }
func (m *mockLifecycleProvider) View(w, h int, theme Theme) string { return "mock" }
func (m *mockLifecycleProvider) Init() error                       { m.inited = true; return nil }
func (m *mockLifecycleProvider) Destroy() error                    { m.destroyed = true; return nil }

func TestDebugProviderWithLifecycle_Interface(t *testing.T) {
	var _ DebugProviderWithLifecycle = (*mockLifecycleProvider)(nil)
}

func TestDevConsole_InitProviders(t *testing.T) {
	dc := newDevConsole()
	lp := &mockLifecycleProvider{name: "lifecycle"}
	dc.providers = append(dc.providers, lp)

	dc.initProviders()

	if !lp.inited {
		t.Error("expected lifecycle provider Init to be called")
	}
}

func TestDevConsole_DestroyProviders(t *testing.T) {
	dc := newDevConsole()
	lp := &mockLifecycleProvider{name: "lifecycle"}
	dc.providers = append(dc.providers, lp)

	dc.destroyProviders()

	if !lp.destroyed {
		t.Error("expected lifecycle provider Destroy to be called")
	}
}

func TestDevConsole_ExportJSON(t *testing.T) {
	dc := newDevConsole()
	dc.active = true

	err := dc.exportJSON()
	if err != nil {
		t.Fatalf("exportJSON: %v", err)
	}
}

func TestDevConsole_ExportJSON_KeyBind(t *testing.T) {
	dc := newDevConsole()
	dc.active = true
	dc.SetTheme(DefaultTheme())
	dc.SetSize(120, 40)

	binds := dc.KeyBindings()
	found := false
	for _, b := range binds {
		if b.Key == "ctrl+e" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected ctrl+e keybind in DevConsole KeyBindings")
	}
}
