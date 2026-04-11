package blit

import (
	"os"
	"sync"
)

// FeatureFlag provides runtime feature toggling for blit applications.
//
// Use feature flags to gate experimental components, A/B test UI behavior,
// or conditionally enable functionality without rebuilding. Flags can be
// set programmatically or via environment variables (BLIT_FLAG_<NAME>=1).
//
// Flag resolution order (first match wins):
//  1. Explicit override set via Set or WithFeatureFlags
//  2. Environment variable BLIT_FLAG_<NAME> (1 = enabled, 0 = disabled)
//  3. Default value from the flag definition
type FeatureFlag struct {
	mu    sync.RWMutex
	flags map[string]bool
	defs  map[string]flagDef
}

type flagDef struct {
	defaultValue bool
	description  string
}

// NewFeatureFlag creates a flag registry with the given definitions.
// Each entry is (name, defaultValue, description).
func NewFeatureFlag(defs ...FlagDef) *FeatureFlag {
	ff := &FeatureFlag{
		flags: make(map[string]bool),
		defs:  make(map[string]flagDef),
	}
	for _, d := range defs {
		ff.defs[d.Name] = flagDef{defaultValue: d.Default, description: d.Description}
	}
	return ff
}

// FlagDef describes a single feature flag.
type FlagDef struct {
	// Name is the flag identifier. Must be uppercase snake_case (e.g. "NEW_TABLE").
	Name string

	// Default is the value when no override or env var is set.
	Default bool

	// Description explains what the flag controls.
	Description string
}

// Enabled returns whether the named flag is active.
//
// Resolution: explicit override → env var BLIT_FLAG_<NAME> → default.
// Returns false for unknown flag names.
func (ff *FeatureFlag) Enabled(name string) bool {
	ff.mu.RLock()
	defer ff.mu.RUnlock()

	// 1. Explicit override
	if v, ok := ff.flags[name]; ok {
		return v
	}

	// 2. Environment variable
	if v, ok := envFlag(name); ok {
		return v
	}

	// 3. Default
	if d, ok := ff.defs[name]; ok {
		return d.defaultValue
	}

	return false
}

// Set overrides a flag value. The value persists until Clear is called.
func (ff *FeatureFlag) Set(name string, enabled bool) {
	ff.mu.Lock()
	defer ff.mu.Unlock()
	ff.flags[name] = enabled
}

// Clear removes an explicit override, reverting to env/default resolution.
func (ff *FeatureFlag) Clear(name string) {
	ff.mu.Lock()
	defer ff.mu.Unlock()
	delete(ff.flags, name)
}

// Names returns all defined flag names.
func (ff *FeatureFlag) Names() []string {
	ff.mu.RLock()
	defer ff.mu.RUnlock()
	names := make([]string, 0, len(ff.defs))
	for n := range ff.defs {
		names = append(names, n)
	}
	return names
}

// Describe returns the description for a flag, or "" if undefined.
func (ff *FeatureFlag) Describe(name string) string {
	ff.mu.RLock()
	defer ff.mu.RUnlock()
	if d, ok := ff.defs[name]; ok {
		return d.description
	}
	return ""
}

// Snapshot returns a map of all flags and their current resolved values.
func (ff *FeatureFlag) Snapshot() map[string]bool {
	ff.mu.RLock()
	defer ff.mu.RUnlock()
	out := make(map[string]bool, len(ff.defs))
	for name := range ff.defs {
		if v, ok := ff.flags[name]; ok {
			out[name] = v
			continue
		}
		if v, ok := envFlag(name); ok {
			out[name] = v
			continue
		}
		out[name] = ff.defs[name].defaultValue
	}
	return out
}

// envFlag checks the BLIT_FLAG_<NAME> environment variable.
// Returns (value, true) if the variable is set, (false, false) otherwise.
func envFlag(name string) (bool, bool) {
	key := "BLIT_FLAG_" + name
	v := os.Getenv(key)
	if v == "1" {
		return true, true
	}
	if v == "0" {
		return false, true
	}
	return false, false
}
