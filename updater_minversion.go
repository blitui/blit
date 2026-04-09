package tuikit

import (
	"regexp"
)

// minimumVersionRE matches a line like "minimum_version: v1.2.3" (case
// insensitive, leading whitespace allowed). The marker may appear anywhere
// in a release body — typically a Markdown metadata line at the top.
var minimumVersionRE = regexp.MustCompile(`(?im)^\s*minimum[_-]version\s*[:=]\s*(v?\d+\.\d+\.\d+)\s*$`)

// ParseMinimumVersion extracts a "minimum_version: vX.Y.Z" marker from a
// release body (typically Markdown). Returns the parsed Version and true on
// match, zero Version and false otherwise. Supports both "minimum_version"
// and "minimum-version", and the marker is case-insensitive.
func ParseMinimumVersion(releaseBody string) (Version, bool) {
	if releaseBody == "" {
		return Version{}, false
	}
	m := minimumVersionRE.FindStringSubmatch(releaseBody)
	if len(m) < 2 {
		return Version{}, false
	}
	v, err := ParseVersion(m[1])
	if err != nil {
		return Version{}, false
	}
	return v, true
}
