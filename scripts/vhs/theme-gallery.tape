# theme-gallery.tape — blit 8-preset theme gallery cycle
# Run: vhs scripts/vhs/theme-gallery.tape
# Requires: vhs (brew install vhs / go install github.com/charmbracelet/vhs@latest)
#
# Size budget: target < 1.5 MB
#   Width 800 / Height 480 / Framerate 12 / PlaybackSpeed 1.0

Output docs/gifs/theme-gallery.gif
Set Width 800
Set Height 480
Set FontSize 13
Set Framerate 12
Set PlaybackSpeed 1.0
Set Theme "Dracula"
Set TypingSpeed 60ms

# Launch dashboard with theme cycling
Type "go run ./examples/dashboard/"
Sleep 500ms
Key Enter
Sleep 2000ms

# Cycle through all 8 theme presets with 'c'
Type "c"
Sleep 700ms
Type "c"
Sleep 700ms
Type "c"
Sleep 700ms
Type "c"
Sleep 700ms
Type "c"
Sleep 700ms
Type "c"
Sleep 700ms
Type "c"
Sleep 700ms
Type "c"
Sleep 700ms

# Back to start
Sleep 500ms

# Quit
Type "q"
Sleep 500ms

Sleep 2s
