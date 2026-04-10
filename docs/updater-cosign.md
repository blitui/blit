# Cosign Signature Verification

blit's self-updater supports optional detached ed25519 signature verification
via `UpdateConfig.CosignPublicKey`. When set, `SelfUpdate` fetches the
`<asset>.sig` release attachment and verifies it before replacing the binary.

## Why ed25519 instead of the full cosign library?

The `github.com/sigstore/cosign` library pulls in OCI registry clients, Fulcio/Rekor
TLS stacks, and OIDC flows — hundreds of transitive dependencies that would bloat every
consumer binary. The common offline case (a maintainer signs release assets with a
static key) only requires verifying a detached signature against a known public key,
which Go's stdlib `crypto/ed25519` handles in ~50 lines with zero new dependencies.

## Setup

### 1. Generate a keypair

```sh
cosign generate-key-pair
# produces cosign.key (private) and cosign.pub (public)
```

Or with OpenSSL:

```sh
openssl genpkey -algorithm ed25519 -out cosign.key
openssl pkey -in cosign.key -pubout -out cosign.pub
```

### 2. Sign release assets in CI

With cosign:

```sh
cosign sign-blob --key cosign.key myapp_1.0.0_linux_amd64.tar.gz \
  --output-signature myapp_1.0.0_linux_amd64.tar.gz.sig
```

Upload each `<asset>.sig` file alongside the asset in your GitHub release.
GoReleaser can automate this via the `signs:` block in `.goreleaser.yaml`.

### 3. Embed the public key in your binary

```go
//go:embed cosign.pub
var cosignPub string

blit.WithAutoUpdate(blit.UpdateConfig{
    Owner:           "myorg",
    Repo:            "myapp",
    BinaryName:      "myapp",
    Version:         version,
    CosignPublicKey: cosignPub, // PEM or bare base64
})
```

## Behavior

| CosignPublicKey | .sig asset present | Result |
|---|---|---|
| empty | any | verification skipped |
| set | present, valid | update proceeds |
| set | present, invalid | error — binary not replaced |
| set | missing | error — binary not replaced |

Verification happens after the SHA256 checksum check and before `replaceBinary`,
so a failed cosign check leaves the running binary untouched (no rollback needed).

## Key formats accepted

- PEM block with type `PUBLIC KEY` (SubjectPublicKeyInfo, as emitted by OpenSSL/cosign)
- Raw base64-encoded 32-byte ed25519 public key (standard or URL-safe encoding)
