package blit_test

import (
	"strings"
	"testing"

	blit "github.com/blitui/blit"
	"github.com/blitui/blit/btest"
)

func TestForcedUpdateScreen_RendersVersions(t *testing.T) {
	res := &blit.UpdateResult{
		CurrentVersion: "v1.0.0",
		LatestVersion:  "v2.0.0",
		ReleaseNotes:   "big changes",
	}
	gate := blit.NewForcedUpdateScreen(res, blit.UpdateConfig{BinaryName: "tool"})
	// Use the model's View directly: lipgloss.Place with big padding +
	// a virtual 80x24 can clip content unpredictably across platforms.
	v := gate.View()
	if !strings.Contains(v, "Required update") {
		t.Errorf("missing title:\n%s", v)
	}
	if !strings.Contains(v, "v1.0.0") || !strings.Contains(v, "v2.0.0") {
		t.Errorf("missing version strings:\n%s", v)
	}
	if !strings.Contains(v, "[u]pdate") || !strings.Contains(v, "[q]uit") {
		t.Errorf("missing action hints:\n%s", v)
	}
}

func TestForcedUpdateScreen_UpdateKey(t *testing.T) {
	res := &blit.UpdateResult{CurrentVersion: "v1.0.0", LatestVersion: "v2.0.0"}
	gate := blit.NewForcedUpdateScreen(res, blit.UpdateConfig{})
	tm := btest.NewTestModel(t, gate, 80, 24)
	tm.SendKey("y")
	if gate.Choice != blit.ForcedChoiceUpdate {
		t.Errorf("expected ForcedChoiceUpdate, got %v", gate.Choice)
	}
}

func TestForcedUpdateScreen_QuitKey(t *testing.T) {
	res := &blit.UpdateResult{CurrentVersion: "v1.0.0", LatestVersion: "v2.0.0"}
	gate := blit.NewForcedUpdateScreen(res, blit.UpdateConfig{})
	tm := btest.NewTestModel(t, gate, 80, 24)
	tm.SendKey("q")
	if gate.Choice != blit.ForcedChoiceQuit {
		t.Errorf("expected ForcedChoiceQuit, got %v", gate.Choice)
	}
}

func TestForcedUpdateScreen_EnterIsUpdate(t *testing.T) {
	res := &blit.UpdateResult{CurrentVersion: "v1.0.0", LatestVersion: "v2.0.0"}
	gate := blit.NewForcedUpdateScreen(res, blit.UpdateConfig{})
	tm := btest.NewTestModel(t, gate, 80, 24)
	tm.SendKey("enter")
	if gate.Choice != blit.ForcedChoiceUpdate {
		t.Errorf("expected ForcedChoiceUpdate, got %v", gate.Choice)
	}
}
