# quickstart.tape — tuikit-go quick-start demo (minimal ListView)
# Run: vhs scripts/vhs/quickstart.tape
# Requires: vhs (brew install vhs / go install github.com/charmbracelet/vhs@latest)
#
# Size budget: target < 1.5 MB
#   Width 640 / Height 384 / Framerate 12 / PlaybackSpeed 1.0

Output docs/gifs/quickstart.gif
Set Width 640
Set Height 384
Set FontSize 13
Set Framerate 12
Set PlaybackSpeed 1.0
Set Theme "Dracula"
Set TypingSpeed 80ms

# Launch the minimal example
Type "go run ./examples/minimal/"
Sleep 500ms
Key Enter
Sleep 1500ms

# Navigate down through the list
Key Down
Sleep 300ms
Key Down
Sleep 300ms
Key Down
Sleep 300ms
Key Down
Sleep 300ms

# Navigate back up
Key Up
Sleep 300ms
Key Up
Sleep 300ms

# Open help overlay
Type "?"
Sleep 800ms

# Dismiss help
Key Escape
Sleep 500ms

# Quit
Type "q"
Sleep 500ms

Sleep 2s
