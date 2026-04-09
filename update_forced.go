package tuikit

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ForcedUpdateScreen is the full-screen gate shown when an update is
// Required (minimum_version marker set). The app cannot proceed until
// the user either updates or quits.
//
// Wiring:
//
//	gate := NewForcedUpdateScreen(result, cfg)
//	if result.Required {
//	    program := tea.NewProgram(gate)
//	    program.Run()
//	}
type ForcedUpdateScreen struct {
	Result  *UpdateResult
	Cfg     UpdateConfig
	Choice  ForcedChoice
	Width   int
	Height  int
	Message string // post-action status (e.g., "updating...", "failed: ...")
}

// ForcedChoice records the user's decision on the forced gate.
type ForcedChoice int

const (
	// ForcedChoicePending means the user has not chosen yet.
	ForcedChoicePending ForcedChoice = iota
	// ForcedChoiceUpdate means the user accepted the update.
	ForcedChoiceUpdate
	// ForcedChoiceQuit means the user declined and wants to quit.
	ForcedChoiceQuit
)

// NewForcedUpdateScreen returns a gate for the given update result and config.
func NewForcedUpdateScreen(result *UpdateResult, cfg UpdateConfig) *ForcedUpdateScreen {
	return &ForcedUpdateScreen{
		Result: result,
		Cfg:    cfg,
		Width:  80,
		Height: 24,
	}
}

// Init implements tea.Model.
func (f *ForcedUpdateScreen) Init() tea.Cmd { return nil }

// Update implements tea.Model. Accepts y/u/enter to update, q/esc/ctrl+c to quit.
func (f *ForcedUpdateScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		f.Width = m.Width
		f.Height = m.Height
	case tea.KeyMsg:
		switch m.String() {
		case "y", "u", "enter":
			f.Choice = ForcedChoiceUpdate
			return f, tea.Quit
		case "q", "esc", "ctrl+c":
			f.Choice = ForcedChoiceQuit
			return f, tea.Quit
		}
	}
	return f, nil
}

// View implements tea.Model. Renders a centered dialog box that fills the
// terminal with a title, release info, and action hint line.
func (f *ForcedUpdateScreen) View() string {
	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("11")).
		Render("!! Required update")
	body := fmt.Sprintf(
		"A new version is required: %s → %s\n"+
			"Your current version can no longer connect. Install the update to continue.",
		f.Result.CurrentVersion, f.Result.LatestVersion,
	)
	hint := lipgloss.NewStyle().Faint(true).
		Render("[u]pdate   [q]uit")

	var notes string
	if rn := strings.TrimSpace(f.Result.ReleaseNotes); rn != "" {
		notes = "\n\nRelease notes:\n" + rn
	}
	content := title + "\n\n" + body + notes + "\n\n" + hint
	if f.Message != "" {
		content += "\n\n" + lipgloss.NewStyle().Italic(true).Render(f.Message)
	}

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("11")).
		Padding(1, 2).
		Render(content)

	return lipgloss.Place(f.Width, f.Height,
		lipgloss.Center, lipgloss.Center, box,
	)
}
