package scaffold_test

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/blitui/blit/internal/scaffold"
)

func TestGenerate_DefaultComponent(t *testing.T) {
	result, err := scaffold.Generate(scaffold.Options{PkgPath: "./mypkg"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	src := string(result.Content)

	for _, want := range []string{
		"package mypkg_test",
		"btest.NewTestModel",
		"func TestMypkg(",
		"stubMypkg",
	} {
		if !strings.Contains(src, want) {
			t.Errorf("generated source missing %q\ngot:\n%s", want, src)
		}
	}

	if result.FileName != "mypkg_test.go" {
		t.Errorf("FileName = %q, want %q", result.FileName, "mypkg_test.go")
	}
}

func TestGenerate_CustomComponent(t *testing.T) {
	result, err := scaffold.Generate(scaffold.Options{
		PkgPath:   "./mypkg",
		Component: "Dashboard",
	})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	src := string(result.Content)

	for _, want := range []string{
		"func TestDashboard(",
		"stubDashboard",
		"btest.NewTestModel",
	} {
		if !strings.Contains(src, want) {
			t.Errorf("generated source missing %q\ngot:\n%s", want, src)
		}
	}

	if result.FileName != "dashboard_test.go" {
		t.Errorf("FileName = %q, want %q", result.FileName, "dashboard_test.go")
	}
}

func TestGenerate_GoVet(t *testing.T) {
	result, err := scaffold.Generate(scaffold.Options{
		PkgPath:   "./mypkg",
		Component: "Widget",
	})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	fset := token.NewFileSet()
	_, parseErr := parser.ParseFile(fset, "widget_test.go", result.Content, parser.AllErrors)
	if parseErr != nil {
		t.Fatalf("generated source does not parse: %v\nsource:\n%s", parseErr, result.Content)
	}
}

func TestGenerate_GoldenFixture(t *testing.T) {
	goldenPath := filepath.Join("testdata", "mypkg_test.go.golden")

	result, err := scaffold.Generate(scaffold.Options{PkgPath: "./mypkg"})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	if os.Getenv("UPDATE_GOLDEN") != "" {
		if err := os.MkdirAll(filepath.Dir(goldenPath), 0o755); err != nil {
			t.Fatalf("mkdir testdata: %v", err)
		}
		if err := os.WriteFile(goldenPath, result.Content, 0o644); err != nil {
			t.Fatalf("write golden: %v", err)
		}
		t.Logf("golden file updated: %s", goldenPath)
		return
	}

	expected, err := os.ReadFile(goldenPath)
	if err != nil {
		if os.IsNotExist(err) {
			// First run: create the golden file.
			if mkErr := os.MkdirAll(filepath.Dir(goldenPath), 0o755); mkErr != nil {
				t.Fatalf("mkdir testdata: %v", mkErr)
			}
			if wErr := os.WriteFile(goldenPath, result.Content, 0o644); wErr != nil {
				t.Fatalf("write golden: %v", wErr)
			}
			t.Logf("golden file created at %s; re-run to verify", goldenPath)
			return
		}
		t.Fatalf("read golden: %v", err)
	}

	if string(expected) != string(result.Content) {
		t.Errorf("generator output does not match golden file %s\nwant:\n%s\ngot:\n%s",
			goldenPath, string(expected), string(result.Content))
	}
}
