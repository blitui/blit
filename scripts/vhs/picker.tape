# picker.tape — blit Picker (fuzzy file picker) demo
# Run: vhs scripts/vhs/picker.tape
# Requires: vhs (brew install vhs / go install github.com/charmbracelet/vhs@latest)
#
# Size budget: target < 1.5 MB
#   Width 800 / Height 480 / Framerate 12 / PlaybackSpeed 1.0

Output docs/gifs/picker.gif
Set Width 800
Set Height 480
Set FontSize 13
Set Framerate 12
Set PlaybackSpeed 1.0
Set Theme "Dracula"
Set TypingSpeed 70ms

# Launch picker example
Type "go run ./examples/picker/"
Sleep 500ms
Key Enter
Sleep 1500ms

# Type to filter — fuzzy search
Type "go"
Sleep 400ms
Type ".m"
Sleep 400ms

# Navigate results
Key Down
Sleep 300ms
Key Down
Sleep 300ms

# View preview pane updates
Sleep 500ms

# Clear filter
Key Ctrl+K
Sleep 400ms

# Filter for something else
Type "app"
Sleep 500ms

# Confirm selection
Key Enter
Sleep 800ms

# Quit
Type "q"
Sleep 500ms

Sleep 2s
