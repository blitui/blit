package blit

import (
	"os"
	"testing"
)

func TestFeatureFlagDefault(t *testing.T) {
	t.Parallel()
	ff := NewFeatureFlag(FlagDef{
		Name: "EXPERIMENTAL_WIDGET", Default: false, Description: "test widget",
	})
	if ff.Enabled("EXPERIMENTAL_WIDGET") {
		t.Error("default should be false")
	}
}

func TestFeatureFlagSet(t *testing.T) {
	t.Parallel()
	ff := NewFeatureFlag(FlagDef{
		Name: "NEW_TABLE", Default: false, Description: "new table impl",
	})
	ff.Set("NEW_TABLE", true)
	if !ff.Enabled("NEW_TABLE") {
		t.Error("override should enable flag")
	}
}

func TestFeatureFlagClear(t *testing.T) {
	t.Parallel()
	ff := NewFeatureFlag(FlagDef{
		Name: "PANEL", Default: true, Description: "panel feature",
	})
	ff.Set("PANEL", false)
	if ff.Enabled("PANEL") {
		t.Error("override should take precedence over default")
	}
	ff.Clear("PANEL")
	if !ff.Enabled("PANEL") {
		t.Error("clear should revert to default (true)")
	}
}

func TestFeatureFlagEnvVar(t *testing.T) {
	t.Parallel()
	os.Setenv("BLIT_FLAG_DARK_MODE", "1")
	defer os.Unsetenv("BLIT_FLAG_DARK_MODE")

	ff := NewFeatureFlag(FlagDef{
		Name: "DARK_MODE", Default: false, Description: "dark mode toggle",
	})
	if !ff.Enabled("DARK_MODE") {
		t.Error("env var BLIT_FLAG_DARK_MODE=1 should enable flag")
	}
}

func TestFeatureFlagEnvVarDisabled(t *testing.T) {
	t.Parallel()
	os.Setenv("BLIT_FLAG_VERBOSE", "0")
	defer os.Unsetenv("BLIT_FLAG_VERBOSE")

	ff := NewFeatureFlag(FlagDef{
		Name: "VERBOSE", Default: true, Description: "verbose output",
	})
	if ff.Enabled("VERBOSE") {
		t.Error("env var BLIT_FLAG_VERBOSE=0 should disable flag even with default true")
	}
}

func TestFeatureFlagOverrideBeatsEnv(t *testing.T) {
	t.Parallel()
	os.Setenv("BLIT_FLAG_ALPHA", "1")
	defer os.Unsetenv("BLIT_FLAG_ALPHA")

	ff := NewFeatureFlag(FlagDef{
		Name: "ALPHA", Default: false, Description: "alpha feature",
	})
	ff.Set("ALPHA", false)
	if ff.Enabled("ALPHA") {
		t.Error("explicit Set should take precedence over env var")
	}
}

func TestFeatureFlagUnknownReturnsFalse(t *testing.T) {
	t.Parallel()
	ff := NewFeatureFlag()
	if ff.Enabled("NONEXISTENT") {
		t.Error("unknown flag should return false")
	}
}

func TestFeatureFlagNames(t *testing.T) {
	t.Parallel()
	ff := NewFeatureFlag(
		FlagDef{Name: "A", Default: false, Description: "a"},
		FlagDef{Name: "B", Default: true, Description: "b"},
	)
	names := ff.Names()
	if len(names) != 2 {
		t.Fatalf("expected 2 names, got %d", len(names))
	}
}

func TestFeatureFlagDescribe(t *testing.T) {
	t.Parallel()
	ff := NewFeatureFlag(FlagDef{
		Name: "METRICS", Default: true, Description: "enable metrics",
	})
	if got := ff.Describe("METRICS"); got != "enable metrics" {
		t.Errorf("expected 'enable metrics', got %q", got)
	}
	if got := ff.Describe("UNKNOWN"); got != "" {
		t.Errorf("expected empty for unknown, got %q", got)
	}
}

func TestFeatureFlagSnapshot(t *testing.T) {
	t.Parallel()
	ff := NewFeatureFlag(
		FlagDef{Name: "FAST", Default: true, Description: "fast mode"},
		FlagDef{Name: "SLOW", Default: false, Description: "slow mode"},
	)
	ff.Set("SLOW", true)
	snap := ff.Snapshot()
	if !snap["FAST"] {
		t.Error("FAST should be true (default)")
	}
	if !snap["SLOW"] {
		t.Error("SLOW should be true (override)")
	}
}
