// Package updatetest provides HTTP mocks and fake binaries for testing
// code that uses blit's auto-update system. It lets consumer apps
// write end-to-end update tests that do not hit the real GitHub API and
// do not touch the user's filesystem.
//
// Typical usage in a consumer test:
//
//	srv := updatetest.NewMockServer(updatetest.Release{
//	    Tag: "v2.0.0", BinaryName: "mytool", Body: "# What's new",
//	})
//	defer srv.Close()
//
//	cfg := blit.UpdateConfig{
//	    Owner: "owner", Repo: "repo",
//	    BinaryName: "mytool", Version: "v1.0.0",
//	    APIBaseURL: srv.URL, CacheDir: t.TempDir(),
//	}
//	res, _ := blit.CheckForUpdate(cfg)
//	if !res.Available { t.Fatal("expected available update") }
package updatetest

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"runtime"
	"strings"
)

// Release describes a single fake GitHub release for the mock server.
// BinaryName is the logical name of the binary inside the tar.gz asset.
// If Assets is empty, NewMockServer auto-generates a tar.gz asset for the
// current GOOS/GOARCH and a matching SHA256SUMS file.
type Release struct {
	// Tag is the release tag name, e.g. "v2.0.0". Required.
	Tag string
	// BinaryName is the name of the binary bundled inside the archive.
	// Required when Assets is empty.
	BinaryName string
	// Body is the release notes Markdown (shown in UpdateResult.ReleaseNotes).
	Body string
	// HTMLURL overrides the release page URL. Optional.
	HTMLURL string
	// Assets lets callers supply custom assets. If nil, defaults are
	// generated for the current GOOS/GOARCH.
	Assets []Asset
	// MinimumVersion, if non-empty, is prepended to Body as a
	// "minimum_version: <v>" marker so tests can exercise forced updates.
	MinimumVersion string
}

// Asset mirrors blit's ReleaseAsset shape for mock server responses.
type Asset struct {
	Name        string `json:"name"`
	DownloadURL string `json:"browser_download_url"`
	ContentType string `json:"content_type"`
}

// NewMockServer returns an httptest.Server that serves GitHub-shaped
// release JSON for the given releases. The latest release (by tag order)
// is exposed at /repos/{owner}/{repo}/releases/latest. All releases are
// available at /repos/{owner}/{repo}/releases. Asset download URLs are
// served from the same server.
//
// Multi-release callers: pass releases in priority order; the FIRST
// release is treated as the "latest".
func NewMockServer(releases ...Release) *httptest.Server {
	mux := http.NewServeMux()
	srv := httptest.NewUnstartedServer(mux)

	// Generate assets + binaries up front so we know the final URLs.
	var prepd []*prepared

	for _, r := range releases {
		p := &prepared{rel: r}
		if len(r.Assets) == 0 {
			bin := NewMockBinary(r.Tag)
			archiveName := fmt.Sprintf("%s_%s_%s_%s.tar.gz",
				r.BinaryName, r.Tag, runtime.GOOS, runtime.GOARCH)
			archive := buildTarGz(r.BinaryName, bin)
			sumsName := fmt.Sprintf("%s_%s_checksums.txt", r.BinaryName, r.Tag)
			sum := sha256.Sum256(archive)
			sums := []byte(fmt.Sprintf("%s  %s\n", hex.EncodeToString(sum[:]), archiveName))
			p.archive = archive
			p.sums = sums
			p.assets = []Asset{
				{Name: archiveName, ContentType: "application/gzip"},
				{Name: sumsName, ContentType: "text/plain"},
			}
		} else {
			p.assets = r.Assets
		}
		prepd = append(prepd, p)
	}

	// Register asset download handlers (URL needs srv to be started).
	mux.HandleFunc("/download/", func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, "/download/")
		for _, p := range prepd {
			for _, a := range p.assets {
				if a.Name != name {
					continue
				}
				if strings.HasSuffix(name, ".tar.gz") && p.archive != nil {
					w.Header().Set("Content-Type", "application/gzip")
					_, _ = w.Write(p.archive)
					return
				}
				if strings.Contains(name, "checksums") && p.sums != nil {
					w.Header().Set("Content-Type", "text/plain")
					_, _ = w.Write(p.sums)
					return
				}
			}
		}
		http.NotFound(w, r)
	})

	mux.HandleFunc("/repos/", func(w http.ResponseWriter, r *http.Request) {
		// Path looks like /repos/{owner}/{repo}/releases[/latest]
		parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/repos/"), "/")
		if len(parts) < 3 || parts[2] != "releases" {
			http.NotFound(w, r)
			return
		}
		wantLatest := len(parts) > 3 && parts[3] == "latest"

		if wantLatest {
			if len(prepd) == 0 {
				http.NotFound(w, r)
				return
			}
			writeReleaseJSON(w, prepd[0], srv.URL)
			return
		}
		// Full list
		out := make([]map[string]interface{}, 0, len(prepd))
		for _, p := range prepd {
			out = append(out, releaseMap(p, srv.URL))
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(out)
	})

	srv.Start()
	return srv
}

// prepared bundles a release with its generated assets so handlers can
// close over them cheaply.
type prepared struct {
	rel     Release
	assets  []Asset
	archive []byte
	sums    []byte
}

func releaseMap(p *prepared, base string) map[string]interface{} {
	body := p.rel.Body
	if p.rel.MinimumVersion != "" {
		body = "minimum_version: " + p.rel.MinimumVersion + "\n\n" + body
	}
	htmlURL := p.rel.HTMLURL
	if htmlURL == "" {
		htmlURL = base + "/release/" + p.rel.Tag
	}
	assets := make([]map[string]interface{}, 0, len(p.assets))
	for _, a := range p.assets {
		assets = append(assets, map[string]interface{}{
			"name":                 a.Name,
			"browser_download_url": base + "/download/" + a.Name,
			"content_type":         a.ContentType,
		})
	}
	return map[string]interface{}{
		"tag_name": p.rel.Tag,
		"html_url": htmlURL,
		"body":     body,
		"assets":   assets,
	}
}

func writeReleaseJSON(w http.ResponseWriter, p *prepared, base string) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(releaseMap(p, base))
}

// NewMockBinary returns a tiny fake binary payload tagged with the given
// version string. It is NOT a real executable — consumers that actually
// exec the returned bytes should instead use go build against a TestMain
// fixture. For checksum and extract-flow tests this is sufficient.
func NewMockBinary(version string) []byte {
	return []byte("mock-binary version=" + version + " padding=" +
		strings.Repeat("x", 128))
}

// buildTarGz packs one file as name into a tar.gz archive.
func buildTarGz(name string, data []byte) []byte {
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)
	hdr := &tar.Header{
		Name: name,
		Mode: 0o755,
		Size: int64(len(data)),
	}
	_ = tw.WriteHeader(hdr)
	_, _ = tw.Write(data)
	_ = tw.Close()
	_ = gw.Close()
	return buf.Bytes()
}
