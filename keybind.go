package tuikit

import tea "github.com/charmbracelet/bubbletea"

// KeyBind defines a keyboard shortcut with its label and group.
// Components return slices of KeyBind from their KeyBindings() method.
// The App collects these for dispatch and auto-generates the help screen.
type KeyBind struct {
	Key        string         // Key name: "q", "ctrl+c", "?", "/", "up", "down"
	Label      string         // Human-readable label: "Quit", "Search", "Help"
	Group      string         // Grouping for help screen: "NAVIGATION", "DATA", "OTHER"
	Handler    func()         // Optional handler for app-level keybindings (nil for component bindings)
	HandlerCmd func() tea.Cmd // Optional handler that returns a tea.Cmd
}

// KeyGroup is a named group of keybindings for help screen rendering.
type KeyGroup struct {
	Name     string
	Bindings []KeyBind
}

// Registry collects keybindings from all components and provides
// grouped access for the help screen and conflict detection.
type Registry struct {
	sources map[string][]KeyBind // component name -> bindings
	order   []string             // insertion order of sources
}

func newRegistry() *Registry {
	return &Registry{
		sources: make(map[string][]KeyBind),
	}
}

// addBindings registers keybindings from a named source (component or "global").
func (r *Registry) addBindings(source string, bindings []KeyBind) {
	if _, exists := r.sources[source]; !exists {
		r.order = append(r.order, source)
	}
	r.sources[source] = bindings
}

// all returns every registered keybinding in insertion order.
func (r *Registry) all() []KeyBind {
	var result []KeyBind
	for _, src := range r.order {
		result = append(result, r.sources[src]...)
	}
	return result
}

// grouped returns keybindings organized by Group, preserving insertion order.
// Handler funcs are stripped from the output to keep help screen data clean.
func (r *Registry) grouped() []KeyGroup {
	seen := make(map[string]int)
	var groups []KeyGroup
	for _, kb := range r.all() {
		entry := KeyBind{Key: kb.Key, Label: kb.Label, Group: kb.Group}
		if idx, ok := seen[kb.Group]; ok {
			groups[idx].Bindings = append(groups[idx].Bindings, entry)
		} else {
			seen[kb.Group] = len(groups)
			groups = append(groups, KeyGroup{
				Name:     kb.Group,
				Bindings: []KeyBind{entry},
			})
		}
	}
	return groups
}

// conflicts returns key names that are bound by more than one source.
func (r *Registry) conflicts() []string {
	counts := make(map[string]int)
	for _, src := range r.order {
		for _, kb := range r.sources[src] {
			counts[kb.Key]++
		}
	}
	var result []string
	for key, count := range counts {
		if count > 1 {
			result = append(result, key)
		}
	}
	return result
}
