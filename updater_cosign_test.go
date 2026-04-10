package blit_test

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"net/http"
	"net/http/httptest"
	"runtime"
	"strings"
	"testing"

	blit "github.com/blitui/blit"
)

// generateTestKeyPair returns a fresh ed25519 key pair for use in tests.
func generateTestKeyPair(t *testing.T) (ed25519.PublicKey, ed25519.PrivateKey) {
	t.Helper()
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("generating ed25519 key pair: %v", err)
	}
	return pub, priv
}

// pubKeyToPEM encodes an ed25519 public key as a minimal SubjectPublicKeyInfo PEM block.
func pubKeyToPEM(pub ed25519.PublicKey) string {
	spkiHeader := []byte{0x30, 0x2a, 0x30, 0x05, 0x06, 0x03, 0x2b, 0x65, 0x70, 0x03, 0x21, 0x00}
	body := make([]byte, 0, len(spkiHeader)+len(pub))
	body = append(body, spkiHeader...)
	body = append(body, pub...)
	block := &pem.Block{Type: "PUBLIC KEY", Bytes: body}
	return string(pem.EncodeToMemory(block))
}

// pubKeyToBase64 encodes an ed25519 public key as bare base64.
func pubKeyToBase64(pub ed25519.PublicKey) string {
	return base64.StdEncoding.EncodeToString(pub)
}

// cosignSign signs data with priv and returns the base64-encoded signature
// (the format written to <asset>.sig by cosign sign-blob).
func cosignSign(t *testing.T, priv ed25519.PrivateKey, data []byte) []byte {
	t.Helper()
	sig := ed25519.Sign(priv, data)
	return []byte(base64.StdEncoding.EncodeToString(sig))
}

// makeChecksumFile returns a GoReleaser-style checksums.txt body.
func makeChecksumFile(archive []byte, assetName string) []byte {
	h := sha256.Sum256(archive)
	return []byte(fmt.Sprintf("%x  %s\n", h, assetName))
}

func TestMatchSigAsset(t *testing.T) {
	assets := []blit.ReleaseAsset{
		{Name: "myapp_1.0.0_linux_amd64.tar.gz", DownloadURL: "https://example.com/asset"},
		{Name: "myapp_1.0.0_linux_amd64.tar.gz.sig", DownloadURL: "https://example.com/asset.sig"},
		{Name: "checksums.txt", DownloadURL: "https://example.com/checksums"},
	}

	t.Run("found", func(t *testing.T) {
		got, err := blit.MatchSigAsset(assets, "myapp_1.0.0_linux_amd64.tar.gz")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.DownloadURL != "https://example.com/asset.sig" {
			t.Errorf("DownloadURL = %q, want sig URL", got.DownloadURL)
		}
	})

	t.Run("not found", func(t *testing.T) {
		_, err := blit.MatchSigAsset(assets, "nonexistent.tar.gz")
		if err == nil {
			t.Error("expected error for missing .sig asset")
		}
	})
}

func TestVerifyCosignSignature(t *testing.T) {
	pub, priv := generateTestKeyPair(t)
	data := []byte("fake release asset bytes")
	sigData := cosignSign(t, priv, data)

	t.Run("valid signature PEM key", func(t *testing.T) {
		err := blit.VerifyCosignSignature(data, sigData, pubKeyToPEM(pub))
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("valid signature base64 key", func(t *testing.T) {
		err := blit.VerifyCosignSignature(data, sigData, pubKeyToBase64(pub))
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("tampered data", func(t *testing.T) {
		err := blit.VerifyCosignSignature([]byte("tampered"), sigData, pubKeyToPEM(pub))
		if err == nil {
			t.Error("expected verification failure for tampered data")
		}
	})

	t.Run("wrong key", func(t *testing.T) {
		pub2, _ := generateTestKeyPair(t)
		err := blit.VerifyCosignSignature(data, sigData, pubKeyToPEM(pub2))
		if err == nil {
			t.Error("expected verification failure for wrong key")
		}
	})

	t.Run("invalid signature bytes", func(t *testing.T) {
		err := blit.VerifyCosignSignature(data, []byte("not-valid-base64!!!"), pubKeyToPEM(pub))
		if err == nil {
			t.Error("expected error for invalid base64 signature")
		}
	})

	t.Run("invalid public key", func(t *testing.T) {
		err := blit.VerifyCosignSignature(data, sigData, "not-a-valid-key")
		if err == nil {
			t.Error("expected error for invalid public key")
		}
	})
}

func TestSelfUpdateCosignVerification(t *testing.T) {
	pub, priv := generateTestKeyPair(t)

	binaryContent := []byte("fake binary for cosign test")
	ext := "tar.gz"
	binaryFile := "myapp"
	if runtime.GOOS == "windows" {
		ext = "zip"
		binaryFile = "myapp.exe"
	}
	assetName := fmt.Sprintf("myapp_1.0.0_%s_%s.%s", runtime.GOOS, runtime.GOARCH, ext)

	var archive []byte
	if ext == "zip" {
		archive = createTestZip(t, binaryFile, binaryContent)
	} else {
		archive = createTestTarGz(t, binaryFile, binaryContent)
	}

	validSig := cosignSign(t, priv, archive)
	_, wrongPriv := generateTestKeyPair(t)
	invalidSig := cosignSign(t, wrongPriv, archive)
	checksumFile := makeChecksumFile(archive, assetName)

	tests := []struct {
		name        string
		sig         []byte
		pubKey      string
		wantCosignErr bool
	}{
		{
			name:          "valid cosign signature passes verification",
			sig:           validSig,
			pubKey:        pubKeyToPEM(pub),
			wantCosignErr: false,
		},
		{
			name:          "invalid cosign signature blocks update",
			sig:           invalidSig,
			pubKey:        pubKeyToPEM(pub),
			wantCosignErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sig := tt.sig
			var srvURL string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/repos/owner/repo/releases/latest":
					w.Header().Set("Content-Type", "application/json")
					fmt.Fprintf(w, `{
						"tag_name": "v1.0.0",
						"html_url": "https://github.com/owner/repo/releases/tag/v1.0.0",
						"body": "",
						"assets": [
							{"name": %q, "browser_download_url": %q},
							{"name": %q, "browser_download_url": %q},
							{"name": "checksums.txt", "browser_download_url": %q}
						]
					}`,
						assetName, srvURL+"/download/asset",
						assetName+".sig", srvURL+"/download/sig",
						srvURL+"/download/checksums",
					)
				case "/download/asset":
					w.Write(archive)
				case "/download/sig":
					w.Write(sig)
				case "/download/checksums":
					w.Write(checksumFile)
				default:
					http.NotFound(w, r)
				}
			}))
			defer srv.Close()
			srvURL = srv.URL

			cfg := blit.UpdateConfig{
				Owner:           "owner",
				Repo:            "repo",
				BinaryName:      "myapp",
				Version:         "v0.9.0",
				CacheDir:        t.TempDir(),
				APIBaseURL:      srv.URL,
				CosignPublicKey: tt.pubKey,
			}

			err := blit.SelfUpdate(cfg)

			if tt.wantCosignErr {
				if err == nil {
					t.Fatalf("expected cosign error, got nil")
				}
				if !strings.Contains(err.Error(), "cosign") {
					t.Errorf("error %q does not mention cosign", err.Error())
				}
			} else if err != nil && strings.Contains(err.Error(), "cosign") {
				// Valid sig: SelfUpdate may still fail at replaceBinary (test env),
				// but must not fail at the cosign step.
				t.Errorf("unexpected cosign error: %v", err)
			}
		})
	}
}

func TestSelfUpdateCosignMissingSigAsset(t *testing.T) {
	// B4: public key set but no .sig asset in release → verification failure.
	pub, _ := generateTestKeyPair(t)

	binaryContent := []byte("fake binary missing sig")
	ext := "tar.gz"
	binaryFile := "myapp"
	if runtime.GOOS == "windows" {
		ext = "zip"
		binaryFile = "myapp.exe"
	}
	assetName := fmt.Sprintf("myapp_1.0.0_%s_%s.%s", runtime.GOOS, runtime.GOARCH, ext)

	var archive []byte
	if ext == "zip" {
		archive = createTestZip(t, binaryFile, binaryContent)
	} else {
		archive = createTestTarGz(t, binaryFile, binaryContent)
	}
	checksumFile := makeChecksumFile(archive, assetName)

	var srvURL string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/owner/repo/releases/latest":
			w.Header().Set("Content-Type", "application/json")
			// No .sig asset listed.
			fmt.Fprintf(w, `{
				"tag_name": "v1.0.0",
				"html_url": "https://github.com/owner/repo/releases/tag/v1.0.0",
				"body": "",
				"assets": [
					{"name": %q, "browser_download_url": %q},
					{"name": "checksums.txt", "browser_download_url": %q}
				]
			}`,
				assetName, srvURL+"/download/asset",
				srvURL+"/download/checksums",
			)
		case "/download/asset":
			w.Write(archive)
		case "/download/checksums":
			w.Write(checksumFile)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()
	srvURL = srv.URL

	cfg := blit.UpdateConfig{
		Owner:           "owner",
		Repo:            "repo",
		BinaryName:      "myapp",
		Version:         "v0.9.0",
		CacheDir:        t.TempDir(),
		APIBaseURL:      srv.URL,
		CosignPublicKey: pubKeyToPEM(pub),
	}

	err := blit.SelfUpdate(cfg)
	if err == nil {
		t.Fatal("expected error when .sig asset is missing, got nil")
	}
	if !strings.Contains(err.Error(), "finding signature asset") {
		t.Errorf("error %q should mention finding signature asset", err.Error())
	}
}

func TestSelfUpdateCosignSkippedWhenKeyEmpty(t *testing.T) {
	binaryContent := []byte("fake binary no cosign")
	ext := "tar.gz"
	binaryFile := "myapp"
	if runtime.GOOS == "windows" {
		ext = "zip"
		binaryFile = "myapp.exe"
	}
	assetName := fmt.Sprintf("myapp_1.0.0_%s_%s.%s", runtime.GOOS, runtime.GOARCH, ext)

	var archive []byte
	if ext == "zip" {
		archive = createTestZip(t, binaryFile, binaryContent)
	} else {
		archive = createTestTarGz(t, binaryFile, binaryContent)
	}
	checksumFile := makeChecksumFile(archive, assetName)

	sigFetched := false
	var srvURL string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/owner/repo/releases/latest":
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, `{
				"tag_name": "v1.0.0",
				"html_url": "https://github.com/owner/repo/releases/tag/v1.0.0",
				"body": "",
				"assets": [
					{"name": %q, "browser_download_url": %q},
					{"name": %q, "browser_download_url": %q},
					{"name": "checksums.txt", "browser_download_url": %q}
				]
			}`,
				assetName, srvURL+"/download/asset",
				assetName+".sig", srvURL+"/download/sig",
				srvURL+"/download/checksums",
			)
		case "/download/asset":
			w.Write(archive)
		case "/download/sig":
			sigFetched = true
			w.Write([]byte("should-not-be-fetched"))
		case "/download/checksums":
			w.Write(checksumFile)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()
	srvURL = srv.URL

	cfg := blit.UpdateConfig{
		Owner:           "owner",
		Repo:            "repo",
		BinaryName:      "myapp",
		Version:         "v0.9.0",
		CacheDir:        t.TempDir(),
		APIBaseURL:      srv.URL,
		CosignPublicKey: "", // empty — skip verification
	}

	_ = blit.SelfUpdate(cfg)

	if sigFetched {
		t.Error("sig asset was fetched even though CosignPublicKey was empty")
	}
}
