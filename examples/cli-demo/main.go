// Package main demonstrates tuikit's cli/ package — interactive CLI primitives
// that work without a full-screen TUI.
package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/moneycaringcoder/tuikit-go/cli"
)

var (
	titleStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205")).Padding(0, 1)
	stepStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("63"))
	labelStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("82"))
	accentStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true)
	boxStyle     = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")).
			Padding(0, 2)
)

func main() {
	fmt.Println()
	fmt.Println(titleStyle.Render("tuikit CLI Primitives Demo"))
	fmt.Println()

	// 1. Confirm
	fmt.Println(stepStyle.Render("1/5") + " " + labelStyle.Render("Confirm"))
	proceed := cli.Confirm("  Run the full demo?", true)
	fmt.Println()
	if !proceed {
		fmt.Println(labelStyle.Render("  Maybe next time!"))
		return
	}

	// 2. Select
	fmt.Println(stepStyle.Render("2/5") + " " + labelStyle.Render("Select"))
	languages := []string{"Go", "Rust", "Python", "TypeScript"}
	lang, _, err := cli.SelectOne("  Pick your favorite language:", languages)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cancelled: %v\n", err)
		os.Exit(1)
	}
	fmt.Println()

	// 3. Input
	fmt.Println(stepStyle.Render("3/5") + " " + labelStyle.Render("Text input"))
	name, err := cli.Input("  Project name:", func(s string) error {
		if s == "" {
			return fmt.Errorf("cannot be empty")
		}
		return nil
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cancelled: %v\n", err)
		os.Exit(1)
	}
	fmt.Println()

	// 4. Spinner
	fmt.Println(stepStyle.Render("4/5") + " " + labelStyle.Render("Spinner"))
	spinner := cli.Spin(fmt.Sprintf("  Scaffolding %s...", accentStyle.Render(name)))
	time.Sleep(2 * time.Second)
	spinner.Stop()
	fmt.Println(successStyle.Render("  ✓ Project scaffolded"))
	fmt.Println()

	// 5. Progress
	fmt.Println(stepStyle.Render("5/5") + " " + labelStyle.Render("Progress bar"))
	deps := 20 + rand.Intn(30)
	bar := cli.NewProgress(deps, "  Installing")
	for i := 0; i < deps; i++ {
		time.Sleep(time.Duration(30+rand.Intn(70)) * time.Millisecond)
		bar.Increment(1)
	}
	bar.Done()
	fmt.Println()

	// Summary
	summary := fmt.Sprintf(
		"%s %s\n%s %s\n%s %s",
		labelStyle.Render("Language:"),
		accentStyle.Render(lang),
		labelStyle.Render("Project: "),
		accentStyle.Render(name),
		labelStyle.Render("Deps:    "),
		accentStyle.Render(fmt.Sprintf("%d installed", deps)),
	)
	fmt.Println(boxStyle.Render(summary))
	fmt.Println()
	fmt.Println(successStyle.Render("  Done! Ready to build."))
	fmt.Println()
}
