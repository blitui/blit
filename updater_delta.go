package blit

import (
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strings"

	"github.com/gabstv/go-bsdiff/pkg/bspatch"
)

// DeltaPatchSuffix is appended to the binary name when building a delta
// patch asset name. A patch asset is expected to be named:
//
//	<binary>_<fromVersion>_to_<toVersion>_<goos>_<goarch>.bsdiff
//
// The <fromVersion> is the version the patch applies against (the caller's
// current version) and <toVersion> is the version the release publishes.
const DeltaPatchSuffix = ".bsdiff"

// MatchDeltaAsset looks for a delta patch asset that upgrades fromVersion to
// toVersion for the given binary, OS, and architecture. Both version strings
// are normalized by stripping a leading "v" so callers can pass either form.
// Returns an error if no matching asset exists.
func MatchDeltaAsset(assets []ReleaseAsset, binaryName, fromVersion, toVersion, goos, goarch string) (ReleaseAsset, error) {
	from := strings.TrimPrefix(fromVersion, "v")
	to := strings.TrimPrefix(toVersion, "v")
	if from == "" || to == "" {
		return ReleaseAsset{}, fmt.Errorf("delta asset requires both from and to versions")
	}
	// Canonical form: <binary>_<from>_to_<to>_<goos>_<goarch>.bsdiff
	want := fmt.Sprintf("%s_%s_to_%s_%s_%s%s", binaryName, from, to, goos, goarch, DeltaPatchSuffix)
	for _, a := range assets {
		if a.Name == want {
			return a, nil
		}
	}
	return ReleaseAsset{}, fmt.Errorf("no delta asset %q in release", want)
}

// ApplyDeltaPatch reconstructs the new binary by applying a bsdiff patch to
// the current on-disk binary at exePath. The returned bytes are the new
// binary contents ready for checksum verification and replacement. On any
// failure the caller should fall back to a full download.
func ApplyDeltaPatch(exePath string, patch []byte) ([]byte, error) {
	if len(patch) == 0 {
		return nil, fmt.Errorf("empty patch")
	}
	oldBinary, err := os.ReadFile(exePath)
	if err != nil {
		return nil, fmt.Errorf("reading current binary: %w", err)
	}
	newBinary, err := bspatch.Bytes(oldBinary, patch)
	if err != nil {
		return nil, fmt.Errorf("applying bsdiff patch: %w", err)
	}
	if len(newBinary) == 0 {
		return nil, fmt.Errorf("patch produced empty binary")
	}
	return newBinary, nil
}

// tryDeltaUpdate attempts the opt-in delta update path. Returns the new
// binary bytes and the asset name used for checksum lookup on success.
// Any error indicates the caller should fall back to full download. This
// function never modifies disk state itself beyond reading the current
// executable.
func tryDeltaUpdate(cfg UpdateConfig, rel *Release, exePath string, client *http.Client) ([]byte, string, error) {
	if !cfg.EnableDeltaUpdates {
		return nil, "", fmt.Errorf("delta updates disabled")
	}
	if cfg.Version == "" || cfg.Version == "dev" {
		return nil, "", fmt.Errorf("delta requires a concrete current version")
	}
	patchAsset, err := MatchDeltaAsset(rel.Assets, cfg.BinaryName, cfg.Version, rel.TagName, runtime.GOOS, runtime.GOARCH)
	if err != nil {
		return nil, "", err
	}
	checksumAsset, err := MatchChecksumAsset(rel.Assets)
	if err != nil {
		return nil, "", fmt.Errorf("finding checksums: %w", err)
	}
	patchData, err := downloadURL(client, patchAsset.DownloadURL)
	if err != nil {
		return nil, "", fmt.Errorf("downloading patch: %w", err)
	}
	checksumData, err := downloadURL(client, checksumAsset.DownloadURL)
	if err != nil {
		return nil, "", fmt.Errorf("downloading checksums: %w", err)
	}
	// Verify the patch asset itself matches checksums.txt before applying.
	if err := VerifyChecksum(patchData, patchAsset.Name, checksumData); err != nil {
		return nil, "", fmt.Errorf("patch checksum failed: %w", err)
	}
	newBinary, err := ApplyDeltaPatch(exePath, patchData)
	if err != nil {
		return nil, "", err
	}
	return newBinary, patchAsset.Name, nil
}

// deltaUpdateLogf logs delta updater events to stderr. Kept minimal so the
// updater package does not take a logging dependency.
func deltaUpdateLogf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "blit updater [delta]: "+format+"\n", args...)
}
