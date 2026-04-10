package blit_test

import (
	"testing"

	"github.com/blitui/blit"
)

// TestContextZeroValue verifies that a zero-value Context is safe to use.
func TestContextZeroValue(t *testing.T) {
	var ctx blit.Context

	// Size fields should be zero
	if ctx.Size.Width != 0 || ctx.Size.Height != 0 {
		t.Errorf("zero Context.Size = %v, want 0x0", ctx.Size)
	}

	// Focus fields should be zero-valued
	if ctx.Focus.Index != 0 {
		t.Errorf("zero Context.Focus.Index = %d, want 0", ctx.Focus.Index)
	}
	if ctx.Focus.Name != "" {
		t.Errorf("zero Context.Focus.Name = %q, want empty", ctx.Focus.Name)
	}

	// Theme should be zero-valued (no panic on access)
	if ctx.Theme.Text != "" {
		t.Errorf("zero Context.Theme.Text = %v, want empty", ctx.Theme.Text)
	}

	// Hotkeys and Logger should be nil-safe
	if ctx.Hotkeys != nil {
		t.Error("zero Context.Hotkeys should be nil")
	}
	if ctx.Logger != nil {
		t.Error("zero Context.Logger should be nil")
	}
	if ctx.Clock != nil {
		t.Error("zero Context.Clock should be nil")
	}
}

// TestContextSizeAccessible verifies Size fields can be set and read.
func TestContextSizeAccessible(t *testing.T) {
	ctx := blit.Context{
		Size: blit.Size{Width: 120, Height: 40},
	}
	if ctx.Size.Width != 120 {
		t.Errorf("Width = %d, want 120", ctx.Size.Width)
	}
	if ctx.Size.Height != 40 {
		t.Errorf("Height = %d, want 40", ctx.Size.Height)
	}
}

// TestContextFocusAccessible verifies Focus fields can be set and read.
func TestContextFocusAccessible(t *testing.T) {
	ctx := blit.Context{
		Focus: blit.Focus{Index: 2, Name: "sidebar"},
	}
	if ctx.Focus.Index != 2 {
		t.Errorf("Focus.Index = %d, want 2", ctx.Focus.Index)
	}
	if ctx.Focus.Name != "sidebar" {
		t.Errorf("Focus.Name = %q, want %q", ctx.Focus.Name, "sidebar")
	}
}
