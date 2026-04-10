package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScaffoldProject(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "myapp")

	opts := initOpts{
		ProjectName: "myapp",
		ModulePath:  "github.com/test/myapp",
		Template:    "starter",
		BinaryName:  "myapp",
	}

	if err := scaffoldProject(target, opts); err != nil {
		t.Fatalf("scaffoldProject: %v", err)
	}

	// Verify all expected files exist
	expected := []string{
		"go.mod",
		".gitignore",
		"Makefile",
		"cmd/myapp/main.go",
		"internal/myapp/app.go",
		"internal/myapp/app_test.go",
	}
	for _, f := range expected {
		path := filepath.Join(target, f)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected file %s to exist", f)
		}
	}

	// Verify go.mod contains module path
	gomod, err := os.ReadFile(filepath.Join(target, "go.mod"))
	if err != nil {
		t.Fatalf("read go.mod: %v", err)
	}
	if got := string(gomod); !contains(got, "github.com/test/myapp") {
		t.Errorf("go.mod should contain module path, got:\n%s", got)
	}

	// Verify main.go imports the internal package
	mainGo, err := os.ReadFile(filepath.Join(target, "cmd/myapp/main.go"))
	if err != nil {
		t.Fatalf("read main.go: %v", err)
	}
	if got := string(mainGo); !contains(got, "github.com/test/myapp/internal/myapp") {
		t.Errorf("main.go should import internal package, got:\n%s", got)
	}

	// Verify app.go contains project name
	appGo, err := os.ReadFile(filepath.Join(target, "internal/myapp/app.go"))
	if err != nil {
		t.Fatalf("read app.go: %v", err)
	}
	if got := string(appGo); !contains(got, "myapp") {
		t.Errorf("app.go should contain project name, got:\n%s", got)
	}

	// Verify app_test.go references the project
	appTest, err := os.ReadFile(filepath.Join(target, "internal/myapp/app_test.go"))
	if err != nil {
		t.Fatalf("read app_test.go: %v", err)
	}
	if got := string(appTest); !contains(got, "myapp") {
		t.Errorf("app_test.go should contain project name, got:\n%s", got)
	}
}

func TestScaffoldProject_Makefile(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "testproj")

	opts := initOpts{
		ProjectName: "testproj",
		ModulePath:  "github.com/test/testproj",
		Template:    "starter",
		BinaryName:  "testproj",
	}

	if err := scaffoldProject(target, opts); err != nil {
		t.Fatalf("scaffoldProject: %v", err)
	}

	makefile, err := os.ReadFile(filepath.Join(target, "Makefile"))
	if err != nil {
		t.Fatalf("read Makefile: %v", err)
	}
	got := string(makefile)
	if !contains(got, "testproj") {
		t.Error("Makefile should contain binary name")
	}
	if !contains(got, "CGO_ENABLED=0") {
		t.Error("Makefile should have CGO_ENABLED=0")
	}
}

func TestScaffoldProject_Gitignore(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "proj")

	opts := initOpts{
		ProjectName: "proj",
		ModulePath:  "github.com/test/proj",
		Template:    "starter",
		BinaryName:  "proj",
	}

	if err := scaffoldProject(target, opts); err != nil {
		t.Fatalf("scaffoldProject: %v", err)
	}

	gi, err := os.ReadFile(filepath.Join(target, ".gitignore"))
	if err != nil {
		t.Fatalf("read .gitignore: %v", err)
	}
	got := string(gi)
	if !contains(got, "coverage.out") {
		t.Error(".gitignore should include coverage.out")
	}
	if !contains(got, ".claude/") {
		t.Error(".gitignore should include .claude/")
	}
}

func TestScaffoldProject_Dashboard(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "dash")

	opts := initOpts{
		ProjectName: "dash",
		ModulePath:  "github.com/test/dash",
		Template:    "dashboard",
		BinaryName:  "dash",
	}

	if err := scaffoldProject(target, opts); err != nil {
		t.Fatalf("scaffoldProject: %v", err)
	}

	// Verify dashboard-specific files exist
	expected := []string{
		"cmd/dash/main.go",
		"internal/dash/app.go",
		"internal/dash/app_test.go",
	}
	for _, f := range expected {
		path := filepath.Join(target, f)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected file %s to exist", f)
		}
	}

	// Verify app.go contains dashboard-specific content
	appGo, err := os.ReadFile(filepath.Join(target, "internal/dash/app.go"))
	if err != nil {
		t.Fatalf("read app.go: %v", err)
	}
	got := string(appGo)
	if !contains(got, "SlotMain") {
		t.Error("dashboard app.go should use SlotMain")
	}
	if !contains(got, "SlotSidebar") {
		t.Error("dashboard app.go should use SlotSidebar")
	}
	if !contains(got, "NewTable") {
		t.Error("dashboard app.go should use NewTable")
	}
}

func TestScaffoldProject_Form(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "myform")

	opts := initOpts{
		ProjectName: "myform",
		ModulePath:  "github.com/test/myform",
		Template:    "form",
		BinaryName:  "myform",
	}

	if err := scaffoldProject(target, opts); err != nil {
		t.Fatalf("scaffoldProject: %v", err)
	}

	expected := []string{
		"cmd/myform/main.go",
		"internal/myform/app.go",
		"internal/myform/app_test.go",
	}
	for _, f := range expected {
		path := filepath.Join(target, f)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected file %s to exist", f)
		}
	}

	appGo, err := os.ReadFile(filepath.Join(target, "internal/myform/app.go"))
	if err != nil {
		t.Fatalf("read app.go: %v", err)
	}
	got := string(appGo)
	if !contains(got, "NewForm") {
		t.Error("form app.go should use NewForm")
	}
	if !contains(got, "NewTextField") {
		t.Error("form app.go should use NewTextField")
	}
	if !contains(got, "EmailValidator") {
		t.Error("form app.go should use EmailValidator")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStr(s, substr))
}

func containsStr(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
