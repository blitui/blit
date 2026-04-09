// Package main demonstrates the tuikit Form component with validators and wizard mode.
//
// Press 'w' to toggle between normal and wizard modes.
// Press 'q' to quit.
package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	tuikit "github.com/moneycaringcoder/tuikit-go"
)

func buildForm(wizard bool, onSubmit func(map[string]string)) *tuikit.Form {
	return tuikit.NewForm(tuikit.FormOpts{
		WizardMode: wizard,
		OnSubmit:   onSubmit,
		Groups: []tuikit.FormGroup{
			{
				Title: "Account",
				Fields: []tuikit.Field{
					tuikit.NewTextField("username", "Username").
						WithPlaceholder("e.g. alice123").
						WithRequired().
						WithValidator(tuikit.ComposeValidators(
							tuikit.MinLength(3),
							tuikit.MaxLength(20),
							tuikit.RegexValidator(`^[a-zA-Z0-9_]+$`, "only letters, digits, and underscores"),
						)),
					tuikit.NewTextField("email", "Email").
						WithPlaceholder("you@example.com").
						WithRequired().
						WithValidator(tuikit.EmailValidator()),
					tuikit.NewPasswordField("password", "Password").
						WithPlaceholder("min 8 characters").
						WithRequired().
						WithValidator(tuikit.MinLength(8)),
				},
			},
			{
				Title: "Profile",
				Fields: []tuikit.Field{
					tuikit.NewSelectField("role", "Role",
						[]string{"Developer", "Designer", "Manager", "Other"}).
						WithHint("Your primary role"),
					tuikit.NewMultiSelectField("interests", "Interests",
						[]string{"Go", "TUI", "CLI", "Web", "DevOps"}).
						WithHint("Select all that apply"),
					tuikit.NewNumberField("age", "Age").
						WithPlaceholder("18-120").
						WithMin(18).
						WithMax(120),
					tuikit.NewConfirmField("newsletter", "Subscribe to newsletter").
						WithDefault(true),
				},
			},
		},
	})
}

type model struct {
	form      *tuikit.Form
	wizard    bool
	submitted bool
	result    map[string]string
	theme     tuikit.Theme
	width     int
	height    int
}

func newModel() model {
	theme := tuikit.DefaultTheme()
	m := model{theme: theme}
	m.form = buildForm(false, func(vals map[string]string) {
		m.submitted = true
		m.result = vals
	})
	m.form.SetTheme(theme)
	m.form.SetFocused(true)
	return m
}

func (m model) Init() tea.Cmd { return m.form.Init() }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.form.SetSize(msg.Width-4, msg.Height-6)
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "w":
			m.wizard = !m.wizard
			m.submitted = false
			m.result = nil
			m.form = buildForm(m.wizard, func(vals map[string]string) {
				m.submitted = true
				m.result = vals
			})
			m.form.SetTheme(m.theme)
			m.form.SetFocused(true)
			m.form.SetSize(m.width-4, m.height-6)
			return m, m.form.Init()
		}
	case tuikit.FormSubmitMsg:
		m.submitted = true
		m.result = msg.Values
		return m, nil
	}
	comp, cmd := m.form.Update(msg, tuikit.Context{})
	m.form = comp.(*tuikit.Form)
	return m, cmd
}

func (m model) View() string {
	th := m.theme
	titleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(th.Accent)).Bold(true).Padding(0, 1)
	modeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(th.Muted))
	hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(th.Muted))

	mode := "normal"
	if m.wizard {
		mode = "wizard"
	}
	header := lipgloss.JoinHorizontal(lipgloss.Center,
		titleStyle.Render("Signup Form"),
		"  ",
		modeStyle.Render(fmt.Sprintf("[mode: %s]", mode)),
	)
	hint := hintStyle.Render("w: toggle wizard  q: quit")

	if m.submitted && m.result != nil {
		var lines []string
		lines = append(lines,
			lipgloss.NewStyle().Foreground(lipgloss.Color(th.Positive)).Bold(true).Render("Form submitted successfully!"))
		lines = append(lines, "")
		for k, v := range m.result {
			if k == "password" {
				v = strings.Repeat("*", len(v))
			}
			lines = append(lines,
				lipgloss.NewStyle().Foreground(lipgloss.Color(th.Text)).
					Render(fmt.Sprintf("  %-14s %s", k+":", v)))
		}
		lines = append(lines, "")
		lines = append(lines, hintStyle.Render("Press 'w' to reset or 'q' to quit"))
		return lipgloss.JoinVertical(lipgloss.Left, header, "", strings.Join(lines, "\n"))
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		"",
		lipgloss.NewStyle().Padding(0, 2).Render(m.form.View()),
		"",
		lipgloss.NewStyle().Padding(0, 2).Render(hint),
	)
}

func main() {
	p := tea.NewProgram(newModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
