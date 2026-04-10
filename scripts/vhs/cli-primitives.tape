# cli-primitives.tape — blit CLI primitives showcase
# Run: vhs scripts/vhs/cli-primitives.tape
# Requires: vhs (brew install vhs / go install github.com/charmbracelet/vhs@latest)
#
# Size budget: target < 1.5 MB
#   Width 720 / Height 432 / Framerate 12 / PlaybackSpeed 1.0

Output docs/gifs/cli-primitives.gif
Set Width 720
Set Height 432
Set FontSize 13
Set Framerate 12
Set PlaybackSpeed 1.0
Set Theme "Dracula"
Set TypingSpeed 70ms

# Launch CLI demo
Type "go run ./examples/cli-demo/"
Sleep 500ms
Key Enter
Sleep 1000ms

# Confirm prompt — answer yes
Type "y"
Sleep 300ms
Key Enter
Sleep 500ms

# Input field
Type "blit"
Sleep 300ms
Key Enter
Sleep 500ms

# SelectOne — move and select
Key Down
Sleep 200ms
Key Down
Sleep 200ms
Key Enter
Sleep 500ms

# MultiSelect — toggle items
Space
Sleep 200ms
Key Down
Sleep 200ms
Space
Sleep 200ms
Key Enter
Sleep 500ms

# Password field
Type "s3cr3t"
Sleep 300ms
Key Enter
Sleep 500ms

# Spinner + progress (auto-advances)
Sleep 2000ms

Type "q"
Sleep 500ms

Sleep 2s
