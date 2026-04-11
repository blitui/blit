# Security Policy

## Reporting a Vulnerability

If you discover a security vulnerability in blit, please report it responsibly.

**Do not open a public issue.**

Instead, use [GitHub's private vulnerability reporting](https://github.com/blitui/blit/security/advisories/new).

We will acknowledge receipt within 48 hours and aim to release a fix within 7 days for critical issues.

## Supported Versions

| Version | Supported |
|---------|-----------|
| 0.2.x   | Yes       |
| < 0.2   | No        |

## Incident Response

### Linter failures in CI

1. Check `golangci-lint run ./...` locally to reproduce
2. Fix the reported issue (do not disable the linter)
3. Run `make check` to verify before pushing

### Broken build or test failures

1. Run `make check` locally (fmt + vet + lint + test)
2. Check the CI job logs for the specific failure
3. If race-related, run `go test -race ./...` locally

### Secret leak detected by gitleaks

1. Rotate the leaked credential immediately
2. Use `git filter-repo` or BFG to remove the secret from history
3. Verify gitleaks no longer flags the commit
