package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/blitui/blit/cli"
)

// templateInfo describes a project template.
type templateInfo struct {
	Name        string
	Description string
}

var projectTemplates = []templateInfo{
	{Name: "starter", Description: "Minimal app — list view, status bar, help overlay"},
	{Name: "dashboard", Description: "Dashboard — table, sidebar, split pane layout"},
	{Name: "form", Description: "Form app — validated inputs, submission flow"},
}

// initOpts holds the collected wizard answers.
type initOpts struct {
	ProjectName string
	ModulePath  string
	Template    string
	BinaryName  string // derived from ProjectName
}

// runInit implements the `blit init` subcommand.
func runInit(args []string) int {
	fmt.Println("blit init — create a new blit TUI project")
	fmt.Println()

	// 1. Project name
	name, err := cli.Input("Project name:", func(v string) error {
		if v == "" {
			return fmt.Errorf("project name cannot be empty")
		}
		if strings.ContainsAny(v, " \t/\\") {
			return fmt.Errorf("project name cannot contain spaces or slashes")
		}
		return nil
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cancelled.\n")
		return 1
	}

	// 2. Module path
	defaultMod := "github.com/" + name
	modPrompt := fmt.Sprintf("Go module path [%s]:", defaultMod)
	modPath, err := cli.Input(modPrompt, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cancelled.\n")
		return 1
	}
	if modPath == "" {
		modPath = defaultMod
	}

	// 3. Template
	templateNames := make([]string, len(projectTemplates))
	for i, t := range projectTemplates {
		templateNames[i] = fmt.Sprintf("%s — %s", t.Name, t.Description)
	}
	_, idx, err := cli.SelectOne("Choose a template:", templateNames)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cancelled.\n")
		return 1
	}
	tmplName := projectTemplates[idx].Name

	opts := initOpts{
		ProjectName: name,
		ModulePath:  modPath,
		Template:    tmplName,
		BinaryName:  name,
	}

	fmt.Printf("\nScaffolding %q with template %q...\n\n", name, tmplName)

	// Create directory structure
	dir := name
	if len(args) > 0 {
		dir = args[0]
	}

	if err := scaffoldProject(dir, opts); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	// Run go mod tidy
	fmt.Println("Running go mod tidy...")
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: go mod tidy failed: %v\n", err)
	}

	// Print next steps
	fmt.Printf(`
Done! Your project is ready at ./%s

Next steps:

  cd %s
  go run ./cmd/%s

Useful commands:

  go test ./...          Run tests
  go build ./cmd/%s   Build binary
  blit -watch            Watch mode testing

`, dir, dir, opts.BinaryName, opts.BinaryName)

	return 0
}

// scaffoldProject creates the project directory and writes all template files.
func scaffoldProject(dir string, opts initOpts) error {
	files := map[string]string{
		"go.mod":    goModTmpl,
		".gitignore": gitignoreTmpl,
		"Makefile":  makefileTmpl,
	}

	// Add template-specific files.
	switch opts.Template {
	case "dashboard":
		files["cmd/"+opts.BinaryName+"/main.go"] = dashboardMainTmpl
		files["internal/"+opts.BinaryName+"/app.go"] = dashboardAppTmpl
		files["internal/"+opts.BinaryName+"/app_test.go"] = dashboardAppTestTmpl
	case "form":
		files["cmd/"+opts.BinaryName+"/main.go"] = formMainTmpl
		files["internal/"+opts.BinaryName+"/app.go"] = formAppTmpl
		files["internal/"+opts.BinaryName+"/app_test.go"] = formAppTestTmpl
	default: // "starter"
		files["cmd/"+opts.BinaryName+"/main.go"] = starterMainTmpl
		files["internal/"+opts.BinaryName+"/app.go"] = starterAppTmpl
		files["internal/"+opts.BinaryName+"/app_test.go"] = starterAppTestTmpl
	}

	for relPath, tmplStr := range files {
		fullPath := filepath.Join(dir, relPath)

		// Create parent directories
		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			return fmt.Errorf("mkdir %s: %w", filepath.Dir(fullPath), err)
		}

		tmpl, err := template.New(relPath).Parse(tmplStr)
		if err != nil {
			return fmt.Errorf("parse template %s: %w", relPath, err)
		}

		f, err := os.Create(fullPath)
		if err != nil {
			return fmt.Errorf("create %s: %w", fullPath, err)
		}

		if err := tmpl.Execute(f, opts); err != nil {
			_ = f.Close()
			return fmt.Errorf("execute template %s: %w", relPath, err)
		}
		if err := f.Close(); err != nil {
			return fmt.Errorf("close %s: %w", fullPath, err)
		}

		fmt.Printf("  created %s\n", relPath)
	}

	return nil
}

// --- Templates ---

const goModTmpl = `module {{.ModulePath}}

go 1.25.0

require github.com/blitui/blit v0.2.22
`

const gitignoreTmpl = `# Binaries
{{.BinaryName}}
*.exe
*.exe~
*.dll
*.so
*.dylib

# Build output
/bin/
/dist/

# Test output
coverage.out
*.test

# IDE
.idea/
.vscode/
*.swp
*.swo
*~

# OS
.DS_Store
Thumbs.db

# Claude
.claude/
`

const makefileTmpl = `.PHONY: build test dev clean

BINARY = {{.BinaryName}}

build:
	CGO_ENABLED=0 go build -o bin/$(BINARY) ./cmd/$(BINARY)

test:
	go test ./...

dev: build
	./bin/$(BINARY)

clean:
	rm -rf bin/
`

const starterMainTmpl = `package main

import (
	"fmt"
	"os"

	app "{{.ModulePath}}/internal/{{.BinaryName}}"
)

func main() {
	a := app.New()
	if err := a.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
`

const starterAppTmpl = `package {{.BinaryName}}

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	blit "github.com/blitui/blit"
)

// New creates and configures the application.
func New() *blit.App {
	items := []string{
		"Hello, world!",
		"Welcome to {{.ProjectName}}",
		"Built with blit",
	}

	list := blit.NewListView(blit.ListViewOpts[string]{
		RenderItem: func(item string, idx int, isCursor bool, theme blit.Theme) string {
			style := lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Text))
			if isCursor {
				style = style.Foreground(lipgloss.Color(theme.Accent)).Bold(true)
			}
			return style.Render(fmt.Sprintf("  %s", item))
		},
		HeaderFunc: func(theme blit.Theme) string {
			return lipgloss.NewStyle().
				Foreground(lipgloss.Color(theme.Accent)).
				Bold(true).
				Render("  {{.ProjectName}}")
		},
	})
	list.SetItems(items)

	return blit.NewApp(
		blit.WithTheme(blit.DefaultTheme()),
		blit.WithComponent("main", list),
		blit.WithStatusBar(
			func() string { return " q quit  ? help" },
			func() string { return fmt.Sprintf(" %d items ", len(items)) },
		),
		blit.WithHelp(),
	)
}
`

const starterAppTestTmpl = `package {{.BinaryName}}_test

import (
	"testing"

	"github.com/blitui/blit/btest"

	app "{{.ModulePath}}/internal/{{.BinaryName}}"
)

func TestApp_Renders(t *testing.T) {
	a := app.New()
	tm := btest.NewTestModel(t, a.Model(), 80, 24)

	tm.RequireScreen(func(t testing.TB, s *btest.Screen) {
		btest.AssertNotEmpty(t, s)
		btest.AssertContains(t, s, "{{.ProjectName}}")
	})
}
`
