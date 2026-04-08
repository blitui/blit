package tuikit_test

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	tuikit "github.com/moneycaringcoder/tuikit-go"
)

func TestParseVersion(t *testing.T) {
	tests := []struct {
		input   string
		major   int
		minor   int
		patch   int
		wantErr bool
	}{
		{"v1.2.3", 1, 2, 3, false},
		{"0.4.0", 0, 4, 0, false},
		{"v0.10.1", 0, 10, 1, false},
		{"v1.0.0", 1, 0, 0, false},
		{"bad", 0, 0, 0, true},
		{"v1.2", 0, 0, 0, true},
		{"", 0, 0, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			v, err := tuikit.ParseVersion(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error for %q", tt.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if v.Major != tt.major || v.Minor != tt.minor || v.Patch != tt.patch {
				t.Errorf("got %d.%d.%d, want %d.%d.%d", v.Major, v.Minor, v.Patch, tt.major, tt.minor, tt.patch)
			}
		})
	}
}

func TestDetectInstallMethod(t *testing.T) {
	tests := []struct {
		path string
		want tuikit.InstallMethod
	}{
		{"/opt/homebrew/Cellar/cryptstream/0.3.0/bin/cryptstream", tuikit.InstallHomebrew},
		{"/home/linuxbrew/.linuxbrew/Cellar/cryptstream/0.3.0/bin/cryptstream", tuikit.InstallHomebrew},
		{"/usr/local/Cellar/cryptstream/0.3.0/bin/cryptstream", tuikit.InstallHomebrew},
		{`C:\Users\user\scoop\apps\cryptstream\current\cryptstream.exe`, tuikit.InstallScoop},
		{"/usr/local/bin/cryptstream", tuikit.InstallManual},
		{`C:\Users\user\go\bin\cryptstream.exe`, tuikit.InstallManual},
		{"/home/user/bin/cryptstream", tuikit.InstallManual},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := tuikit.DetectInstallMethod(tt.path)
			if got != tt.want {
				t.Errorf("DetectInstallMethod(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestVersionNewerThan(t *testing.T) {
	tests := []struct {
		a, b string
		want bool
	}{
		{"v0.4.0", "v0.3.0", true},
		{"v0.3.0", "v0.4.0", false},
		{"v0.4.0", "v0.4.0", false},
		{"v1.0.0", "v0.99.0", true},
		{"v0.4.1", "v0.4.0", true},
		{"v2.0.0", "v1.9.9", true},
	}

	for _, tt := range tests {
		t.Run(tt.a+"_vs_"+tt.b, func(t *testing.T) {
			a, _ := tuikit.ParseVersion(tt.a)
			b, _ := tuikit.ParseVersion(tt.b)
			if got := a.NewerThan(b); got != tt.want {
				t.Errorf("(%s).NewerThan(%s) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestCacheWriteAndRead(t *testing.T) {
	dir := t.TempDir()
	cache := tuikit.UpdateCache{
		CheckedAt:     time.Now().UTC().Truncate(time.Second),
		LatestVersion: "v0.5.0",
		ReleaseURL:    "https://github.com/owner/repo/releases/tag/v0.5.0",
		ReleaseNotes:  "Bug fixes",
	}

	path := filepath.Join(dir, "update-check.json")
	if err := tuikit.WriteCache(path, cache); err != nil {
		t.Fatalf("WriteCache: %v", err)
	}

	got, err := tuikit.ReadCache(path)
	if err != nil {
		t.Fatalf("ReadCache: %v", err)
	}
	if got.LatestVersion != cache.LatestVersion {
		t.Errorf("LatestVersion = %q, want %q", got.LatestVersion, cache.LatestVersion)
	}
	if got.ReleaseURL != cache.ReleaseURL {
		t.Errorf("ReleaseURL = %q, want %q", got.ReleaseURL, cache.ReleaseURL)
	}
	if !got.CheckedAt.Equal(cache.CheckedAt) {
		t.Errorf("CheckedAt = %v, want %v", got.CheckedAt, cache.CheckedAt)
	}
}

func TestCacheFreshness(t *testing.T) {
	fresh := tuikit.UpdateCache{CheckedAt: time.Now().UTC()}
	stale := tuikit.UpdateCache{CheckedAt: time.Now().UTC().Add(-25 * time.Hour)}

	ttl := 24 * time.Hour
	if !fresh.IsFresh(ttl) {
		t.Error("expected fresh cache to be fresh")
	}
	if stale.IsFresh(ttl) {
		t.Error("expected stale cache to not be fresh")
	}
}

func TestReadCacheMissingFile(t *testing.T) {
	_, err := tuikit.ReadCache(filepath.Join(t.TempDir(), "nonexistent.json"))
	if err == nil {
		t.Error("expected error for missing file")
	}
}

func TestFetchLatestRelease(t *testing.T) {
	responseJSON := `{
		"tag_name": "v0.5.0",
		"html_url": "https://github.com/owner/repo/releases/tag/v0.5.0",
		"body": "Bug fixes and improvements",
		"assets": [
			{"name": "myapp_0.5.0_linux_amd64.tar.gz", "browser_download_url": "https://example.com/myapp_0.5.0_linux_amd64.tar.gz"},
			{"name": "myapp_0.5.0_darwin_arm64.tar.gz", "browser_download_url": "https://example.com/myapp_0.5.0_darwin_arm64.tar.gz"},
			{"name": "myapp_0.5.0_windows_amd64.zip", "browser_download_url": "https://example.com/myapp_0.5.0_windows_amd64.zip"},
			{"name": "checksums.txt", "browser_download_url": "https://example.com/checksums.txt"}
		]
	}`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/owner/repo/releases/latest" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(responseJSON))
	}))
	defer srv.Close()

	rel, err := tuikit.FetchLatestRelease(srv.URL, "owner", "repo")
	if err != nil {
		t.Fatalf("FetchLatestRelease: %v", err)
	}
	if rel.TagName != "v0.5.0" {
		t.Errorf("TagName = %q, want %q", rel.TagName, "v0.5.0")
	}
	if len(rel.Assets) != 4 {
		t.Errorf("got %d assets, want 4", len(rel.Assets))
	}
}

func TestFetchLatestReleaseTimeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(6 * time.Second)
	}))
	defer srv.Close()

	_, err := tuikit.FetchLatestRelease(srv.URL, "owner", "repo")
	if err == nil {
		t.Error("expected timeout error")
	}
}

func TestMatchAsset(t *testing.T) {
	assets := []tuikit.ReleaseAsset{
		{Name: "myapp_0.5.0_linux_amd64.tar.gz", DownloadURL: "https://example.com/linux_amd64.tar.gz"},
		{Name: "myapp_0.5.0_darwin_arm64.tar.gz", DownloadURL: "https://example.com/darwin_arm64.tar.gz"},
		{Name: "myapp_0.5.0_windows_amd64.zip", DownloadURL: "https://example.com/windows_amd64.zip"},
		{Name: "checksums.txt", DownloadURL: "https://example.com/checksums.txt"},
	}

	tests := []struct {
		binary  string
		goos    string
		goarch  string
		wantURL string
		wantErr bool
	}{
		{"myapp", "linux", "amd64", "https://example.com/linux_amd64.tar.gz", false},
		{"myapp", "darwin", "arm64", "https://example.com/darwin_arm64.tar.gz", false},
		{"myapp", "windows", "amd64", "https://example.com/windows_amd64.zip", false},
		{"myapp", "freebsd", "amd64", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.goos+"_"+tt.goarch, func(t *testing.T) {
			got, err := tuikit.MatchAsset(assets, tt.binary, tt.goos, tt.goarch)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.DownloadURL != tt.wantURL {
				t.Errorf("DownloadURL = %q, want %q", got.DownloadURL, tt.wantURL)
			}
		})
	}
}

func TestMatchChecksumAsset(t *testing.T) {
	assets := []tuikit.ReleaseAsset{
		{Name: "myapp_0.5.0_linux_amd64.tar.gz", DownloadURL: "https://example.com/linux.tar.gz"},
		{Name: "checksums.txt", DownloadURL: "https://example.com/checksums.txt"},
	}

	got, err := tuikit.MatchChecksumAsset(assets)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Name != "checksums.txt" {
		t.Errorf("Name = %q, want %q", got.Name, "checksums.txt")
	}
}

