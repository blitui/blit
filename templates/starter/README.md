# myapp

A minimal [tuikit-go](https://github.com/moneycaringcoder/tuikit-go) starter application.

## Getting started

1. Clone or copy this directory.
2. Replace every `OWNER` placeholder with your GitHub username.
3. Replace the module path in `go.mod`:

   ```
   module github.com/OWNER/myapp
   ```

4. Run the app:

   ```bash
   go run ./cmd/myapp
   ```

## Structure

```
cmd/myapp/
  main.go        — entry point, app wiring
  app_test.go    — tuitest session tests
go.mod
.goreleaser.yaml
.github/workflows/
  ci.yml         — build + test on every push / PR
  release.yml    — GoReleaser on v* tags
```

## Tests

```bash
go test ./...
```

Tests use [tuitest](https://github.com/moneycaringcoder/tuikit-go/tree/main/tuitest), the built-in TUI testing harness. See `cmd/myapp/app_test.go` for example assertions.

## Release

Tag a commit to trigger the release workflow:

```bash
git tag v0.1.0
git push origin v0.1.0
```

GoReleaser will build cross-platform binaries and publish them to GitHub Releases, Homebrew, and Scoop (configure tap/bucket repos first).

## License

MIT
