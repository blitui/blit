package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunGen_Basic(t *testing.T) {
	dir := t.TempDir()
	pkgDir := filepath.Join(dir, "mypkg")
	if err := os.Mkdir(pkgDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	code := runGen([]string{pkgDir})
	if code != 0 {
		t.Fatalf("runGen returned %d, want 0", code)
	}

	outFile := filepath.Join(pkgDir, "mypkg_test.go")
	data, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("output file not created: %v", err)
	}
	src := string(data)
	for _, want := range []string{
		"package mypkg_test",
		"btest.NewTestModel",
		"func TestMypkg(",
	} {
		if !strings.Contains(src, want) {
			t.Errorf("generated file missing %q", want)
		}
	}
}

func TestRunGen_CustomComponent(t *testing.T) {
	dir := t.TempDir()
	pkgDir := filepath.Join(dir, "mypkg")
	if err := os.Mkdir(pkgDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	code := runGen([]string{"--component", "Dashboard", pkgDir})
	if code != 0 {
		t.Fatalf("runGen returned %d, want 0", code)
	}

	outFile := filepath.Join(pkgDir, "dashboard_test.go")
	data, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("output file not created: %v", err)
	}
	if !strings.Contains(string(data), "func TestDashboard(") {
		t.Errorf("generated file missing TestDashboard function")
	}
}

func TestRunGen_RefusesOverwrite(t *testing.T) {
	dir := t.TempDir()
	pkgDir := filepath.Join(dir, "mypkg")
	if err := os.Mkdir(pkgDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	// Create the file first.
	existing := filepath.Join(pkgDir, "mypkg_test.go")
	if err := os.WriteFile(existing, []byte("// existing\n"), 0o644); err != nil {
		t.Fatalf("write existing: %v", err)
	}

	code := runGen([]string{pkgDir})
	if code == 0 {
		t.Fatal("runGen should have failed with exit 1 when file exists")
	}

	// File must be unchanged.
	data, _ := os.ReadFile(existing)
	if string(data) != "// existing\n" {
		t.Error("existing file was overwritten without --force")
	}
}

func TestRunGen_ForceOverwrite(t *testing.T) {
	dir := t.TempDir()
	pkgDir := filepath.Join(dir, "mypkg")
	if err := os.Mkdir(pkgDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	existing := filepath.Join(pkgDir, "mypkg_test.go")
	if err := os.WriteFile(existing, []byte("// existing\n"), 0o644); err != nil {
		t.Fatalf("write existing: %v", err)
	}

	code := runGen([]string{"--force", pkgDir})
	if code != 0 {
		t.Fatalf("runGen with --force returned %d, want 0", code)
	}

	data, err := os.ReadFile(existing)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	if string(data) == "// existing\n" {
		t.Error("file was not overwritten with --force")
	}
	if !strings.Contains(string(data), "btest.NewTestModel") {
		t.Error("overwritten file missing blit boilerplate")
	}
}

func TestRunGen_MissingArg(t *testing.T) {
	code := runGen([]string{})
	if code != 2 {
		t.Fatalf("runGen with no args returned %d, want 2", code)
	}
}
