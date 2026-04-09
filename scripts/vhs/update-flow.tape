# update-flow.tape — tuikit-go self-update progress flow demo
# Run: vhs scripts/vhs/update-flow.tape
# Requires: vhs (brew install vhs / go install github.com/charmbracelet/vhs@latest)
#
# Size budget: target < 1.5 MB
#   Width 800 / Height 480 / Framerate 15 / PlaybackSpeed 1.0

Output docs/gifs/update-flow.gif
Set Width 800
Set Height 480
Set FontSize 13
Set Framerate 15
Set PlaybackSpeed 1.0
Set Theme "Dracula"
Set TypingSpeed 60ms

# Launch monitor example which shows the update progress bar
Type "go run ./examples/monitor/"
Sleep 500ms
Key Enter
Sleep 2000ms

# Navigate through the monitor dashboard
Key Down
Sleep 300ms
Key Down
Sleep 300ms
Key Down
Sleep 300ms

# Watch for update shimmer on progress bar
Sleep 1500ms

# Navigate back up
Key Up
Sleep 300ms
Key Up
Sleep 300ms

# Quit
Type "q"
Sleep 500ms

Sleep 2s
