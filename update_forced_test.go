package tuikit_test

import (
	"strings"
	"testing"

	tuikit "github.com/moneycaringcoder/tuikit-go"
	"github.com/moneycaringcoder/tuikit-go/tuitest"
)

func TestForcedUpdateScreen_RendersVersions(t *testing.T) {
	res := &tuikit.UpdateResult{
		CurrentVersion: "v1.0.0",
		LatestVersion:  "v2.0.0",
		ReleaseNotes:   "big changes",
	}
	gate := tuikit.NewForcedUpdateScreen(res, tuikit.UpdateConfig{BinaryName: "tool"})
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
	res := &tuikit.UpdateResult{CurrentVersion: "v1.0.0", LatestVersion: "v2.0.0"}
	gate := tuikit.NewForcedUpdateScreen(res, tuikit.UpdateConfig{})
	tm := tuitest.NewTestModel(t, gate, 80, 24)
	tm.SendKey("y")
	if gate.Choice != tuikit.ForcedChoiceUpdate {
		t.Errorf("expected ForcedChoiceUpdate, got %v", gate.Choice)
	}
}

func TestForcedUpdateScreen_QuitKey(t *testing.T) {
	res := &tuikit.UpdateResult{CurrentVersion: "v1.0.0", LatestVersion: "v2.0.0"}
	gate := tuikit.NewForcedUpdateScreen(res, tuikit.UpdateConfig{})
	tm := tuitest.NewTestModel(t, gate, 80, 24)
	tm.SendKey("q")
	if gate.Choice != tuikit.ForcedChoiceQuit {
		t.Errorf("expected ForcedChoiceQuit, got %v", gate.Choice)
	}
}

func TestForcedUpdateScreen_EnterIsUpdate(t *testing.T) {
	res := &tuikit.UpdateResult{CurrentVersion: "v1.0.0", LatestVersion: "v2.0.0"}
	gate := tuikit.NewForcedUpdateScreen(res, tuikit.UpdateConfig{})
	tm := tuitest.NewTestModel(t, gate, 80, 24)
	tm.SendKey("enter")
	if gate.Choice != tuikit.ForcedChoiceUpdate {
		t.Errorf("expected ForcedChoiceUpdate, got %v", gate.Choice)
	}
}
