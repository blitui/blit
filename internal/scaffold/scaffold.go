// Package scaffold generates blit boilerplate for a given package path.
package scaffold

import (
	"bytes"
	"fmt"
	"go/format"
	"path/filepath"
	"strings"
	"text/template"
	"unicode"
)

// Options controls the output of Generate.
type Options struct {
	// PkgPath is the import path or directory path of the package under test.
	PkgPath string
	// Component is an optional component type name. When non-empty the
	// boilerplate is tailored to that type.
	Component string
}

// Result holds the generated file content and the suggested output path.
type Result struct {
	// FileName is the suggested file name (not the full path).
	FileName string
	// Content is the gofmt-formatted Go source.
	Content []byte
}

// Generate produces a _test.go file with btest.NewTestModel boilerplate.
// The generated source passes go vet immediately.
func Generate(opts Options) (*Result, error) {
	pkgName := pkgNameFromPath(opts.PkgPath)
	testPkgName := pkgName + "_test"

	component := opts.Component
	if component == "" {
		component = exported(pkgName)
	}

	data := tmplData{
		TestPkg:   testPkgName,
		Component: component,
		ModPkg:    opts.PkgPath,
	}

	var buf bytes.Buffer
	if err := genTemplate.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("scaffold: template execute: %w", err)
	}

	src, err := format.Source(buf.Bytes())
	if err != nil {
		return nil, fmt.Errorf("scaffold: gofmt: %w: raw source:\n%s", err, buf.String())
	}

	fileName := strings.ToLower(component) + "_test.go"
	return &Result{
		FileName: fileName,
		Content:  src,
	}, nil
}

type tmplData struct {
	TestPkg   string
	Component string
	ModPkg    string
}

// pkgNameFromPath derives a short package name from an import path or directory.
//
//	"./mypkg"         → "mypkg"
//	"github.com/x/y" → "y"
//	"."               → "main"
func pkgNameFromPath(p string) string {
	p = filepath.ToSlash(p)
	p = strings.TrimPrefix(p, "./")
	p = strings.TrimPrefix(p, "../")
	if p == "" || p == "." {
		return "main"
	}
	parts := strings.Split(p, "/")
	name := parts[len(parts)-1]
	if name == "" {
		return "main"
	}
	return name
}

// exported converts a name to an exported Go identifier (capitalises the first
// letter and replaces non-letter/digit runes with nothing).
func exported(s string) string {
	if s == "" {
		return "Component"
	}
	runes := []rune(s)
	runes[0] = unicode.ToUpper(runes[0])
	var out []rune
	for _, r := range runes {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			out = append(out, r)
		}
	}
	if len(out) == 0 {
		return "Component"
	}
	return string(out)
}

var genTemplate = template.Must(template.New("scaffold").Parse(`package {{.TestPkg}}

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/blitui/blit/btest"
)

// Test{{.Component}} is a generated test stub. Replace the stub model with a
// real tea.Model from your package and extend the assertions.
func Test{{.Component}}(t *testing.T) {
	// TODO: replace &stub{{.Component}}{} with your model, e.g. mypkg.New{{.Component}}().
	tm := btest.NewTestModel(t, &stub{{.Component}}{}, 80, 24)

	tm.RequireScreen(func(t testing.TB, s *btest.Screen) {
		if s.String() == "" {
			t.Log("screen is empty — wire up your model and add assertions here")
		}
	})
}

// stub{{.Component}} satisfies tea.Model so the generated file compiles and
// passes go vet immediately without importing the package under test.
type stub{{.Component}} struct{}

func (s *stub{{.Component}}) Init() tea.Cmd                           { return nil }
func (s *stub{{.Component}}) Update(msg tea.Msg) (tea.Model, tea.Cmd) { return s, nil }
func (s *stub{{.Component}}) View() string                            { return "" }
`))
