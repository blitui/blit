package btest

// CellStyle represents the visual attributes of a terminal cell.
type CellStyle struct {
	Fg        string // foreground color (empty = default)
	Bg        string // background color (empty = default)
	Bold      bool
	Italic    bool
	Underline bool
	Reverse   bool
}
