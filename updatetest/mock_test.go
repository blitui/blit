package updatetest_test

import (
	"io"
	"net/http"
	"strings"
	"testing"

	tuikit "github.com/moneycaringcoder/tuikit-go"
	"github.com/moneycaringcoder/tuikit-go/updatetest"
)

func TestNewMockServer_LatestRelease(t *testing.T) {
	srv := updatetest.NewMockServer(updatetest.Release{
		Tag: "v2.0.0", BinaryName: "tool", Body: "# What's new\n\nShiny",
	})
	defer srv.Close()

	cfg := tuikit.UpdateConfig{
		Owner:      "octocat",
		Repo:       "tool",
		BinaryName: "tool",
		Version:    "v1.0.0",
		APIBaseURL: srv.URL,
		CacheDir:   t.TempDir(),
	}
	res, err := tuikit.CheckForUpdate(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if !res.Available {
		t.Error("expected Available=true for v1.0.0 → v2.0.0")
	}
	if res.LatestVersion != "v2.0.0" {
		t.Errorf("LatestVersion = %q, want v2.0.0", res.LatestVersion)
	}
	if !strings.Contains(res.ReleaseNotes, "Shiny") {
		t.Errorf("ReleaseNotes missing body: %q", res.ReleaseNotes)
	}
}

func TestNewMockServer_CurrentEqualsLatest(t *testing.T) {
	srv := updatetest.NewMockServer(updatetest.Release{
		Tag: "v1.0.0", BinaryName: "tool",
	})
	defer srv.Close()

	cfg := tuikit.UpdateConfig{
		Owner: "o", Repo: "r", BinaryName: "tool", Version: "v1.0.0",
		APIBaseURL: srv.URL, CacheDir: t.TempDir(),
	}
	res, err := tuikit.CheckForUpdate(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if res.Available {
		t.Error("equal versions should not be Available")
	}
}

func TestNewMockServer_ChecksumAssetDownloads(t *testing.T) {
	srv := updatetest.NewMockServer(updatetest.Release{
		Tag: "v3.0.0", BinaryName: "tool",
	})
	defer srv.Close()

	// Fetch the release JSON, locate the checksum asset URL, download it.
	resp, err := http.Get(srv.URL + "/repos/o/r/releases/latest")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "tool_v3.0.0_checksums.txt") {
		t.Errorf("release body missing checksum asset name: %s", body)
	}

	dl, err := http.Get(srv.URL + "/download/tool_v3.0.0_checksums.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer dl.Body.Close()
	sums, _ := io.ReadAll(dl.Body)
	if len(sums) < 40 {
		t.Errorf("checksum file looks empty: %q", sums)
	}
}

func TestNewMockServer_MinimumVersionInBody(t *testing.T) {
	srv := updatetest.NewMockServer(updatetest.Release{
		Tag: "v2.0.0", BinaryName: "tool",
		MinimumVersion: "v1.5.0",
		Body:           "other notes",
	})
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/repos/o/r/releases/latest")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "minimum_version: v1.5.0") {
		t.Errorf("release body missing minimum_version marker: %s", body)
	}
}

func TestNewMockBinary_Deterministic(t *testing.T) {
	a := updatetest.NewMockBinary("v1.0.0")
	b := updatetest.NewMockBinary("v1.0.0")
	if string(a) != string(b) {
		t.Error("NewMockBinary should be deterministic for the same version")
	}
	if string(a) == string(updatetest.NewMockBinary("v2.0.0")) {
		t.Error("NewMockBinary should differ across versions")
	}
}

func TestNewMockServer_MultipleReleases(t *testing.T) {
	srv := updatetest.NewMockServer(
		updatetest.Release{Tag: "v2.0.0", BinaryName: "tool"},
		updatetest.Release{Tag: "v1.0.0", BinaryName: "tool"},
	)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/repos/o/r/releases")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "v2.0.0") || !strings.Contains(string(body), "v1.0.0") {
		t.Errorf("both releases should appear in list: %s", body)
	}
}
