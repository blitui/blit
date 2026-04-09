# table.tape — tuikit-go Table component demo
# Run: vhs scripts/vhs/table.tape
# Requires: vhs (brew install vhs / go install github.com/charmbracelet/vhs@latest)
#
# Size budget: target < 1.5 MB
#   Width 800 / Height 480 / Framerate 12 / PlaybackSpeed 1.0

Output docs/gifs/table.gif
Set Width 800
Set Height 480
Set FontSize 13
Set Framerate 12
Set PlaybackSpeed 1.0
Set Theme "Dracula"
Set TypingSpeed 60ms

# Launch dashboard which contains the table
Type "go run ./examples/dashboard/"
Sleep 500ms
Key Enter
Sleep 2000ms

# Sort by first column
Type "s"
Sleep 400ms
Type "s"
Sleep 400ms

# Filter
Type "/"
Sleep 300ms
Type "ali"
Sleep 600ms
Key Escape
Sleep 400ms

# Navigate rows
Key Down
Sleep 200ms
Key Down
Sleep 200ms
Key Down
Sleep 200ms
Key Up
Sleep 200ms

# Open detail overlay with Enter
Key Enter
Sleep 800ms
Key Escape
Sleep 400ms

# Quit
Type "q"
Sleep 500ms

Sleep 2s
