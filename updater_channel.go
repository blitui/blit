package tuikit

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Update channels.
const (
	ChannelStable     = "stable"
	ChannelBeta       = "beta"
	ChannelPrerelease = "prerelease"
)

// FetchReleases returns the full list of releases for the repo, newest first,
// using the rate-limited fetcher.
func FetchReleases(baseURL, owner, repo string) ([]Release, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/releases", baseURL, owner, repo)
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := fetchWithBackoff(client, url, 3)
	if err != nil {
		return nil, fmt.Errorf("fetching releases: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}
	var out []Release
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, fmt.Errorf("parsing releases: %w", err)
	}
	return out, nil
}

// PickReleaseForChannel filters releases according to the update channel:
//   - stable: first non-prerelease release (tag without a pre-release suffix)
//   - beta: first release whose tag contains "-beta"
//   - prerelease: first release marked Prerelease=true or tag containing "-"
//
// Returns nil if no matching release is found.
func PickReleaseForChannel(releases []Release, channel string) *Release {
	if channel == "" {
		channel = ChannelStable
	}
	for i := range releases {
		r := &releases[i]
		switch channel {
		case ChannelStable:
			if !r.Prerelease && !strings.Contains(r.TagName, "-") {
				return r
			}
		case ChannelBeta:
			if strings.Contains(strings.ToLower(r.TagName), "-beta") {
				return r
			}
		case ChannelPrerelease:
			if r.Prerelease || strings.Contains(r.TagName, "-") {
				return r
			}
		}
	}
	return nil
}

// FetchLatestReleaseForChannel selects the appropriate release for the
// configured channel. Stable falls through to FetchLatestRelease for backward
// compatibility; non-stable channels walk /releases.
func FetchLatestReleaseForChannel(baseURL, owner, repo, channel string) (*Release, error) {
	if channel == "" || channel == ChannelStable {
		return FetchLatestRelease(baseURL, owner, repo)
	}
	releases, err := FetchReleases(baseURL, owner, repo)
	if err != nil {
		return nil, err
	}
	rel := PickReleaseForChannel(releases, channel)
	if rel == nil {
		return nil, fmt.Errorf("no release found for channel %q", channel)
	}
	return rel, nil
}
