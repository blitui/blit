package blit

import "testing"

func TestRegistryCollect(t *testing.T) {
	r := newRegistry()
	r.addBindings("table", []KeyBind{
		{Key: "up", Label: "Move up", Group: "NAVIGATION"},
		{Key: "down", Label: "Move down", Group: "NAVIGATION"},
		{Key: "s", Label: "Sort", Group: "DATA"},
	})
	r.addBindings("global", []KeyBind{
		{Key: "q", Label: "Quit", Group: "OTHER"},
		{Key: "?", Label: "Help", Group: "OTHER"},
	})
	all := r.all()
	if len(all) != 5 {
		t.Fatalf("expected 5 bindings, got %d", len(all))
	}
}

func TestRegistryGrouped(t *testing.T) {
	r := newRegistry()
	r.addBindings("test", []KeyBind{
		{Key: "up", Label: "Move up", Group: "NAVIGATION"},
		{Key: "s", Label: "Sort", Group: "DATA"},
		{Key: "q", Label: "Quit", Group: "OTHER"},
	})
	groups := r.grouped()
	if len(groups) != 3 {
		t.Fatalf("expected 3 groups, got %d", len(groups))
	}
	if groups[0].Name != "NAVIGATION" {
		t.Errorf("expected first group NAVIGATION, got %s", groups[0].Name)
	}
}

func TestRegistryConflict(t *testing.T) {
	r := newRegistry()
	r.addBindings("comp1", []KeyBind{
		{Key: "s", Label: "Sort", Group: "DATA"},
	})
	r.addBindings("comp2", []KeyBind{
		{Key: "s", Label: "Search", Group: "DATA"},
	})
	conflicts := r.conflicts()
	if len(conflicts) != 1 {
		t.Fatalf("expected 1 conflict, got %d", len(conflicts))
	}
	if conflicts[0] != "s" {
		t.Errorf("expected conflict on 's', got %s", conflicts[0])
	}
}
