# tuitest-runner.tape — tuikit-go tuitest vitest-style reporter demo
# Run: vhs scripts/vhs/tuitest-runner.tape
# Requires: vhs (brew install vhs / go install github.com/charmbracelet/vhs@latest)
#
# Size budget: target < 1.5 MB
#   Width 800 / Height 480 / Framerate 12 / PlaybackSpeed 1.0

Output docs/gifs/tuitest-runner.gif
Set Width 800
Set Height 480
Set FontSize 13
Set Framerate 12
Set PlaybackSpeed 1.0
Set Theme "Dracula"
Set TypingSpeed 60ms

# Run the tuitest suite via the CLI with verbose output
Type "go test -v ./tuitest/... 2>&1 | head -60"
Sleep 500ms
Key Enter
Sleep 4000ms

# Show it completing
Sleep 1000ms

Sleep 2s
