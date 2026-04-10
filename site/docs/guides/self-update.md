# Self-Update

blit ships a binary self-update system designed for GoReleaser-published CLIs. It checks GitHub Releases, verifies SHA256 checksums against GoReleaser's `checksums.txt`, replaces the running binary atomically, rolls back on verify failure, and detects Homebrew/Scoop installs so package-managed binaries are left alone.

## Update Modes

| Mode | Behaviour |
|------|-----------|
| `UpdateNotify` | Non-blocking banner inside the TUI after startup |
| `UpdateBlocking` | Prompt in stdout before the TUI starts |
| `UpdateForced` | Full-screen gate for mandatory upgrades |
| `UpdateSilent` | Check + cache, no UI |
| `UpdateDryRun` | Verify without replacing the binary |

## App Integration

Add one option to `NewApp`:

```go
app := blit.NewApp(
    blit.WithAutoUpdate(blit.UpdateConfig{
        Owner:      "myorg",
        Repo:       "mytool",
        BinaryName: "mytool",
        Version:    version, // set via ldflags: -X main.version=v1.2.3
        Mode:       blit.UpdateNotify,
        CacheTTL:   24 * time.Hour,
    }),
)
```

Dev builds (`version == ""` or `"dev"`) are skipped automatically. Results are cached to avoid hitting the GitHub API on every launch.

## Manual Update Command

Call `SelfUpdate` directly to implement an explicit `--update` flag:

```go
if err := blit.SelfUpdate(cfg); err != nil {
    fmt.Fprintln(os.Stderr, "update failed:", err)
    os.Exit(1)
}
```

Add `CleanupOldBinary()` near the top of `main()` to remove the `.old` backup left by a previous update.

## Install Method Detection

```go
method := blit.DetectInstallMethod(os.Args[0])
// Returns: InstallManual, InstallHomebrew, or InstallScoop
```

Use this to skip `SelfUpdate` when the binary is managed by a package manager.

## Extras

- **Skip version** — users can skip a specific release; it won't be offered again
- **Minimum version markers** — add `minimum_version: v1.5.0` to release notes to auto-promote to `UpdateForced`
- **Update channels** — stable / beta / nightly via `Channel` field in `UpdateConfig`
- **Rate-limit backoff** — exponential backoff when GitHub API returns 429
- **Kill switch** — set `TUIKIT_UPDATE_DISABLE=1` to disable all update checks
