package blit

import (
	"os"
	"path/filepath"
	"testing"
)

type testModuleConfig struct {
	Interval int      `yaml:"interval" blit:"label=Poll interval,group=Polling,default=30,min=5,max=300"`
	Theme    string   `yaml:"theme"    blit:"label=Theme,group=Appearance,default=dark,options=dark|light|auto"`
	Repos    []string `yaml:"repos"    blit:"label=Repos,group=Data,hint=owner/repo format"`
	Debug    bool     `yaml:"debug"    blit:"label=Debug mode,group=Advanced,default=false"`
	Secret   string   `yaml:"secret"`  // No blit tag — should be skipped by Editor
}

func TestParseBlitTag(t *testing.T) {
	tests := []struct {
		name string
		tag  string
		want fieldMeta
	}{
		{
			name: "empty tag",
			tag:  "",
			want: fieldMeta{},
		},
		{
			name: "label only",
			tag:  "label=Foo",
			want: fieldMeta{Label: "Foo"},
		},
		{
			name: "all fields",
			tag:  "label=Poll interval,group=Polling,hint=seconds,default=30,min=5,max=300",
			want: fieldMeta{
				Label:   "Poll interval",
				Group:   "Polling",
				Hint:    "seconds",
				Default: "30",
				Min:     5,
				Max:     300,
			},
		},
		{
			name: "options with pipe",
			tag:  "label=Theme,options=dark|light|auto",
			want: fieldMeta{
				Label:   "Theme",
				Options: []string{"dark", "light", "auto"},
			},
		},
		{
			name: "readonly",
			tag:  "label=Version,readonly",
			want: fieldMeta{Label: "Version", ReadOnly: true},
		},
		{
			name: "group and hint",
			tag:  "group=Advanced,hint=enable verbose logging",
			want: fieldMeta{Group: "Advanced", Hint: "enable verbose logging"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseBlitTag(tt.tag)
			if got.Label != tt.want.Label {
				t.Errorf("Label = %q, want %q", got.Label, tt.want.Label)
			}
			if got.Group != tt.want.Group {
				t.Errorf("Group = %q, want %q", got.Group, tt.want.Group)
			}
			if got.Hint != tt.want.Hint {
				t.Errorf("Hint = %q, want %q", got.Hint, tt.want.Hint)
			}
			if got.Default != tt.want.Default {
				t.Errorf("Default = %q, want %q", got.Default, tt.want.Default)
			}
			if got.Min != tt.want.Min {
				t.Errorf("Min = %d, want %d", got.Min, tt.want.Min)
			}
			if got.Max != tt.want.Max {
				t.Errorf("Max = %d, want %d", got.Max, tt.want.Max)
			}
			if got.ReadOnly != tt.want.ReadOnly {
				t.Errorf("ReadOnly = %v, want %v", got.ReadOnly, tt.want.ReadOnly)
			}
			if len(tt.want.Options) > 0 {
				if len(got.Options) != len(tt.want.Options) {
					t.Fatalf("Options len = %d, want %d", len(got.Options), len(tt.want.Options))
				}
				for i, o := range tt.want.Options {
					if got.Options[i] != o {
						t.Errorf("Options[%d] = %q, want %q", i, got.Options[i], o)
					}
				}
			}
		})
	}
}

func TestLoadConfig_Defaults(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	// No file exists — defaults should apply
	cfg, err := LoadConfig[testModuleConfig]("test", WithConfigPath(path))
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if cfg.Value.Interval != 30 {
		t.Errorf("Interval = %d, want 30", cfg.Value.Interval)
	}
	if cfg.Value.Theme != "dark" {
		t.Errorf("Theme = %q, want %q", cfg.Value.Theme, "dark")
	}
	if cfg.Value.Debug != false {
		t.Errorf("Debug = %v, want false", cfg.Value.Debug)
	}
}

func TestLoadConfig_FromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	content := "interval: 60\ntheme: light\nrepos:\n  - owner/repo1\n  - owner/repo2\ndebug: true\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	cfg, err := LoadConfig[testModuleConfig]("test", WithConfigPath(path))
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if cfg.Value.Interval != 60 {
		t.Errorf("Interval = %d, want 60", cfg.Value.Interval)
	}
	if cfg.Value.Theme != "light" {
		t.Errorf("Theme = %q, want %q", cfg.Value.Theme, "light")
	}
	if len(cfg.Value.Repos) != 2 {
		t.Fatalf("Repos len = %d, want 2", len(cfg.Value.Repos))
	}
	if cfg.Value.Repos[0] != "owner/repo1" {
		t.Errorf("Repos[0] = %q, want %q", cfg.Value.Repos[0], "owner/repo1")
	}
	if cfg.Value.Debug != true {
		t.Errorf("Debug = %v, want true", cfg.Value.Debug)
	}
}

func TestLoadConfig_CustomPath(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "custom", "myconfig.yaml")

	cfg, err := LoadConfig[testModuleConfig]("test", WithConfigPath(path))
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	if cfg.Path() != path {
		t.Errorf("Path = %q, want %q", cfg.Path(), path)
	}
}

func TestConfig_Save_Roundtrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	cfg, err := LoadConfig[testModuleConfig]("test", WithConfigPath(path))
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}
	cfg.Value.Interval = 120
	cfg.Value.Theme = "light"
	cfg.Value.Repos = []string{"a/b", "c/d"}
	cfg.Value.Debug = true

	if err := cfg.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}

	cfg2, err := LoadConfig[testModuleConfig]("test", WithConfigPath(path))
	if err != nil {
		t.Fatalf("LoadConfig round-trip: %v", err)
	}
	if cfg2.Value.Interval != 120 {
		t.Errorf("Interval = %d, want 120", cfg2.Value.Interval)
	}
	if cfg2.Value.Theme != "light" {
		t.Errorf("Theme = %q, want %q", cfg2.Value.Theme, "light")
	}
	if len(cfg2.Value.Repos) != 2 {
		t.Fatalf("Repos len = %d, want 2", len(cfg2.Value.Repos))
	}
	if cfg2.Value.Debug != true {
		t.Errorf("Debug = %v, want true", cfg2.Value.Debug)
	}
}

func TestConfig_SetValue(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	cfg, err := LoadConfig[testModuleConfig]("test", WithConfigPath(path))
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}

	if cfg.IsDirty() {
		t.Error("expected clean config after load")
	}

	if err := cfg.SetValue(func(v *testModuleConfig) {
		v.Interval = 99
	}); err != nil {
		t.Fatalf("SetValue: %v", err)
	}

	if !cfg.IsDirty() {
		t.Error("expected dirty config after SetValue")
	}
	if cfg.Value.Interval != 99 {
		t.Errorf("Interval = %d, want 99", cfg.Value.Interval)
	}
}

func TestConfig_Editor(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	cfg, err := LoadConfig[testModuleConfig]("test", WithConfigPath(path))
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}

	editor := cfg.Editor()
	if editor == nil {
		t.Fatal("Editor returned nil")
	}

	// testConfig has 4 tagged fields (Secret has no blit tag)
	if len(editor.fields) != 4 {
		t.Errorf("editor field count = %d, want 4", len(editor.fields))
	}
}

func TestConfig_Editor_GetSet(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	cfg, err := LoadConfig[testModuleConfig]("test", WithConfigPath(path))
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}

	editor := cfg.Editor()

	tests := []struct {
		label    string
		initVal  string
		setVal   string
		checkVal string
	}{
		{"Poll interval", "30", "60", "60"},
		{"Theme", "dark", "light", "light"},
		{"Repos", "", "a/b,c/d", "a/b,c/d"},
		{"Debug mode", "false", "true", "true"},
	}

	for _, tt := range tests {
		t.Run(tt.label, func(t *testing.T) {
			var field *ConfigField
			for i := range editor.fields {
				if editor.fields[i].Label == tt.label {
					field = &editor.fields[i]
					break
				}
			}
			if field == nil {
				t.Fatalf("field %q not found", tt.label)
			}

			got := field.Get()
			if got != tt.initVal {
				t.Errorf("Get() = %q, want %q", got, tt.initVal)
			}

			if field.Set != nil {
				if err := field.Set(tt.setVal); err != nil {
					t.Fatalf("Set(%q): %v", tt.setVal, err)
				}
				got = field.Get()
				if got != tt.checkVal {
					t.Errorf("after Set, Get() = %q, want %q", got, tt.checkVal)
				}
			}
		})
	}
}

func TestConfig_Editor_Validate(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	cfg, err := LoadConfig[testModuleConfig]("test", WithConfigPath(path))
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}

	editor := cfg.Editor()

	// Find the Interval field (has min=5, max=300)
	var intervalField *ConfigField
	for i := range editor.fields {
		if editor.fields[i].Label == "Poll interval" {
			intervalField = &editor.fields[i]
			break
		}
	}
	if intervalField == nil {
		t.Fatal("interval field not found")
	}

	// Below min
	if err := intervalField.Validate("3"); err == nil {
		t.Error("expected error for value below min")
	}
	// Above max
	if err := intervalField.Validate("500"); err == nil {
		t.Error("expected error for value above max")
	}
	// Valid
	if err := intervalField.Validate("60"); err != nil {
		t.Errorf("unexpected error for valid value: %v", err)
	}
	// Not an integer
	if err := intervalField.Validate("abc"); err == nil {
		t.Error("expected error for non-integer")
	}

	// Find Theme field (has options)
	var themeField *ConfigField
	for i := range editor.fields {
		if editor.fields[i].Label == "Theme" {
			themeField = &editor.fields[i]
			break
		}
	}
	if themeField == nil {
		t.Fatal("theme field not found")
	}

	if err := themeField.Validate("dark"); err != nil {
		t.Errorf("unexpected error for valid option: %v", err)
	}
	if err := themeField.Validate("neon"); err == nil {
		t.Error("expected error for invalid option")
	}
}

func TestConfig_Reset(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	cfg, err := LoadConfig[testModuleConfig]("test", WithConfigPath(path))
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}

	cfg.Value.Interval = 999
	cfg.Value.Theme = "neon"

	if err := cfg.Reset(); err != nil {
		t.Fatalf("Reset: %v", err)
	}

	if cfg.Value.Interval != 30 {
		t.Errorf("Interval = %d, want 30", cfg.Value.Interval)
	}
	if cfg.Value.Theme != "dark" {
		t.Errorf("Theme = %q, want %q", cfg.Value.Theme, "dark")
	}
	if cfg.IsDirty() {
		t.Error("expected clean config after Reset+Save")
	}

	// Verify persisted
	cfg2, err := LoadConfig[testModuleConfig]("test", WithConfigPath(path))
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	if cfg2.Value.Interval != 30 {
		t.Errorf("reloaded Interval = %d, want 30", cfg2.Value.Interval)
	}
}

func TestConfig_Defaults(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	cfg, err := LoadConfig[testModuleConfig]("test", WithConfigPath(path))
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}

	cfg.Value.Interval = 999
	d := cfg.Defaults()

	if d.Interval != 30 {
		t.Errorf("Defaults().Interval = %d, want 30", d.Interval)
	}
	if d.Theme != "dark" {
		t.Errorf("Defaults().Theme = %q, want %q", d.Theme, "dark")
	}
	// Original should be unchanged
	if cfg.Value.Interval != 999 {
		t.Errorf("original Interval changed to %d", cfg.Value.Interval)
	}
}
