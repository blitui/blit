package blit_test

import (
	"testing"

	blit "github.com/blitui/blit"
	"github.com/blitui/blit/updatetest"
)

func TestCheckForUpdate_MinimumVersionSetsRequired(t *testing.T) {
	srv := updatetest.NewMockServer(updatetest.Release{
		Tag:            "v2.0.0",
		BinaryName:     "tool",
		MinimumVersion: "v1.5.0",
		Body:           "breaking changes",
	})
	defer srv.Close()

	cfg := blit.UpdateConfig{
		Owner: "o", Repo: "r", BinaryName: "tool",
		Version:    "v1.0.0",
		APIBaseURL: srv.URL,
		CacheDir:   t.TempDir(),
	}
	res, err := blit.CheckForUpdate(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if !res.Required {
		t.Error("expected Required=true when current < minimum_version")
	}
	if !res.Available {
		t.Error("expected Available=true when forced")
	}
}

func TestCheckForUpdate_MinimumVersionNotRequired(t *testing.T) {
	srv := updatetest.NewMockServer(updatetest.Release{
		Tag:            "v2.0.0",
		BinaryName:     "tool",
		MinimumVersion: "v1.5.0",
	})
	defer srv.Close()

	cfg := blit.UpdateConfig{
		Owner: "o", Repo: "r", BinaryName: "tool",
		Version:    "v1.6.0",
		APIBaseURL: srv.URL,
		CacheDir:   t.TempDir(),
	}
	res, err := blit.CheckForUpdate(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if res.Required {
		t.Error("expected Required=false when current >= minimum_version")
	}
	if !res.Available {
		t.Error("expected Available=true (v1.6.0 < v2.0.0)")
	}
}

func TestCheckForUpdate_SkippedVersionSuppressed(t *testing.T) {
	srv := updatetest.NewMockServer(updatetest.Release{
		Tag: "v2.0.0", BinaryName: "tool",
	})
	defer srv.Close()

	cfg := blit.UpdateConfig{
		Owner: "o", Repo: "r", BinaryName: "tool",
		Version:    "v1.0.0",
		APIBaseURL: srv.URL,
		CacheDir:   t.TempDir(),
	}
	if err := blit.SkipVersion(cfg, "v2.0.0"); err != nil {
		t.Fatal(err)
	}
	res, err := blit.CheckForUpdate(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if res.Available {
		t.Error("expected Available=false for skipped version")
	}
}

func TestCheckForUpdate_SkipDoesNotOverrideRequired(t *testing.T) {
	srv := updatetest.NewMockServer(updatetest.Release{
		Tag:            "v2.0.0",
		BinaryName:     "tool",
		MinimumVersion: "v1.5.0",
	})
	defer srv.Close()

	cfg := blit.UpdateConfig{
		Owner: "o", Repo: "r", BinaryName: "tool",
		Version:    "v1.0.0",
		APIBaseURL: srv.URL,
		CacheDir:   t.TempDir(),
	}
	_ = blit.SkipVersion(cfg, "v2.0.0")
	res, err := blit.CheckForUpdate(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if !res.Required || !res.Available {
		t.Errorf("forced update should ignore skip: Required=%v Available=%v",
			res.Required, res.Available)
	}
}
