package blit

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
)

func TestRelativeTime(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 4, 8, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		name string
		t    time.Time
		want string
	}{
		{"seconds", now.Add(-30 * time.Second), "30s ago"},
		{"minutes", now.Add(-5 * time.Minute), "5m ago"},
		{"hours", now.Add(-3 * time.Hour), "3h ago"},
		{"days", now.Add(-48 * time.Hour), "2d ago"},
		{"future", now.Add(10 * time.Second), "now"},
		{"zero", now, "0s ago"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RelativeTime(tt.t, now)
			if got != tt.want {
				t.Errorf("RelativeTime() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestOSC8Link(t *testing.T) {
	t.Parallel()
	t.Run("with url", func(t *testing.T) {
		got := OSC8Link("https://example.com", "click me")
		if !strings.Contains(got, "https://example.com") {
			t.Errorf("OSC8Link should contain URL, got %q", got)
		}
		if !strings.Contains(got, "click me") {
			t.Errorf("OSC8Link should contain text, got %q", got)
		}
		if !strings.HasPrefix(got, "\x1b]8;;") {
			t.Errorf("OSC8Link should start with OSC8 escape, got %q", got)
		}
	})

	t.Run("empty url", func(t *testing.T) {
		got := OSC8Link("", "plain text")
		if got != "plain text" {
			t.Errorf("OSC8Link with empty URL should return plain text, got %q", got)
		}
	})
}

func TestSparkline(t *testing.T) {
	t.Parallel()
	t.Run("basic", func(t *testing.T) {
		data := []float64{1, 2, 3, 4, 5, 4, 3, 2, 1}
		result, n := Sparkline(data, 20, nil)
		if n != 9 {
			t.Errorf("Sparkline n = %d, want 9", n)
		}
		if result == "" {
			t.Error("Sparkline should return non-empty string")
		}
	})

	t.Run("too few points", func(t *testing.T) {
		result, n := Sparkline([]float64{1}, 20, nil)
		if n != 0 || result != "" {
			t.Errorf("Sparkline with <2 points should return empty, got n=%d result=%q", n, result)
		}
	})

	t.Run("zero width", func(t *testing.T) {
		result, n := Sparkline([]float64{1, 2, 3}, 0, nil)
		if n != 0 || result != "" {
			t.Errorf("Sparkline with zero width should return empty, got n=%d", n)
		}
	})

	t.Run("samples when exceeding width", func(t *testing.T) {
		data := make([]float64, 100)
		for i := range data {
			data[i] = float64(i)
		}
		_, n := Sparkline(data, 10, nil)
		if n != 10 {
			t.Errorf("Sparkline should sample to maxWidth, got n=%d, want 10", n)
		}
	})

	t.Run("flat data", func(t *testing.T) {
		data := []float64{5, 5, 5, 5}
		result, n := Sparkline(data, 10, nil)
		if n != 4 {
			t.Errorf("Sparkline n = %d, want 4", n)
		}
		if result == "" {
			t.Error("Sparkline should handle flat data")
		}
	})

	t.Run("mono mode", func(t *testing.T) {
		data := []float64{1, 3, 2, 4}
		result, _ := Sparkline(data, 10, &SparklineOpts{Mono: true})
		if result == "" {
			t.Error("Sparkline mono mode should return non-empty string")
		}
	})
}

func TestTruncateUtil(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    string
		maxWidth int
		want     string
	}{
		{"fits", "hello", 10, "hello"},
		{"exact", "hello", 5, "hello"},
		{"truncated", "hello world", 8, "hello w…"},
		{"very short", "abcdef", 2, "a…"},
		{"single char", "abcdef", 1, "…"},
		{"empty input", "", 5, ""},
		{"zero width", "hello", 0, "…"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := Truncate(tc.input, tc.maxWidth)
			if lipgloss.Width(got) > tc.maxWidth && tc.maxWidth > 0 {
				t.Errorf("Truncate(%q, %d) width=%d exceeds max", tc.input, tc.maxWidth, lipgloss.Width(got))
			}
			if got != tc.want {
				t.Errorf("Truncate(%q, %d) = %q, want %q", tc.input, tc.maxWidth, got, tc.want)
			}
		})
	}
}

func TestDivider(t *testing.T) {
	t.Parallel()
	theme := DefaultTheme()
	d := Divider(20, theme)
	if !strings.Contains(d, "─") {
		t.Errorf("Divider should contain ─ chars, got %q", d)
	}

	d0 := Divider(0, theme)
	if strings.Contains(d0, "─") {
		t.Errorf("Divider(0) should be empty, got %q", d0)
	}
}

func TestBadge(t *testing.T) {
	t.Parallel()
	t.Run("basic", func(t *testing.T) {
		b := Badge("INFO", lipgloss.Color("#00ff00"), false)
		if !strings.Contains(b, "INFO") {
			t.Errorf("Badge should contain text, got %q", b)
		}
	})

	t.Run("bold", func(t *testing.T) {
		b := Badge("WARN", lipgloss.Color("#ff0000"), true)
		if !strings.Contains(b, "WARN") {
			t.Errorf("Bold badge should contain text, got %q", b)
		}
	})

	t.Run("empty text", func(t *testing.T) {
		b := Badge("", lipgloss.Color("#000000"), false)
		if strings.Contains(b, "INFO") {
			t.Error("Empty badge should not contain text")
		}
	})
}
