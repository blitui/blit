package cli

import (
	"fmt"
	"strings"
	"sync"
)

const (
	barWidth  = 20
	barFilled = "█"
	barEmpty  = "░"
)

// Progress shows a progress bar that updates as work completes.
type Progress struct {
	total   int
	current int
	label   string
	mu      sync.Mutex
	done    bool
}

// NewProgress creates a progress bar with the given total and label.
func NewProgress(total int, label string) *Progress {
	p := &Progress{total: total, label: label}
	p.render()
	return p
}

// Increment advances the progress bar by n.
func (p *Progress) Increment(n int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.done {
		return
	}
	p.current += n
	if p.current > p.total {
		p.current = p.total
	}
	p.render()
}

// Done completes the progress bar and moves to the next line.
func (p *Progress) Done() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.done {
		return
	}
	p.current = p.total
	p.done = true
	p.render()
	fmt.Println()
}

// render writes the bar to stdout, overwriting the current line.
// Must be called with p.mu held.
func (p *Progress) render() {
	pct := 0
	if p.total > 0 {
		pct = p.current * 100 / p.total
	}
	filled := 0
	if p.total > 0 {
		filled = p.current * barWidth / p.total
	}
	if filled > barWidth {
		filled = barWidth
	}
	bar := strings.Repeat(barFilled, filled) + strings.Repeat(barEmpty, barWidth-filled)
	fmt.Printf("\r%-20s [%s] %3d%%", p.label, bar, pct)
}
