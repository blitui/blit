package tuikit

import (
	"strings"
	"testing"
	"time"
)

func TestNewUpdateProgress_Defaults(t *testing.T) {
	p := NewUpdateProgress("tool", "v1.0.0", 1024)
	if p.Binary != "tool" || p.Version != "v1.0.0" || p.Total != 1024 {
		t.Errorf("bad fields: %+v", p)
	}
	if p.Width < 10 {
		t.Errorf("default width too small: %d", p.Width)
	}
	if p.StartedAt.IsZero() {
		t.Error("StartedAt should be set")
	}
}

func TestUpdateProgress_Percent(t *testing.T) {
	p := NewUpdateProgress("t", "v1", 100)
	p.Downloaded = 25
	if got := p.Percent(); got != 0.25 {
		t.Errorf("got %v, want 0.25", got)
	}
	p.Downloaded = 200
	if got := p.Percent(); got != 1.0 {
		t.Errorf("clamped got %v, want 1.0", got)
	}
}

func TestUpdateProgress_PercentUnknownTotal(t *testing.T) {
	p := NewUpdateProgress("t", "v1", 0)
	if got := p.Percent(); got != 0 {
		t.Errorf("unknown total should give 0, got %v", got)
	}
}

func TestUpdateProgress_View_Contains(t *testing.T) {
	p := NewUpdateProgress("mytool", "v2.0.0", 1024)
	p.Downloaded = 512
	v := p.View()
	for _, want := range []string{"mytool", "v2.0.0", "50%", "KiB"} {
		if !strings.Contains(v, want) {
			t.Errorf("View missing %q:\n%s", want, v)
		}
	}
}

func TestUpdateProgress_View_Error(t *testing.T) {
	p := NewUpdateProgress("t", "v1", 100)
	p.Update(UpdateProgressMsg{Err: errTestBoom})
	v := p.View()
	if !strings.Contains(v, "error") || !strings.Contains(v, "boom") {
		t.Errorf("error view missing error text: %s", v)
	}
}

func TestUpdateProgress_UpdateMsg(t *testing.T) {
	p := NewUpdateProgress("t", "v1", 1000)
	p.Update(UpdateProgressMsg{Downloaded: 500})
	if p.Downloaded != 500 {
		t.Errorf("downloaded = %d", p.Downloaded)
	}
	p.Update(UpdateProgressMsg{Done: true})
	if !p.Done {
		t.Error("Done not set")
	}
}

func TestHumanBytes(t *testing.T) {
	cases := map[int64]string{
		0:               "0 B",
		512:             "512 B",
		2048:            "2.0 KiB",
		5 * 1024 * 1024: "5.00 MiB",
	}
	for n, want := range cases {
		if got := humanBytes(n); got != want {
			t.Errorf("humanBytes(%d) = %q, want %q", n, got, want)
		}
	}
}

func TestUpdateProgress_ETA(t *testing.T) {
	p := NewUpdateProgress("t", "v1", 1000)
	p.StartedAt = time.Now().Add(-1 * time.Second)
	p.Downloaded = 500
	if eta := p.ETA(); eta <= 0 || eta > 5*time.Second {
		t.Errorf("ETA looks wrong: %v", eta)
	}
}

var errTestBoom = &testError{"boom"}

type testError struct{ msg string }

func (e *testError) Error() string { return e.msg }
