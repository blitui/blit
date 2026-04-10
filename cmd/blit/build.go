package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// runBuild implements the `blit build` subcommand.
// It wraps goreleaser with simplified flags for common operations.
//
// Usage:
//
//	blit build                  build snapshot (local only, no publish)
//	blit build --release        full release build (requires GITHUB_TOKEN)
//	blit build --clean          remove dist/ before building
func runBuild(args []string) int {
	fs := flag.NewFlagSet("build", flag.ExitOnError)
	release := fs.Bool("release", false, "run a full release (requires GITHUB_TOKEN and a git tag)")
	clean := fs.Bool("clean", true, "remove dist/ before building")
	snapshot := fs.Bool("snapshot", true, "build a local snapshot (no publish)")
	_ = fs.Parse(args)

	if !goreleaserInstalled() {
		fmt.Fprintln(os.Stderr, "[blit build] goreleaser is not installed")
		fmt.Fprintln(os.Stderr, "  Install: https://goreleaser.com/install/")
		fmt.Fprintln(os.Stderr, "  Or: go install github.com/goreleaser/goreleaser/v2@latest")
		return 1
	}

	if !hasGoreleaserConfig() {
		fmt.Fprintln(os.Stderr, "[blit build] no .goreleaser.yaml found in current directory")
		fmt.Fprintln(os.Stderr, "  Run 'blit init' to scaffold a project with goreleaser config")
		return 1
	}

	var buildArgs []string
	if *release {
		buildArgs = append(buildArgs, "release")
	} else {
		buildArgs = append(buildArgs, "build")
	}

	if *clean {
		buildArgs = append(buildArgs, "--clean")
	}
	if *snapshot && !*release {
		buildArgs = append(buildArgs, "--snapshot")
	}

	fmt.Printf("[blit build] goreleaser %s\n", strings.Join(buildArgs, " "))
	cmd := exec.Command("goreleaser", buildArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()

	if err := cmd.Run(); err != nil {
		if exit, ok := err.(*exec.ExitError); ok {
			return exit.ExitCode()
		}
		fmt.Fprintf(os.Stderr, "[blit build] error: %v\n", err)
		return 1
	}

	fmt.Println("[blit build] done")
	return 0
}

func goreleaserInstalled() bool {
	_, err := exec.LookPath("goreleaser")
	return err == nil
}

func hasGoreleaserConfig() bool {
	candidates := []string{".goreleaser.yaml", ".goreleaser.yml"}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return true
		}
	}
	return false
}
