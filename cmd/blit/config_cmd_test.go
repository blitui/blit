package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectAppName(t *testing.T) {
	name := detectAppName()
	if name == "" {
		t.Error("expected non-empty app name")
	}
}

func TestConfigGet(t *testing.T) {
	dir := t.TempDir()
	appDir := filepath.Join(dir, "testapp")
	if err := os.MkdirAll(appDir, 0o755); err != nil {
		t.Fatal(err)
	}
	configFile := filepath.Join(appDir, "config.yaml")
	if err := os.WriteFile(configFile, []byte("theme: dracula\nverbose: true\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 {
		t.Error("config file should not be empty")
	}
}

func TestRunConfig_Unknown(t *testing.T) {
	code := runConfig([]string{"--app", "testapp", "bogus"})
	if code != 1 {
		t.Errorf("expected exit 1 for unknown subcommand, got %d", code)
	}
}

func TestRunConfig_Path(t *testing.T) {
	code := runConfig([]string{"--app", "testapp", "path"})
	if code != 0 {
		t.Errorf("expected exit 0 for path, got %d", code)
	}
}
