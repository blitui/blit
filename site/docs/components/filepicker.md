# FilePicker

File system browser with tree navigation, search, and optional preview pane. Built on top of the Tree component. Implements `Component` and `Themed`.

## Construction

```go
fp := blit.NewFilePicker(blit.FilePickerOpts{
    Root:        "/home/user/projects",
    ShowHidden:  false,
    PreviewPane: true,
    OnSelect: func(path string) {
        fmt.Println("Selected:", path)
    },
})
```

## FilePickerOpts

```go
type FilePickerOpts struct {
    Root        string         // Starting directory (default: ".")
    PreviewPane bool           // Show file preview on the right
    ShowHidden  bool           // Show dot-prefixed files
    OnSelect    func(path string) // Called on Enter
    OnCancel    func()         // Called on Esc
}
```

## Keyboard

| Key | Action |
|-----|--------|
| `j` / `k` | Move cursor down / up |
| `Enter` / `l` | Open directory or select file |
| `h` | Go to parent directory |
| `/` | Search files |
| `Esc` | Cancel |

## Example

```bash
go run ./examples/filetree/
```
