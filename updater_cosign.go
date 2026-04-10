// Package blit provides TUI components and update utilities.
//
// Cosign signature verification — implementation notes:
//
// We deliberately avoid the full sigstore/cosign library (github.com/sigstore/cosign)
// because it pulls in hundreds of transitive dependencies (OCI registry clients,
// Fulcio/Rekor TLS stacks, OIDC flows) that would bloat every consumer binary.
//
// Instead we implement the narrow subset that covers the common offline case:
// a detached ed25519 signature over the raw asset bytes, stored as a <asset>.sig
// release attachment and signed with `cosign sign-blob --key <ed25519.key>`.
// The corresponding public key (PEM or bare base64) is embedded in UpdateConfig.
//
// If CosignPublicKey is empty the verification step is skipped entirely, keeping
// the feature strictly opt-in and zero-cost for existing consumers.
package blit

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"strings"
)

// MatchSigAsset finds the detached signature asset for the given asset name.
// GoReleaser cosign integration appends ".sig" to each asset name.
func MatchSigAsset(assets []ReleaseAsset, assetName string) (ReleaseAsset, error) {
	want := assetName + ".sig"
	for _, a := range assets {
		if strings.EqualFold(a.Name, want) {
			return a, nil
		}
	}
	return ReleaseAsset{}, fmt.Errorf("no .sig asset found for %s", assetName)
}

// parseEd25519PublicKey decodes a raw ed25519 public key from either:
//   - PEM block with type "PUBLIC KEY" containing a bare 32-byte ed25519 key
//   - bare base64-encoded 32-byte key (standard or URL-safe encoding)
func parseEd25519PublicKey(keyData string) (ed25519.PublicKey, error) {
	keyData = strings.TrimSpace(keyData)

	// Try PEM first.
	if strings.HasPrefix(keyData, "-----") {
		block, _ := pem.Decode([]byte(keyData))
		if block == nil {
			return nil, fmt.Errorf("failed to decode PEM block")
		}
		// Accept a bare 32-byte DER body (just the raw key, no ASN.1 wrapper)
		// as well as a 32-byte body — cosign sign-blob --key ed25519.pub emits this.
		raw := block.Bytes
		// Strip a minimal SubjectPublicKeyInfo prefix if present (44 bytes total:
		// 12-byte ASN.1 header + 32-byte key). We recognise it by length and the
		// standard ed25519 OID prefix bytes.
		const spkiLen = 44
		ed25519OIDPrefix := []byte{0x30, 0x2a, 0x30, 0x05, 0x06, 0x03, 0x2b, 0x65, 0x70, 0x03, 0x21, 0x00}
		if len(raw) == spkiLen && len(raw) > len(ed25519OIDPrefix) {
			prefixMatch := true
			for i, b := range ed25519OIDPrefix {
				if raw[i] != b {
					prefixMatch = false
					break
				}
			}
			if prefixMatch {
				raw = raw[len(ed25519OIDPrefix):]
			}
		}
		if len(raw) != ed25519.PublicKeySize {
			return nil, fmt.Errorf("PEM key body is %d bytes, want %d", len(raw), ed25519.PublicKeySize)
		}
		return ed25519.PublicKey(raw), nil
	}

	// Try base64 (standard then URL-safe).
	decoded, err := base64.StdEncoding.DecodeString(keyData)
	if err != nil {
		decoded, err = base64.URLEncoding.DecodeString(keyData)
		if err != nil {
			decoded, err = base64.RawStdEncoding.DecodeString(keyData)
			if err != nil {
				decoded, err = base64.RawURLEncoding.DecodeString(keyData)
				if err != nil {
					return nil, fmt.Errorf("key is not valid PEM or base64: %w", err)
				}
			}
		}
	}
	if len(decoded) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("base64 key is %d bytes, want %d", len(decoded), ed25519.PublicKeySize)
	}
	return ed25519.PublicKey(decoded), nil
}

// VerifyCosignSignature verifies a detached ed25519 signature over assetData.
//
// sigData is the raw content of the <asset>.sig file as produced by
// `cosign sign-blob --key <ed25519.key>` — a base64-encoded signature.
// pubKeyPEM is the signer's public key in PEM or bare base64 format.
func VerifyCosignSignature(assetData []byte, sigData []byte, pubKeyPEM string) error {
	pub, err := parseEd25519PublicKey(pubKeyPEM)
	if err != nil {
		return fmt.Errorf("parsing cosign public key: %w", err)
	}

	// The .sig file produced by cosign is a base64-encoded raw signature.
	sigBytes, err := base64.StdEncoding.DecodeString(strings.TrimSpace(string(sigData)))
	if err != nil {
		// Try URL-safe encoding as a fallback.
		sigBytes, err = base64.URLEncoding.DecodeString(strings.TrimSpace(string(sigData)))
		if err != nil {
			return fmt.Errorf("decoding signature: %w", err)
		}
	}

	if !ed25519.Verify(pub, assetData, sigBytes) {
		return fmt.Errorf("cosign signature verification failed: signature does not match")
	}
	return nil
}
