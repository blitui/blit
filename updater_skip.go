package tuikit

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

// skipVersionMu serializes concurrent reads/writes to the skip-version file.
// The file is tiny (one map of version strings) so a package-level mutex
// is sufficient — we avoid a more elaborate lock manager.
var skipVersionMu sync.Mutex

// skippedVersionsFile returns the on-disk path where skipped versions are
// persisted for this cfg. Uses the same cache directory as the update check.
func skippedVersionsFile(cfg UpdateConfig) string {
	dir := cfg.CacheDir
	if dir == "" {
		base, err := os.UserConfigDir()
		if err != nil {
			base = os.TempDir()
		}
		dir = filepath.Join(base, cfg.BinaryName)
	}
	return filepath.Join(dir, "skipped-versions.json")
}

// skipRecord is the on-disk schema for skipped-versions.json.
// A map is used for O(1) membership checks; the bool is always true.
type skipRecord struct {
	Versions map[string]bool `json:"versions"`
}

// SkipVersion marks v as skipped so future CheckForUpdate calls return
// Available=false for that exact tag. Writes to
// <CacheDir>/skipped-versions.json. Errors are returned but not logged.
func SkipVersion(cfg UpdateConfig, v string) error {
	if v == "" {
		return nil
	}
	skipVersionMu.Lock()
	defer skipVersionMu.Unlock()

	path := skippedVersionsFile(cfg)
	rec := readSkipRecordLocked(path)
	if rec.Versions == nil {
		rec.Versions = make(map[string]bool)
	}
	rec.Versions[v] = true

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(rec, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// IsVersionSkipped reports whether v has been previously passed to
// SkipVersion for this cfg. Missing file or unreadable file returns false.
func IsVersionSkipped(cfg UpdateConfig, v string) bool {
	if v == "" {
		return false
	}
	skipVersionMu.Lock()
	defer skipVersionMu.Unlock()
	rec := readSkipRecordLocked(skippedVersionsFile(cfg))
	return rec.Versions[v]
}

// ClearSkippedVersions removes all skip entries. Useful for tests and for a
// "reset" menu item in a consumer app.
func ClearSkippedVersions(cfg UpdateConfig) error {
	skipVersionMu.Lock()
	defer skipVersionMu.Unlock()
	path := skippedVersionsFile(cfg)
	err := os.Remove(path)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

// readSkipRecordLocked reads the skip file; assumes skipVersionMu is held.
// Never returns an error: missing/bad files yield an empty record so
// callers can treat "no skips" as the default.
func readSkipRecordLocked(path string) skipRecord {
	var rec skipRecord
	data, err := os.ReadFile(path)
	if err != nil {
		return rec
	}
	_ = json.Unmarshal(data, &rec)
	if rec.Versions == nil {
		rec.Versions = make(map[string]bool)
	}
	return rec
}
