# Theming

## Semantic Color Tokens

`Theme` maps human-intent names to `lipgloss.Color` values. Components reference tokens rather than raw hex codes so the entire UI recolors by swapping one struct.

| Token | Meaning |
|-------|---------|
| `Positive` | Gains, success, online — default `#22c55e` |
| `Negative` | Losses, errors, offline — default `#ef4444` |
| `Accent` | Highlights, active tab, cursor marker |
| `Muted` | Dimmed text, secondary info, timestamps |
| `Text` | Primary content text |
| `TextInverse` | Text on colored backgrounds |
| `Cursor` | Row/item selection highlight background |
| `Border` | Borders, separators, dividers |
| `Flash` | Temporary notification / warning background |

## Built-In Themes

```go
blit.DefaultTheme() // dark — teal accent on dark background
blit.LightTheme()   // light — dark text on white background
```

## Theme From a Color Map

Parse theme colors from YAML/JSON/TOML config without hand-constructing the struct:

```go
theme := blit.ThemeFromMap(map[string]string{
    "positive": "#22c55e",
    "negative": "#ef4444",
    "accent":   "#3b82f6",
    "muted":    "#6b7280",
    "text":     "#f1f5f9",
    // Unknown keys go into Theme.Extra automatically
    "push":   "#22c55e",
    "pr":     "#3b82f6",
    "review": "#a855f7",
})
```

## App-Specific Colors via Extra

Extend the theme with domain tokens for your application:

```go
// Look up with fallback
color := theme.Color("push", theme.Positive) // returns #22c55e
color  = theme.Color("missing", theme.Muted) // returns Muted fallback
```

## Theme Presets

blit ships several built-in presets accessible via `ThemePreset`:

```go
theme := blit.ThemePreset("dracula")
theme  = blit.ThemePreset("nord")
theme  = blit.ThemePreset("gruvbox")
```

## Hot Reload

Set `TUIKIT_THEME` environment variable to a JSON file path. blit watches the file and reloads the theme without restarting the app.

## Custom Glyphs

Override the cursor marker, flash marker, and other glyphs:

```go
theme := blit.DefaultTheme()
theme.Glyphs = &blit.Glyphs{
    CursorMarker: "▶",
    FlashMarker:  "★",
}
```
