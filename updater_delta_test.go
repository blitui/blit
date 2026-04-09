package tuikit

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/gabstv/go-bsdiff/pkg/bsdiff"
	"github.com/gabstv/go-bsdiff/pkg/bspatch"
)

// makeBlob returns a deterministic pseudo-binary payload of size n seeded
// by seed. Using a seeded reader keeps the test hermetic while still giving
// bsdiff non-trivial input to compress against.
func makeBlob(t *testing.T, n int, seed byte) []byte {
	t.Helper()
	buf := make([]byte, n)
	for i := range buf {
		buf[i] = seed + byte(i%251)
	}
	return buf
}

func TestDeltaPatchRoundTrip(t *testing.T) {
	oldBinary := makeBlob(t, 8*1024, 0x10)
	newBinary := make([]byte, len(oldBinary))
	copy(newBinary, oldBinary)
	// Mutate a small window so the patch is meaningfully smaller than the
	// full binary but bspatch still has to do real work.
	for i := 100; i < 200; i++ {
		newBinary[i] ^= 0xFF
	}
	// Append a tail so the new binary differs in length too.
	newBinary = append(newBinary, []byte("v2-tail-bytes")...)

	patch, err := bsdiff.Bytes(oldBinary, newBinary)
	if err != nil {
		t.Fatalf("bsdiff.Bytes: %v", err)
	}
	if len(patch) == 0 {
		t.Fatal("empty patch")
	}

	tmp := t.TempDir()
	exePath := filepath.Join(tmp, "fakebin")
	if err := os.WriteFile(exePath, oldBinary, 0o755); err != nil {
		t.Fatalf("write old binary: %v", err)
	}

	got, err := ApplyDeltaPatch(exePath, patch)
	if err != nil {
		t.Fatalf("ApplyDeltaPatch: %v", err)
	}
	if !bytes.Equal(got, newBinary) {
		t.Fatalf("patched binary mismatch: got %d bytes, want %d", len(got), len(newBinary))
	}
}

func TestApplyDeltaPatchErrors(t *testing.T) {
	tmp := t.TempDir()
	exePath := filepath.Join(tmp, "fakebin")
	if err := os.WriteFile(exePath, []byte("original"), 0o755); err != nil {
		t.Fatalf("write: %v", err)
	}

	if _, err := ApplyDeltaPatch(exePath, nil); err == nil {
		t.Error("ApplyDeltaPatch(nil) should error")
	}
	if _, err := ApplyDeltaPatch(exePath, []byte("not a bsdiff patch")); err == nil {
		t.Error("ApplyDeltaPatch(garbage) should error")
	}
	if _, err := ApplyDeltaPatch(filepath.Join(tmp, "missing"), []byte("x")); err == nil {
		t.Error("ApplyDeltaPatch(missing file) should error")
	}
}

func TestMatchDeltaAsset(t *testing.T) {
	goos, goarch := runtime.GOOS, runtime.GOARCH
	assets := []ReleaseAsset{
		{Name: fmt.Sprintf("tuikit_0.11.0_to_0.12.0_%s_%s.bsdiff", goos, goarch), DownloadURL: "https://example.com/p"},
		{Name: "tuikit_0.12.0_linux_amd64.tar.gz"},
		{Name: "checksums.txt"},
	}

	got, err := MatchDeltaAsset(assets, "tuikit", "v0.11.0", "v0.12.0", goos, goarch)
	if err != nil {
		t.Fatalf("MatchDeltaAsset: %v", err)
	}
	if got.DownloadURL != "https://example.com/p" {
		t.Errorf("got wrong asset: %+v", got)
	}

	// No "v" prefix — should still match.
	if _, err := MatchDeltaAsset(assets, "tuikit", "0.11.0", "0.12.0", goos, goarch); err != nil {
		t.Errorf("bare versions should match: %v", err)
	}

	// Missing patch asset.
	if _, err := MatchDeltaAsset(assets, "tuikit", "v0.10.0", "v0.12.0", goos, goarch); err == nil {
		t.Error("expected error for missing patch asset")
	}

	// Missing versions.
	if _, err := MatchDeltaAsset(assets, "tuikit", "", "v0.12.0", goos, goarch); err == nil {
		t.Error("expected error for empty fromVersion")
	}
}

func TestDeltaSmallerThanFull(t *testing.T) {
	// Sanity check: a tiny mutation produces a patch much smaller than
	// the full binary. This is the whole point of delta updates.
	oldBinary := make([]byte, 128*1024)
	if _, err := rand.Read(oldBinary); err != nil {
		t.Fatalf("rand: %v", err)
	}
	newBinary := make([]byte, len(oldBinary))
	copy(newBinary, oldBinary)
	// Flip 16 bytes in the middle.
	for i := 1000; i < 1016; i++ {
		newBinary[i] ^= 0xA5
	}

	patch, err := bsdiff.Bytes(oldBinary, newBinary)
	if err != nil {
		t.Fatalf("bsdiff: %v", err)
	}
	if len(patch) >= len(newBinary) {
		t.Errorf("patch (%d) not smaller than full binary (%d)", len(patch), len(newBinary))
	}
	t.Logf("delta size: patch=%d full=%d ratio=%.2f%%", len(patch), len(newBinary), 100*float64(len(patch))/float64(len(newBinary)))

	// And it round-trips.
	got, err := bspatch.Bytes(oldBinary, patch)
	if err != nil {
		t.Fatalf("bspatch: %v", err)
	}
	if !bytes.Equal(got, newBinary) {
		t.Fatal("round trip mismatch")
	}
}

func BenchmarkDeltaApply(b *testing.B) {
	oldBinary := make([]byte, 1024*1024)
	for i := range oldBinary {
		oldBinary[i] = byte(i)
	}
	newBinary := make([]byte, len(oldBinary))
	copy(newBinary, oldBinary)
	for i := 500; i < 600; i++ {
		newBinary[i] ^= 0xFF
	}
	patch, err := bsdiff.Bytes(oldBinary, newBinary)
	if err != nil {
		b.Fatalf("bsdiff: %v", err)
	}
	b.ReportMetric(float64(len(patch)), "patch_bytes")
	b.ReportMetric(float64(len(newBinary)), "full_bytes")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := bspatch.Bytes(oldBinary, patch); err != nil {
			b.Fatal(err)
		}
	}
}
