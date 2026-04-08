package tuikit

import (
	"fmt"
	"strconv"
	"strings"
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
