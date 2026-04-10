package tuikit

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Gradient defines a linear color gradient between two lipgloss colors.
type Gradient struct {
	Start lipgloss.Color
	End   lipgloss.Color
}

// RenderAt returns the interpolated color at position t (0.0 to 1.0).
func (g Gradient) RenderAt(t float64) lipgloss.Color {
	t = math.Max(0, math.Min(1, t))
	sr, sg, sb := parseHex(string(g.Start))
	er, eg, eb := parseHex(string(g.End))
	r := lerpU8(sr, er, t)
	gv := lerpU8(sg, eg, t)
	b := lerpU8(sb, eb, t)
	return lipgloss.Color(fmt.Sprintf("#%02x%02x%02x", r, gv, b))
}

// RenderText applies the gradient across each character of the string.
func (g Gradient) RenderText(s string) string {
	runes := []rune(s)
	n := len(runes)
	if n == 0 {
		return ""
	}
	var sb strings.Builder
	for i, r := range runes {
		t := 0.0
		if n > 1 {
			t = float64(i) / float64(n-1)
		}
		col := g.RenderAt(t)
		sb.WriteString(lipgloss.NewStyle().Foreground(col).Render(string(r)))
	}
	return sb.String()
}

// RenderGradient renders text with the given gradient. Convenience wrapper for Gradient.RenderText.
func RenderGradient(text string, g Gradient) string {
	return g.RenderText(text)
}

func parseHex(s string) (uint8, uint8, uint8) {
	s = strings.TrimPrefix(s, "#")
	if len(s) != 6 {
		return 0, 0, 0
	}
	r, err1 := strconv.ParseUint(s[0:2], 16, 8)
	g, err2 := strconv.ParseUint(s[2:4], 16, 8)
	b, err3 := strconv.ParseUint(s[4:6], 16, 8)
	if err1 != nil || err2 != nil || err3 != nil {
		return 0, 0, 0
	}
	return uint8(r), uint8(g), uint8(b)
}

func lerpU8(a, b uint8, t float64) uint8 {
	return uint8(math.Round(float64(a) + (float64(b)-float64(a))*t))
}
