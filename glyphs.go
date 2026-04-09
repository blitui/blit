package tuikit

// Glyphs defines terminal symbols used by tuikit components.
// Two packs are provided: DefaultGlyphs (Unicode) and AsciiGlyphs.
// Attach a Glyphs instance to a Theme to customise all component symbols at once.
type Glyphs struct {
	// List / tree connectors
	TreeBranch string
	TreeLast   string
	TreePipe   string
	TreeEmpty  string

	// Cursor / selection
	CursorMarker     string
	FlashMarker      string
	SelectedBullet   string
	UnselectedBullet string
	CollapsedArrow   string
	ExpandedArrow    string

	// Progress bar
	BarFilled string
	BarEmpty  string

	// Spinner frames (cycled sequentially)
	SpinnerFrames []string

	// Status / feedback
	Check string
	Cross string
	Info  string
	Warn  string
	Star  string
	Dot   string
}

// DefaultGlyphs returns the standard Unicode glyph set.
func DefaultGlyphs() Glyphs {
	return Glyphs{
		TreeBranch:       "├─",
		TreeLast:         "└─",
		TreePipe:         "│",
		TreeEmpty:        "  ",
		CursorMarker:     "▌",
		FlashMarker:      "▐",
		SelectedBullet:   "●",
		UnselectedBullet: "○",
		CollapsedArrow:   "▸",
		ExpandedArrow:    "▾",
		BarFilled:        "█",
		BarEmpty:         "░",
		SpinnerFrames:    []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
		Check:            "✓",
		Cross:            "✗",
		Info:             "ℹ",
		Warn:             "!",
		Star:             "★",
		Dot:              "·",
	}
}

// AsciiGlyphs returns a pure-ASCII glyph set for terminals without Unicode support.
func AsciiGlyphs() Glyphs {
	return Glyphs{
		TreeBranch:       "+-",
		TreeLast:         "\\-",
		TreePipe:         "|",
		TreeEmpty:        "  ",
		CursorMarker:     ">",
		FlashMarker:      "*",
		SelectedBullet:   "*",
		UnselectedBullet: ".",
		CollapsedArrow:   ">",
		ExpandedArrow:    "v",
		BarFilled:        "#",
		BarEmpty:         "-",
		SpinnerFrames:    []string{"|", "/", "-", "\\"},
		Check:            "v",
		Cross:            "x",
		Info:             "i",
		Warn:             "!",
		Star:             "*",
		Dot:              ".",
	}
}
