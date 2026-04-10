package blit

// SlotName identifies a named region in the App shell.
//
// Slots are the canonical way to assemble an App as of v0.10. Instead of
// juggling separate options for components, layouts, status bars, and
// overlays, you declare each region by name:
//
//	app := blit.NewApp(
//	    blit.WithSlot(blit.SlotMain, table),
//	    blit.WithSlot(blit.SlotSidebar, panel),
//	    blit.WithSlot(blit.SlotFooter, statusBar),
//	    blit.WithSlot(blit.SlotOverlay, configEditor),
//	)
//
// The legacy options (WithComponent, WithLayout, WithStatusBar, WithOverlay)
// continue to work and are routed through the slot registry internally, so
// both styles can be mixed freely during migration. Legacy mapping:
//
//	WithComponent(name, c)              -> SlotMain (first call) or stacked component list
//	WithLayout(&DualPane{Main:m,Side:s}) -> SlotMain = m, SlotSidebar = s
//	WithStatusBar(l, r)                 -> SlotFooter = StatusBar
//	WithStatusBarSignal(l, r)           -> SlotFooter = StatusBar
//	WithOverlay(name, key, o)           -> SlotOverlay (keyed by name)
//
// Slot names are stable identifiers. Use the SlotXxx constants below rather
// than hard-coding strings so a typo surfaces at compile time.
type SlotName string

const (
	// SlotHeader renders above the main content area.
	SlotHeader SlotName = "header"

	// SlotMain holds the primary component. When combined with SlotSidebar,
	// the two form a DualPane automatically.
	SlotMain SlotName = "main"

	// SlotSidebar holds the side panel, paired with SlotMain in a DualPane.
	SlotSidebar SlotName = "sidebar"

	// SlotFooter renders below the main content, above the toast/overlay
	// areas. The built-in StatusBar is placed here by WithStatusBar.
	SlotFooter SlotName = "footer"

	// SlotOverlay holds stacked modal overlays. Multiple WithSlot calls to
	// SlotOverlay push additional overlays onto the stack in order.
	SlotOverlay SlotName = "overlay"

	// SlotToast is the toast stack region. Currently drives the existing
	// toastManager; adding a custom component here is reserved for a
	// future release.
	SlotToast SlotName = "toast"
)

// slotEntry is one occupant of a named slot. Overlays accumulate; other
// slots keep the most-recently-set value.
type slotEntry struct {
	name      SlotName
	component Component
	// metadata for overlays — empty for non-overlay slots.
	overlayKey  string
	overlayName string
}

// slotRegistry tracks which components are bound to each slot. It is the
// single source of truth the App consults during setup(); legacy options
// write into it first, then setup() materialises the familiar dualPane /
// components / statusBar / namedOverlays fields.
type slotRegistry struct {
	entries map[SlotName][]slotEntry
	order   []SlotName // insertion order, for deterministic iteration in tests
}

func newSlotRegistry() *slotRegistry {
	return &slotRegistry{entries: make(map[SlotName][]slotEntry)}
}

// set replaces any existing entries for a single-occupant slot.
func (r *slotRegistry) set(name SlotName, c Component) {
	if _, ok := r.entries[name]; !ok {
		r.order = append(r.order, name)
	}
	r.entries[name] = []slotEntry{{name: name, component: c}}
}

// push appends an entry to a multi-occupant slot (overlays).
func (r *slotRegistry) push(entry slotEntry) {
	if _, ok := r.entries[entry.name]; !ok {
		r.order = append(r.order, entry.name)
	}
	r.entries[entry.name] = append(r.entries[entry.name], entry)
}

// get returns the single component bound to a slot, or nil.
func (r *slotRegistry) get(name SlotName) Component {
	es := r.entries[name]
	if len(es) == 0 {
		return nil
	}
	return es[0].component
}

// all returns every entry for a slot (used for overlays).
func (r *slotRegistry) all(name SlotName) []slotEntry {
	return r.entries[name]
}

// has reports whether the slot contains at least one entry.
func (r *slotRegistry) has(name SlotName) bool {
	return len(r.entries[name]) > 0
}

// WithSlot binds a component to a named slot. This is the canonical way to
// compose an App shell in v0.10+.
//
// Behaviour per slot:
//   - SlotMain, SlotSidebar, SlotHeader, SlotFooter, SlotToast — single
//     occupant; calling WithSlot twice replaces the previous value.
//   - SlotOverlay — stacks; each WithSlot call pushes another overlay.
//     Use WithSlotOverlay to attach a trigger key.
//
// When both SlotMain and SlotSidebar are populated, the App implicitly
// wires them into a DualPane. Use WithLayout directly if you need to
// customise widths, toggle keys, or side placement.
func WithSlot(name SlotName, c Component) Option {
	return func(a *appModel) {
		if c == nil {
			return
		}
		switch name {
		case SlotOverlay:
			if o, ok := c.(Overlay); ok {
				a.slots.push(slotEntry{name: SlotOverlay, component: c, overlayName: string(SlotOverlay)})
				// Also register through the legacy namedOverlay path so the
				// existing key-dispatch and activation logic keeps working.
				a.namedOverlays = append(a.namedOverlays, namedOverlay{
					name:    string(SlotOverlay),
					overlay: o,
				})
			}
		default:
			a.slots.set(name, c)
		}
	}
}

// WithSlotOverlay pushes an overlay into SlotOverlay with an associated
// trigger key and display name. It is the slot-native equivalent of
// WithOverlay.
func WithSlotOverlay(name, triggerKey string, o Overlay) Option {
	return func(a *appModel) {
		if o == nil {
			return
		}
		a.slots.push(slotEntry{
			name:        SlotOverlay,
			component:   o,
			overlayKey:  triggerKey,
			overlayName: name,
		})
		a.namedOverlays = append(a.namedOverlays, namedOverlay{
			name:       name,
			triggerKey: triggerKey,
			overlay:    o,
		})
	}
}

// materialiseSlots promotes slot registry entries into the concrete
// appModel fields the rest of the App still consults (components,
// dualPane, statusBar). This runs after all options are applied so that
// legacy and slot-based options compose in any order.
func (a *appModel) materialiseSlots() {
	if a.slots == nil {
		return
	}

	main := a.slots.get(SlotMain)
	side := a.slots.get(SlotSidebar)

	// If the caller used slot-based composition for Main/Side and has not
	// already supplied a DualPane, synthesise one so the existing layout
	// path takes over.
	if main != nil && side != nil && a.dualPane == nil {
		a.dualPane = &DualPane{
			Main:         main,
			Side:         side,
			SideWidth:    32,
			MinMainWidth: 40,
			SideRight:    true,
		}
	} else if main != nil && a.dualPane == nil && len(a.components) == 0 {
		// Main-only slot composition -> single focused component.
		a.components = append(a.components, namedComponent{
			name:      string(SlotMain),
			component: main,
		})
	}

	// Footer slot: if the caller passed a StatusBar component via slots,
	// adopt it as the footer.
	if a.statusBar == nil {
		if f := a.slots.get(SlotFooter); f != nil {
			if sb, ok := f.(*StatusBar); ok {
				a.statusBar = sb
			}
		}
	}
}
