// review.go implements `tuitest review` — an interactive TUI that presents
// a queue of pending .golden.new snapshot diffs so the developer can accept
// or reject each one without leaving the terminal.
//
// Key bindings:
//
//	a  accept current item (rename .golden.new → .golden atomically)
//	r  reject current item (delete .golden.new)
//	s  skip current item (move to next without changing disk)
//	n  next item
//	p  prev item
//	q  quit
package main

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/moneycaringcoder/tuikit-go/tuitest"
)

// ----------------------------------------------------------------------------
// Messages
// ----------------------------------------------------------------------------

type reviewLoadedMsg struct {
	items []tuitest.PendingGolden
}

type reviewActionMsg struct {
	accepted bool // true=accept, false=reject
	idx      int
}

type reviewErrMsg struct{ err error }

// ----------------------------------------------------------------------------
// Model
// ----------------------------------------------------------------------------

// reviewModel is the bubbletea model for the review TUI.
type reviewModel struct {
	items     []tuitest.PendingGolden
	cursor    int
	width     int
	height    int
	viewer    *tuitest.DiffViewer
	done      bool
	statusMsg string
}

func newReviewModel(items []tuitest.PendingGolden) reviewModel {
	m := reviewModel{items: items}
	if len(items) > 0 {
		m.viewer = buildViewer(items[0])
	}
	return m
}

func buildViewer(pg tuitest.PendingGolden) *tuitest.DiffViewer {
	fc := &tuitest.FailureCapture{
		TestName:       pg.TestName(),
		Kind:           tuitest.FailureGolden,
		GoldenPath:     pg.GoldenPath,
		GoldenExpected: pg.Expected,
		GoldenActual:   pg.Actual,
	}
	dv := tuitest.NewDiffViewer(fc)
	return dv
}

func (m reviewModel) Init() tea.Cmd { return nil }

func (m reviewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if m.viewer != nil {
			m.viewer.SetSize(m.width, m.height-5)
		}
		return m, nil

	case tea.KeyMsg:
		if m.done || len(m.items) == 0 {
			if msg.String() == "q" || msg.String() == "ctrl+c" {
				return m, tea.Quit
			}
			return m, nil
		}
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "a":
			idx := m.cursor
			item := m.items[idx]
			if err := item.Accept(); err != nil {
				m.statusMsg = fmt.Sprintf("error: %v", err)
				return m, nil
			}
			m.statusMsg = fmt.Sprintf("accepted: %s", item.TestName())
			m.items = append(m.items[:idx], m.items[idx+1:]...)
			m = m.clampCursor()
		case "r":
			idx := m.cursor
			item := m.items[idx]
			if err := item.Reject(); err != nil {
				m.statusMsg = fmt.Sprintf("error: %v", err)
				return m, nil
			}
			m.statusMsg = fmt.Sprintf("rejected: %s", item.TestName())
			m.items = append(m.items[:idx], m.items[idx+1:]...)
			m = m.clampCursor()
		case "s", "n":
			m.statusMsg = ""
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}
		case "p":
			m.statusMsg = ""
			if m.cursor > 0 {
				m.cursor--
			}
		default:
			// Forward other keys to the diff viewer (scroll etc.)
			if m.viewer != nil {
				updated, cmd := m.viewer.Update(msg, nil)
				m.viewer = updated
				return m, cmd
			}
		}
		// Rebuild viewer after navigation or action.
		if len(m.items) == 0 {
			m.done = true
			m.viewer = nil
		} else {
			m.viewer = buildViewer(m.items[m.cursor])
			if m.viewer != nil {
				m.viewer.SetSize(m.width, m.height-5)
			}
		}
		return m, nil
	}
	return m, nil
}

func (m reviewModel) clampCursor() reviewModel {
	if len(m.items) == 0 {
		m.cursor = 0
		return m
	}
	if m.cursor >= len(m.items) {
		m.cursor = len(m.items) - 1
	}
	return m
}

// ----------------------------------------------------------------------------
// View
// ----------------------------------------------------------------------------

var (
	reviewHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#ffffff")).
				Background(lipgloss.Color("#5555ff")).
				PaddingLeft(1).
				PaddingRight(1)

	reviewStatusStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#aaaaaa")).
				PaddingLeft(1)

	reviewHelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")).
			PaddingLeft(1)

	reviewDoneStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#55ff55"))

	reviewItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ffffff"))

	reviewCursorStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#ffff55"))
)

func (m reviewModel) View() string {
	var b strings.Builder

	total := len(m.items)

	if m.done || total == 0 {
		b.WriteString(reviewHeaderStyle.Render(" tuitest review ") + "\n\n")
		b.WriteString(reviewDoneStyle.Render("  All pending snapshots reviewed.") + "\n\n")
		b.WriteString(reviewHelpStyle.Render(" q quit") + "\n")
		return b.String()
	}

	// Header
	header := fmt.Sprintf(" tuitest review  %d/%d pending ", m.cursor+1, total)
	if m.width > 0 {
		reviewHeaderStyle = reviewHeaderStyle.Width(m.width)
	}
	b.WriteString(reviewHeaderStyle.Render(header) + "\n")

	// Queue list (up to 5 items)
	listHeight := 5
	start := m.cursor - 2
	if start < 0 {
		start = 0
	}
	end := start + listHeight
	if end > total {
		end = total
	}
	for i := start; i < end; i++ {
		prefix := "  "
		style := reviewItemStyle
		if i == m.cursor {
			prefix = "▶ "
			style = reviewCursorStyle
		}
		b.WriteString(style.Render(fmt.Sprintf("%s%s", prefix, m.items[i].TestName())) + "\n")
	}
	b.WriteString("\n")

	// Diff viewer
	if m.viewer != nil {
		b.WriteString(m.viewer.View())
		b.WriteString("\n")
	}

	// Status + help
	if m.statusMsg != "" {
		b.WriteString(reviewStatusStyle.Render(m.statusMsg) + "\n")
	}
	b.WriteString(reviewHelpStyle.Render(" a accept  r reject  s skip  n next  p prev  q quit") + "\n")

	return b.String()
}

// ----------------------------------------------------------------------------
// Entry point
// ----------------------------------------------------------------------------

// runReview implements the `tuitest review` subcommand. It scans the given
// root (default ".") for .golden.new files and launches the review TUI.
func runReview(args []string) int {
	root := "."
	if len(args) > 0 {
		root = args[0]
	}

	items, err := tuitest.FindPendingGoldens(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[tuitest review] scan error: %v\n", err)
		return 1
	}
	if len(items) == 0 {
		fmt.Println("[tuitest review] no pending snapshots found")
		return 0
	}

	m := newReviewModel(items)
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "[tuitest review] TUI error: %v\n", err)
		return 1
	}
	return 0
}
