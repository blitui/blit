# tuikit-go

The pragmatic TUI toolkit for shipping CLI tools fast. Wraps [Bubble Tea](https://github.com/charmbracelet/bubbletea) + [Lip Gloss](https://github.com/charmbracelet/lipgloss) with reusable components, a layout engine, a keybinding registry, a theme system, and built-in binary self-update. Build a complete TUI app in under 20 lines.

![tuikit-go quick start](https://raw.githubusercontent.com/moneycaringcoder/tuikit-go/main/docs/gifs/quickstart.gif)

## Features

- Table with sorting, filtering, custom cell rendering, and mouse support
- ListView, StatusBar, Help screen, ConfigEditor, CommandBar, DetailOverlay, CollapsibleSection
- Dual-pane layout engine with collapsible sidebar
- Keybinding registry with auto-generated help screen
- Dark and light themes with semantic color tokens + extensible `Extra` color map
- Poller for background data with automatic tick-driven refresh
- **CLI primitives** — Confirm, SelectOne, MultiSelect, Input, Password, Spinner, Progress, and styled message helpers for non-TUI workflows
- **tuitest** — virtual-terminal testing framework with 30+ assertions, screen diffing, golden files, snapshot update, session record/replay, JUnit + HTML reporters, and a vitest-style console runner
- Utilities: Sparkline, RelativeTime, OpenURL, Hyperlink
- **Self-update built in** — binary replacement with SHA256 checksum verification, skip/forced/notify modes, rollback on verify failure, rate-limit backoff, update channels, and Homebrew/Scoop detection
- **tuitest CLI** — `go install` or grab a prebuilt binary to run test suites with snapshot update, JUnit/HTML reports, filtering, parallelism, and watch mode

## Install

![install and get started](https://raw.githubusercontent.com/moneycaringcoder/tuikit-go/main/docs/gifs/tuitest-runner.gif)

Library:

```bash
go get github.com/moneycaringcoder/tuikit-go
```

`tuitest` CLI (optional, for running tuitest-based test suites):

```bash
# Homebrew
brew install moneycaringcoder/tap/tuitest

# Scoop
scoop bucket add moneycaringcoder https://github.com/moneycaringcoder/scoop-bucket
scoop install tuitest

# Go
go install github.com/moneycaringcoder/tuikit-go/cmd/tuitest@latest
```

## Quick Start

![quick start demo](https://raw.githubusercontent.com/moneycaringcoder/tuikit-go/main/docs/gifs/quickstart.gif)

```go
package main

import (
    "fmt"
    tuikit "github.com/moneycaringcoder/tuikit-go"
)

func main() {
    table := tuikit.NewTable(
        []tuikit.Column{
            {Title: "Name", Width: 20, Sortable: true},
            {Title: "Status", Width: 15},
        },
        []tuikit.Row{
            {"Alice", "Online"},
            {"Bob", "Away"},
        },
        tuikit.TableOpts{Sortable: true, Filterable: true},
    )

    app := tuikit.NewApp(
        tuikit.WithTheme(tuikit.DefaultTheme()),
        tuikit.WithComponent("main", table),
        tuikit.WithStatusBar(
            func() string { return " ? help  q quit" },
            func() string { return fmt.Sprintf(" %d items", 2) },
        ),
        tuikit.WithHelp(),
    )

    app.Run()
}
```

See the examples for complete apps showing components together:

```bash
go run ./examples/minimal/     # Simple ListView in ~30 lines
go run ./examples/dashboard/   # Full Table + DualPane + ConfigEditor
go run ./examples/monitor/     # Service fleet dashboard with all components
go run ./examples/cli-demo/    # Interactive CLI primitives showcase
```

## Components

![components overview](https://raw.githubusercontent.com/moneycaringcoder/tuikit-go/main/docs/gifs/table.gif)

### Table

Adaptive table with responsive columns, sorting, filtering, and cursor navigation.

```go
columns := []tuikit.Column{
    {Title: "Name", Width: 20, Sortable: true},
    {Title: "Score", Width: 10, Align: tuikit.Right, Sortable: true},
    {Title: "Extra", Width: 15, MinWidth: 100}, // hides below 100 cols
}

table := tuikit.NewTable(columns, rows, tuikit.TableOpts{
    Sortable:   true, // 's' to cycle sort
    Filterable: true, // '/' to search
})

table.SetRows(newRows) // update data dynamically
```

**Custom cell rendering** — full control over per-cell styling:

```go
tuikit.TableOpts{
    CellRenderer: func(row tuikit.Row, colIdx int, isCursor bool, theme tuikit.Theme) string {
        val := row[colIdx]
        if colIdx == 2 && val == "Online" {
            return lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Positive)).Render(val)
        }
        return val
    },
}
```

**Custom sort** — numeric, time-based, or any logic:

```go
tuikit.TableOpts{
    SortFunc: func(a, b tuikit.Row, sortCol int, sortAsc bool) bool {
        va, _ := strconv.ParseFloat(a[sortCol], 64)
        vb, _ := strconv.ParseFloat(b[sortCol], 64)
        if sortAsc { return va < vb }
        return va > vb
    },
}
```

**Predicate filter** — programmatic row filtering alongside text search:

```go
table.SetFilter(func(row tuikit.Row) bool {
    return row[1] == "online"
})
table.SetFilter(nil) // clear
```

Mouse scroll and click are handled automatically when `tuikit.WithMouseSupport()` is set.

### StatusBar

Footer with left-aligned hints and right-aligned status.

```go
tuikit.WithStatusBar(
    func() string { return " ? help  q quit" },
    func() string { return " 42 items" },
)
```

### Help

Auto-generated from all registered keybindings. Zero configuration — press `?` to toggle.

```go
tuikit.WithHelp()
```

### ConfigEditor

Declarative settings overlay with grouped fields and validation.

```go
editor := tuikit.NewConfigEditor([]tuikit.ConfigField{
    {
        Label: "Refresh Interval",
        Group: "General",
        Hint:  "seconds, min 5",
        Get:   func() string { return fmt.Sprint(cfg.Interval) },
        Set: func(v string) error {
            n, _ := strconv.Atoi(v)
            if n < 5 { return fmt.Errorf("must be >= 5") }
            cfg.Interval = n
            return nil
        },
    },
})

tuikit.WithOverlay("Settings", "c", editor) // press 'c' to open
```

### CommandBar

Inline command input with completion and dispatch.

### DetailOverlay

Full-screen detail view for a selected row or item.

### CollapsibleSection

Expandable section for grouping content in a panel.

### Poller

Background data polling with tick-driven refresh.

### App-Level Keybindings

```go
tuikit.WithKeyBind(tuikit.KeyBind{
    Key:   "f",
    Label: "Cycle filter",
    Group: "DATA",
    Handler: func() {
        filterIdx = (filterIdx + 1) % len(modes)
        table.SetRows(rows)
    },
})
```

Registered keybindings appear in the help screen automatically.

### Tick / Timer

```go
tuikit.WithTickInterval(100 * time.Millisecond)
```

Components receive `tuikit.TickMsg` in their `Update` method.

### External Data

Push data into the app from WebSocket streams, API polling, or any goroutine:

```go
app := tuikit.NewApp(...)
go func() {
    for data := range stream {
        app.Send(MyDataMsg{data})
    }
}()
app.Run()
```

Unknown message types are forwarded to all components via `Update`.

## CLI Primitives

![CLI primitives showcase](https://raw.githubusercontent.com/moneycaringcoder/tuikit-go/main/docs/gifs/cli-primitives.gif)

The `cli` package provides interactive prompts for tools that need more than `fmt.Print` but less than a full TUI. Each primitive runs a minimal Bubble Tea program, captures input, and returns the result.

```go
import "github.com/moneycaringcoder/tuikit-go/cli"

// Yes/No
proceed := cli.Confirm("Deploy to production?", false)

// Single select (type-to-filter on 10+ items)
lang, idx, err := cli.SelectOne("Language:", []string{"Go", "Rust", "Python"})

// Multi select with checkboxes
selected, indices, err := cli.MultiSelect("Features:", []string{"Auth", "DB", "Cache"})

// Text input with validation
name, err := cli.Input("Project name:", func(s string) error {
    if s == "" { return fmt.Errorf("required") }
    return nil
})

// Masked password input
secret, err := cli.Password("API token:", nil)

// Spinner (runs in background goroutine)
s := cli.Spin("Installing...")
// ... do work ...
s.Stop()

// Progress bar
bar := cli.NewProgress(100, "Downloading")
bar.Increment(25)
bar.Done()
```

**Styled message helpers** for consistent CLI output:

```go
cli.Title("Setup Wizard")          // Bold underlined title
cli.Step(1, 3, "Installing deps")  // [1/3] numbered step
cli.Success("Build complete")      // ✓ green
cli.Warning("Deprecated flag")     // ! yellow
cli.Error("Connection failed")     // ✗ red
cli.Info("Using defaults")         // ℹ blue
cli.Separator()                    // ────────────
cli.KeyValue("Version", "1.2.3")   // dimmed key: value
```

## Testing with tuitest

![tuitest vitest-style reporter](https://raw.githubusercontent.com/moneycaringcoder/tuikit-go/main/docs/gifs/tuitest-runner.gif)

The `tuitest` package provides a virtual terminal and assertion helpers for testing Bubble Tea models without launching a real terminal.

```go
import "github.com/moneycaringcoder/tuikit-go/tuitest"

func TestMyApp(t *testing.T) {
    tm := tuitest.NewTestModel(t, myModel{}, 80, 24)

    // Interact
    tm.SendKey("down")
    tm.SendKeys("j", "j", "enter")
    tm.Type("hello")
    tm.SendResize(120, 40)
    tm.SendMsg(myCustomMsg{})

    // Assert on rendered screen
    scr := tm.Screen()
    tuitest.AssertContains(t, scr, "Expected text")
    tuitest.AssertRowContains(t, scr, 0, "Header")
    tuitest.AssertFgAt(t, scr, 2, 0, "red")
    tuitest.AssertBoldAt(t, scr, 0, 0)
    tuitest.AssertMatches(t, scr, `\d+ items`)
    tuitest.AssertRowCount(t, scr, 5)

    // Region-scoped assertions
    tuitest.AssertRegionContains(t, scr, 0, 40, 20, 10, "sidebar text")

    // Screen comparison
    before := tm.Screen()
    tm.SendKey("enter")
    after := tm.Screen()
    tuitest.AssertScreensNotEqual(t, before, after)

    // Golden file testing
    tuitest.AssertGolden(t, scr, "my-test") // compares against testdata/my-test.golden

    // Wait for async state
    ok := tm.WaitFor(tuitest.UntilContains("loaded"), 10)
}
```

### tuitest CLI

The `tuitest` binary is a thin wrapper around `go test` that adds flags for snapshot updates, JUnit/HTML reports, filtering, parallelism, and watch mode.

**Install:**

```bash
# Homebrew
brew install moneycaringcoder/tap/tuitest

# Scoop
scoop bucket add moneycaringcoder https://github.com/moneycaringcoder/scoop-bucket
scoop install tuitest

# Go
go install github.com/moneycaringcoder/tuikit-go/cmd/tuitest@latest
```

Prebuilt linux/darwin/windows (amd64 + arm64) archives are attached to every [GitHub Release](https://github.com/moneycaringcoder/tuikit-go/releases).

**Usage:**

```bash
tuitest                                    # go test ./...
tuitest -filter TestHarness ./tuitest/...  # run tests matching a regexp
tuitest -update ./tuitest/...              # regenerate golden snapshots
tuitest -junit out/junit.xml -parallel 4   # parallel run + JUnit report
tuitest -html out/report.html              # HTML report
tuitest -watch                             # re-run on file changes (1s poll)
```

The `-update` flag drives the same `-tuitest.update` hook that `AssertGolden` and `AssertScreenSnapshot` use. JUnit and HTML reporters are opt-in via the `JUnitReporter` / `HTMLReporter` wiring in `tuitest`.

**Vitest-like test reporter** — run with `-v` for grouped, color-coded output:

```
  tuitest · terminal test toolkit

  Screen
    ✓ PlainText 0.000ms
    ✓ Contains 0.000ms
  Assert
    ✓ ContainsPass 0.000ms
    ✓ RowMatchesPass 0.000ms

  PASS 96 tests (3ms)
```

## Utilities

### Sparkline

Unicode block sparkline from a `[]float64` slice. Bars are colored by direction (up/down/neutral) or rendered mono.

```go
line, width := tuikit.Sparkline(prices, 40, &tuikit.SparklineOpts{Mono: false})
```

### RelativeTime

Short human-readable duration string.

```go
tuikit.RelativeTime(event.Time, time.Now()) // "3m ago", "2h ago", "5d ago"
```

### OpenURL

Opens a URL in the user's default browser. Runs asynchronously, does not block.

```go
tuikit.OpenURL("https://github.com/moneycaringcoder/tuikit-go")
```

### Hyperlink

Wraps text in a clickable terminal hyperlink (OSC8). Supported in modern terminals.

```go
tuikit.Hyperlink("https://example.com", "click here")
```

## Layout & Theming

![theme gallery — 8 presets](https://raw.githubusercontent.com/moneycaringcoder/tuikit-go/main/docs/gifs/theme-gallery.gif)

### Layout

Single pane or dual pane with collapsible sidebar.

```go
tuikit.WithLayout(&tuikit.DualPane{
    Main:         table,
    Side:         panel,
    SideWidth:    30,
    MinMainWidth: 60, // sidebar auto-hides below this
    SideRight:    true,
    ToggleKey:    "p",
})
```

### Theming

Built-in dark and light themes, or create your own from a color map.

```go
tuikit.DefaultTheme()
tuikit.LightTheme()

// From parsed config (YAML/JSON/TOML)
tuikit.ThemeFromMap(map[string]string{
    "positive": "#00ff00",
    "negative": "#ff0000",
    "accent":   "#0000ff",
})
```

Semantic tokens: `Positive`, `Negative`, `Accent`, `Muted`, `Text`, `TextInverse`, `Cursor`, `Border`, `Flash`.

**App-specific colors** — extend the theme with domain-specific tokens via the `Extra` map:

```go
theme := tuikit.ThemeFromMap(map[string]string{
    "positive": "#22c55e",
    "push":     "#22c55e",  // unknown keys go into Extra
    "pr":       "#3b82f6",
    "review":   "#a855f7",
})

// Look up with fallback
color := theme.Color("push", theme.Positive) // returns #22c55e
color = theme.Color("missing", theme.Muted)  // returns Muted fallback
```

Components receive theme updates automatically if they implement `Themed`:

```go
type Themed interface {
    SetTheme(tuikit.Theme)
}
```

## Self-Update

![self-update progress flow](https://raw.githubusercontent.com/moneycaringcoder/tuikit-go/main/docs/gifs/update-flow.gif)

tuikit-go ships a binary self-update system designed for GoReleaser-published CLIs. It checks GitHub Releases, verifies SHA256 checksums against GoReleaser's `checksums.txt`, replaces the running binary atomically, rolls back on verify failure, and detects Homebrew/Scoop installs so package-managed binaries are left alone.

**Modes:** `UpdateNotify` (non-blocking banner), `UpdateBlocking` (prompt before TUI starts), `UpdateForced` (full-screen gate for mandatory upgrades), `UpdateSilent` (check + cache, no UI), `UpdateDryRun` (verify without replacing).

**Extras:** skip-version support, `minimum_version:` markers in release notes auto-promote to forced, rate-limit backoff, update channels (stable/beta/nightly), and `TUIKIT_UPDATE_DISABLE=1` kill switch.

### Built-in app integration

Add one option to your `NewApp` call:

```go
app := tuikit.NewApp(
    tuikit.WithAutoUpdate(tuikit.UpdateConfig{
        Owner:      "myorg",
        Repo:       "mytool",
        BinaryName: "mytool",
        Version:    version, // set via ldflags: -X main.version=v1.2.3
        Mode:       tuikit.UpdateNotify,   // or UpdateBlocking
        CacheTTL:   24 * time.Hour,
    }),
)
```

- `UpdateNotify` — shows a non-blocking banner inside the TUI after startup.
- `UpdateBlocking` — prompts in stdout before the TUI starts.
- Dev builds (`version == ""` or `"dev"`) are skipped automatically.
- Results are cached to avoid hitting the GitHub API on every launch.

### Manual update command

Call `SelfUpdate` directly to implement an explicit `--update` flag or menu action:

```go
if err := tuikit.SelfUpdate(cfg); err != nil {
    fmt.Fprintln(os.Stderr, "update failed:", err)
    os.Exit(1)
}
```

`SelfUpdate` downloads the matching release asset, verifies its checksum, extracts the binary, and replaces the running executable. Add `CleanupOldBinary()` near the top of `main()` to remove the `.old` backup left by a previous update.

### Install method detection

```go
method := tuikit.DetectInstallMethod(os.Args[0])
// InstallManual, InstallHomebrew, or InstallScoop
```

Use this to skip `SelfUpdate` when the binary is managed by a package manager.

## Predictable API

tuikit follows consistent patterns throughout:

- All components implement `tuikit.Component` — `Init`, `Update`, `View`, `KeyBindings`, `SetSize`, `Focused`, `SetFocused`
- App is configured via functional options: `WithTheme`, `WithComponent`, `WithLayout`, `WithAutoUpdate`, etc.
- Key dispatch order: overlay stack → built-in globals (`q` / `tab` / `?`) → pane toggle → overlay triggers → app keybindings → focused component
- Return `tuikit.Consumed()` from a component's `Update` to stop further dispatch
- Modal overlays implement the `Overlay` interface and register with a trigger key via `WithOverlay`
- See `examples/dashboard/main.go` for a complete reference

## Used By

- [cryptstream-tui](https://github.com/moneycaringcoder/cryptstream-tui) — Live cryptocurrency ticker
- [gitstream-tui](https://github.com/moneycaringcoder/gitstream-tui) — GitHub activity dashboard

## Dependencies

Charm ecosystem only:

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) — TUI framework
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) — Styling
- [Bubbles](https://github.com/charmbracelet/bubbles) — Component primitives

## License

MIT
