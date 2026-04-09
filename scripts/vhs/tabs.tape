# tabs.tape — tuikit-go Tabs component demo
# Run: vhs scripts/vhs/tabs.tape
# Requires: vhs (brew install vhs / go install github.com/charmbracelet/vhs@latest)
#
# Size budget: target < 1.5 MB
#   Width 800 / Height 480 / Framerate 12 / PlaybackSpeed 1.0

Output docs/gifs/tabs.gif
Set Width 800
Set Height 480
Set FontSize 13
Set Framerate 12
Set PlaybackSpeed 1.0
Set Theme "Dracula"
Set TypingSpeed 60ms

# Launch tabs example
Type "go run ./examples/tabs/"
Sleep 500ms
Key Enter
Sleep 1500ms

# Cycle through tabs with Tab key
Key Tab
Sleep 500ms
Key Tab
Sleep 500ms
Key Tab
Sleep 500ms

# Jump directly to tab 1 with number key
Type "1"
Sleep 400ms

# Jump to tab 2
Type "2"
Sleep 400ms

# Jump to tab 3
Type "3"
Sleep 400ms

# Navigate back with Shift+Tab
Key Shift+Tab
Sleep 400ms
Key Shift+Tab
Sleep 400ms

# Interact with content on current tab (navigate down)
Key Down
Sleep 200ms
Key Down
Sleep 200ms
Key Down
Sleep 200ms

Type "q"
Sleep 500ms

Sleep 2s
