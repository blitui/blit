package blit

import (
	"testing"
)

func TestFromGogh_PartialFields(t *testing.T) {
	// Only foreground and color1 set.
	data := []byte(`{"foreground":"#aabbcc","color1":"#ff0000"}`)
	theme, err := FromGogh(data)
	if err != nil {
		t.Fatalf("FromGogh: %v", err)
	}
	if string(theme.Text) != "#aabbcc" {
		t.Errorf("Text = %q, want #aabbcc", theme.Text)
	}
	if string(theme.Negative) != "#ff0000" {
		t.Errorf("Negative = %q, want #ff0000", theme.Negative)
	}
	// Unset fields should keep defaults.
	def := DefaultTheme()
	if theme.Positive != def.Positive {
		t.Errorf("Positive should be default, got %q", theme.Positive)
	}
}

func TestFromGogh_EmptyJSON(t *testing.T) {
	theme, err := FromGogh([]byte(`{}`))
	if err != nil {
		t.Fatalf("FromGogh: %v", err)
	}
	def := DefaultTheme()
	if theme.Text != def.Text {
		t.Errorf("empty JSON should keep default Text")
	}
}

func TestFromAlacritty_0xPrefix(t *testing.T) {
	data := `
[colors.primary]
foreground = 0xAABBCC
background = "0x112233"

[colors.normal]
red = "#ff5555"
green = 0x50fa7b
blue = "#7aa2f7"
magenta = "#bb9af7"
yellow = "#e0af68"
black = "#282a36"
`
	theme, err := FromAlacritty([]byte(data))
	if err != nil {
		t.Fatalf("FromAlacritty: %v", err)
	}
	if string(theme.Text) != "#AABBCC" {
		t.Errorf("Text = %q, want #AABBCC", theme.Text)
	}
	if string(theme.TextInverse) != "#112233" {
		t.Errorf("TextInverse = %q, want #112233", theme.TextInverse)
	}
	if string(theme.Positive) != "#50fa7b" {
		t.Errorf("Positive = %q, want #50fa7b", theme.Positive)
	}
	if string(theme.Border) != "#282a36" {
		t.Errorf("Border = %q, want #282a36", theme.Border)
	}
}

func TestFromAlacritty_CursorSection(t *testing.T) {
	data := `
[colors.cursor]
cursor = "#ff79c6"
`
	theme, err := FromAlacritty([]byte(data))
	if err != nil {
		t.Fatalf("FromAlacritty: %v", err)
	}
	if string(theme.Cursor) != "#ff79c6" {
		t.Errorf("Cursor = %q, want #ff79c6", theme.Cursor)
	}
}

func TestFromAlacritty_CommentsAndBlanks(t *testing.T) {
	data := `
# This is a comment

[colors.primary]
# Another comment
foreground = "#ffffff"

`
	theme, err := FromAlacritty([]byte(data))
	if err != nil {
		t.Fatalf("FromAlacritty: %v", err)
	}
	if string(theme.Text) != "#ffffff" {
		t.Errorf("Text = %q, want #ffffff", theme.Text)
	}
}

func TestFromAlacritty_NonColorValues(t *testing.T) {
	data := `
[colors.primary]
foreground = "not-a-color"
background = "#112233"
`
	theme, err := FromAlacritty([]byte(data))
	if err != nil {
		t.Fatalf("FromAlacritty: %v", err)
	}
	// foreground should keep default since value doesn't start with #.
	if string(theme.TextInverse) != "#112233" {
		t.Errorf("TextInverse = %q, want #112233", theme.TextInverse)
	}
}

func TestFromIterm2_InlineXML(t *testing.T) {
	data := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Foreground Color</key>
	<dict>
		<key>Red Component</key>
		<real>0.8</real>
		<key>Green Component</key>
		<real>0.9</real>
		<key>Blue Component</key>
		<real>1.0</real>
	</dict>
	<key>Background Color</key>
	<dict>
		<key>Red Component</key>
		<real>0.1</real>
		<key>Green Component</key>
		<real>0.1</real>
		<key>Blue Component</key>
		<real>0.15</real>
	</dict>
	<key>Ansi 1 Color</key>
	<dict>
		<key>Red Component</key>
		<real>0.95</real>
		<key>Green Component</key>
		<real>0.3</real>
		<key>Blue Component</key>
		<real>0.35</real>
	</dict>
	<key>Ansi 2 Color</key>
	<dict>
		<key>Red Component</key>
		<real>0.4</real>
		<key>Green Component</key>
		<real>0.85</real>
		<key>Blue Component</key>
		<real>0.5</real>
	</dict>
	<key>Ansi 3 Color</key>
	<dict>
		<key>Red Component</key>
		<real>0.9</real>
		<key>Green Component</key>
		<real>0.75</real>
		<key>Blue Component</key>
		<real>0.3</real>
	</dict>
	<key>Ansi 4 Color</key>
	<dict>
		<key>Red Component</key>
		<real>0.3</real>
		<key>Green Component</key>
		<real>0.5</real>
		<key>Blue Component</key>
		<real>0.9</real>
	</dict>
	<key>Ansi 5 Color</key>
	<dict>
		<key>Red Component</key>
		<real>0.7</real>
		<key>Green Component</key>
		<real>0.4</real>
		<key>Blue Component</key>
		<real>0.8</real>
	</dict>
	<key>Ansi 8 Color</key>
	<dict>
		<key>Red Component</key>
		<real>0.35</real>
		<key>Green Component</key>
		<real>0.38</real>
		<key>Blue Component</key>
		<real>0.45</real>
	</dict>
</dict>
</plist>`

	theme, err := FromIterm2([]byte(data))
	if err != nil {
		t.Fatalf("FromIterm2: %v", err)
	}
	if string(theme.Text) == "" {
		t.Error("Text should be set from Foreground Color")
	}
	if string(theme.TextInverse) == "" {
		t.Error("TextInverse should be set from Background Color")
	}
	if string(theme.Negative) == "" {
		t.Error("Negative should be set from Ansi 1 Color")
	}
	if string(theme.Positive) == "" {
		t.Error("Positive should be set from Ansi 2 Color")
	}
	if string(theme.Flash) == "" {
		t.Error("Flash should be set from Ansi 3 Color")
	}
	if string(theme.Accent) == "" {
		t.Error("Accent should be set from Ansi 4 Color")
	}
	if string(theme.Cursor) == "" {
		t.Error("Cursor should be set from Ansi 5 Color")
	}
	if string(theme.Muted) == "" {
		t.Error("Muted should be set from Ansi 8 Color")
	}
}

func TestFromIterm2_InvalidXML(t *testing.T) {
	_, err := FromIterm2([]byte("<<<not xml"))
	// Should not panic, may or may not return an error depending on parser tolerance.
	_ = err
}

func TestFromIterm2_EmptyPlist(t *testing.T) {
	data := `<?xml version="1.0" encoding="UTF-8"?>
<plist version="1.0">
<dict>
</dict>
</plist>`
	theme, err := FromIterm2([]byte(data))
	if err != nil {
		t.Fatalf("FromIterm2: %v", err)
	}
	// Should return defaults.
	def := DefaultTheme()
	if theme.Text != def.Text {
		t.Errorf("empty plist should keep default Text")
	}
}

func TestClampF_EdgeCases(t *testing.T) {
	if clampF(-0.5) != 0 {
		t.Error("clampF(-0.5) should be 0")
	}
	if clampF(1.5) != 1 {
		t.Error("clampF(1.5) should be 1")
	}
	if clampF(0.5) != 0.5 {
		t.Error("clampF(0.5) should be 0.5")
	}
	if clampF(0) != 0 {
		t.Error("clampF(0) should be 0")
	}
	if clampF(1) != 1 {
		t.Error("clampF(1) should be 1")
	}
}
