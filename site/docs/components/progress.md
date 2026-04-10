# Progress (UpdateProgress)

A download/operation progress bar component used by the self-update system. Can also be embedded in custom update flows.

## Construction

```go
p := blit.NewUpdateProgress(binary, version string, total int64)
```

- `binary` — name shown in the bar label
- `version` — target version string
- `total` — expected byte count; pass `0` for indeterminate mode

## Driving Progress

Send `UpdateProgressMsg` updates into the model:

```go
p.Update(blit.UpdateProgressMsg{Downloaded: bytesReceived})
// When complete:
p.Update(blit.UpdateProgressMsg{Done: true})
// On error:
p.Update(blit.UpdateProgressMsg{Err: err})
```

## Rendering

```go
fmt.Println(p.View())
```

The bar renders a filled progress track with a shimmer highlight sweep animation (driven by `animTickMsg`). In indeterminate mode (total == 0) it shows a spinner-style sweep.

## Fields

| Field | Type | Description |
|-------|------|-------------|
| `Binary` | `string` | Binary name label |
| `Version` | `string` | Target version |
| `Total` | `int64` | Expected bytes (0 = indeterminate) |
| `Downloaded` | `int64` | Bytes received so far |
| `Width` | `int` | Bar width in columns (default 40) |
| `Done` | `bool` | Set to true when complete |
| `Err` | `error` | Set on failure |
| `StartedAt` | `time.Time` | Used to compute elapsed time |

## CLI Progress (cli package)

For interactive CLI workflows outside a full TUI, use the `cli` package's progress bar:

```go
import "github.com/blitui/blit/cli"

bar := cli.NewProgress(100, "Downloading")
bar.Increment(25)
bar.Increment(75)
bar.Done()
```

See [CLI Primitives](../guides/cli-primitives.md) for the full API.
