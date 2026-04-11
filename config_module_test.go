package blit

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

type testModuleConfig struct {
	Interval int      `yaml:"interval" blit:"label=Poll interval,group=Polling,default=30,min=5,max=300"`
	Theme    string   `yaml:"theme"    blit:"label=Theme,group=Appearance,default=dark,options=dark|light|auto"`
	Repos    []string `yaml:"repos"    blit:"label=Repos,group=Data,hint=owner/repo format"`
	Debug    bool     `yaml:"debug"    blit:"label=Debug mode,group=Advanced,default=false"`
	Secret   string   `yaml:"secret"` // No blit tag — should be skipped by Editor
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

func TestCamelToUpperSnake(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"Interval", "INTERVAL"},
		{"Theme", "THEME"},
		{"MaxRetries", "MAX_RETRIES"},
		{"appName", "APP_NAME"},
		{"myapp", "MYAPP"},
		{"LogLevel", "LOG_LEVEL"},
		{"X", "X"},
		{"", ""},
	}
	for _, tt := range tests {
		got := camelToUpperSnake(tt.in)
		if got != tt.want {
			t.Errorf("camelToUpperSnake(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestLoadConfig_EnvOverrides(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	// Write a config with interval=30
	if err := os.WriteFile(path, []byte("interval: 30\ntheme: dark\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Set env var to override interval
	t.Setenv("TEST_INTERVAL", "120")

	cfg, err := LoadConfig[testModuleConfig]("test", WithConfigPath(path))
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}

	if cfg.Value.Interval != 120 {
		t.Errorf("Interval = %d, want 120 (from env)", cfg.Value.Interval)
	}
	// Theme should remain from file since no env var set.
	if cfg.Value.Theme != "dark" {
		t.Errorf("Theme = %q, want %q", cfg.Value.Theme, "dark")
	}
}

func TestLoadConfig_EnvOverridesDefault(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	// No file — defaults apply, then env override on top.
	t.Setenv("TEST_THEME", "light")

	cfg, err := LoadConfig[testModuleConfig]("test", WithConfigPath(path))
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}

	// Interval should have default 30 (no env override).
	if cfg.Value.Interval != 30 {
		t.Errorf("Interval = %d, want 30 (default)", cfg.Value.Interval)
	}
	// Theme should be overridden by env var.
	if cfg.Value.Theme != "light" {
		t.Errorf("Theme = %q, want %q (from env)", cfg.Value.Theme, "light")
	}
}

func TestLoadConfig_EnvOverridesBool(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	t.Setenv("TEST_DEBUG", "true")

	cfg, err := LoadConfig[testModuleConfig]("test", WithConfigPath(path))
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}

	if !cfg.Value.Debug {
		t.Error("Debug = false, want true (from env)")
	}
}

func TestConfig_CLICommands_Get(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	cfg, err := LoadConfig[testModuleConfig]("test", WithConfigPath(path))
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}

	cmds := cfg.CLICommands()

	// get should succeed for known fields.
	if err := cmds["get"]([]string{"interval"}); err != nil {
		t.Errorf("get interval: %v", err)
	}

	// get should fail for unknown fields.
	if err := cmds["get"]([]string{"nonexistent"}); err == nil {
		t.Error("get nonexistent: expected error")
	}

	// get should fail with no args.
	if err := cmds["get"](nil); err == nil {
		t.Error("get with no args: expected error")
	}
}

func TestConfig_CLICommands_Set(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	cfg, err := LoadConfig[testModuleConfig]("test", WithConfigPath(path))
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}

	cmds := cfg.CLICommands()

	// set should update the value and save.
	if err := cmds["set"]([]string{"interval", "99"}); err != nil {
		t.Fatalf("set interval: %v", err)
	}
	if cfg.Value.Interval != 99 {
		t.Errorf("Interval = %d, want 99", cfg.Value.Interval)
	}

	// Verify persisted by reloading.
	cfg2, err := LoadConfig[testModuleConfig]("test", WithConfigPath(path))
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	if cfg2.Value.Interval != 99 {
		t.Errorf("reloaded Interval = %d, want 99", cfg2.Value.Interval)
	}

	// set should fail for unknown fields.
	if err := cmds["set"]([]string{"nonexistent", "val"}); err == nil {
		t.Error("set nonexistent: expected error")
	}

	// set should fail with insufficient args.
	if err := cmds["set"]([]string{"interval"}); err == nil {
		t.Error("set with 1 arg: expected error")
	}
}

func TestConfig_CLICommands_List(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	cfg, err := LoadConfig[testModuleConfig]("test", WithConfigPath(path))
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}

	cmds := cfg.CLICommands()

	// list should not return an error.
	if err := cmds["list"](nil); err != nil {
		t.Errorf("list: %v", err)
	}
}

func TestYamlFieldName(t *testing.T) {
	rt := reflect.TypeOf(testModuleConfig{})

	// Interval has yaml:"interval" tag.
	f, _ := rt.FieldByName("Interval")
	if got := yamlFieldName(f); got != "interval" {
		t.Errorf("yamlFieldName(Interval) = %q, want %q", got, "interval")
	}

	// Secret has yaml:"secret" tag.
	f, _ = rt.FieldByName("Secret")
	if got := yamlFieldName(f); got != "secret" {
		t.Errorf("yamlFieldName(Secret) = %q, want %q", got, "secret")
	}
}

func TestConfig_AsSignal(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	cfg, err := LoadConfig[testModuleConfig]("test", WithConfigPath(path))
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}

	sig := cfg.AsSignal()
	if sig == nil {
		t.Fatal("AsSignal returned nil")
	}

	// Should return the same signal on subsequent calls.
	sig2 := cfg.AsSignal()
	if sig != sig2 {
		t.Error("AsSignal returned different signal on second call")
	}

	// Signal should reflect current config value.
	if sig.Get().Interval != 30 {
		t.Errorf("signal Interval = %d, want 30", sig.Get().Interval)
	}
	if sig.Get().Theme != "dark" {
		t.Errorf("signal Theme = %q, want %q", sig.Get().Theme, "dark")
	}
}

func TestConfig_AsSignal_SetValueEmits(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	cfg, err := LoadConfig[testModuleConfig]("test", WithConfigPath(path))
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}

	sig := cfg.AsSignal()

	if err := cfg.SetValue(func(v *testModuleConfig) {
		v.Interval = 77
		v.Theme = "light"
	}); err != nil {
		t.Fatalf("SetValue: %v", err)
	}

	// Signal should have the updated value.
	if sig.Get().Interval != 77 {
		t.Errorf("signal Interval = %d, want 77", sig.Get().Interval)
	}
	if sig.Get().Theme != "light" {
		t.Errorf("signal Theme = %q, want %q", sig.Get().Theme, "light")
	}
}

func TestConfig_AsSignal_WatchFileEmits(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	if err := os.WriteFile(path, []byte("interval: 30\ntheme: dark\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := &Config[testModuleConfig]{
		path:    path,
		appName: "test",
		Value:   testModuleConfig{Interval: 30, Theme: "dark"},
	}

	sig := cfg.AsSignal()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		_ = cfg.WatchFile(ctx, nil)
	}()

	// Give the watcher time to start.
	time.Sleep(100 * time.Millisecond)

	// Modify the file.
	if err := os.WriteFile(path, []byte("interval: 88\ntheme: light\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Poll for the signal to update (debounce is 200ms).
	deadline := time.After(3 * time.Second)
	for sig.Get().Interval != 88 {
		select {
		case <-deadline:
			t.Fatalf("timed out: signal Interval = %d, want 88", sig.Get().Interval)
		case <-time.After(50 * time.Millisecond):
		}
	}

	if sig.Get().Theme != "light" {
		t.Errorf("signal Theme = %q, want %q", sig.Get().Theme, "light")
	}

	cancel()
}

func TestConfig_WatchFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	// Write initial config.
	initial := "interval: 30\ntheme: dark\n"
	if err := os.WriteFile(path, []byte(initial), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := &Config[testModuleConfig]{
		path:    path,
		appName: "test",
		Value:   testModuleConfig{Interval: 30, Theme: "dark"},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	changed := make(chan testModuleConfig, 1)
	errCh := make(chan error, 1)

	go func() {
		errCh <- cfg.WatchFile(ctx, func(v testModuleConfig) {
			select {
			case changed <- v:
			default:
			}
		})
	}()

	// Give the watcher time to start.
	time.Sleep(100 * time.Millisecond)

	// Modify the file.
	updated := "interval: 60\ntheme: light\n"
	if err := os.WriteFile(path, []byte(updated), 0o644); err != nil {
		t.Fatal(err)
	}

	// Wait for the onChange callback (debounce is 200ms).
	select {
	case v := <-changed:
		if v.Interval != 60 {
			t.Errorf("Interval = %d, want 60", v.Interval)
		}
		if v.Theme != "light" {
			t.Errorf("Theme = %q, want %q", v.Theme, "light")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for config change")
	}

	// Verify cfg.Value was updated.
	if cfg.Value.Interval != 60 {
		t.Errorf("cfg.Value.Interval = %d, want 60", cfg.Value.Interval)
	}

	// Cancel and verify clean shutdown.
	cancel()
	select {
	case err := <-errCh:
		if err != nil {
			t.Errorf("WatchFile returned error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("WatchFile did not return after cancel")
	}
}

func TestConfig_WatchFile_NoPath(t *testing.T) {
	cfg := &Config[testModuleConfig]{}

	err := cfg.WatchFile(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for empty path")
	}
}

func TestConfig_WatchFile_AppliesDefaults(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	// Write config with only interval (no theme).
	if err := os.WriteFile(path, []byte("interval: 99\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg := &Config[testModuleConfig]{
		path:    path,
		appName: "test",
		Value:   testModuleConfig{Interval: 1, Theme: "old"},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	changed := make(chan testModuleConfig, 1)

	go func() {
		_ = cfg.WatchFile(ctx, func(v testModuleConfig) {
			select {
			case changed <- v:
			default:
			}
		})
	}()

	time.Sleep(100 * time.Millisecond)

	// Rewrite with only interval — theme should get default "dark".
	if err := os.WriteFile(path, []byte("interval: 42\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	select {
	case v := <-changed:
		if v.Interval != 42 {
			t.Errorf("Interval = %d, want 42", v.Interval)
		}
		if v.Theme != "dark" {
			t.Errorf("Theme = %q, want default %q", v.Theme, "dark")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for config change")
	}

	cancel()
}
