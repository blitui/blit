package tuikit

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestConsumed(t *testing.T) {
	cmd := Consumed()
	msg := cmd()
	if _, ok := msg.(consumedMsg); !ok {
		t.Fatalf("expected consumedMsg, got %T", msg)
	}
}

func TestIsConsumed(t *testing.T) {
	tests := []struct {
		name string
		cmd  tea.Cmd
		want bool
	}{
		{"nil cmd", nil, false},
		{"consumed cmd", Consumed(), true},
		{"other cmd", func() tea.Msg { return "hello" }, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isConsumed(tt.cmd)
			if got != tt.want {
				t.Errorf("isConsumed() = %v, want %v", got, tt.want)
			}
		})
	}
}
