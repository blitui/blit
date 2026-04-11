package blit

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// Re-exports of bubbletea message and command types so that blit consumers
// never need to import bubbletea directly. The type aliases are zero-cost
// and fully compatible with the original types.

// --- Bubble Tea message types ------------------------------------------------

// Msg is an alias for tea.Msg, the universal message interface.
type Msg = tea.Msg

// Cmd is an alias for tea.Cmd, a function that returns a message.
type Cmd = tea.Cmd

// KeyMsg is an alias for tea.KeyMsg, sent on key presses.
type KeyMsg = tea.KeyMsg

// MouseMsg is an alias for tea.MouseMsg, sent on mouse events.
type MouseMsg = tea.MouseMsg

// WindowSizeMsg is an alias for tea.WindowSizeMsg, sent on terminal resize.
type WindowSizeMsg = tea.WindowSizeMsg

// Key is an alias for tea.Key, the underlying key data structure.
type Key = tea.Key

// KeyType is an alias for tea.KeyType, the type of a key event.
type KeyType = tea.KeyType

// --- Bubble Tea key constants ------------------------------------------------

// Arrow and navigation keys.
var (
	KeyUp        = tea.KeyUp
	KeyDown      = tea.KeyDown
	KeyLeft      = tea.KeyLeft
	KeyRight     = tea.KeyRight
	KeyHome      = tea.KeyHome
	KeyEnd       = tea.KeyEnd
	KeyPgUp      = tea.KeyPgUp
	KeyPgDown    = tea.KeyPgDown
	KeyDelete    = tea.KeyDelete
	KeyInsert    = tea.KeyInsert
	KeyBackspace = tea.KeyBackspace
	KeySpace     = tea.KeySpace
	KeyTab       = tea.KeyTab
	KeyShiftTab  = tea.KeyShiftTab
	KeyEnter     = tea.KeyEnter
	KeyEscape    = tea.KeyEscape
	KeyRunes     = tea.KeyRunes
)

// Control keys.
var (
	KeyCtrlA = tea.KeyCtrlA
	KeyCtrlB = tea.KeyCtrlB
	KeyCtrlC = tea.KeyCtrlC
	KeyCtrlD = tea.KeyCtrlD
	KeyCtrlE = tea.KeyCtrlE
	KeyCtrlF = tea.KeyCtrlF
	KeyCtrlG = tea.KeyCtrlG
	KeyCtrlH = tea.KeyCtrlH
	KeyCtrlJ = tea.KeyCtrlJ
	KeyCtrlK = tea.KeyCtrlK
	KeyCtrlL = tea.KeyCtrlL
	KeyCtrlN = tea.KeyCtrlN
	KeyCtrlO = tea.KeyCtrlO
	KeyCtrlP = tea.KeyCtrlP
	KeyCtrlQ = tea.KeyCtrlQ
	KeyCtrlR = tea.KeyCtrlR
	KeyCtrlS = tea.KeyCtrlS
	KeyCtrlT = tea.KeyCtrlT
	KeyCtrlU = tea.KeyCtrlU
	KeyCtrlV = tea.KeyCtrlV
	KeyCtrlW = tea.KeyCtrlW
	KeyCtrlX = tea.KeyCtrlX
	KeyCtrlY = tea.KeyCtrlY
	KeyCtrlZ = tea.KeyCtrlZ
)

// Function keys.
var (
	F1  = tea.KeyF1
	F2  = tea.KeyF2
	F3  = tea.KeyF3
	F4  = tea.KeyF4
	F5  = tea.KeyF5
	F6  = tea.KeyF6
	F7  = tea.KeyF7
	F8  = tea.KeyF8
	F9  = tea.KeyF9
	F10 = tea.KeyF10
	F11 = tea.KeyF11
	F12 = tea.KeyF12
)

// Special keys.
var (
	KeyCtrlBackslash = tea.KeyCtrlBackslash
	KeyShiftDown     = tea.KeyShiftDown
	KeyShiftUp       = tea.KeyShiftUp
	KeyShiftLeft     = tea.KeyShiftLeft
	KeyShiftRight    = tea.KeyShiftRight
)

// --- Bubble Tea command constructors ------------------------------------------

var (
	// Batch performs the given commands simultaneously with no ordering
	// guarantees about the results.
	Batch = tea.Batch

	// Quit is a command that tells the Bubble Tea runtime to exit.
	Quit = tea.Quit

	// Sequence runs the given commands one at a time, in order.
	Sequence = tea.Sequence
)

// --- Lipgloss type aliases ---------------------------------------------------

// Color is an alias for lipgloss.Color, a terminal color value.
type Color = lipgloss.Color

// Style is an alias for lipgloss.Style, a terminal styling primitive.
type Style = lipgloss.Style

// Border is an alias for lipgloss.Border, a set of border characters.
type Border = lipgloss.Border

// Position is an alias for lipgloss.Position, used for layout placement.
type Position = lipgloss.Position

// --- Lipgloss constructors and layout functions --------------------------------

var (
	// NewStyle creates a new Style.
	NewStyle = lipgloss.NewStyle

	// Width measures the visual width of a string, accounting for ANSI
	// escape sequences and double-width runes.
	Width = lipgloss.Width

	// Height measures the visual height of a string (number of newlines).
	Height = lipgloss.Height

	// JoinVertical joins strings vertically, aligned along the given
	// position.
	JoinVertical = lipgloss.JoinVertical

	// JoinHorizontal joins strings horizontally, aligned along the given
	// position.
	JoinHorizontal = lipgloss.JoinHorizontal

	// Place places a string or text block in a box of the given dimensions.
	Place = lipgloss.Place
)

// --- Lipgloss position constants ----------------------------------------------
//
// Note: Left, Right, and Center are already defined in blit as Alignment
// constants (iota). For lipgloss position values, use the qualified form:
//
//	blit.Place(w, h, lipgloss.Center, lipgloss.Top, s)
//
// Or use the helpers below which avoid the name collision.

var (
	// PosCenter is lipgloss.Center, for use with Place and JoinVertical.
	PosCenter = lipgloss.Center
	// PosTop is lipgloss.Top, for use with Place and JoinVertical.
	PosTop = lipgloss.Top
	// PosBottom is lipgloss.Bottom, for use with Place and JoinVertical.
	PosBottom = lipgloss.Bottom
	// PosLeft is lipgloss.Left, for use with Place and JoinHorizontal.
	PosLeft = lipgloss.Left
	// PosRight is lipgloss.Right, for use with Place and JoinHorizontal.
	PosRight = lipgloss.Right
)

// --- Lipgloss border presets --------------------------------------------------

var (
	RoundedBorder = lipgloss.RoundedBorder
	DoubleBorder  = lipgloss.DoubleBorder
	ThickBorder   = lipgloss.ThickBorder
	ASCIIBorder   = lipgloss.ASCIIBorder
	NormalBorder  = lipgloss.NormalBorder
)

// --- ANSI helpers (from charmbracelet/x/ansi) --------------------------------

var (
	// StringWidth measures the visual width of a string, accounting for
	// ANSI escape sequences and East Asian wide characters.
	StringWidth = ansi.StringWidth

	// TruncateWith truncates a string to the given visual width, appending
	// the tail string if truncation occurs. Use this when you need a custom
	// tail indicator. For the common case (tail = "…"), use blit.Truncate.
	TruncateWith = ansi.Truncate
)
