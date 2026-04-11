package blit

import (
	"os"
	"path/filepath"
	"testing"
)

type testConfig struct {
	Name    string `yaml:"name"`
	Port    int    `yaml:"port"`
	Verbose bool   `yaml:"verbose"`
}

func TestSaveAndLoadYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	orig := testConfig{Name: "myapp", Port: 8080, Verbose: true}
	if err := SaveYAML(path, &orig); err != nil {
		t.Fatalf("SaveYAML: %v", err)
	}

	var loaded testConfig
	if err := LoadYAML(path, &loaded); err != nil {
		t.Fatalf("LoadYAML: %v", err)
	}
	if loaded != orig {
		t.Errorf("round-trip mismatch: got %+v, want %+v", loaded, orig)
	}
}

func TestLoadYAML_MissingFile(t *testing.T) {
	var cfg testConfig
	err := LoadYAML("/nonexistent/path/config.yaml", &cfg)
	if err != nil {
		t.Errorf("LoadYAML on missing file should return nil, got %v", err)
	}
	if cfg.Name != "" || cfg.Port != 0 {
		t.Error("cfg should remain zero-valued for missing file")
	}
}

func TestLoadYAML_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.yaml")
	os.WriteFile(path, []byte(":::invalid"), 0o644)

	var cfg testConfig
	err := LoadYAML(path, &cfg)
	if err == nil {
		t.Error("LoadYAML should return error for invalid YAML")
	}
}

func TestSaveYAML_CreatesDir(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "nested", "config.yaml")

	cfg := testConfig{Name: "test"}
	if err := SaveYAML(path, &cfg); err != nil {
		t.Fatalf("SaveYAML with nested dirs: %v", err)
	}

	if _, err := os.Stat(path); err != nil {
		t.Errorf("file should exist after SaveYAML: %v", err)
	}
}

func TestSaveYAML_Overwrite(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	first := testConfig{Name: "first", Port: 1}
	SaveYAML(path, &first)

	second := testConfig{Name: "second", Port: 2}
	SaveYAML(path, &second)

	var loaded testConfig
	LoadYAML(path, &loaded)
	if loaded.Name != "second" || loaded.Port != 2 {
		t.Errorf("overwrite failed: got %+v", loaded)
	}
}

func TestDefaultConfigPath(t *testing.T) {
	path, err := DefaultConfigPath("testapp")
	if err != nil {
		t.Fatalf("DefaultConfigPath: %v", err)
	}
	if filepath.Base(path) != "config.yaml" {
		t.Errorf("expected config.yaml, got %s", filepath.Base(path))
	}
	if !containsDir(path, "testapp") {
		t.Errorf("path should contain appName dir: %s", path)
	}
}

func TestEnsureConfigDir(t *testing.T) {
	dir, err := EnsureConfigDir("blit-test-ensure")
	if err != nil {
		t.Fatalf("EnsureConfigDir: %v", err)
	}
	if !containsDir(dir, "blit-test-ensure") {
		t.Errorf("dir should contain app name: %s", dir)
	}
	// Clean up
	os.Remove(dir)
}

func containsDir(path, name string) bool {
	for path != "" {
		dir := filepath.Base(path)
		if dir == name {
			return true
		}
		parent := filepath.Dir(path)
		if parent == path {
			break
		}
		path = parent
	}
	return false
}
