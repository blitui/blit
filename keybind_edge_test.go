package blit

import "testing"

// TestRegistryEmpty verifies that an empty registry returns empty results.
func TestRegistryEmpty(t *testing.T) {
	r := newRegistry()
	if len(r.all()) != 0 {
		t.Errorf("empty registry all() = %d, want 0", len(r.all()))
	}
	if len(r.grouped()) != 0 {
		t.Errorf("empty registry grouped() = %d, want 0", len(r.grouped()))
	}
	if len(r.conflicts()) != 0 {
		t.Errorf("empty registry conflicts() = %d, want 0", len(r.conflicts()))
	}
}

// TestRegistryAddBindingsReplace verifies that adding bindings for the
// same source replaces the previous set.
func TestRegistryAddBindingsReplace(t *testing.T) {
	r := newRegistry()
	r.addBindings("comp", []KeyBind{{Key: "a", Label: "A", Group: "G"}})
	r.addBindings("comp", []KeyBind{{Key: "b", Label: "B", Group: "G"}})

	all := r.all()
	if len(all) != 1 {
		t.Fatalf("expected 1 binding after replace, got %d", len(all))
	}
	if all[0].Key != "b" {
		t.Errorf("expected key 'b', got %q", all[0].Key)
	}
}

// TestRegistryInsertionOrder verifies that bindings come out in source
// insertion order.
func TestRegistryInsertionOrder(t *testing.T) {
	r := newRegistry()
	r.addBindings("alpha", []KeyBind{{Key: "a", Group: "G"}})
	r.addBindings("beta", []KeyBind{{Key: "b", Group: "G"}})
	r.addBindings("gamma", []KeyBind{{Key: "c", Group: "G"}})

	all := r.all()
	if len(all) != 3 {
		t.Fatalf("expected 3 bindings, got %d", len(all))
	}
	expected := []string{"a", "b", "c"}
	for i, want := range expected {
		if all[i].Key != want {
			t.Errorf("all()[%d].Key = %q, want %q", i, all[i].Key, want)
		}
	}
}

// TestRegistryNoConflicts verifies that unique keys produce no conflicts.
func TestRegistryNoConflicts(t *testing.T) {
	r := newRegistry()
	r.addBindings("a", []KeyBind{{Key: "x", Group: "G"}})
	r.addBindings("b", []KeyBind{{Key: "y", Group: "G"}})

	if len(r.conflicts()) != 0 {
		t.Errorf("expected no conflicts, got %v", r.conflicts())
	}
}

// TestRegistryGroupedStripsHandlers verifies that grouped output does not
// leak Handler funcs.
func TestRegistryGroupedStripsHandlers(t *testing.T) {
	r := newRegistry()
	r.addBindings("test", []KeyBind{
		{Key: "q", Label: "Quit", Group: "OTHER", Handler: func() {}},
	})

	groups := r.grouped()
	if len(groups) != 1 {
		t.Fatalf("expected 1 group, got %d", len(groups))
	}
	if groups[0].Bindings[0].Handler != nil {
		t.Error("grouped() should strip Handler from bindings")
	}
}

// TestRegistryMultipleConflicts verifies detection of multiple conflicting keys.
func TestRegistryMultipleConflicts(t *testing.T) {
	r := newRegistry()
	r.addBindings("a", []KeyBind{{Key: "s"}, {Key: "x"}})
	r.addBindings("b", []KeyBind{{Key: "s"}, {Key: "x"}})

	conflicts := r.conflicts()
	if len(conflicts) != 2 {
		t.Errorf("expected 2 conflicts, got %d: %v", len(conflicts), conflicts)
	}
}
