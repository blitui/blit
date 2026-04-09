package tuitest

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// DiffMode controls how the DiffViewer renders the comparison.
type DiffMode int

const (
	// DiffModeSideBySide shows expected on the left, actual on the right.
	DiffModeSideBySide DiffMode = iota
	// DiffModeUnified shows a unified diff with +/- prefix lines.
	DiffModeUnified
	// DiffModeCellsOnly lists only the differing cells.
	DiffModeCellsOnly
)

// DiffViewer is a tuikit Component that renders a side-by-side (or unified /
// cells-only) comparison of two screens captured from a failing tuitest
// assertion. It implements the tuikit Component interface so it can be
// embedded in any App layout.
//
// Keybindings:
//
//	s  — side-by-side mode
//	u  — unified mode
//	d  — cells-only mode
//	q  — signal back-to-runner (emits DiffViewerBackMsg)
type DiffViewer struct {
	capture *FailureCapture
	mode    DiffMode
	width   int
	height  int
	focused bool
	scroll  int // vertical scroll offset for unified / cells-only views
}

// DiffViewerBackMsg is sent when the user presses q to return to the runner.
type DiffViewerBackMsg struct{}

// NewDiffViewer constructs a DiffViewer from a persisted FailureCapture.
func NewDiffViewer(fc *FailureCapture) *DiffViewer {
	return &DiffViewer{
		capture: fc,
		mode:    DiffModeSideBySide,
	}
}

// --- Component interface ---

// Init implements tuikit.Component.
func (dv *DiffViewer) Init() tea.Cmd { return nil }

// Update implements tuikit.Component. ctx is the ambient tuikit.Context but
// the DiffViewer only needs key messages, so it accepts tea.Msg directly and
// ignores the ctx value (kept as interface{} to avoid importing tuikit from
// inside the tuitest sub-package, which would create a cycle).
func (dv *DiffViewer) Update(msg tea.Msg, _ interface{}) (*DiffViewer, tea.Cmd) {
	switch m := msg.(type) {
	case tea.KeyMsg:
		switch m.String() {
		case "s":
			dv.mode = DiffModeSideBySide
			dv.scroll = 0
		case "u":
			dv.mode = DiffModeUnified
			dv.scroll = 0
		case "d":
			dv.mode = DiffModeCellsOnly
			dv.scroll = 0
		case "q":
			return dv, func() tea.Msg { return DiffViewerBackMsg{} }
		case "up", "k":
			if dv.scroll > 0 {
				dv.scroll--
			}
		case "down", "j":
			dv.scroll++
		case "pgup":
			dv.scroll -= dv.viewHeight()
			if dv.scroll < 0 {
				dv.scroll = 0
			}
		case "pgdown":
			dv.scroll += dv.viewHeight()
		}
	}
	return dv, nil
}

// View implements tuikit.Component.
func (dv *DiffViewer) View() string {
	if dv.capture == nil || dv.width < 4 || dv.height < 3 {
		return ""
	}
	header := dv.renderHeader()
	body := dv.renderBody()
	help := dv.renderHelp()
	return strings.Join([]string{header, body, help}, "\n")
}

// KeyBindings implements tuikit.Component.
func (dv *DiffViewer) KeyBindings() []interface{} { return nil }

// SetSize implements tuikit.Component.
func (dv *DiffViewer) SetSize(w, h int) { dv.width = w; dv.height = h }

// Focused implements tuikit.Component.
func (dv *DiffViewer) Focused() bool { return dv.focused }

// SetFocused implements tuikit.Component.
func (dv *DiffViewer) SetFocused(f bool) { dv.focused = f }

// Mode returns the current display mode.
func (dv *DiffViewer) Mode() DiffMode { return dv.mode }

// SetMode sets the display mode directly (used by the CLI one-shot renderer).
func (dv *DiffViewer) SetMode(m DiffMode) { dv.mode = m; dv.scroll = 0 }

// --- rendering helpers ---

func (dv *DiffViewer) viewHeight() int {
	h := dv.height - 3 // header + help
	if h < 1 {
		return 1
	}
	return h
}

func (dv *DiffViewer) renderHeader() string {
	modeLabel := map[DiffMode]string{
		DiffModeSideBySide: "side-by-side",
		DiffModeUnified:    "unified",
		DiffModeCellsOnly:  "cells-only",
	}[dv.mode]

	title := fmt.Sprintf(" diff: %s  [%s] ", dv.capture.TestName, modeLabel)
	style := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#ffffff")).
		Background(lipgloss.Color("#444444")).
		Width(dv.width)
	return style.Render(title)
}

func (dv *DiffViewer) renderHelp() string {
	s := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	return s.Render(" s side-by-side  u unified  d cells-only  q back ")
}

func (dv *DiffViewer) renderBody() string {
	switch dv.mode {
	case DiffModeSideBySide:
		return dv.renderSideBySide()
	case DiffModeUnified:
		return dv.renderUnified()
	case DiffModeCellsOnly:
		return dv.renderCellsOnly()
	}
	return ""
}

// colorForKind returns the lipgloss color used to highlight a cell.
func colorForKind(k CellKind) lipgloss.Color {
	switch k {
	case CellTextDiffer:
		return lipgloss.Color("#ff5555") // red
	case CellStyleDiffer:
		return lipgloss.Color("#ffff55") // yellow
	default:
		return lipgloss.Color("#55ff55") // green
	}
}

func (dv *DiffViewer) renderSideBySide() string {
	exp := dv.capture.ExpectedScreen
	act := dv.capture.ActualScreen
	if dv.capture.Kind == FailureGolden {
		exp = dv.capture.GoldenExpected
		act = dv.capture.GoldenActual
	}

	expLines := strings.Split(exp, "\n")
	actLines := strings.Split(act, "\n")

	maxLines := len(expLines)
	if len(actLines) > maxLines {
		maxLines = len(actLines)
	}

	half := (dv.width - 3) / 2 // -3 for separator
	if half < 4 {
		half = 4
	}

	sep := lipgloss.NewStyle().Foreground(lipgloss.Color("#444444")).Render("│")
	labelStyle := lipgloss.NewStyle().Bold(true).Width(half)
	expLabel := labelStyle.Foreground(lipgloss.Color("#55ff55")).Render("EXPECTED")
	actLabel := labelStyle.Foreground(lipgloss.Color("#ff5555")).Render("ACTUAL")
	header := expLabel + sep + actLabel

	viewH := dv.viewHeight() - 1 // -1 for column header
	if viewH < 1 {
		viewH = 1
	}
	offset := dv.scroll
	if offset > maxLines-viewH && maxLines > viewH {
		offset = maxLines - viewH
	}
	if offset < 0 {
		offset = 0
	}

	matchStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#55ff55")).Width(half)
	diffStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#ff5555")).Width(half)

	rows := []string{header}
	for i := offset; i < offset+viewH && i < maxLines; i++ {
		var el, al string
		if i < len(expLines) {
			el = expLines[i]
		}
		if i < len(actLines) {
			al = actLines[i]
		}
		same := el == al
		var eStyle, aStyle lipgloss.Style
		if same {
			eStyle = matchStyle
			aStyle = matchStyle
		} else {
			eStyle = diffStyle
			aStyle = diffStyle
		}
		eTrunc := truncateStr(el, half)
		aTrunc := truncateStr(al, half)
		row := eStyle.Render(eTrunc) + sep + aStyle.Render(aTrunc)
		rows = append(rows, row)
	}
	return strings.Join(rows, "\n")
}

func (dv *DiffViewer) renderUnified() string {
	exp := dv.capture.ExpectedScreen
	act := dv.capture.ActualScreen
	if dv.capture.Kind == FailureGolden {
		exp = dv.capture.GoldenExpected
		act = dv.capture.GoldenActual
	}

	expLines := strings.Split(exp, "\n")
	actLines := strings.Split(act, "\n")
	maxLines := len(expLines)
	if len(actLines) > maxLines {
		maxLines = len(actLines)
	}

	addStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#55ff55"))
	delStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#ff5555"))
	ctxStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#aaaaaa"))

	var allLines []string
	for i := 0; i < maxLines; i++ {
		var el, al string
		if i < len(expLines) {
			el = expLines[i]
		}
		if i < len(actLines) {
			al = actLines[i]
		}
		if el == al {
			allLines = append(allLines, ctxStyle.Render("  "+truncateStr(el, dv.width-2)))
		} else {
			allLines = append(allLines, delStyle.Render("- "+truncateStr(el, dv.width-2)))
			allLines = append(allLines, addStyle.Render("+ "+truncateStr(al, dv.width-2)))
		}
	}

	viewH := dv.viewHeight()
	offset := dv.scroll
	if offset > len(allLines)-viewH && len(allLines) > viewH {
		offset = len(allLines) - viewH
	}
	if offset < 0 {
		offset = 0
	}
	end := offset + viewH
	if end > len(allLines) {
		end = len(allLines)
	}
	return strings.Join(allLines[offset:end], "\n")
}

func (dv *DiffViewer) renderCellsOnly() string {
	// Build two screens from the stored text to extract per-cell diffs.
	exp := dv.capture.ExpectedScreen
	act := dv.capture.ActualScreen
	if dv.capture.Kind == FailureGolden {
		exp = dv.capture.GoldenExpected
		act = dv.capture.GoldenActual
	}

	expLines := strings.Split(exp, "\n")
	actLines := strings.Split(act, "\n")
	rows := len(expLines)
	if len(actLines) > rows {
		rows = len(actLines)
	}
	cols := 0
	for _, l := range expLines {
		if len(l) > cols {
			cols = len(l)
		}
	}
	for _, l := range actLines {
		if len(l) > cols {
			cols = len(l)
		}
	}
	if cols < 1 {
		cols = 80
	}

	eScr := NewScreen(cols, rows)
	aScr := NewScreen(cols, rows)
	eScr.Render(exp)
	aScr.Render(act)

	diffs := ScreenCellDiff(eScr, aScr)

	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#aaaaaa"))
	var allLines []string
	for _, cd := range diffs {
		color := colorForKind(cd.Kind)
		label := "text"
		if cd.Kind == CellStyleDiffer {
			label = "style"
		}
		line := fmt.Sprintf("(%d,%d) %-5s  exp:%q act:%q", cd.Row, cd.Col, label, cd.ExpectedText, cd.ActualText)
		allLines = append(allLines, lipgloss.NewStyle().Foreground(color).Render(line))
	}
	if len(allLines) == 0 {
		allLines = []string{labelStyle.Render("  (no cell differences)")}
	}

	viewH := dv.viewHeight()
	offset := dv.scroll
	if offset > len(allLines)-viewH && len(allLines) > viewH {
		offset = len(allLines) - viewH
	}
	if offset < 0 {
		offset = 0
	}
	end := offset + viewH
	if end > len(allLines) {
		end = len(allLines)
	}
	return strings.Join(allLines[offset:end], "\n")
}

// truncateStr truncates s to at most maxW bytes (ASCII-safe shortcut).
func truncateStr(s string, maxW int) string {
	if len(s) <= maxW {
		return s
	}
	if maxW <= 3 {
		return s[:maxW]
	}
	return s[:maxW-1] + "…"
}
