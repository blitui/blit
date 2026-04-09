package tuikit

import (
	"testing"
)

func testSkipCfg(t *testing.T) UpdateConfig {
	t.Helper()
	return UpdateConfig{
		BinaryName: "tuitest-binary",
		CacheDir:   t.TempDir(),
	}
}

func TestSkipVersion_RoundTrip(t *testing.T) {
	cfg := testSkipCfg(t)
	if IsVersionSkipped(cfg, "v1.0.0") {
		t.Fatal("fresh cache should have no skipped versions")
	}
	if err := SkipVersion(cfg, "v1.0.0"); err != nil {
		t.Fatal(err)
	}
	if !IsVersionSkipped(cfg, "v1.0.0") {
		t.Error("v1.0.0 should be marked skipped after SkipVersion")
	}
	if IsVersionSkipped(cfg, "v2.0.0") {
		t.Error("unrelated version should not be marked skipped")
	}
}

func TestSkipVersion_MultipleVersions(t *testing.T) {
	cfg := testSkipCfg(t)
	versions := []string{"v1.0.0", "v1.1.0", "v2.0.0"}
	for _, v := range versions {
		if err := SkipVersion(cfg, v); err != nil {
			t.Fatal(err)
		}
	}
	for _, v := range versions {
		if !IsVersionSkipped(cfg, v) {
			t.Errorf("%s should be skipped", v)
		}
	}
}

func TestSkipVersion_EmptyString(t *testing.T) {
	cfg := testSkipCfg(t)
	if err := SkipVersion(cfg, ""); err != nil {
		t.Errorf("empty version should be a no-op, got %v", err)
	}
	if IsVersionSkipped(cfg, "") {
		t.Error("empty string should never be 'skipped'")
	}
}

func TestSkipVersion_Clear(t *testing.T) {
	cfg := testSkipCfg(t)
	_ = SkipVersion(cfg, "v1.0.0")
	_ = SkipVersion(cfg, "v2.0.0")
	if err := ClearSkippedVersions(cfg); err != nil {
		t.Fatal(err)
	}
	if IsVersionSkipped(cfg, "v1.0.0") {
		t.Error("ClearSkippedVersions did not clear v1.0.0")
	}
}

func TestSkipVersion_ClearWhenMissing(t *testing.T) {
	cfg := testSkipCfg(t)
	if err := ClearSkippedVersions(cfg); err != nil {
		t.Errorf("Clear on empty dir should not error, got %v", err)
	}
}

func TestParseMinimumVersion(t *testing.T) {
	cases := []struct {
		name    string
		body    string
		wantOK  bool
		wantStr string
	}{
		{"plain", "minimum_version: v1.2.3", true, "v1.2.3"},
		{"dash", "minimum-version: v1.2.3", true, "v1.2.3"},
		{"no v", "minimum_version: 1.2.3", true, "v1.2.3"},
		{"uppercase", "MINIMUM_VERSION: v1.2.3", true, "v1.2.3"},
		{"embedded", "Release notes:\n\nminimum_version: v0.5.0\n\nbullet", true, "v0.5.0"},
		{"equals", "minimum_version = v1.0.0", true, "v1.0.0"},
		{"missing", "Just some notes", false, ""},
		{"empty", "", false, ""},
		{"garbage", "minimum_version: notaversion", false, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := ParseMinimumVersion(tc.body)
			if ok != tc.wantOK {
				t.Errorf("ok = %v, want %v", ok, tc.wantOK)
			}
			if ok && got.String() != tc.wantStr {
				t.Errorf("version = %q, want %q", got.String(), tc.wantStr)
			}
		})
	}
}
