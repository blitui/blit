package tuikit

import (
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
