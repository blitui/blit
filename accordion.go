package blit

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// AccordionSection is a collapsible section in an Accordion.
type AccordionSection struct {
	// Title is the header text.
	Title string
	// Content is the body text shown when expanded.
	Content string
	// Expanded controls whether this section is open.
	Expanded bool
}

// AccordionOpts configures an Accordion.
type AccordionOpts struct {
	// Exclusive limits expansion to one section at a time (true = only one open).
	Exclusive bool

	// OnToggle is called when a section is expanded or collapsed.
	OnToggle func(index int, expanded bool)
}

// Accordion is a vertically stacked set of collapsible sections.
// Implements Component and Themed.
type Accordion struct {
	opts     AccordionOpts
	sections []AccordionSection
	cursor   int
	theme    Theme
	focused  bool
	width    int
	height   int
}

// NewAccordion creates an Accordion with the given sections and options.
func NewAccordion(sections []AccordionSection, opts AccordionOpts) *Accordion {
	return &Accordion{
		opts:     opts,
		sections: sections,
	}
}

// Sections returns the current sections.
func (a *Accordion) Sections() []AccordionSection { return a.sections }

// CursorIndex returns the currently highlighted section index.
func (a *Accordion) CursorIndex() int { return a.cursor }

// Init implements Component.
func (a *Accordion) Init() tea.Cmd { return nil }

// Update implements Component.
func (a *Accordion) Update(msg tea.Msg, ctx Context) (Component, tea.Cmd) {
	if !a.focused {
		return a, nil
	}
	km, ok := msg.(tea.KeyMsg)
	if !ok {
		return a, nil
	}
	switch km.String() {
	case "up", "k":
		if a.cursor > 0 {
			a.cursor--
		}
		return a, Consumed()
	case "down", "j":
		if a.cursor < len(a.sections)-1 {
			a.cursor++
		}
		return a, Consumed()
	case "enter", " ":
		a.toggleSection(a.cursor)
		return a, Consumed()
	}
	return a, nil
}

// View implements Component.
func (a *Accordion) View() string {
	if a.width == 0 || a.height == 0 || len(a.sections) == 0 {
		return ""
	}

	g := a.theme.glyphsOrDefault()
	textColor := lipgloss.Color(a.theme.Text)
	accentColor := lipgloss.Color(a.theme.Accent)
	mutedColor := lipgloss.Color(a.theme.Muted)
	cursorBg := lipgloss.Color(a.theme.Cursor)
	inverseFg := lipgloss.Color(a.theme.TextInverse)

	var lines []string
	for i, sec := range a.sections {
		isCursor := i == a.cursor

		// Arrow indicator.
		var arrow string
		if sec.Expanded {
			arrow = g.ExpandedArrow
		} else {
			arrow = g.CollapsedArrow
		}

		// Header line.
		var header string
		if isCursor {
			arrowStyled := lipgloss.NewStyle().Foreground(accentColor).Render(arrow)
			titleStyled := lipgloss.NewStyle().
				Background(cursorBg).
				Foreground(inverseFg).
				Render(" " + sec.Title + " ")
			header = arrowStyled + " " + titleStyled
		} else {
			arrowStyled := lipgloss.NewStyle().Foreground(mutedColor).Render(arrow)
			titleStyled := lipgloss.NewStyle().Foreground(textColor).Render(sec.Title)
			header = arrowStyled + " " + titleStyled
		}
		lines = append(lines, header)

		// Content (indented) when expanded.
		if sec.Expanded && sec.Content != "" {
			contentStyle := lipgloss.NewStyle().
				Foreground(textColor).
				Width(a.width - 4).
				PaddingLeft(3)
			lines = append(lines, contentStyle.Render(sec.Content))
		}

		// Separator between sections.
		if i < len(a.sections)-1 {
			sep := lipgloss.NewStyle().
				Foreground(mutedColor).
				Render(strings.Repeat("─", a.width))
			lines = append(lines, sep)
		}
	}

	return strings.Join(lines, "\n")
}

// KeyBindings implements Component.
func (a *Accordion) KeyBindings() []KeyBind {
	return []KeyBind{
		{Key: "up/k", Label: "Previous section", Group: "ACCORDION"},
		{Key: "down/j", Label: "Next section", Group: "ACCORDION"},
		{Key: "enter/space", Label: "Toggle section", Group: "ACCORDION"},
	}
}

// SetSize implements Component.
func (a *Accordion) SetSize(w, h int) {
	a.width = w
	a.height = h
}

// Focused implements Component.
func (a *Accordion) Focused() bool { return a.focused }

// SetFocused implements Component.
func (a *Accordion) SetFocused(f bool) { a.focused = f }

// SetTheme implements Themed.
func (a *Accordion) SetTheme(theme Theme) { a.theme = theme }

func (a *Accordion) toggleSection(idx int) {
	if idx < 0 || idx >= len(a.sections) {
		return
	}
	sec := &a.sections[idx]
	sec.Expanded = !sec.Expanded

	// In exclusive mode, collapse all others when expanding.
	if a.opts.Exclusive && sec.Expanded {
		for i := range a.sections {
			if i != idx {
				a.sections[i].Expanded = false
			}
		}
	}

	if a.opts.OnToggle != nil {
		a.opts.OnToggle(idx, sec.Expanded)
	}
}
