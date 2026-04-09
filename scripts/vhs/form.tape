# form.tape — tuikit-go Form component / signup flow demo
# Run: vhs scripts/vhs/form.tape
# Requires: vhs (brew install vhs / go install github.com/charmbracelet/vhs@latest)
#
# Size budget: target < 1.5 MB
#   Width 720 / Height 432 / Framerate 12 / PlaybackSpeed 1.0

Output docs/gifs/form.gif
Set Width 720
Set Height 432
Set FontSize 13
Set Framerate 12
Set PlaybackSpeed 1.0
Set Theme "Dracula"
Set TypingSpeed 70ms

# Launch form example
Type "go run ./examples/form/"
Sleep 500ms
Key Enter
Sleep 1500ms

# Fill in name field
Type "Alice"
Sleep 300ms
Key Tab
Sleep 300ms

# Email field
Type "alice@example.com"
Sleep 300ms
Key Tab
Sleep 300ms

# Password field
Type "hunter2"
Sleep 300ms
Key Tab
Sleep 300ms

# Select role via SelectField
Key Down
Sleep 200ms
Key Down
Sleep 200ms
Key Enter
Sleep 300ms
Key Tab
Sleep 300ms

# Confirm field
Type "y"
Sleep 300ms
Key Tab
Sleep 300ms

# Submit
Key Enter
Sleep 1000ms

# Switch to wizard mode if available
Sleep 500ms

Type "q"
Sleep 500ms

Sleep 2s
