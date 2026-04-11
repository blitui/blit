---
name: codebase-health
description: Maintain code quality, run linting, testing, and coverage checks for the blit Go TUI framework
---

# Codebase Health

## Commands
- Build: `go build ./...`
- Test: `go test -race ./...` or `make test`
- Lint: `golangci-lint run ./...` or `make lint`
- Format: `gofmt -w .` or `make fix`
- Full check: `make check` (fmt + vet + lint + test)
- Coverage: `make cover`

## Conventions
- `gofmt` is law
- Conventional commits: `feat:`, `fix:`, `refactor:`, `docs:`, `chore:`, `test:`
- Wrap errors: `fmt.Errorf("context: %w", err)`
- Godoc on all exported types/functions
- No `panic`/`log.Fatal` in library code
- Table-driven tests, standard `testing` package only
