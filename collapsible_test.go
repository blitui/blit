package blit

import (
	"strings"
	"testing"
)

func TestCollapsibleSectionDefaultExpanded(t *testing.T) {
	s := NewCollapsibleSection("VOL SPIKES")
	if s.Collapsed {
		t.Error("section should be expanded by default")
	}
}

func TestCollapsibleSectionToggle(t *testing.T) {
	s := NewCollapsibleSection("VOL SPIKES")
	s.Toggle()
	if !s.Collapsed {
		t.Error("section should be collapsed after toggle")
	}
	s.Toggle()
	if s.Collapsed {
		t.Error("section should be expanded after second toggle")
	}
}

func TestCollapsibleSectionRenderExpanded(t *testing.T) {
	s := NewCollapsibleSection("VOL SPIKES")
	theme := DefaultTheme()
	output := s.Render(theme, func() string {
		return "line1\nline2"
	})
	if !strings.Contains(output, "▾") {
		t.Error("expanded section should contain ▾ arrow")
	}
	if !strings.Contains(output, "VOL SPIKES") {
		t.Error("expanded section should contain title")
	}
	if !strings.Contains(output, "line1") {
		t.Error("expanded section should contain content")
	}
}

func TestCollapsibleSectionRenderCollapsed(t *testing.T) {
	s := NewCollapsibleSection("VOL SPIKES")
	s.Toggle()
	theme := DefaultTheme()
	called := false
	output := s.Render(theme, func() string {
		called = true
		return "line1\nline2"
	})
	if called {
		t.Error("contentFunc should not be called when collapsed")
	}
	if !strings.Contains(output, "▸") {
		t.Error("collapsed section should contain ▸ arrow")
	}
	if !strings.Contains(output, "VOL SPIKES") {
		t.Error("collapsed section should contain title")
	}
	if strings.Contains(output, "line1") {
		t.Error("collapsed section should not contain content")
	}
}
