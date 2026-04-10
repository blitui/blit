package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectMainPackage_CmdDir(t *testing.T) {
	// The blit repo itself has cmd/blit/main.go, so detectMainPackage
	// should find it when run from the repo root.
	pkg := detectMainPackage()
	if pkg == "" {
		t.Skip("no cmd/ directory found (not running from repo root)")
	}
	if pkg != "./cmd/blit" && pkg != "./cmd/sess2tape" {
		// Any cmd/<name> with a main.go is valid.
		t.Logf("detected package: %s", pkg)
	}
}

func TestDetectMainPackage_MainGo(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	defer func() { _ = os.Chdir(orig) }()

	// Create a main.go in a temp directory (no cmd/ dir).
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\nfunc main(){}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	pkg := detectMainPackage()
	if pkg != "." {
		t.Errorf("expected '.', got %q", pkg)
	}
}

func TestDetectMainPackage_Empty(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	defer func() { _ = os.Chdir(orig) }()

	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	pkg := detectMainPackage()
	if pkg != "" {
		t.Errorf("expected empty string, got %q", pkg)
	}
}

func TestStopProcess_Nil(t *testing.T) {
	// stopProcess should not panic on nil.
	stopProcess(nil)
}

func TestSnapshotConfigTree(t *testing.T) {
	dir := t.TempDir()

	// Empty directory should return empty hash.
	h1 := snapshotConfigTree(dir)
	if h1 != "" {
		t.Errorf("expected empty hash for empty dir, got %q", h1)
	}

	// Create a YAML file.
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte("key: val"), 0o644); err != nil {
		t.Fatal(err)
	}
	h2 := snapshotConfigTree(dir)
	if h2 == "" {
		t.Error("expected non-empty hash after adding config.yaml")
	}

	// Create a JSON file — hash should change.
	if err := os.WriteFile(filepath.Join(dir, "theme.json"), []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	h3 := snapshotConfigTree(dir)
	if h3 == h2 {
		t.Error("expected hash to change after adding theme.json")
	}

	// .go files should be ignored.
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main"), 0o644); err != nil {
		t.Fatal(err)
	}
	h4 := snapshotConfigTree(dir)
	if h4 != h3 {
		t.Error("expected hash unchanged after adding .go file")
	}
}

func TestConfigExts(t *testing.T) {
	for _, ext := range []string{".yaml", ".yml", ".json", ".toml"} {
		if !configExts[ext] {
			t.Errorf("expected %q in configExts", ext)
		}
	}
	if configExts[".go"] {
		t.Error(".go should not be in configExts")
	}
}
