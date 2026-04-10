# Cookbook

Five complete recipes for common blit patterns.

---

## 1. Dashboard in 5 Minutes

A full dashboard with a table, dual-pane sidebar, and status bar:

```go
package main

import (
    "fmt"
    "time"

    "github.com/charmbracelet/lipgloss"
    blit "github.com/blitui/blit"
)

func main() {
    table := blit.NewTable(
        []blit.Column{
            {Title: "Service", Width: 20, Sortable: true},
            {Title: "Status",  Width: 12},
            {Title: "Latency", Width: 10, Align: blit.Right, Sortable: true},
        },
        []blit.Row{
            {"api-gateway",  "online",  "12ms"},
            {"auth-service", "online",  "8ms"},
            {"db-primary",   "online",  "2ms"},
            {"cache",        "degraded","45ms"},
        },
        blit.TableOpts{
            Sortable:   true,
            Filterable: true,
            CellRenderer: func(row blit.Row, col int, isCursor bool, th blit.Theme) string {
                if col == 1 {
                    color := th.Positive
                    if row[1] != "online" { color = th.Flash }
                    return lipgloss.NewStyle().Foreground(lipgloss.Color(color)).Render(row[1])
                }
                return row[col]
            },
        },
    )

    detail := blit.NewListView[string](blit.ListViewOpts[string]{
        RenderItem: func(s string, _ int, _ bool, th blit.Theme) string { return s },
    })

    table.Opts().OnCursorChange = func(row blit.Row, _ int) {
        detail.SetItems([]string{
            "Service: " + row[0],
            "Status:  " + row[1],
            "Latency: " + row[2],
        })
    }

    app := blit.NewApp(
        blit.WithTheme(blit.DefaultTheme()),
        blit.WithLayout(&blit.DualPane{
            Main:         table,
            Side:         detail,
            SideWidth:    28,
            MinMainWidth: 60,
            SideRight:    true,
            ToggleKey:    "p",
        }),
        blit.WithStatusBar(
            func() string { return " ? help  s sort  / search  p panel  q quit" },
            func() string { return fmt.Sprintf(" %d services", table.RowCount()) },
        ),
        blit.WithHelp(),
        blit.WithTickInterval(5*time.Second),
    )
    app.Run()
}
```

---

## 2. Self-Updating CLI

Wire binary self-update into an existing app with three lines:

```go
app := blit.NewApp(
    blit.WithTheme(blit.DefaultTheme()),
    blit.WithComponent("main", myMainComponent),
    blit.WithAutoUpdate(blit.UpdateConfig{
        Owner:      "myorg",
        Repo:       "mytool",
        BinaryName: "mytool",
        Version:    version, // injected via: -ldflags "-X main.version=v1.2.3"
        Mode:       blit.UpdateNotify,
        CacheTTL:   24 * time.Hour,
    }),
)
```

Add cleanup at the top of `main()` for the `.old` backup left by a previous update:

```go
func main() {
    blit.CleanupOldBinary()
    // ...
}
```

---

## 3. Testing a TUI

Use `blit` to assert on rendered screen content without a real terminal:

```go
func TestTable(t *testing.T) {
    table := blit.NewTable(
        []blit.Column{{Title: "Name", Width: 20}},
        []blit.Row{{"Alice"}, {"Bob"}},
        blit.TableOpts{Filterable: true},
    )

    tm := blit.NewTestModel(t, wrapInApp(table), 80, 24)

    // Header visible
    blit.AssertRowContains(t, tm.Screen(), 0, "Name")

    // Navigate down
    tm.SendKey("j")
    blit.AssertContains(t, tm.Screen(), "Bob")

    // Filter
    tm.SendKey("/")
    tm.Type("ali")
    blit.AssertContains(t, tm.Screen(), "Alice")
    blit.AssertNotContains(t, tm.Screen(), "Bob")

    // Golden snapshot
    blit.AssertGolden(t, tm.Screen(), "table-filtered")
}
```

Regenerate golden files:

```bash
blit -update ./...
```

---

## 4. Importing a Theme

Load a theme from a TOML/YAML config file at startup:

```go
// config.toml:
// [theme]
// accent  = "#7aa2f7"
// positive = "#9ece6a"

cfg := loadConfig("config.toml")
theme := blit.ThemeFromMap(cfg.Theme)

app := blit.NewApp(
    blit.WithTheme(theme),
    ...
)
```

For live hot-reload during development:

```bash
BLIT_THEME=./mytheme.json go run .
```

Edit `mytheme.json` and the app reloads the theme without restart.

---

## 5. SSH-Served TUI

Serve your blit app over SSH using [Wish](https://github.com/charmbracelet/wish) so remote users can access it without installing anything:

```go
package main

import (
    "context"
    "os"
    "os/signal"
    "syscall"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/ssh"
    "github.com/charmbracelet/wish"
    "github.com/charmbracelet/wish/bubbletea"
    blit "github.com/blitui/blit"
)

func main() {
    s, _ := wish.NewServer(
        wish.WithAddress(":2222"),
        wish.WithHostKeyPath(".ssh/id_ed25519"),
        wish.WithMiddleware(
            bubbletea.Middleware(func(s ssh.Session) (tea.Model, []tea.ProgramOption) {
                table := blit.NewTable(columns, rows, blit.TableOpts{Sortable: true})
                app := blit.NewApp(
                    blit.WithTheme(blit.DefaultTheme()),
                    blit.WithComponent("main", table),
                )
                return app.Model(), []tea.ProgramOption{tea.WithAltScreen()}
            }),
        ),
    )

    done := make(chan os.Signal, 1)
    signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
    go s.ListenAndServe()
    <-done
    s.Shutdown(context.Background())
}
```

Connect:

```bash
ssh localhost -p 2222
```
