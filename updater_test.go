package tuikit_test

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"runtime"
	"strings"
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
		{"v0.6.2+dirty", 0, 6, 2, false},
		{"v1.0.0+incompatible", 1, 0, 0, false},
		{"v2.1.0-rc.1", 2, 1, 0, false},
		{"v1.3.0-beta.2+build.123", 1, 3, 0, false},
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

func TestVerifyChecksum(t *testing.T) {
	data := []byte("hello world binary content")
	hash := sha256.Sum256(data)
	hexHash := fmt.Sprintf("%x", hash)

	checksumFile := fmt.Sprintf("%s  myapp_0.5.0_linux_amd64.tar.gz\nabc123  other_file.zip\n", hexHash)

	t.Run("valid checksum", func(t *testing.T) {
		err := tuikit.VerifyChecksum(data, "myapp_0.5.0_linux_amd64.tar.gz", []byte(checksumFile))
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("invalid checksum", func(t *testing.T) {
		err := tuikit.VerifyChecksum([]byte("tampered"), "myapp_0.5.0_linux_amd64.tar.gz", []byte(checksumFile))
		if err == nil {
			t.Error("expected checksum mismatch error")
		}
	})

	t.Run("missing asset in checksums", func(t *testing.T) {
		err := tuikit.VerifyChecksum(data, "nonexistent.tar.gz", []byte(checksumFile))
		if err == nil {
			t.Error("expected missing asset error")
		}
	})
}

func TestWithAutoUpdateNotifyMode(t *testing.T) {
	responseJSON := `{
		"tag_name": "v0.5.0",
		"html_url": "https://github.com/owner/repo/releases/tag/v0.5.0",
		"body": "New stuff",
		"assets": []
	}`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(responseJSON))
	}))
	defer srv.Close()

	cfg := tuikit.UpdateConfig{
		Owner:      "owner",
		Repo:       "repo",
		BinaryName: "myapp",
		Version:    "v0.3.0",
		Mode:       tuikit.UpdateNotify,
		CacheTTL:   24 * time.Hour,
		CacheDir:   t.TempDir(),
		APIBaseURL: srv.URL,
	}

	// Verify the option can be passed to NewApp without panic
	app := tuikit.NewApp(
		tuikit.WithAutoUpdate(cfg),
	)
	if app == nil {
		t.Fatal("expected non-nil app")
	}
}

func TestCheckForUpdate(t *testing.T) {
	responseJSON := `{
		"tag_name": "v0.5.0",
		"html_url": "https://github.com/owner/repo/releases/tag/v0.5.0",
		"body": "Release notes",
		"assets": []
	}`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(responseJSON))
	}))
	defer srv.Close()

	dir := t.TempDir()

	t.Run("update available", func(t *testing.T) {
		cfg := tuikit.UpdateConfig{
			Owner:      "owner",
			Repo:       "repo",
			BinaryName: "myapp",
			Version:    "v0.3.0",
			CacheTTL:   24 * time.Hour,
			CacheDir:   dir,
			APIBaseURL: srv.URL,
		}
		result, err := tuikit.CheckForUpdate(cfg)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.Available {
			t.Error("expected update to be available")
		}
		if result.LatestVersion != "v0.5.0" {
			t.Errorf("LatestVersion = %q, want %q", result.LatestVersion, "v0.5.0")
		}
	})

	t.Run("no update when current is latest", func(t *testing.T) {
		cfg := tuikit.UpdateConfig{
			Owner:      "owner",
			Repo:       "repo",
			BinaryName: "myapp",
			Version:    "v0.5.0",
			CacheTTL:   24 * time.Hour,
			CacheDir:   filepath.Join(dir, "no-update"),
			APIBaseURL: srv.URL,
		}
		result, err := tuikit.CheckForUpdate(cfg)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Available {
			t.Error("expected no update available")
		}
	})

	t.Run("skip when version is dev", func(t *testing.T) {
		cfg := tuikit.UpdateConfig{
			Owner:   "owner",
			Repo:    "repo",
			Version: "dev",
		}
		result, err := tuikit.CheckForUpdate(cfg)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Available {
			t.Error("expected no update for dev version")
		}
	})

	t.Run("skip when version is empty", func(t *testing.T) {
		cfg := tuikit.UpdateConfig{
			Owner: "owner",
			Repo:  "repo",
		}
		result, err := tuikit.CheckForUpdate(cfg)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Available {
			t.Error("expected no update for empty version")
		}
	})

	t.Run("uses cache when fresh", func(t *testing.T) {
		cacheDir := filepath.Join(dir, "cached")
		cachePath := filepath.Join(cacheDir, "update-check.json")
		cache := tuikit.UpdateCache{
			CheckedAt:     time.Now().UTC(),
			LatestVersion: "v0.6.0",
			ReleaseURL:    "https://example.com",
		}
		if err := tuikit.WriteCache(cachePath, cache); err != nil {
			t.Fatalf("WriteCache: %v", err)
		}

		// Server would return v0.5.0, but cache says v0.6.0
		cfg := tuikit.UpdateConfig{
			Owner:      "owner",
			Repo:       "repo",
			BinaryName: "myapp",
			Version:    "v0.3.0",
			CacheTTL:   24 * time.Hour,
			CacheDir:   cacheDir,
			APIBaseURL: srv.URL,
		}
		result, err := tuikit.CheckForUpdate(cfg)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// Should use cached v0.6.0, not server's v0.5.0
		if result.LatestVersion != "v0.6.0" {
			t.Errorf("LatestVersion = %q, want %q (from cache)", result.LatestVersion, "v0.6.0")
		}
	})
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		cfg     tuikit.UpdateConfig
		wantErr string
	}{
		{
			name: "valid config",
			cfg: tuikit.UpdateConfig{
				Owner:      "owner",
				Repo:       "repo",
				BinaryName: "myapp",
				Version:    "v1.0.0",
			},
		},
		{
			name:    "missing owner",
			cfg:     tuikit.UpdateConfig{Repo: "repo", BinaryName: "myapp", Version: "v1.0.0"},
			wantErr: "Owner",
		},
		{
			name:    "missing repo",
			cfg:     tuikit.UpdateConfig{Owner: "owner", BinaryName: "myapp", Version: "v1.0.0"},
			wantErr: "Repo",
		},
		{
			name:    "missing binary name",
			cfg:     tuikit.UpdateConfig{Owner: "owner", Repo: "repo", Version: "v1.0.0"},
			wantErr: "BinaryName",
		},
		{
			name:    "missing version",
			cfg:     tuikit.UpdateConfig{Owner: "owner", Repo: "repo", BinaryName: "myapp"},
			wantErr: "Version",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tuikit.ValidateConfig(tt.cfg)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error %q does not contain %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestCheckForUpdateValidatesConfig(t *testing.T) {
	// Version is non-empty and non-dev, so ValidateConfig runs. Missing Owner should error.
	cfg := tuikit.UpdateConfig{
		Version:    "v1.0.0",
		Repo:       "repo",
		BinaryName: "myapp",
	}
	_, err := tuikit.CheckForUpdate(cfg)
	if err == nil {
		t.Fatal("expected error for missing Owner, got nil")
	}
}

func TestSelfUpdateValidatesConfig(t *testing.T) {
	cfg := tuikit.UpdateConfig{
		Version:    "v1.0.0",
		Repo:       "repo",
		BinaryName: "myapp",
		// Owner missing
	}
	err := tuikit.SelfUpdate(cfg)
	if err == nil {
		t.Fatal("expected error for missing Owner, got nil")
	}
	if !strings.Contains(err.Error(), "Owner") {
		t.Errorf("error %q should mention Owner", err.Error())
	}
}

func TestProgressFuncCalledDuringDownload(t *testing.T) {
	binaryContent := []byte("fake binary content for progress test")

	// Build archive matching the current OS/arch so MatchAsset succeeds.
	ext := "tar.gz"
	if runtime.GOOS == "windows" {
		ext = "zip"
	}
	assetName := fmt.Sprintf("myapp_1.0.0_%s_%s.%s", runtime.GOOS, runtime.GOARCH, ext)

	var archive []byte
	if ext == "zip" {
		archive = createTestZip(t, "myapp.exe", binaryContent)
	} else {
		archive = createTestTarGz(t, "myapp", binaryContent)
	}

	hash := sha256.Sum256(archive)
	checksumFile := fmt.Sprintf("%x  %s\n", hash, assetName)

	// srvURL is set after the server starts; the handler closure captures the pointer.
	var srvURL string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/owner/repo/releases/latest":
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, `{
				"tag_name": "v1.0.0",
				"html_url": "https://github.com/owner/repo/releases/tag/v1.0.0",
				"body": "",
				"assets": [
					{"name": %q, "browser_download_url": %q},
					{"name": "checksums.txt", "browser_download_url": %q}
				]
			}`,
				assetName,
				srvURL+"/download/asset",
				srvURL+"/download/checksums",
			)
		case "/download/asset":
			w.Header().Set("Content-Length", fmt.Sprintf("%d", len(archive)))
			w.Write(archive)
		case "/download/checksums":
			w.Write([]byte(checksumFile))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()
	srvURL = srv.URL

	var calls []struct{ received, total int64 }
	cfg := tuikit.UpdateConfig{
		Owner:      "owner",
		Repo:       "repo",
		BinaryName: "myapp",
		Version:    "v0.9.0",
		CacheDir:   t.TempDir(),
		APIBaseURL: srv.URL,
		OnProgress: func(received, total int64) {
			calls = append(calls, struct{ received, total int64 }{received, total})
		},
	}

	// SelfUpdate will fail at replaceBinary (no real exe path in tests), but
	// progress must have been called before that point.
	_ = tuikit.SelfUpdate(cfg)

	if len(calls) == 0 {
		t.Fatal("OnProgress was never called")
	}
	// Final call should report full size received.
	last := calls[len(calls)-1]
	if last.received != int64(len(archive)) {
		t.Errorf("final received = %d, want %d", last.received, len(archive))
	}
	if last.total != int64(len(archive)) {
		t.Errorf("final total = %d, want %d", last.total, len(archive))
	}
	// Progress should be monotonically non-decreasing.
	for i := 1; i < len(calls); i++ {
		if calls[i].received < calls[i-1].received {
			t.Errorf("progress went backwards at call %d: %d < %d", i, calls[i].received, calls[i-1].received)
		}
	}
}

func TestProgressFuncNilIsNoOp(t *testing.T) {
	// Ensure nil OnProgress does not panic — covered by existing SelfUpdate tests,
	// but we explicitly verify downloadURL (which calls downloadURLWithProgress(nil)) here.
	content := []byte("data")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(content)
	}))
	defer srv.Close()

	// Use CheckForUpdate path which internally uses downloadURL (no progress).
	// This is a compile+runtime sanity check that nil progress doesn't panic.
	cfg := tuikit.UpdateConfig{
		Owner:      "owner",
		Repo:       "repo",
		BinaryName: "myapp",
		Version:    "v0.1.0",
		CacheDir:   t.TempDir(),
		APIBaseURL: srv.URL,
		// OnProgress deliberately nil
	}
	// Server returns invalid JSON — result will be silently ignored. We just want no panic.
	result, err := tuikit.CheckForUpdate(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = result
}

func createTestTarGz(t *testing.T, binaryName string, content []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)

	hdr := &tar.Header{
		Name: binaryName,
		Mode: 0o755,
		Size: int64(len(content)),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		t.Fatal(err)
	}
	if _, err := tw.Write(content); err != nil {
		t.Fatal(err)
	}
	tw.Close()
	gw.Close()
	return buf.Bytes()
}

func TestExtractBinaryFromTarGz(t *testing.T) {
	content := []byte("#!/bin/sh\necho hello")
	archive := createTestTarGz(t, "myapp", content)

	got, err := tuikit.ExtractBinary(archive, "myapp", "tar.gz")
	if err != nil {
		t.Fatalf("ExtractBinary: %v", err)
	}
	if !bytes.Equal(got, content) {
		t.Errorf("extracted content mismatch")
	}
}

func TestExtractBinaryFromTarGzMissing(t *testing.T) {
	archive := createTestTarGz(t, "other", []byte("content"))

	_, err := tuikit.ExtractBinary(archive, "myapp", "tar.gz")
	if err == nil {
		t.Error("expected error for missing binary")
	}
}

func createTestZip(t *testing.T, binaryName string, content []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	fw, err := zw.Create(binaryName)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fw.Write(content); err != nil {
		t.Fatal(err)
	}
	zw.Close()
	return buf.Bytes()
}

func TestExtractBinaryFromZip(t *testing.T) {
	content := []byte("windows binary content")
	archive := createTestZip(t, "myapp.exe", content)

	got, err := tuikit.ExtractBinary(archive, "myapp", "zip")
	if err != nil {
		t.Fatalf("ExtractBinary: %v", err)
	}
	if !bytes.Equal(got, content) {
		t.Errorf("extracted content mismatch")
	}
}

func TestExtractBinaryFromZipMissing(t *testing.T) {
	archive := createTestZip(t, "other.exe", []byte("content"))

	_, err := tuikit.ExtractBinary(archive, "myapp", "zip")
	if err == nil {
		t.Error("expected error for missing binary")
	}
}
