package blit

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// paneSize holds computed dimensions for a layout pane.
type paneSize struct {
	width  int
	height int
}

// Layout defines how components are arranged within the terminal.
type Layout interface {
	compute(totalWidth, totalHeight int) (paneSize, paneSize, bool)
}

// DualPane is a layout with a main component and a collapsible sidebar.
type DualPane struct {
	Main         Component // Main pane component
	Side         Component // Sidebar component
	MainName     string    // Display name for focus badge (default: "Main")
	SideName     string    // Display name for focus badge (default: "Side")
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

func (d *DualPane) compute(totalWidth, totalHeight int) (paneSize, paneSize, bool) {
	if d.sideHidden {
		return paneSize{totalWidth, totalHeight}, paneSize{}, false
	}

	// Auto-hide if terminal too narrow
	if totalWidth < d.MinMainWidth+d.SideWidth+3 {
		return paneSize{totalWidth, totalHeight}, paneSize{}, false
	}

	mainW := totalWidth - d.SideWidth - 3 // 3 for " │ " separator
	return paneSize{mainW, totalHeight}, paneSize{d.SideWidth, totalHeight}, true
}

// ----------------------------------------------------------------------------
// Flex layout primitives
// ----------------------------------------------------------------------------

// FlexAlign controls cross-axis alignment for HBox/VBox children.
type FlexAlign int

const (
	// FlexAlignStart aligns children to the start of the cross axis.
	FlexAlignStart FlexAlign = iota
	// FlexAlignCenter centers children on the cross axis.
	FlexAlignCenter
	// FlexAlignEnd aligns children to the end of the cross axis.
	FlexAlignEnd
	// FlexAlignStretch stretches children to fill the cross axis.
	FlexAlignStretch
)

// FlexJustify controls main-axis distribution for HBox/VBox children.
type FlexJustify int

const (
	// FlexJustifyStart packs children toward the start.
	FlexJustifyStart FlexJustify = iota
	// FlexJustifyCenter packs children in the center.
	FlexJustifyCenter
	// FlexJustifyEnd packs children toward the end.
	FlexJustifyEnd
	// FlexJustifySpaceBetween distributes children with equal gaps between them.
	FlexJustifySpaceBetween
	// FlexJustifySpaceAround distributes children with equal space around each.
	FlexJustifySpaceAround
)

// Sized wraps a Component with a fixed size on the main axis.
// In HBox the size is a fixed width; in VBox it is a fixed height.
type Sized struct {
	W int       // Fixed width (HBox) or height (VBox) in columns/rows.
	C Component // The wrapped component.
}

// Init initializes the Sized component.
func (s Sized) Init() tea.Cmd { return s.C.Init() }
// Update handles incoming messages and updates Sized state.
func (s Sized) Update(msg tea.Msg, ctx Context) (Component, tea.Cmd) {
	c, cmd := s.C.Update(msg, ctx)
	return Sized{W: s.W, C: c}, cmd
}
// View renders the Sized as a string.
func (s Sized) View() string           { return s.C.View() }
// KeyBindings returns the key bindings for the Sized.
func (s Sized) KeyBindings() []KeyBind { return s.C.KeyBindings() }
// SetSize sets the width and height of the Sized.
func (s Sized) SetSize(w, h int)       { s.C.SetSize(w, h) }
// Focused reports whether the Sized is focused.
func (s Sized) Focused() bool          { return s.C.Focused() }
// SetFocused sets the focus state of the Sized.
func (s Sized) SetFocused(f bool)      { s.C.SetFocused(f) }

// SetTheme delegates theme propagation to the inner component if it
// implements Themed. This ensures runtime theme switching reaches
// children inside HBox/VBox layouts.
func (s Sized) SetTheme(t Theme) {
	if themed, ok := s.C.(Themed); ok {
		themed.SetTheme(t)
	}
}

// Flex wraps a Component that grows proportionally to fill remaining space.
// Grow is the relative weight — a child with Grow=2 gets twice as much space
// as one with Grow=1.
type Flex struct {
	Grow int       // Growth weight (≥1).
	C    Component // The wrapped component.
}

// Init initializes the Flex component.
func (f Flex) Init() tea.Cmd { return f.C.Init() }
// Update handles incoming messages and updates Flex state.
func (f Flex) Update(msg tea.Msg, ctx Context) (Component, tea.Cmd) {
	c, cmd := f.C.Update(msg, ctx)
	return Flex{Grow: f.Grow, C: c}, cmd
}
// View renders the Flex as a string.
func (f Flex) View() string           { return f.C.View() }
// KeyBindings returns the key bindings for the Flex.
func (f Flex) KeyBindings() []KeyBind { return f.C.KeyBindings() }
// SetSize sets the width and height of the Flex.
func (f Flex) SetSize(w, h int)       { f.C.SetSize(w, h) }
// Focused reports whether the Flex is focused.
func (f Flex) Focused() bool          { return f.C.Focused() }
// SetFocused sets the focus state of the Flex.
func (f Flex) SetFocused(foc bool)    { f.C.SetFocused(foc) }

// SetTheme implements Themed by delegating to the wrapped component.
func (f Flex) SetTheme(th Theme) {
	if t, ok := f.C.(Themed); ok {
		t.SetTheme(th)
	}
}

// flexItem is an internal resolved slot used during layout computation.
type flexItem struct {
	c       Component
	fixedSz int // 0 = auto / will be computed
	grow    int // 0 = fixed
}

// HBox lays out children horizontally (left → right).
// It implements the full Component interface and can be nested.
type HBox struct {
	Gap     int         // Gap in columns between children.
	Align   FlexAlign   // Cross-axis (vertical) alignment.
	Justify FlexJustify // Main-axis (horizontal) distribution.
	Items   []Component // Children — may be plain Component, Sized, or Flex.

	width   int
	height  int
	focused bool
}

// VBox lays out children vertically (top → bottom).
// It implements the full Component interface and can be nested.
type VBox struct {
	Gap     int         // Gap in rows between children.
	Align   FlexAlign   // Cross-axis (horizontal) alignment.
	Justify FlexJustify // Main-axis (vertical) distribution.
	Items   []Component // Children — may be plain Component, Sized, or Flex.

	width   int
	height  int
	focused bool
}

// Init implements Component.
func (h *HBox) Init() tea.Cmd {
	var cmds []tea.Cmd
	for _, item := range h.Items {
		if c := unwrapFlexComponent(item); c != nil {
			if cmd := c.Init(); cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
	}
	return tea.Batch(cmds...)
}

// Update implements Component.
func (h *HBox) Update(msg tea.Msg, ctx Context) (Component, tea.Cmd) {
	var cmds []tea.Cmd
	for i, item := range h.Items {
		c := unwrapFlexComponent(item)
		if c == nil {
			continue
		}
		updated, cmd := c.Update(msg, ctx)
		h.Items[i] = rewrapFlexComponent(item, updated)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}
	return h, tea.Batch(cmds...)
}

// View implements Component.
func (h *HBox) View() string {
	if len(h.Items) == 0 {
		return ""
	}
	slots := resolveHBoxSlots(h)
	return renderHBox(h, slots)
}

// KeyBindings implements Component.
func (h *HBox) KeyBindings() []KeyBind { return nil }

// SetSize implements Component.
func (h *HBox) SetSize(width, height int) {
	h.width = width
	h.height = height
	slots := resolveHBoxSlots(h)
	for i, item := range h.Items {
		c := unwrapFlexComponent(item)
		if c == nil || i >= len(slots) {
			continue
		}
		childH := h.height
		c.SetSize(slots[i].fixedSz, childH)
	}
}

// Focused implements Component.
func (h *HBox) Focused() bool { return h.focused }

// SetFocused implements Component.
func (h *HBox) SetFocused(focused bool) {
	h.focused = focused
	for _, item := range h.Items {
		if c := unwrapFlexComponent(item); c != nil {
			c.SetFocused(focused)
		}
	}
}

// SetTheme implements Themed by propagating to all children.
func (h *HBox) SetTheme(th Theme) {
	for _, item := range h.Items {
		if c := unwrapFlexComponent(item); c != nil {
			applyTheme(c, th)
		}
	}
}

// Init implements Component.
func (v *VBox) Init() tea.Cmd {
	var cmds []tea.Cmd
	for _, item := range v.Items {
		if c := unwrapFlexComponent(item); c != nil {
			if cmd := c.Init(); cmd != nil {
				cmds = append(cmds, cmd)
			}
		}
	}
	return tea.Batch(cmds...)
}

// Update implements Component.
func (v *VBox) Update(msg tea.Msg, ctx Context) (Component, tea.Cmd) {
	var cmds []tea.Cmd
	for i, item := range v.Items {
		c := unwrapFlexComponent(item)
		if c == nil {
			continue
		}
		updated, cmd := c.Update(msg, ctx)
		v.Items[i] = rewrapFlexComponent(item, updated)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}
	return v, tea.Batch(cmds...)
}

// View implements Component.
func (v *VBox) View() string {
	if len(v.Items) == 0 {
		return ""
	}
	slots := resolveVBoxSlots(v)
	return renderVBox(v, slots)
}

// KeyBindings implements Component.
func (v *VBox) KeyBindings() []KeyBind { return nil }

// SetSize implements Component.
func (v *VBox) SetSize(width, height int) {
	v.width = width
	v.height = height
	slots := resolveVBoxSlots(v)
	for i, item := range v.Items {
		c := unwrapFlexComponent(item)
		if c == nil || i >= len(slots) {
			continue
		}
		childW := v.width
		c.SetSize(childW, slots[i].fixedSz)
	}
}

// Focused implements Component.
func (v *VBox) Focused() bool { return v.focused }

// SetFocused implements Component.
func (v *VBox) SetFocused(focused bool) {
	v.focused = focused
	for _, item := range v.Items {
		if c := unwrapFlexComponent(item); c != nil {
			c.SetFocused(focused)
		}
	}
}

// SetTheme implements Themed by propagating to all children.
func (v *VBox) SetTheme(th Theme) {
	for _, item := range v.Items {
		if c := unwrapFlexComponent(item); c != nil {
			applyTheme(c, th)
		}
	}
}

// ----------------------------------------------------------------------------
// Internal helpers
// ----------------------------------------------------------------------------

// unwrapFlexComponent returns the underlying Component regardless of whether
// item is a plain Component, a Sized wrapper, or a Flex wrapper.
func unwrapFlexComponent(item Component) Component {
	switch v := item.(type) {
	case Sized:
		return v.C
	case *Sized:
		return v.C
	case Flex:
		return v.C
	case *Flex:
		return v.C
	default:
		return item
	}
}

// rewrapFlexComponent puts updated back into the same wrapper type as original.
func rewrapFlexComponent(original, updated Component) Component {
	switch v := original.(type) {
	case Sized:
		v.C = updated
		return v
	case *Sized:
		v.C = updated
		return v
	case Flex:
		v.C = updated
		return v
	case *Flex:
		v.C = updated
		return v
	default:
		return updated
	}
}

// toFlexItem converts a Component (possibly Sized/Flex wrapped) to a flexItem.
func toFlexItem(item Component) flexItem {
	switch v := item.(type) {
	case Sized:
		return flexItem{c: v.C, fixedSz: v.W}
	case *Sized:
		return flexItem{c: v.C, fixedSz: v.W}
	case Flex:
		g := v.Grow
		if g < 1 {
			g = 1
		}
		return flexItem{c: v.C, grow: g}
	case *Flex:
		g := v.Grow
		if g < 1 {
			g = 1
		}
		return flexItem{c: v.C, grow: g}
	default:
		// Plain component — auto sizing: treated as Flex{Grow:1}
		return flexItem{c: item, grow: 1}
	}
}

// resolveHBoxSlots computes the pixel widths for each HBox child.
func resolveHBoxSlots(h *HBox) []flexItem {
	n := len(h.Items)
	items := make([]flexItem, n)
	for i, it := range h.Items {
		items[i] = toFlexItem(it)
	}

	totalGap := 0
	if n > 1 {
		totalGap = h.Gap * (n - 1)
	}
	available := h.width - totalGap
	if available < 0 {
		available = 0
	}

	totalGrow := 0
	fixed := 0
	for _, it := range items {
		if it.grow > 0 {
			totalGrow += it.grow
		} else {
			fixed += it.fixedSz
		}
	}

	flexSpace := available - fixed
	if flexSpace < 0 {
		flexSpace = 0
	}

	if totalGrow > 0 {
		distributed := 0
		for i, it := range items {
			if it.grow > 0 {
				w := (flexSpace * it.grow) / totalGrow
				items[i].fixedSz = w
				distributed += w
			}
		}
		rem := flexSpace - distributed
		if rem > 0 {
			for i := len(items) - 1; i >= 0; i-- {
				if items[i].grow > 0 {
					items[i].fixedSz += rem
					break
				}
			}
		}
	}

	return items
}

// resolveVBoxSlots computes the row heights for each VBox child.
func resolveVBoxSlots(v *VBox) []flexItem {
	n := len(v.Items)
	items := make([]flexItem, n)
	for i, it := range v.Items {
		items[i] = toFlexItem(it)
	}

	totalGap := 0
	if n > 1 {
		totalGap = v.Gap * (n - 1)
	}
	available := v.height - totalGap
	if available < 0 {
		available = 0
	}

	totalGrow := 0
	fixed := 0
	for _, it := range items {
		if it.grow > 0 {
			totalGrow += it.grow
		} else {
			fixed += it.fixedSz
		}
	}

	flexSpace := available - fixed
	if flexSpace < 0 {
		flexSpace = 0
	}

	if totalGrow > 0 {
		distributed := 0
		for i, it := range items {
			if it.grow > 0 {
				h := (flexSpace * it.grow) / totalGrow
				items[i].fixedSz = h
				distributed += h
			}
		}
		rem := flexSpace - distributed
		if rem > 0 {
			for i := len(items) - 1; i >= 0; i-- {
				if items[i].grow > 0 {
					items[i].fixedSz += rem
					break
				}
			}
		}
	}

	return items
}

// renderHBox renders an HBox's children into a single joined string.
func renderHBox(h *HBox, slots []flexItem) string {
	views := make([]string, 0, len(slots))
	for _, slot := range slots {
		if slot.c == nil {
			continue
		}
		childH := h.height
		slot.c.SetSize(slot.fixedSz, childH)
		raw := slot.c.View()
		view := alignCrossHBox(raw, slot.fixedSz, childH, h.Align)
		views = append(views, view)
	}

	if len(views) == 0 {
		return ""
	}

	gap := strings.Repeat(" ", h.Gap)
	return applyJustifyHBox(views, h.Justify, h.width, h.Gap, gap)
}

// renderVBox renders a VBox's children into a single joined string.
func renderVBox(v *VBox, slots []flexItem) string {
	views := make([]string, 0, len(slots))
	for _, slot := range slots {
		if slot.c == nil {
			continue
		}
		childW := v.width
		slot.c.SetSize(childW, slot.fixedSz)
		raw := slot.c.View()
		view := alignCrossVBox(raw, childW, v.Align)
		views = append(views, view)
	}

	if len(views) == 0 {
		return ""
	}

	return applyJustifyVBox(views, v.Justify, v.width, v.height, v.Gap)
}

// alignCrossHBox handles vertical (cross-axis) alignment of a single HBox child cell.
func alignCrossHBox(content string, w, totalH int, align FlexAlign) string {
	lines := strings.Split(content, "\n")
	padded := make([]string, len(lines))
	for i, l := range lines {
		padded[i] = lipgloss.NewStyle().Width(w).Render(l)
	}

	switch align {
	case FlexAlignStretch:
		for len(padded) < totalH {
			padded = append(padded, lipgloss.NewStyle().Width(w).Render(""))
		}
	case FlexAlignCenter:
		pad := totalH - len(padded)
		if pad > 0 {
			top := pad / 2
			bottom := pad - top
			empty := lipgloss.NewStyle().Width(w).Render("")
			topLines := make([]string, top)
			for i := range topLines {
				topLines[i] = empty
			}
			botLines := make([]string, bottom)
			for i := range botLines {
				botLines[i] = empty
			}
			padded = append(topLines, append(padded, botLines...)...) //nolint:gocritic // intentional multi-append for vertical padding
		}
	case FlexAlignEnd:
		pad := totalH - len(padded)
		if pad > 0 {
			empty := lipgloss.NewStyle().Width(w).Render("")
			topLines := make([]string, pad)
			for i := range topLines {
				topLines[i] = empty
			}
			padded = append(topLines, padded...)
		}
	default: // FlexAlignStart
	}

	return strings.Join(padded, "\n")
}

// alignCrossVBox handles horizontal (cross-axis) alignment of a single VBox child.
func alignCrossVBox(content string, totalW int, align FlexAlign) string {
	lines := strings.Split(content, "\n")
	result := make([]string, len(lines))
	for i, l := range lines {
		switch align {
		case FlexAlignCenter:
			result[i] = lipgloss.NewStyle().Width(totalW).Align(lipgloss.Center).Render(l)
		case FlexAlignEnd:
			result[i] = lipgloss.NewStyle().Width(totalW).Align(lipgloss.Right).Render(l)
		default: // Start, Stretch
			result[i] = lipgloss.NewStyle().Width(totalW).Render(l)
		}
	}
	return strings.Join(result, "\n")
}

// applyJustifyHBox places the column views side-by-side with justify distribution.
func applyJustifyHBox(views []string, justify FlexJustify, totalW, gap int, gapStr string) string {
	n := len(views)
	if n == 0 {
		return ""
	}

	lineSlices := make([][]string, n)
	maxLines := 0
	for i, v := range views {
		lineSlices[i] = strings.Split(v, "\n")
		if len(lineSlices[i]) > maxLines {
			maxLines = len(lineSlices[i])
		}
	}

	for i, ls := range lineSlices {
		for len(ls) < maxLines {
			ls = append(ls, "")
		}
		lineSlices[i] = ls
	}

	usedW := 0
	for _, v := range views {
		if len(v) > 0 {
			firstLine := strings.Split(v, "\n")[0]
			usedW += lipgloss.Width(firstLine)
		}
	}
	if n > 1 {
		usedW += gap * (n - 1)
	}
	spare := totalW - usedW
	if spare < 0 {
		spare = 0
	}

	var leading, between int
	switch justify {
	case FlexJustifyCenter:
		leading = spare / 2
	case FlexJustifyEnd:
		leading = spare
	case FlexJustifySpaceBetween:
		if n > 1 {
			between = spare / (n - 1)
		} else {
			leading = spare / 2
		}
	case FlexJustifySpaceAround:
		if n > 0 {
			around := spare / n
			leading = around / 2
			between = around
		}
	default: // FlexJustifyStart
	}

	leadingStr := strings.Repeat(" ", leading)
	betweenStr := strings.Repeat(" ", between)

	result := make([]string, maxLines)
	for row := 0; row < maxLines; row++ {
		var sb strings.Builder
		sb.WriteString(leadingStr)
		for i, ls := range lineSlices {
			line := ""
			if row < len(ls) {
				line = ls[row]
			}
			sb.WriteString(line)
			if i < n-1 {
				sb.WriteString(gapStr)
				sb.WriteString(betweenStr)
			}
		}
		result[row] = sb.String()
	}

	return strings.Join(result, "\n")
}

// applyJustifyVBox places child views vertically with justify distribution.
func applyJustifyVBox(views []string, justify FlexJustify, totalW, totalH, gap int) string {
	n := len(views)
	if n == 0 {
		return ""
	}

	usedLines := 0
	for _, v := range views {
		usedLines += strings.Count(v, "\n") + 1
	}
	if n > 1 {
		usedLines += gap * (n - 1)
	}
	spare := totalH - usedLines
	if spare < 0 {
		spare = 0
	}

	blankLine := strings.Repeat(" ", totalW)

	var leading, between int
	switch justify {
	case FlexJustifyCenter:
		leading = spare / 2
	case FlexJustifyEnd:
		leading = spare
	case FlexJustifySpaceBetween:
		if n > 1 {
			between = spare / (n - 1)
		} else {
			leading = spare / 2
		}
	case FlexJustifySpaceAround:
		if n > 0 {
			around := spare / n
			leading = around / 2
			between = around
		}
	default: // FlexJustifyStart
	}

	gapLines := make([]string, gap)
	for i := range gapLines {
		gapLines[i] = blankLine
	}
	betweenLines := make([]string, between)
	for i := range betweenLines {
		betweenLines[i] = blankLine
	}

	var parts []string
	for range leading {
		parts = append(parts, blankLine)
	}
	for i, v := range views {
		parts = append(parts, v)
		if i < n-1 {
			parts = append(parts, gapLines...)
			parts = append(parts, betweenLines...)
		}
	}

	return strings.Join(parts, "\n")
}
