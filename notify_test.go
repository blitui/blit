package tuikit

import (
	"testing"
)

func TestNotifyCmd_ZeroDuration(t *testing.T) {
	cmd := NotifyCmd("", 0)
	if cmd == nil {
		t.Fatal("NotifyCmd returned nil")
	}
	msg, ok := cmd().(NotifyMsg)
	if !ok {
		t.Fatalf("expected NotifyMsg, got %T", cmd())
	}
	if msg.Text != "" || msg.Duration != 0 {
		t.Errorf("zero-value NotifyMsg got %+v", msg)
	}
}

func TestNotifyCmd_Nil(t *testing.T) {
	cmd := NotifyCmd("hi", 0)
	// Call twice: verify it's idempotent (pure function).
	a := cmd().(NotifyMsg)
	b := cmd().(NotifyMsg)
	if a != b {
		t.Errorf("NotifyCmd produced different messages on repeated calls: %v vs %v", a, b)
	}
}
