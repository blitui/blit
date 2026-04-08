// Package main demonstrates tuikit's cli/ package — interactive CLI primitives
// that work without a full-screen TUI. Confirm, Select, Input, Spinner, and ProgressBar
// in a single sequential flow.
package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/moneycaringcoder/tuikit-go/cli"
)

func main() {
	fmt.Println("╔══════════════════════════════════╗")
	fmt.Println("║   tuikit CLI Primitives Demo     ║")
	fmt.Println("╚══════════════════════════════════╝")
	fmt.Println()

	// 1. Confirm
	proceed := cli.Confirm("Would you like to run the demo?", true)
	fmt.Println()
	if !proceed {
		fmt.Println("Maybe next time!")
		return
	}

	// 2. Select
	languages := []string{"Go", "Rust", "Python", "TypeScript"}
	lang, _, err := cli.SelectOne("Pick your favorite language:", languages)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cancelled: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("  → %s, nice choice!\n\n", lang)

	// 3. Input with validation
	name, err := cli.Input("Enter your project name:", func(s string) error {
		if s == "" {
			return fmt.Errorf("project name cannot be empty")
		}
		return nil
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cancelled: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("  → Project: %s\n\n", name)

	// 4. Spinner with fake delay
	spinner := cli.Spin(fmt.Sprintf("Setting up %s...", name))
	time.Sleep(2 * time.Second)
	spinner.Stop()
	fmt.Printf("  ✓ Project scaffolded\n\n")

	// 5. Progress bar with fake install
	deps := 20 + rand.Intn(30)
	bar := cli.NewProgress(deps, "Installing deps")
	for i := 0; i < deps; i++ {
		time.Sleep(time.Duration(30+rand.Intn(70)) * time.Millisecond)
		bar.Increment(1)
	}
	bar.Done()
	fmt.Println()

	// 6. Summary
	fmt.Println("┌─────────────────────────────────┐")
	fmt.Printf("│  Language: %-21s│\n", lang)
	fmt.Printf("│  Project:  %-21s│\n", name)
	fmt.Printf("│  Deps:     %-21d│\n", deps)
	fmt.Println("│  Status:   Ready to go! 🚀      │")
	fmt.Println("└─────────────────────────────────┘")
}
