package main

const formMainTmpl = `package main

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

const formAppTmpl = `package {{.BinaryName}}

import (
	"fmt"

	blit "github.com/blitui/blit"
)

// New creates a form application with validated inputs.
func New() *blit.App {
	form := blit.NewForm(blit.FormOpts{
		OnSubmit: func(vals map[string]string) {
			fmt.Printf("Submitted: %v\n", vals)
		},
		Groups: []blit.FormGroup{
			{
				Title: "Details",
				Fields: []blit.Field{
					blit.NewTextField("name", "Name").
						WithPlaceholder("Your name").
						WithRequired().
						WithValidator(blit.MinLength(2)),
					blit.NewTextField("email", "Email").
						WithPlaceholder("you@example.com").
						WithRequired().
						WithValidator(blit.EmailValidator()),
					blit.NewSelectField("role", "Role",
						[]string{"Developer", "Designer", "Manager", "Other"}).
						WithHint("Your primary role"),
					blit.NewConfirmField("newsletter", "Subscribe to updates").
						WithDefault(true),
				},
			},
		},
	})

	return blit.NewApp(
		blit.WithTheme(blit.DefaultTheme()),
		blit.WithComponent("form", form),
		blit.WithStatusBar(
			func() string { return " tab next  shift+tab prev  enter submit  ? help  q quit" },
			func() string { return " {{.ProjectName}} " },
		),
		blit.WithHelp(),
	)
}
`

const formAppTestTmpl = `package {{.BinaryName}}_test

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
		btest.AssertContains(t, s, "Name")
	})
}
`
