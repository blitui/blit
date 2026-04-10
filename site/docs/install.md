# Install

## Library

```bash
go get github.com/blitui/blit
```

Requires Go 1.24+. No CGO required (`CGO_ENABLED=0` is safe).

## blit CLI

The `blit` binary is a thin wrapper around `go test` that adds snapshot update, JUnit/HTML reports, watch mode, and a vitest-style reporter.

=== "Homebrew"

    ```bash
    brew install blitui/tap/blit
    ```

=== "Scoop"

    ```bash
    scoop bucket add blitui https://github.com/blitui/scoop-bucket
    scoop install blit
    ```

=== "Go install"

    ```bash
    go install github.com/blitui/blit/cmd/blit@latest
    ```

=== "Pre-built binary"

    Download linux/darwin/windows archives (amd64 + arm64) from the
    [GitHub Releases](https://github.com/blitui/blit/releases) page.

## Verify

```bash
blit --version
```

## pkg.go.dev

Full Go API reference is available at [pkg.go.dev/github.com/blitui/blit](https://pkg.go.dev/github.com/blitui/blit).
