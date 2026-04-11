package blit

import (
	"fmt"
	"os"
	"strings"
)

// CLIApp is a lightweight subcommand router for TUI applications.
// Many TUI apps need a CLI layer for config management (add/remove/list)
// alongside the main TUI. CLIApp handles this pattern uniformly.
//
// Usage:
//
//	app := blit.NewCLIApp(blit.CLIAppOpts{
//	    Name:    "myapp",
//	    Version: "1.0.0",
//	    Commands: []blit.Subcommand{
//	        {Name: "add", Usage: "add <item>", Run: func(args []string) error { ... }},
//	        {Name: "list", Aliases: []string{"ls"}, Run: func(args []string) error { ... }},
//	    },
//	    RunTUI: func() error { return runTUI() },
//	})
//	os.Exit(app.Execute(os.Args[1:]))
type CLIApp struct {
	Name     string
	Version  string
	Commands []Subcommand
	RunTUI   func() error
}

// Subcommand defines a single CLI subcommand.
type Subcommand struct {
	Name    string                    // Primary name: "add", "remove"
	Aliases []string                  // Alternatives: ["rm"] for "remove"
	Usage   string                    // Usage text: "add <owner/repo>"
	Run     func(args []string) error // Handler, receives remaining args
}

// CLIAppOpts configures a CLIApp.
type CLIAppOpts struct {
	Name     string
	Version  string
	Commands []Subcommand
	RunTUI   func() error
}

// NewCLIApp creates a CLIApp with the given options.
func NewCLIApp(opts CLIAppOpts) *CLIApp {
	return &CLIApp{
		Name:     opts.Name,
		Version:  opts.Version,
		Commands: opts.Commands,
		RunTUI:   opts.RunTUI,
	}
}

// Execute dispatches a subcommand or launches the TUI. It returns an exit
// code: 0 for success, 1 for error. When no subcommand matches, it calls
// RunTUI. When args is empty, it also calls RunTUI.
func (a *CLIApp) Execute(args []string) int {
	if len(args) == 0 {
		if a.RunTUI != nil {
			if err := a.RunTUI(); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				return 1
			}
			return 0
		}
		a.PrintHelp()
		return 0
	}

	cmd := args[0]
	rest := args[1:]

	// Help flags
	switch cmd {
	case "help", "--help", "-h":
		a.PrintHelp()
		return 0
	case "version", "--version", "-v":
		fmt.Println(a.Name + " " + a.Version)
		return 0
	}

	// Look up subcommand
	for _, c := range a.Commands {
		if strings.EqualFold(c.Name, cmd) {
			if c.Run != nil {
				if err := c.Run(rest); err != nil {
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
					return 1
				}
				return 0
			}
			return 0
		}
		for _, alias := range c.Aliases {
			if strings.EqualFold(alias, cmd) {
				if c.Run != nil {
					if err := c.Run(rest); err != nil {
						fmt.Fprintf(os.Stderr, "Error: %v\n", err)
						return 1
					}
					return 0
				}
				return 0
			}
		}
	}

	// Unknown subcommand — try TUI
	if a.RunTUI != nil {
		if err := a.RunTUI(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 1
		}
		return 0
	}

	fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmd)
	a.PrintHelp()
	return 1
}

// PrintHelp prints usage information for all registered subcommands.
func (a *CLIApp) PrintHelp() {
	fmt.Printf("%s - %s\n\n", a.Name, "a terminal UI application")
	fmt.Println("Usage:")
	fmt.Printf("  %s              Launch the TUI\n", a.Name)
	for _, c := range a.Commands {
		label := c.Name
		if len(c.Aliases) > 0 {
			label += " (" + strings.Join(c.Aliases, ", ") + ")"
		}
		if c.Usage != "" {
			fmt.Printf("  %s %s    %s\n", a.Name, label, c.Usage)
		} else {
			fmt.Printf("  %s %s\n", a.Name, label)
		}
	}
	fmt.Printf("  %s help         Show this help\n", a.Name)
	fmt.Printf("  %s version      Show version\n", a.Name)
}
