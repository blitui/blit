package blit

import (
	"sync"

	"github.com/charmbracelet/lipgloss"
)

// StyleSet groups the visual variants of a named component style.
// Components select the appropriate variant based on their current interaction state.
type StyleSet struct {
	Base     lipgloss.Style
	Hover    lipgloss.Style
	Focus    lipgloss.Style
	Disabled lipgloss.Style
	Active   lipgloss.Style
}

// styleRegistry holds named StyleSets, keyed by component.variant strings such
// as "button.primary" or "row.cursor". It is embedded inside Theme.
type styleRegistry struct {
	mu     sync.RWMutex
	styles map[string]StyleSet
}

func newStyleRegistry() *styleRegistry {
	return &styleRegistry{styles: make(map[string]StyleSet)}
}

// set stores (or replaces) the StyleSet for the given name.
func (r *styleRegistry) set(name string, s StyleSet) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.styles[name] = s
}

// get returns the StyleSet for the given name and whether it was found.
func (r *styleRegistry) get(name string) (StyleSet, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.styles[name]
	return s, ok
}

// --- Theme integration ---

// Style returns the StyleSet registered under name. If no override has been
// registered the built-in default derived from the theme's color tokens is
// returned. The second return value is false only when name is completely
// unknown (neither registered nor a built-in).
func (t *Theme) Style(name string) (StyleSet, bool) {
	if t.registry != nil {
		if s, ok := t.registry.get(name); ok {
			return s, true
		}
	}
	return builtinStyle(name, *t)
}

// RegisterStyle stores a custom StyleSet override for name on this Theme
// instance. Subsequent calls to Style(name) will return this override.
func (t *Theme) RegisterStyle(name string, s StyleSet) {
	if t.registry == nil {
		t.registry = newStyleRegistry()
	}
	t.registry.set(name, s)
}

// builtinStyle returns the default StyleSet for the well-known named styles
// that ship with blit. Returns (StyleSet{}, false) for unknown names.
func builtinStyle(name string, t Theme) (StyleSet, bool) {
	switch name {
	case "button.primary":
		base := lipgloss.NewStyle().
			Background(lipgloss.Color(t.Accent)).
			Foreground(lipgloss.Color(t.TextInverse)).
			Padding(0, 1)
		return StyleSet{
			Base:     base,
			Hover:    base.Background(lipgloss.Color(t.Cursor)),
			Focus:    base.Bold(true),
			Disabled: lipgloss.NewStyle().Foreground(lipgloss.Color(t.Muted)).Padding(0, 1),
			Active:   base.Background(lipgloss.Color(t.Positive)),
		}, true

	case "button.ghost":
		base := lipgloss.NewStyle().
			Foreground(lipgloss.Color(t.Accent)).
			Padding(0, 1)
		return StyleSet{
			Base:     base,
			Hover:    base.Foreground(lipgloss.Color(t.Cursor)),
			Focus:    base.Bold(true).Underline(true),
			Disabled: lipgloss.NewStyle().Foreground(lipgloss.Color(t.Muted)).Padding(0, 1),
			Active:   base.Foreground(lipgloss.Color(t.Positive)),
		}, true

	case "input.text":
		base := lipgloss.NewStyle().Foreground(lipgloss.Color(t.Text))
		return StyleSet{
			Base:     base,
			Hover:    base,
			Focus:    base,
			Disabled: lipgloss.NewStyle().Foreground(lipgloss.Color(t.Muted)),
			Active:   base,
		}, true

	case "input.focus":
		base := lipgloss.NewStyle().Foreground(lipgloss.Color(t.Accent))
		return StyleSet{
			Base:     lipgloss.NewStyle().Foreground(lipgloss.Color(t.Text)),
			Hover:    base,
			Focus:    base.Bold(true),
			Disabled: lipgloss.NewStyle().Foreground(lipgloss.Color(t.Muted)),
			Active:   base,
		}, true

	case "label.hint":
		base := lipgloss.NewStyle().Foreground(lipgloss.Color(t.Muted))
		return StyleSet{
			Base:     base,
			Hover:    base,
			Focus:    base,
			Disabled: base,
			Active:   base,
		}, true

	case "badge.info":
		base := lipgloss.NewStyle().
			Background(lipgloss.Color(t.Accent)).
			Foreground(lipgloss.Color(t.TextInverse)).
			Padding(0, 1)
		return StyleSet{
			Base: base, Hover: base, Focus: base, Disabled: base, Active: base,
		}, true

	case "badge.warn":
		base := lipgloss.NewStyle().
			Background(lipgloss.Color(t.Flash)).
			Foreground(lipgloss.Color(t.TextInverse)).
			Padding(0, 1)
		return StyleSet{
			Base: base, Hover: base, Focus: base, Disabled: base, Active: base,
		}, true

	case "badge.error":
		base := lipgloss.NewStyle().
			Background(lipgloss.Color(t.Negative)).
			Foreground(lipgloss.Color(t.TextInverse)).
			Padding(0, 1)
		return StyleSet{
			Base: base, Hover: base, Focus: base, Disabled: base, Active: base,
		}, true

	case "row.cursor":
		base := lipgloss.NewStyle().
			Background(lipgloss.Color(t.Cursor)).
			Foreground(lipgloss.Color(t.TextInverse))
		return StyleSet{
			Base: base, Hover: base, Focus: base, Disabled: base, Active: base,
		}, true

	case "row.selected":
		base := lipgloss.NewStyle().
			Background(lipgloss.Color(t.Accent)).
			Foreground(lipgloss.Color(t.TextInverse))
		return StyleSet{
			Base: base, Hover: base, Focus: base, Disabled: base, Active: base,
		}, true
	}

	return StyleSet{}, false
}
