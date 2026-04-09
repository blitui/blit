// Command tuitest is a thin wrapper around `go test` for tuitest-powered
// test suites. It adds flags that map to tuitest features: -update to
// regenerate snapshots, -junit/-html to emit reports, -filter to pick
// specific tests, -parallel to set parallelism, and -watch to re-run on
// file changes.
//
// Usage:
//
//	tuitest [flags] [packages...]
//	tuitest record <name> -- <command> [args...]
//	tuitest replay [--speed 1x] <name>
//
// Packages default to "./..." when none are provided. The default reporter
// is the vitest-style runner already wired into the test code.
//
// Examples:
//
//	tuitest                                   # go test ./...
//	tuitest -filter TestHarness ./tuitest/... # run tests matching TestHarness
//	tuitest -update ./tuitest/...             # regenerate snapshots
//	tuitest -junit out/junit.xml -parallel 4  # parallel run + junit output
//	tuitest -watch                            # re-run on file changes (1s poll)
//	tuitest record dashboard -- ./bin/dashboard
//	tuitest replay dashboard --speed 2x
package main

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	// Sub-commands handled before flag parsing so that subcommand flags
	// don't collide with the top-level flag set.
	if len(os.Args) >= 2 {
		switch os.Args[1] {
		case "record":
			os.Exit(runRecord(os.Args[2:]))
		case "replay":
			os.Exit(runReplay(os.Args[2:]))
		case "gen":
			os.Exit(runGen(os.Args[2:]))
		case "history":
			fs := flag.NewFlagSet("history", flag.ExitOnError)
			keep := fs.Int("keep", defaultKeep, "number of recent runs to display")
			_ = fs.Parse(os.Args[2:])
			os.Exit(cmdHistory(*keep))
		case "report":
			fs := flag.NewFlagSet("report", flag.ExitOnError)
			out := fs.String("out", "report.html", "output path for the HTML report")
			_ = fs.Parse(os.Args[2:])
			os.Exit(cmdReport(*out))
		case "coverage":
			os.Exit(readCoverage())
		}
	}

	var (
		filter   = flag.String("filter", "", "run only tests matching regexp (maps to go test -run)")
		update   = flag.Bool("update", false, "regenerate tuitest snapshots (passes -tuitest.update to tests)")
		junit    = flag.String("junit", "", "write JUnit XML report to path (informational; tests must use JUnitReporter to populate it)")
		htmlOut  = flag.String("html", "", "write HTML report to path (informational; tests must use HTMLReporter to populate it)")
		parallel = flag.Int("parallel", 0, "maximum number of tests to run in parallel (maps to go test -parallel)")
		watch    = flag.Bool("watch", false, "watch the working tree for changes and re-run on modification")
		verbose  = flag.Bool("v", false, "verbose go test output")
		keep     = flag.Int("keep", defaultKeep, "max history entries to keep (prune older runs)")
		coverage = flag.Bool("coverage", false, "run go test with -coverprofile and display a coverage summary panel")
	)
	flag.Parse()

	packages := flag.Args()
	if len(packages) == 0 {
		packages = []string{"./..."}
	}

	if *coverage {
		os.Exit(runCoverage(packages))
	}

	runOnce := func() int {
		return runGoTest(*filter, *update, *junit, *htmlOut, *parallel, *verbose, *keep, packages)
	}

	if !*watch {
		os.Exit(runOnce())
	}

	// Watch mode: simple mtime poll on .go files under cwd.
	fmt.Fprintln(os.Stderr, "[tuitest] watch mode (polling every 1s, Ctrl+C to stop)")
	lastHash := snapshotTree(".")
	runOnce()
	for {
		time.Sleep(time.Second)
		h := snapshotTree(".")
		if h != lastHash {
			fmt.Fprintln(os.Stderr, "[tuitest] change detected, re-running")
			lastHash = h
			runOnce()
		}
	}
}

func runGoTest(filter string, update bool, junit, htmlOut string, parallel int, verbose bool, keep int, packages []string) int {
	args := []string{"test"}
	if verbose {
		args = append(args, "-v")
	}
	if filter != "" {
		args = append(args, "-run", filter)
	}
	if parallel > 0 {
		args = append(args, "-parallel", fmt.Sprintf("%d", parallel))
	}
	args = append(args, packages...)
	if update || junit != "" || htmlOut != "" {
		args = append(args, "-args")
		if update {
			args = append(args, "-tuitest.update")
		}
		if junit != "" {
			args = append(args, "-tuitest.junit="+junit)
		}
		if htmlOut != "" {
			args = append(args, "-tuitest.html="+htmlOut)
		}
	}

	start := time.Now()
	cmd := exec.Command("go", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	runErr := cmd.Run()
	duration := time.Since(start).Seconds()

	exitCode := 0
	failed := 0
	if runErr != nil {
		if exit, ok := runErr.(*exec.ExitError); ok {
			exitCode = exit.ExitCode()
		} else {
			fmt.Fprintf(os.Stderr, "[tuitest] run failed: %v\n", runErr)
			return 1
		}
		if exitCode != 0 {
			failed = 1 // at least one failure; exact counts require JSON output parsing
		}
	}

	passed := 0
	if failed == 0 {
		passed = 1
	}
	rec := RunRecord{
		RunAt:    time.Now(),
		Duration: duration,
		Passed:   passed,
		Failed:   failed,
		Total:    passed + failed,
		Packages: packages,
	}
	// Best-effort: ignore history write errors so tests still work without a writable FS.
	_ = writeHistory(rec, keep)

	return exitCode
}

// snapshotTree returns a coarse hash of .go file modification times under
// root. Used only to detect "something changed" in watch mode.
func snapshotTree(root string) string {
	var sb strings.Builder
	_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			name := d.Name()
			if name == ".git" || name == "node_modules" || name == "vendor" || strings.HasPrefix(name, ".omc") {
				return fs.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return nil
		}
		fmt.Fprintf(&sb, "%s:%d\n", path, info.ModTime().UnixNano())
		return nil
	})
	return sb.String()
}
