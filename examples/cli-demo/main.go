// Package main demonstrates blit's cli/ package — interactive CLI primitives
// that work without a full-screen TUI.
package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/blitui/blit/cli"
)

func main() {
	cli.Title("blit CLI Primitives Demo")

	// 1. Confirm
	cli.Step(1, 7, "Confirm prompt")
	proceed := cli.Confirm("  Run the full demo?", true)
	fmt.Println()
	if !proceed {
		cli.Warning("Maybe next time!")
		return
	}

	// 2. Select
	cli.Step(2, 7, "Single select")
	languages := []string{"Go", "Rust", "Python", "TypeScript"}
	lang, _, err := cli.SelectOne("  Pick your favorite language:", languages)
	if err != nil {
		cli.Error(fmt.Sprintf("Cancelled: %v", err))
		os.Exit(1)
	}
	fmt.Println()

	// 3. MultiSelect
	cli.Step(3, 7, "Multi select")
	features := []string{"Table", "ListView", "ConfigEditor", "Auto-Update", "CLI Primitives", "blit"}
	selected, _, err := cli.MultiSelect("  Which blit features do you use?", features)
	if err != nil {
		cli.Error(fmt.Sprintf("Cancelled: %v", err))
		os.Exit(1)
	}
	fmt.Println()

	// 4. Input
	cli.Step(4, 7, "Text input")
	name, err := cli.Input("  Project name:", func(s string) error {
		if s == "" {
			return fmt.Errorf("cannot be empty")
		}
		return nil
	})
	if err != nil {
		cli.Error(fmt.Sprintf("Cancelled: %v", err))
		os.Exit(1)
	}
	fmt.Println()

	// 5. Password
	cli.Step(5, 7, "Password input")
	secret, err := cli.Password("  API token:", func(s string) error {
		if len(s) < 4 {
			return fmt.Errorf("must be at least 4 characters")
		}
		return nil
	})
	if err != nil {
		cli.Error(fmt.Sprintf("Cancelled: %v", err))
		os.Exit(1)
	}
	_ = secret
	fmt.Println()

	// 6. Spinner
	cli.Step(6, 7, "Spinner")
	spinner := cli.Spin(fmt.Sprintf("  Scaffolding %s...", name))
	time.Sleep(2 * time.Second)
	spinner.Stop()
	cli.Success("Project scaffolded")
	fmt.Println()

	// 7. Progress
	cli.Step(7, 7, "Progress bar")
	deps := 20 + rand.Intn(30)
	bar := cli.NewProgress(deps, "  Installing")
	for i := 0; i < deps; i++ {
		time.Sleep(time.Duration(30+rand.Intn(70)) * time.Millisecond)
		bar.Increment(1)
	}
	bar.Done()
	fmt.Println()

	// Summary
	cli.Separator()
	cli.Section("Summary")
	fmt.Println()
	cli.KeyValue("Language", lang)
	cli.KeyValue("Project", name)
	cli.KeyValue("Features", fmt.Sprintf("%d selected", len(selected)))
	cli.KeyValue("Deps", fmt.Sprintf("%d installed", deps))
	fmt.Println()
	cli.Success("Done! Ready to build.")
	fmt.Println()
}
