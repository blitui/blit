# Viewport

Scrollable content pane with keyboard and mouse navigation and a scrollbar indicator. Implements `Component` and `Themed`.

## Construction

```go
func NewViewport() *Viewport
```

## Usage

```go
vp := blit.NewViewport()
vp.SetContent(longTextContent)
```

Set content at any time — the scroll position resets to the top.

## Scrollbar

Viewport renders a scrollbar on the right edge using track and thumb glyphs from the theme. The thumb size and position reflect the visible portion of the content.

## Keybindings

| Key | Action |
|-----|--------|
| `up` / `k` | Scroll up 1 line |
| `down` / `j` | Scroll down 1 line |
| `pgup` | Page up |
| `pgdn` | Page down |
| `home` | Scroll to top |
| `end` | Scroll to bottom |
| `ctrl+u` | Half page up |
| `ctrl+d` | Half page down |
| Mouse wheel | Scroll ±3 lines |

## State

| Method | Description |
|--------|-------------|
| `SetContent(content)` | Set text content, reset scroll |
| `ScrollBy(delta)` | Scroll by lines (positive = down) |
| `GotoTop()` | Scroll to top |
| `GotoBottom()` | Scroll to bottom |
| `YOffset()` | Current scroll offset |
| `AtTop()` | Whether scrolled to top |
| `AtBottom()` | Whether scrolled to bottom |
