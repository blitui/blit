package tuikit

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// UpdateProgress is a reusable tea.Model component that renders a
// download progress bar for a self-update. Consumers drive it by sending
// UpdateProgressMsg updates; the model handles its own render.
//
// Typical wiring inside UpdateBlocking mode:
//
//	p := NewUpdateProgress("tool", "v2.0.0", totalBytes)
//	// ... each time OnProgress fires:
//	prog.Update(UpdateProgressMsg{Downloaded: n})
//	fmt.Println(prog.View())
type UpdateProgress struct {
	Binary     string
	Version    string
	Total      int64
	Downloaded int64
	StartedAt  time.Time
	Width      int
	Done       bool
	Err        error
}

// NewUpdateProgress constructs a progress component for binary/version
// with an expected total byte count. If total is zero or unknown, the
// bar degrades gracefully to an indeterminate spinner-like view.
func NewUpdateProgress(binary, version string, total int64) *UpdateProgress {
	return &UpdateProgress{
		Binary:    binary,
		Version:   version,
		Total:     total,
		Width:     40,
		StartedAt: time.Now(),
	}
}

// UpdateProgressMsg carries incremental progress updates into the model.
type UpdateProgressMsg struct {
	Downloaded int64
	Done       bool
	Err        error
}

// Init implements tea.Model.
func (p *UpdateProgress) Init() tea.Cmd { return nil }

// Update handles UpdateProgressMsg events and returns the model unchanged
// for any other message.
func (p *UpdateProgress) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m, ok := msg.(UpdateProgressMsg); ok {
		if m.Downloaded > 0 {
			p.Downloaded = m.Downloaded
		}
		if m.Done {
			p.Done = true
		}
		if m.Err != nil {
			p.Err = m.Err
		}
	}
	return p, nil
}

// Percent returns the download completion ratio in [0, 1]. If total is
// unknown, returns 0.
func (p *UpdateProgress) Percent() float64 {
	if p.Total <= 0 {
		return 0
	}
	r := float64(p.Downloaded) / float64(p.Total)
	if r < 0 {
		return 0
	}
	if r > 1 {
		return 1
	}
	return r
}

// Speed returns the average bytes/sec since StartedAt.
func (p *UpdateProgress) Speed() float64 {
	elapsed := time.Since(p.StartedAt).Seconds()
	if elapsed <= 0 {
		return 0
	}
	return float64(p.Downloaded) / elapsed
}

// ETA returns the estimated time remaining based on current speed.
// Returns 0 if speed is 0 or total is unknown.
func (p *UpdateProgress) ETA() time.Duration {
	speed := p.Speed()
	if speed <= 0 || p.Total <= 0 {
		return 0
	}
	remaining := p.Total - p.Downloaded
	if remaining <= 0 {
		return 0
	}
	return time.Duration(float64(remaining)/speed) * time.Second
}

// View implements tea.Model. Renders a styled progress bar with bytes,
// percent, speed, and ETA.
func (p *UpdateProgress) View() string {
	var b strings.Builder
	header := lipgloss.NewStyle().Bold(true).Render(
		fmt.Sprintf("Updating %s → %s", p.Binary, p.Version),
	)
	b.WriteString(header)
	b.WriteString("\n")

	if p.Err != nil {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render(
			fmt.Sprintf("  error: %v", p.Err),
		))
		return b.String()
	}

	// Bar.
	width := p.Width
	if width < 10 {
		width = 10
	}
	pct := p.Percent()
	filled := int(float64(width) * pct)
	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	b.WriteString("  [")
	b.WriteString(bar)
	b.WriteString("] ")
	if p.Total > 0 {
		fmt.Fprintf(&b, "%3.0f%%", pct*100)
	} else {
		b.WriteString(" …  ")
	}
	b.WriteString("\n")

	// Stats line.
	fmt.Fprintf(&b, "  %s / %s", humanBytes(p.Downloaded), humanBytes(p.Total))
	if speed := p.Speed(); speed > 0 {
		fmt.Fprintf(&b, "  •  %s/s", humanBytes(int64(speed)))
	}
	if eta := p.ETA(); eta > 0 {
		fmt.Fprintf(&b, "  •  ETA %s", eta.Round(time.Second))
	}
	if p.Done {
		b.WriteString("  •  done")
	}
	return b.String()
}

// humanBytes renders a byte count as KiB / MiB / GiB.
func humanBytes(n int64) string {
	if n <= 0 {
		return "0 B"
	}
	const (
		kib = 1024
		mib = kib * 1024
		gib = mib * 1024
	)
	switch {
	case n >= gib:
		return fmt.Sprintf("%.2f GiB", float64(n)/gib)
	case n >= mib:
		return fmt.Sprintf("%.2f MiB", float64(n)/mib)
	case n >= kib:
		return fmt.Sprintf("%.1f KiB", float64(n)/kib)
	default:
		return fmt.Sprintf("%d B", n)
	}
}
