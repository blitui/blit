package tuikit

import (
	"bufio"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
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
func ParseVersion(s string) (Version, error) {
	s = strings.TrimPrefix(s, "v")
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

// UpdateConfig configures the auto-update system.
type UpdateConfig struct {
	Owner      string        // GitHub repo owner
	Repo       string        // GitHub repo name
	BinaryName string        // Binary name in release assets (e.g. "cryptstream")
	Version    string        // Current version (set via ldflags; "dev" or "" skips check)
	Mode       UpdateMode    // UpdateNotify or UpdateBlocking
	CacheTTL   time.Duration // How long to cache the check result (default: 24h)
	CacheDir   string        // Override cache directory (default: os.UserConfigDir()/<BinaryName>)

	// APIBaseURL overrides the GitHub API URL. Leave empty for production.
	// Exported for testing; not intended for consumer use.
	APIBaseURL string
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
		return 24 * time.Hour
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

// CheckForUpdate checks GitHub Releases for a newer version.
// Returns a zero-value UpdateResult (Available=false) if the version is "dev" or empty.
// Network/API errors return a zero-value result with no error (fail-silent).
func CheckForUpdate(cfg UpdateConfig) (*UpdateResult, error) {
	result := &UpdateResult{CurrentVersion: cfg.Version}

	// Skip for dev builds
	if cfg.Version == "" || cfg.Version == "dev" {
		return result, nil
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
