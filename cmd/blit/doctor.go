package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// runDoctor implements the `blit doctor` subcommand.
// It runs diagnostic checks to verify the development environment is set up
// correctly for blit TUI development.
func runDoctor(_ []string) int {
	fmt.Println("[blit doctor] checking development environment...")
	fmt.Println()

	passed := 0
	failed := 0

	checks := []struct {
		name string
		fn   func() (string, bool)
	}{
		{"go", checkGo},
		{"go.mod", checkGoMod},
		{"blit dependency", checkBlitDep},
		{"goreleaser", checkGoreleaser},
		{"terminal", checkTerminal},
		{"git", checkGit},
	}

	for _, c := range checks {
		detail, ok := c.fn()
		if ok {
			fmt.Printf("  ✓ %s — %s\n", c.name, detail)
			passed++
		} else {
			fmt.Printf("  ✗ %s — %s\n", c.name, detail)
			failed++
		}
	}

	fmt.Println()
	fmt.Printf("[blit doctor] %d passed, %d issues\n", passed, failed)

	if failed > 0 {
		return 1
	}
	return 0
}

func checkGo() (string, bool) {
	out, err := exec.Command("go", "version").Output()
	if err != nil {
		return "go not found in PATH", false
	}
	return strings.TrimSpace(string(out)), true
}

func checkGoMod() (string, bool) {
	if _, err := os.Stat("go.mod"); err != nil {
		return "no go.mod found (not in a Go module)", false
	}
	return "go.mod present", true
}

func checkBlitDep() (string, bool) {
	data, err := os.ReadFile("go.mod")
	if err != nil {
		return "cannot read go.mod", false
	}
	content := string(data)
	if strings.Contains(content, "module github.com/blitui/blit") {
		return "this IS the blit module", true
	}
	if strings.Contains(content, "github.com/blitui/blit") {
		return "github.com/blitui/blit found in go.mod", true
	}
	return "github.com/blitui/blit not found in go.mod", false
}

func checkGoreleaser() (string, bool) {
	path, err := exec.LookPath("goreleaser")
	if err != nil {
		return "not installed (optional, needed for 'blit build')", false
	}
	out, err := exec.Command(path, "--version").Output()
	if err != nil {
		return "installed but version check failed", true
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) > 0 {
		return lines[0], true
	}
	return "installed", true
}

func checkTerminal() (string, bool) {
	term := os.Getenv("TERM")
	colorTerm := os.Getenv("COLORTERM")

	if runtime.GOOS == "windows" {
		wt := os.Getenv("WT_SESSION")
		if wt != "" {
			return "Windows Terminal (truecolor)", true
		}
		return fmt.Sprintf("TERM=%q (Windows)", term), true
	}

	if colorTerm == "truecolor" || colorTerm == "24bit" {
		return fmt.Sprintf("TERM=%s, COLORTERM=%s (truecolor)", term, colorTerm), true
	}
	if term == "" {
		return "TERM not set (may have limited color support)", false
	}
	if strings.Contains(term, "256color") {
		return fmt.Sprintf("TERM=%s (256-color)", term), true
	}
	return fmt.Sprintf("TERM=%s", term), true
}

func checkGit() (string, bool) {
	out, err := exec.Command("git", "--version").Output()
	if err != nil {
		return "git not found in PATH", false
	}
	return strings.TrimSpace(string(out)), true
}
