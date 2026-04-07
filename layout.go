package tuikit

// paneSize holds computed dimensions for a layout pane.
type paneSize struct {
	width  int
	height int
}

// Layout defines how components are arranged within the terminal.
type Layout interface {
	compute(totalWidth, totalHeight int) (paneSize, paneSize, bool)
}

// SinglePane is a layout with one component filling all available space.
type SinglePane struct{}

func (s SinglePane) compute(totalWidth, totalHeight int) (paneSize, paneSize, bool) {
	return paneSize{totalWidth, totalHeight}, paneSize{}, false
}

// DualPane is a layout with a main component and a collapsible sidebar.
type DualPane struct {
	Main         Component // Main pane component
	Side         Component // Sidebar component
	SideWidth    int       // Fixed width of the sidebar
	MinMainWidth int       // Sidebar auto-hides below this main width
	SideRight    bool      // true = sidebar on right, false = on left
	ToggleKey    string    // Key to toggle sidebar visibility

	sideHidden bool // User toggled sidebar off
}

// Toggle flips sidebar visibility on/off.
func (d *DualPane) Toggle() {
	d.sideHidden = !d.sideHidden
}

func (d DualPane) compute(totalWidth, totalHeight int) (paneSize, paneSize, bool) {
	if d.sideHidden {
		return paneSize{totalWidth, totalHeight}, paneSize{}, false
	}

	// Auto-hide if terminal too narrow
	if totalWidth < d.MinMainWidth+d.SideWidth+1 {
		return paneSize{totalWidth, totalHeight}, paneSize{}, false
	}

	mainW := totalWidth - d.SideWidth - 1 // 1 for border/separator
	return paneSize{mainW, totalHeight}, paneSize{d.SideWidth, totalHeight}, true
}
