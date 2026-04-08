package tuikit

import (
	"archive/tar"
	"archive/zip"
	"bufio"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"time"
)

// Version represents a parsed semantic version.
type Version struct {
	Major int
	Minor int
	Patch int
}

// ParseVersion parses a semver string like "v1.2.3" or "1.2.3".
// Pre-release (-rc.1) and build metadata (+dirty, +incompatible) suffixes are stripped.
func ParseVersion(s string) (Version, error) {
	s = strings.TrimPrefix(s, "v")
	// Strip build metadata (+dirty, +incompatible) and pre-release (-rc.1)
	if i := strings.IndexByte(s, '+'); i != -1 {
		s = s[:i]
	}
	if i := strings.IndexByte(s, '-'); i != -1 {
		s = s[:i]
	}
	parts := strings.Split(s, ".")
	if len(parts) != 3 {
		return Version{}, fmt.Errorf("invalid version %q: expected MAJOR.MINOR.PATCH", s)
	}
	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return Version{}, fmt.Errorf("invalid major version %q: %w", parts[0], err)
	}
	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return Version{}, fmt.Errorf("invalid minor version %q: %w", parts[1], err)
	}
	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		return Version{}, fmt.Errorf("invalid patch version %q: %w", parts[2], err)
	}
	return Version{Major: major, Minor: minor, Patch: patch}, nil
}

// NewerThan reports whether v is a newer version than other.
func (v Version) NewerThan(other Version) bool {
	if v.Major != other.Major {
		return v.Major > other.Major
	}
	if v.Minor != other.Minor {
		return v.Minor > other.Minor
	}
	return v.Patch > other.Patch
}

// String returns the version as "vMAJOR.MINOR.PATCH".
func (v Version) String() string {
	return fmt.Sprintf("v%d.%d.%d", v.Major, v.Minor, v.Patch)
}

// InstallMethod describes how the binary was installed.
type InstallMethod int

const (
	// InstallManual means the binary was installed manually or via go install.
	InstallManual InstallMethod = iota
	// InstallHomebrew means the binary was installed via Homebrew.
	InstallHomebrew
	// InstallScoop means the binary was installed via Scoop.
	InstallScoop
)

// DetectInstallMethod inspects a binary path to determine how it was installed.
func DetectInstallMethod(path string) InstallMethod {
	lower := strings.ToLower(path)
	if strings.Contains(lower, "cellar") || strings.Contains(lower, "homebrew") || strings.Contains(lower, "linuxbrew") {
		return InstallHomebrew
	}
	if strings.Contains(lower, "scoop") {
		return InstallScoop
	}
	return InstallManual
}

// UpdateCache is the cached result of an update check.
type UpdateCache struct {
	CheckedAt     time.Time `json:"checked_at"`
	LatestVersion string    `json:"latest_version"`
	ReleaseURL    string    `json:"release_url"`
	ReleaseNotes  string    `json:"release_notes"`
}

// IsFresh reports whether the cache is still within the TTL.
func (c UpdateCache) IsFresh(ttl time.Duration) bool {
	return time.Since(c.CheckedAt) < ttl
}

// ReadCache reads the cached update check from disk.
func ReadCache(path string) (UpdateCache, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return UpdateCache{}, fmt.Errorf("reading cache: %w", err)
	}
	var cache UpdateCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return UpdateCache{}, fmt.Errorf("parsing cache: %w", err)
	}
	return cache, nil
}

// WriteCache writes the update check result to disk.
func WriteCache(path string, cache UpdateCache) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("creating cache dir: %w", err)
	}
	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling cache: %w", err)
	}
	return os.WriteFile(path, data, 0o644)
}

// ReleaseAsset represents a single asset in a GitHub release.
type ReleaseAsset struct {
	Name        string `json:"name"`
	DownloadURL string `json:"browser_download_url"`
}

// Release represents a GitHub release.
type Release struct {
	TagName string         `json:"tag_name"`
	HTMLURL string         `json:"html_url"`
	Body    string         `json:"body"`
	Assets  []ReleaseAsset `json:"assets"`
}

// FetchLatestRelease fetches the latest release from GitHub.
// baseURL is the API base (e.g. "https://api.github.com" or a test server URL).
func FetchLatestRelease(baseURL, owner, repo string) (*Release, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/releases/latest", baseURL, owner, repo)
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetching release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	var rel Release
	if err := json.Unmarshal(body, &rel); err != nil {
		return nil, fmt.Errorf("parsing release: %w", err)
	}
	return &rel, nil
}

// MatchAsset finds the release asset matching the given binary name, OS, and architecture.
func MatchAsset(assets []ReleaseAsset, binaryName, goos, goarch string) (ReleaseAsset, error) {
	suffix := fmt.Sprintf("_%s_%s.", goos, goarch)
	for _, a := range assets {
		if strings.Contains(a.Name, suffix) && strings.HasPrefix(a.Name, binaryName+"_") {
			return a, nil
		}
	}
	return ReleaseAsset{}, fmt.Errorf("no asset found for %s/%s", goos, goarch)
}

// MatchChecksumAsset finds the checksums.txt asset in a release.
func MatchChecksumAsset(assets []ReleaseAsset) (ReleaseAsset, error) {
	for _, a := range assets {
		if strings.EqualFold(a.Name, "checksums.txt") {
			return a, nil
		}
	}
	return ReleaseAsset{}, fmt.Errorf("no checksums.txt asset found")
}

// VerifyChecksum verifies the SHA256 checksum of data against a GoReleaser checksums.txt file.
// The checksums file format is: "<hex>  <filename>\n" per line.
func VerifyChecksum(data []byte, assetName string, checksumFile []byte) error {
	expected := ""
	scanner := bufio.NewScanner(strings.NewReader(string(checksumFile)))
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)
		if len(parts) == 2 && parts[1] == assetName {
			expected = parts[0]
			break
		}
	}
	if expected == "" {
		return fmt.Errorf("no checksum found for %s", assetName)
	}

	actual := fmt.Sprintf("%x", sha256.Sum256(data))
	if actual != expected {
		return fmt.Errorf("checksum mismatch: got %s, want %s", actual, expected)
	}
	return nil
}

// UpdateMode controls how the update prompt behaves.
type UpdateMode int

const (
	// UpdateNotify shows a non-blocking notification in the TUI after startup.
	UpdateNotify UpdateMode = iota
	// UpdateBlocking prompts the user in stdout before the TUI starts.
	UpdateBlocking
)

// ProgressFunc is called during download with bytes received and total bytes.
// If total is -1, the content length is unknown.
type ProgressFunc func(received, total int64)

// UpdateConfig configures the auto-update system.
type UpdateConfig struct {
	Owner      string        // GitHub repo owner
	Repo       string        // GitHub repo name
	BinaryName string        // Binary name in release assets (e.g. "cryptstream")
	Version    string        // Current version (set via ldflags; "dev" or "" skips check)
	Mode       UpdateMode    // UpdateNotify or UpdateBlocking
	CacheTTL   time.Duration // How long to cache the check result (default: 1h)
	CacheDir   string        // Override cache directory (default: os.UserConfigDir()/<BinaryName>)

	// OnProgress is called during binary asset download with bytes received and total bytes.
	// If nil, no progress reporting occurs.
	OnProgress ProgressFunc

	// APIBaseURL overrides the GitHub API URL. Leave empty for production.
	// Exported for testing; not intended for consumer use.
	APIBaseURL string
}

// ValidateConfig checks that cfg has all required fields set.
// Returns a descriptive error if any required field is missing.
func ValidateConfig(cfg UpdateConfig) error {
	if cfg.Owner == "" {
		return fmt.Errorf("UpdateConfig.Owner is required")
	}
	if cfg.Repo == "" {
		return fmt.Errorf("UpdateConfig.Repo is required")
	}
	if cfg.BinaryName == "" {
		return fmt.Errorf("UpdateConfig.BinaryName is required")
	}
	if cfg.Version == "" {
		return fmt.Errorf("UpdateConfig.Version is required")
	}
	return nil
}

func (c UpdateConfig) githubBaseURL() string {
	if c.APIBaseURL != "" {
		return c.APIBaseURL
	}
	return "https://api.github.com"
}

func (c UpdateConfig) cachePath() string {
	dir := c.CacheDir
	if dir == "" {
		base, err := os.UserConfigDir()
		if err != nil {
			base = os.TempDir()
		}
		dir = filepath.Join(base, c.BinaryName)
	}
	return filepath.Join(dir, "update-check.json")
}

func (c UpdateConfig) ttl() time.Duration {
	if c.CacheTTL <= 0 {
		return 1 * time.Hour
	}
	return c.CacheTTL
}

// UpdateResult holds the result of an update check.
type UpdateResult struct {
	Available      bool
	CurrentVersion string
	LatestVersion  string
	ReleaseURL     string
	ReleaseNotes   string
	InstallMethod  InstallMethod
}

// versionFromBuildInfo reads the module version embedded by go install.
// Returns "" if unavailable or if the version is "(devel)".
func versionFromBuildInfo() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return ""
	}
	// go install sets Main.Version to the module version (e.g. "v0.6.0").
	// Local builds set it to "(devel)".
	v := info.Main.Version
	if v == "" || v == "(devel)" {
		return ""
	}
	return v
}

// CheckForUpdate checks GitHub Releases for a newer version.
// If the version is "dev" or empty, it attempts to read the module version
// from Go's embedded build info (set by go install). Returns a zero-value
// UpdateResult (Available=false) if no version can be determined.
// Network/API errors return a zero-value result with no error (fail-silent).
func CheckForUpdate(cfg UpdateConfig) (*UpdateResult, error) {
	// Resolve version from build info when not set via ldflags
	if cfg.Version == "" || cfg.Version == "dev" {
		if v := versionFromBuildInfo(); v != "" {
			cfg.Version = v
		}
	}

	result := &UpdateResult{CurrentVersion: cfg.Version}

	// Skip if still no usable version
	if cfg.Version == "" || cfg.Version == "dev" {
		return result, nil
	}

	if err := ValidateConfig(cfg); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	current, err := ParseVersion(cfg.Version)
	if err != nil {
		return result, nil // unparseable version, skip silently
	}

	// Check cache
	cachePath := cfg.cachePath()
	cache, cacheErr := ReadCache(cachePath)
	if cacheErr == nil && cache.IsFresh(cfg.ttl()) {
		latest, err := ParseVersion(cache.LatestVersion)
		if err == nil {
			result.LatestVersion = cache.LatestVersion
			result.ReleaseURL = cache.ReleaseURL
			result.ReleaseNotes = cache.ReleaseNotes
			result.Available = latest.NewerThan(current)
		}
		return result, nil
	}

	// Fetch from GitHub
	rel, err := FetchLatestRelease(cfg.githubBaseURL(), cfg.Owner, cfg.Repo)
	if err != nil {
		return result, nil // network error, skip silently
	}

	// Write to cache
	newCache := UpdateCache{
		CheckedAt:     time.Now().UTC(),
		LatestVersion: rel.TagName,
		ReleaseURL:    rel.HTMLURL,
		ReleaseNotes:  rel.Body,
	}
	_ = WriteCache(cachePath, newCache) // best-effort cache write

	latest, err := ParseVersion(rel.TagName)
	if err != nil {
		return result, nil
	}

	result.LatestVersion = rel.TagName
	result.ReleaseURL = rel.HTMLURL
	result.ReleaseNotes = rel.Body
	result.Available = latest.NewerThan(current)

	return result, nil
}

// WithAutoUpdate enables automatic update checking on app startup.
// In UpdateNotify mode, a notification is shown if an update is available.
// In UpdateBlocking mode, the user is prompted in stdout before the TUI starts.
func WithAutoUpdate(cfg UpdateConfig) Option {
	return func(a *appModel) {
		a.updateConfig = &cfg
	}
}

// ExtractBinary extracts the named binary from a tar.gz or zip archive.
func ExtractBinary(archive []byte, binaryName, format string) ([]byte, error) {
	switch format {
	case "tar.gz":
		return extractFromTarGz(archive, binaryName)
	case "zip":
		return extractFromZip(archive, binaryName)
	default:
		return nil, fmt.Errorf("unsupported archive format: %s", format)
	}
}

func extractFromTarGz(data []byte, binaryName string) ([]byte, error) {
	gr, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("opening gzip: %w", err)
	}
	defer gr.Close()

	tr := tar.NewReader(gr)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("reading tar: %w", err)
		}
		name := filepath.Base(hdr.Name)
		if name == binaryName || name == binaryName+".exe" {
			return io.ReadAll(tr)
		}
	}
	return nil, fmt.Errorf("binary %q not found in archive", binaryName)
}

func extractFromZip(data []byte, binaryName string) ([]byte, error) {
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("opening zip: %w", err)
	}
	for _, f := range r.File {
		name := filepath.Base(f.Name)
		if name == binaryName || name == binaryName+".exe" {
			rc, err := f.Open()
			if err != nil {
				return nil, fmt.Errorf("opening %s: %w", f.Name, err)
			}
			defer rc.Close()
			return io.ReadAll(rc)
		}
	}
	return nil, fmt.Errorf("binary %q not found in archive", binaryName)
}

// SelfUpdate downloads the latest release and replaces the current binary.
// Only works for manual installs (not Homebrew/Scoop).
func SelfUpdate(cfg UpdateConfig) error {
	if err := ValidateConfig(cfg); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	rel, err := FetchLatestRelease(cfg.githubBaseURL(), cfg.Owner, cfg.Repo)
	if err != nil {
		return fmt.Errorf("fetching release: %w", err)
	}

	asset, err := MatchAsset(rel.Assets, cfg.BinaryName, runtime.GOOS, runtime.GOARCH)
	if err != nil {
		return fmt.Errorf("matching asset: %w", err)
	}

	checksumAsset, err := MatchChecksumAsset(rel.Assets)
	if err != nil {
		return fmt.Errorf("finding checksums: %w", err)
	}

	fmt.Printf("Downloading %s...\n", asset.Name)

	client := &http.Client{Timeout: 120 * time.Second}
	assetData, err := downloadURLWithProgress(client, asset.DownloadURL, cfg.OnProgress)
	if err != nil {
		return fmt.Errorf("downloading asset: %w", err)
	}

	checksumData, err := downloadURL(client, checksumAsset.DownloadURL)
	if err != nil {
		return fmt.Errorf("downloading checksums: %w", err)
	}

	if err := VerifyChecksum(assetData, asset.Name, checksumData); err != nil {
		return fmt.Errorf("checksum verification failed: %w", err)
	}

	format := "tar.gz"
	if strings.HasSuffix(asset.Name, ".zip") {
		format = "zip"
	}

	binaryData, err := ExtractBinary(assetData, cfg.BinaryName, format)
	if err != nil {
		return fmt.Errorf("extracting binary: %w", err)
	}

	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("finding executable path: %w", err)
	}
	exePath, err = filepath.EvalSymlinks(exePath)
	if err != nil {
		return fmt.Errorf("resolving symlinks: %w", err)
	}

	return replaceBinary(exePath, binaryData)
}

func downloadURL(client *http.Client, url string) ([]byte, error) {
	return downloadURLWithProgress(client, url, nil)
}

func downloadURLWithProgress(client *http.Client, url string, onProgress ProgressFunc) ([]byte, error) {
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("GET %s: %w", url, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	if onProgress == nil {
		return io.ReadAll(resp.Body)
	}

	total := int64(-1)
	if cl := resp.ContentLength; cl > 0 {
		total = cl
	}

	var buf bytes.Buffer
	var received int64
	chunk := make([]byte, 32*1024)
	for {
		n, readErr := resp.Body.Read(chunk)
		if n > 0 {
			buf.Write(chunk[:n])
			received += int64(n)
			onProgress(received, total)
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return nil, fmt.Errorf("reading response body: %w", readErr)
		}
	}
	return buf.Bytes(), nil
}

func replaceBinary(exePath string, newBinary []byte) error {
	info, err := os.Stat(exePath)
	if err != nil {
		return fmt.Errorf("stat %s: %w", exePath, err)
	}
	mode := info.Mode()

	newPath := exePath + ".new"
	oldPath := exePath + ".old"

	if err := os.WriteFile(newPath, newBinary, mode); err != nil {
		return fmt.Errorf("writing new binary: %w", err)
	}

	if err := os.Rename(exePath, oldPath); err != nil {
		os.Remove(newPath)
		return fmt.Errorf("backing up current binary: %w", err)
	}

	if err := os.Rename(newPath, exePath); err != nil {
		os.Rename(oldPath, exePath)
		return fmt.Errorf("replacing binary: %w", err)
	}

	os.Remove(oldPath)
	return nil
}

// CleanupOldBinary removes a leftover .old binary from a previous update.
// Call this early in main() or in the update check flow.
func CleanupOldBinary() {
	exe, err := os.Executable()
	if err != nil {
		return
	}
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return
	}
	os.Remove(exe + ".old")
}
