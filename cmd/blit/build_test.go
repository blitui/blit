package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGoreleaserInstalled(t *testing.T) {
	// Just verify the function doesn't panic.
	_ = goreleaserInstalled()
}

func TestHasGoreleaserConfig(t *testing.T) {
	// From the repo root (where .goreleaser.yaml exists), this should be true
	// when run via go test from the project root. But since tests run in
	// cmd/blit/, create a temp config to verify.
	dir := t.TempDir()
	orig, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Skip("cannot chdir")
	}
	defer func() { _ = os.Chdir(orig) }()

	if hasGoreleaserConfig() {
		t.Error("expected false in empty dir")
	}

	if err := os.WriteFile(filepath.Join(dir, ".goreleaser.yaml"), []byte("version: 2\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if !hasGoreleaserConfig() {
		t.Error("expected true with .goreleaser.yaml present")
	}
}
