package blit

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// FromGogh imports a Theme from a Gogh terminal theme JSON file.
func FromGogh(data []byte) (Theme, error) {
	var raw struct {
		Foreground string `json:"foreground"`
		Background string `json:"background"`
		Color1     string `json:"color1"`
		Color2     string `json:"color2"`
		Color3     string `json:"color3"`
		Color4     string `json:"color4"`
		Color5     string `json:"color5"`
		Color8     string `json:"color8"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return Theme{}, fmt.Errorf("FromGogh: %w", err)
	}
	t := DefaultTheme()
	if raw.Foreground != "" {
		t.Text = lipgloss.Color(raw.Foreground)
	}
	if raw.Background != "" {
		t.TextInverse = lipgloss.Color(raw.Background)
	}
	if raw.Color1 != "" {
		t.Negative = lipgloss.Color(raw.Color1)
	}
	if raw.Color2 != "" {
		t.Positive = lipgloss.Color(raw.Color2)
	}
	if raw.Color3 != "" {
		t.Flash = lipgloss.Color(raw.Color3)
	}
	if raw.Color4 != "" {
		t.Accent = lipgloss.Color(raw.Color4)
	}
	if raw.Color5 != "" {
		t.Cursor = lipgloss.Color(raw.Color5)
	}
	if raw.Color8 != "" {
		t.Muted = lipgloss.Color(raw.Color8)
	}
	return t, nil
}

// FromAlacritty imports a Theme from an Alacritty TOML config file.
func FromAlacritty(data []byte) (Theme, error) {
	t := DefaultTheme()
	var section string
	for _, raw := range strings.Split(string(data), "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "[") {
			section = strings.Trim(line, "[]")
			continue
		}
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		val = strings.TrimSpace(val)
		val = strings.Trim(val, `"'`)
		if strings.HasPrefix(val, "0x") || strings.HasPrefix(val, "0X") {
			val = "#" + val[2:]
		}
		if !strings.HasPrefix(val, "#") {
			continue
		}
		switch section {
		case "colors.primary":
			switch key {
			case "foreground":
				t.Text = lipgloss.Color(val)
			case "background":
				t.TextInverse = lipgloss.Color(val)
			}
		case "colors.normal":
			switch key {
			case "red":
				t.Negative = lipgloss.Color(val)
			case "green":
				t.Positive = lipgloss.Color(val)
			case "blue":
				t.Accent = lipgloss.Color(val)
			case "magenta":
				t.Cursor = lipgloss.Color(val)
			case "yellow":
				t.Flash = lipgloss.Color(val)
			case "black":
				t.Border = lipgloss.Color(val)
			}
		case "colors.cursor":
			if key == "cursor" {
				t.Cursor = lipgloss.Color(val)
			}
		}
	}
	return t, nil
}

// FromIterm2 imports a Theme from an iTerm2 color preset XML (.itermcolors) file.
func FromIterm2(data []byte) (Theme, error) {
	colors := map[string]iterm2Color{}

	dec := xml.NewDecoder(strings.NewReader(string(data)))
	if err := parseIterm2Plist(dec, colors); err != nil {
		return Theme{}, fmt.Errorf("FromIterm2: %w", err)
	}

	toHex := func(e iterm2Color) lipgloss.Color {
		r := uint8(clampF(e.r) * 255)
		g := uint8(clampF(e.g) * 255)
		b := uint8(clampF(e.b) * 255)
		return lipgloss.Color(fmt.Sprintf("#%02x%02x%02x", r, g, b))
	}

	t := DefaultTheme()
	if e, ok := colors["Foreground Color"]; ok {
		t.Text = toHex(e)
	}
	if e, ok := colors["Background Color"]; ok {
		t.TextInverse = toHex(e)
	}
	if e, ok := colors["Ansi 1 Color"]; ok {
		t.Negative = toHex(e)
	}
	if e, ok := colors["Ansi 2 Color"]; ok {
		t.Positive = toHex(e)
	}
	if e, ok := colors["Ansi 4 Color"]; ok {
		t.Accent = toHex(e)
	}
	if e, ok := colors["Ansi 5 Color"]; ok {
		t.Cursor = toHex(e)
	}
	if e, ok := colors["Ansi 3 Color"]; ok {
		t.Flash = toHex(e)
	}
	if e, ok := colors["Ansi 8 Color"]; ok {
		t.Muted = toHex(e)
	}
	return t, nil
}

type iterm2Color struct{ r, g, b float64 }

// parseIterm2Plist parses an iTerm2 plist XML into a map of color name → RGB.
func parseIterm2Plist(dec *xml.Decoder, out map[string]iterm2Color) error {
	var outerKey string
	depth := 0
	for {
		tok, err := dec.Token()
		if err != nil {
			break
		}
		switch t := tok.(type) {
		case xml.StartElement:
			depth++
			if depth == 3 && t.Name.Local == "dict" {
				entry, err := parseColorDict(dec)
				if err != nil {
					return err
				}
				depth--
				if outerKey != "" {
					out[outerKey] = entry
					outerKey = ""
				}
			}
		case xml.EndElement:
			depth--
		case xml.CharData:
			s := strings.TrimSpace(string(t))
			if s != "" && depth == 2 {
				outerKey = s
			}
		}
	}
	return nil
}

// parseColorDict reads a single colour sub-dict until its closing </dict>.
func parseColorDict(dec *xml.Decoder) (iterm2Color, error) {
	var c iterm2Color
	var key string
	for {
		tok, err := dec.Token()
		if err != nil {
			return c, err
		}
		switch t := tok.(type) {
		case xml.EndElement:
			if t.Name.Local == "dict" {
				return c, nil
			}
		case xml.CharData:
			s := strings.TrimSpace(string(t))
			if s == "" {
				continue
			}
			if key == "" {
				key = s
			} else {
				var v float64
				_, _ = fmt.Sscanf(s, "%f", &v)
				switch key {
				case "Red Component":
					c.r = v
				case "Green Component":
					c.g = v
				case "Blue Component":
					c.b = v
				}
				key = ""
			}
		}
	}
}

func clampF(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
