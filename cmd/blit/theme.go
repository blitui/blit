package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"

	blit "github.com/blitui/blit"
)

// runTheme implements the `blit theme` subcommand.
// Usage:
//
//	blit theme              list available themes
//	blit theme <name>       preview a specific theme
//	blit theme --all        preview all themes
func runTheme(args []string) int {
	fs := flag.NewFlagSet("theme", flag.ExitOnError)
	all := fs.Bool("all", false, "preview all available themes")
	_ = fs.Parse(args)

	presets := blit.Presets()
	// Also add built-in themes that aren't in the registry.
	presets["default"] = blit.DefaultTheme()
	presets["light"] = blit.LightTheme()

	names := sortedThemeNames(presets)

	if *all {
		for _, name := range names {
			printThemePreview(name, presets[name])
			fmt.Println()
		}
		return 0
	}

	remaining := fs.Args()
	if len(remaining) == 0 {
		fmt.Println("Available themes:")
		fmt.Println()
		for _, name := range names {
			fmt.Printf("  %s\n", name)
		}
		fmt.Println()
		fmt.Println("Usage: blit theme <name>    preview a theme")
		fmt.Println("       blit theme --all     preview all themes")
		return 0
	}

	name := remaining[0]
	theme, ok := presets[name]
	if !ok {
		fmt.Fprintf(os.Stderr, "[blit theme] unknown theme %q\n", name)
		fmt.Fprintf(os.Stderr, "Available: %s\n", strings.Join(names, ", "))
		return 1
	}

	printThemePreview(name, theme)
	return 0
}

func sortedThemeNames(presets map[string]blit.Theme) []string {
	names := make([]string, 0, len(presets))
	for name := range presets {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func printThemePreview(name string, theme blit.Theme) {
	title := lipgloss.NewStyle().Bold(true).Render(name)
	fmt.Println(title)
	fmt.Println(strings.Repeat("─", len(name)+4))

	tokens := []struct {
		label string
		color lipgloss.Color
	}{
		{"Positive", theme.Positive},
		{"Negative", theme.Negative},
		{"Accent", theme.Accent},
		{"Muted", theme.Muted},
		{"Text", theme.Text},
		{"TextInverse", theme.TextInverse},
		{"Cursor", theme.Cursor},
		{"Border", theme.Border},
		{"Flash", theme.Flash},
		{"Warn", theme.Warn},
	}

	for _, tok := range tokens {
		swatch := lipgloss.NewStyle().
			Background(tok.color).
			Render("    ")
		label := lipgloss.NewStyle().
			Width(14).
			Render(tok.label)
		hex := string(tok.color)
		colored := lipgloss.NewStyle().
			Foreground(tok.color).
			Render(hex)
		fmt.Printf("  %s %s  %s\n", swatch, label, colored)
	}

	if len(theme.Extra) > 0 {
		fmt.Println()
		fmt.Println("  Extra tokens:")
		extraNames := make([]string, 0, len(theme.Extra))
		for k := range theme.Extra {
			extraNames = append(extraNames, k)
		}
		sort.Strings(extraNames)
		for _, k := range extraNames {
			c := theme.Extra[k]
			swatch := lipgloss.NewStyle().
				Background(c).
				Render("    ")
			label := lipgloss.NewStyle().
				Width(14).
				Render(k)
			colored := lipgloss.NewStyle().
				Foreground(c).
				Render(string(c))
			fmt.Printf("    %s %s  %s\n", swatch, label, colored)
		}
	}
}
