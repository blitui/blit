package blit_test

import (
	"strings"
	"testing"

	blit "github.com/blitui/blit"
)

func TestProgressBar_NewAndDefaults(t *testing.T) {
	pb := blit.NewProgressBar(blit.ProgressBarOpts{}, 0.5)
	if pb.Value() != 0.5 {
		t.Fatalf("Value() = %f, want 0.5", pb.Value())
	}
}

func TestProgressBar_ClampValue(t *testing.T) {
	pb := blit.NewProgressBar(blit.ProgressBarOpts{}, -0.5)
	if pb.Value() != 0 {
		t.Fatalf("negative value should clamp to 0, got %f", pb.Value())
	}

	pb = blit.NewProgressBar(blit.ProgressBarOpts{}, 1.5)
	if pb.Value() != 1 {
		t.Fatalf("value > 1 should clamp to 1, got %f", pb.Value())
	}
}

func TestProgressBar_SetValue(t *testing.T) {
	pb := blit.NewProgressBar(blit.ProgressBarOpts{}, 0)
	pb.SetValue(0.75)
	if pb.Value() != 0.75 {
		t.Fatalf("Value() = %f, want 0.75", pb.Value())
	}

	pb.SetValue(2.0)
	if pb.Value() != 1.0 {
		t.Fatalf("Value() = %f, want 1.0 (clamped)", pb.Value())
	}
}

func TestProgressBar_Increment(t *testing.T) {
	pb := blit.NewProgressBar(blit.ProgressBarOpts{}, 0.5)
	pb.Increment(0.3)
	if pb.Value() != 0.8 {
		t.Fatalf("Value() = %f, want 0.8", pb.Value())
	}

	// Increment past 1.0 should clamp.
	pb.Increment(0.5)
	if pb.Value() != 1.0 {
		t.Fatalf("Value() = %f, want 1.0 (clamped)", pb.Value())
	}

	// Negative increment.
	pb.Increment(-0.5)
	if pb.Value() != 0.5 {
		t.Fatalf("Value() = %f, want 0.5", pb.Value())
	}
}

func TestProgressBar_ViewEmpty(t *testing.T) {
	pb := blit.NewProgressBar(blit.ProgressBarOpts{}, 0)
	pb.SetTheme(blit.DefaultTheme())
	pb.SetSize(20, 1)

	view := pb.View()
	if view == "" {
		t.Fatal("View() should not be empty")
	}
	g := blit.DefaultGlyphs()
	if strings.Contains(view, g.BarFilled) {
		t.Fatal("0% bar should not contain filled glyphs")
	}
}

func TestProgressBar_ViewFull(t *testing.T) {
	pb := blit.NewProgressBar(blit.ProgressBarOpts{}, 1.0)
	pb.SetTheme(blit.DefaultTheme())
	pb.SetSize(20, 1)

	view := pb.View()
	g := blit.DefaultGlyphs()
	if strings.Contains(view, g.BarEmpty) {
		t.Fatal("100% bar should not contain empty glyphs")
	}
}

func TestProgressBar_ViewWithLabel(t *testing.T) {
	pb := blit.NewProgressBar(blit.ProgressBarOpts{
		Label: "Loading",
	}, 0.5)
	pb.SetTheme(blit.DefaultTheme())
	pb.SetSize(40, 1)

	view := pb.View()
	if !strings.Contains(view, "Loading") {
		t.Fatalf("view should contain label:\n%s", view)
	}
}

func TestProgressBar_ViewWithPercent(t *testing.T) {
	pb := blit.NewProgressBar(blit.ProgressBarOpts{
		ShowPercent: true,
	}, 0.5)
	pb.SetTheme(blit.DefaultTheme())
	pb.SetSize(40, 1)

	view := pb.View()
	if !strings.Contains(view, "50%") {
		t.Fatalf("view should contain 50%%:\n%s", view)
	}
}

func TestProgressBar_ViewWithLabelAndPercent(t *testing.T) {
	pb := blit.NewProgressBar(blit.ProgressBarOpts{
		Label:       "Upload",
		ShowPercent: true,
	}, 0.33)
	pb.SetTheme(blit.DefaultTheme())
	pb.SetSize(50, 1)

	view := pb.View()
	if !strings.Contains(view, "Upload") {
		t.Fatalf("view should contain label:\n%s", view)
	}
	if !strings.Contains(view, "33%") {
		t.Fatalf("view should contain 33%%:\n%s", view)
	}
}

func TestProgressBar_FixedWidth(t *testing.T) {
	pb := blit.NewProgressBar(blit.ProgressBarOpts{
		Width: 10,
	}, 0.5)
	pb.SetTheme(blit.DefaultTheme())
	pb.SetSize(80, 1) // SetSize width should be ignored.

	view := pb.View()
	if view == "" {
		t.Fatal("View() should not be empty")
	}
}

func TestProgressBar_ComponentInterface(t *testing.T) {
	pb := blit.NewProgressBar(blit.ProgressBarOpts{}, 0.5)
	pb.SetTheme(blit.DefaultTheme())
	pb.SetSize(20, 1)

	// Init returns nil.
	if cmd := pb.Init(); cmd != nil {
		t.Fatal("Init() should return nil")
	}

	// KeyBindings returns nil.
	if binds := pb.KeyBindings(); binds != nil {
		t.Fatal("KeyBindings() should return nil")
	}

	// Focused / SetFocused.
	if pb.Focused() {
		t.Fatal("should not be focused by default")
	}
	pb.SetFocused(true)
	if !pb.Focused() {
		t.Fatal("should be focused after SetFocused(true)")
	}

	// Update returns self unchanged.
	updated, cmd := pb.Update(nil, blit.Context{})
	if updated != pb {
		t.Fatal("Update should return same pointer")
	}
	if cmd != nil {
		t.Fatal("Update should return nil cmd")
	}
}

func TestProgressBar_NarrowWidth(t *testing.T) {
	// Even with very narrow width, should not panic.
	pb := blit.NewProgressBar(blit.ProgressBarOpts{
		Label:       "Downloading",
		ShowPercent: true,
	}, 0.5)
	pb.SetTheme(blit.DefaultTheme())
	pb.SetSize(5, 1)

	view := pb.View()
	if view == "" {
		t.Fatal("View() should not be empty even when narrow")
	}
}
