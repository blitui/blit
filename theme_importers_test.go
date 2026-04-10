package blit

import (
	"os"
	"testing"
)

func TestFromGogh(t *testing.T) {
	data, err := os.ReadFile("testdata/themes/gogh.json")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	theme, err := FromGogh(data)
	if err != nil {
		t.Fatalf("FromGogh: %v", err)
	}
	if string(theme.Text) != "#f8f8f2" {
		t.Errorf("Text = %q, want #f8f8f2", theme.Text)
	}
	if string(theme.TextInverse) != "#282a36" {
		t.Errorf("TextInverse = %q, want #282a36", theme.TextInverse)
	}
	if string(theme.Negative) != "#ff5555" {
		t.Errorf("Negative = %q, want #ff5555", theme.Negative)
	}
	if string(theme.Positive) != "#50fa7b" {
		t.Errorf("Positive = %q, want #50fa7b", theme.Positive)
	}
}

func TestFromAlacritty(t *testing.T) {
	data, err := os.ReadFile("testdata/themes/alacritty.toml")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	theme, err := FromAlacritty(data)
	if err != nil {
		t.Fatalf("FromAlacritty: %v", err)
	}
	if string(theme.Text) != "#cdd6f4" {
		t.Errorf("Text = %q, want #cdd6f4", theme.Text)
	}
	if string(theme.TextInverse) != "#1e1e2e" {
		t.Errorf("TextInverse = %q, want #1e1e2e", theme.TextInverse)
	}
	if string(theme.Negative) != "#f38ba8" {
		t.Errorf("Negative = %q, want #f38ba8", theme.Negative)
	}
	if string(theme.Positive) != "#a6e3a1" {
		t.Errorf("Positive = %q, want #a6e3a1", theme.Positive)
	}
	if string(theme.Cursor) != "#89b4fa" {
		t.Errorf("Cursor = %q, want #89b4fa", theme.Cursor)
	}
}

func TestFromIterm2(t *testing.T) {
	data, err := os.ReadFile("testdata/themes/iterm2.xml")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	theme, err := FromIterm2(data)
	if err != nil {
		t.Fatalf("FromIterm2: %v", err)
	}
	if theme.Text == "" {
		t.Error("expected non-empty Text color from iTerm2 fixture")
	}
	if theme.TextInverse == "" {
		t.Error("expected non-empty TextInverse from iTerm2 fixture")
	}
}

func TestFromGoghInvalidJSON(t *testing.T) {
	_, err := FromGogh([]byte("not json"))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestFromAlacrittyEmpty(t *testing.T) {
	theme, err := FromAlacritty([]byte(""))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	def := DefaultTheme()
	if theme.Text != def.Text {
		t.Errorf("empty TOML should fall back to default Text, got %q", theme.Text)
	}
}
