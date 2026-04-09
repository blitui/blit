# VHS tape scripts

This directory contains [VHS](https://github.com/charmbracelet/vhs) tape scripts that generate the demo GIFs embedded in the project README.

## Size budget

Each generated GIF must be **< 1.5 MB**. Settings that control file size:

| VHS setting | Effect |
|---|---|
| `Framerate` | Lower = fewer frames = smaller file. 12 fps is a good default for demos. |
| `PlaybackSpeed` | Higher = shorter recording = smaller file. |
| `Width` / `Height` | Smaller terminal = fewer pixels = smaller file. |
| `Set Theme` | Minimal color variance helps GIF palette compression. |

If a GIF exceeds the budget, reduce `Framerate` first (try 10), then `PlaybackSpeed` (try 1.5), then terminal dimensions.

## Generating GIFs

VHS must be installed:

```bash
# macOS / Linux
brew install vhs

# Go toolchain
go install github.com/charmbracelet/vhs@latest

# Windows (Scoop)
scoop install vhs
```

Then from the repo root:

```bash
# Render all tapes
./scripts/gen-gifs.sh

# Render specific tapes
./scripts/gen-gifs.sh table form picker
```

GIFs are written to `docs/gifs/` which is git-ignored (they are regenerated from the tape scripts on every release).

## Tape inventory

| Tape | Component / feature | Output |
|---|---|---|
| `quickstart.tape` | Quick Start — minimal ListView | `docs/gifs/quickstart.gif` |
| `table.tape` | Table — sort, filter, detail overlay | `docs/gifs/table.gif` |
| `form.tape` | Form — signup flow, all field types | `docs/gifs/form.gif` |
| `tabs.tape` | Tabs — horizontal tabs, nested components | `docs/gifs/tabs.gif` |
| `picker.tape` | Picker — fuzzy file picker with preview | `docs/gifs/picker.gif` |
| `toasts.tape` | Toast manager — stacked notifications | `docs/gifs/toasts.gif` |
| `theme-gallery.tape` | Theme gallery — 8-preset live cycle | `docs/gifs/theme-gallery.gif` |
| `cli-primitives.tape` | CLI primitives — Confirm, Input, Select… | `docs/gifs/cli-primitives.gif` |
| `update-flow.tape` | Self-update progress + shimmer bar | `docs/gifs/update-flow.gif` |
| `tuitest-runner.tape` | tuitest vitest-style reporter output | `docs/gifs/tuitest-runner.gif` |

## sess2tape

The `cmd/sess2tape` tool converts a `.tuisess` session file (produced by `tuitest.SessionRecorder`) into a VHS tape script:

```bash
# Build
go build -o sess2tape ./cmd/sess2tape

# Convert
./sess2tape testdata/sessions/my-flow.tuisess
# Writes testdata/sessions/my-flow.tape

# Pipe
cat my-flow.tuisess | ./sess2tape -out my-flow.tape -
```

This is useful for recording real interaction sessions in tests and then replaying them as GIFs.
