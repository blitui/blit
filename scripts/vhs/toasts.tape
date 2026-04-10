# toasts.tape — blit Toast / notification manager demo
# Run: vhs scripts/vhs/toasts.tape
# Requires: vhs (brew install vhs / go install github.com/charmbracelet/vhs@latest)
#
# Size budget: target < 1.5 MB
#   Width 800 / Height 480 / Framerate 15 / PlaybackSpeed 1.0

Output docs/gifs/toasts.gif
Set Width 800
Set Height 480
Set FontSize 13
Set Framerate 15
Set PlaybackSpeed 1.0
Set Theme "Dracula"
Set TypingSpeed 60ms

# Launch dashboard which fires toasts on startup
Type "go run ./examples/dashboard/"
Sleep 500ms
Key Enter
Sleep 2500ms

# Wait for auto-dismiss of first toasts then fire more manually
Type "n"
Sleep 800ms
Type "n"
Sleep 800ms
Type "n"
Sleep 800ms

# Dismiss top toast with Escape
Key Escape
Sleep 400ms
Key Escape
Sleep 400ms

# Quit
Type "q"
Sleep 500ms

Sleep 2s
