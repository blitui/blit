// Package tuikit is a pragmatic TUI toolkit for shipping Go CLI tools fast.
//
// It wraps Bubble Tea and Lip Gloss with reusable components (Table, ListView,
// Tabs, Picker, Form, Tree, LogViewer), a layout engine (DualPane, HBox, VBox,
// Split), a keybinding registry with auto-generated help, a dark/light theme
// system with hot-reload, and built-in binary self-update.
//
// Quick start:
//
//	app := tuikit.NewApp(
//	    tuikit.WithTheme(tuikit.DefaultTheme()),
//	    tuikit.WithComponent("main", myTable),
//	    tuikit.WithHelp(),
//	)
//	app.Run()
//
// See https://moneycaringcoder.github.io/tuikit-go/ for guides and examples.
package tuikit
