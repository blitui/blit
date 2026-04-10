# Contributing to blit

## Prerequisites

- Go 1.24+ (see `go.mod` for exact version)
- Git

## Build and test

```bash
go build ./...
go test ./...
```

For TUI component tests, blit uses a virtual terminal framework (`blit/`) — no real terminal needed. Run the blit CLI for a nicer test runner experience:

```bash
go run ./cmd/blit ./...
```

## Commits

Use [conventional commits](https://www.conventionalcommits.org/):

```
feat: add frobnicate support
fix: correct off-by-one in table scroll
refactor: simplify key dispatch logic
docs: update install instructions
chore: bump dependencies
test: add coverage for delta updates
```

Keep the subject line under 72 characters. One logical change per commit.

## Branches

- `feat/<short-description>` for new features
- `fix/<short-description>` for bug fixes
- `chore/<short-description>` for maintenance

Don't commit directly to `main` — open a PR.

## Code style

- `gofmt` is law
- Wrap errors with context: `fmt.Errorf("fetch release: %w", err)`
- Godoc comments on all exported types and functions
- No `panic` in library code
- Table-driven tests, standard `testing` package, no testify

## Examples

See `examples/` for runnable demos. If your change adds a new component or significantly changes behavior, consider adding or updating an example.

## Documentation

The docs site lives in `site/` and is built with MkDocs Material. If your change affects user-facing behavior, update the relevant page in `site/docs/`.
