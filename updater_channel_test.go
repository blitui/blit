package tuikit

import "testing"

func TestPickReleaseForChannel_Stable(t *testing.T) {
	releases := []Release{
		{TagName: "v2.0.0-beta.1", Prerelease: true},
		{TagName: "v2.0.0-rc.1", Prerelease: true},
		{TagName: "v1.9.0"},
		{TagName: "v1.8.0"},
	}
	r := PickReleaseForChannel(releases, ChannelStable)
	if r == nil || r.TagName != "v1.9.0" {
		t.Errorf("stable got %v, want v1.9.0", r)
	}
}

func TestPickReleaseForChannel_Beta(t *testing.T) {
	releases := []Release{
		{TagName: "v2.0.0-rc.1"},
		{TagName: "v2.0.0-beta.2"},
		{TagName: "v1.9.0"},
	}
	r := PickReleaseForChannel(releases, ChannelBeta)
	if r == nil || r.TagName != "v2.0.0-beta.2" {
		t.Errorf("beta got %v, want v2.0.0-beta.2", r)
	}
}

func TestPickReleaseForChannel_Prerelease(t *testing.T) {
	releases := []Release{
		{TagName: "v1.9.0"},
		{TagName: "v2.0.0-rc.1", Prerelease: true},
	}
	r := PickReleaseForChannel(releases, ChannelPrerelease)
	if r == nil || r.TagName != "v2.0.0-rc.1" {
		t.Errorf("prerelease got %v, want v2.0.0-rc.1", r)
	}
}

func TestPickReleaseForChannel_EmptyFallsBackToStable(t *testing.T) {
	releases := []Release{{TagName: "v1.0.0"}}
	r := PickReleaseForChannel(releases, "")
	if r == nil || r.TagName != "v1.0.0" {
		t.Errorf("empty channel got %v, want v1.0.0", r)
	}
}

func TestPickReleaseForChannel_None(t *testing.T) {
	releases := []Release{{TagName: "v1.0.0"}}
	if r := PickReleaseForChannel(releases, ChannelBeta); r != nil {
		t.Errorf("expected nil for no matching release, got %v", r)
	}
}
