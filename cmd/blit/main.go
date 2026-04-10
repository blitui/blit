// Command blit is a thin wrapper around `go test` for blit-powered
// test suites. It adds flags that map to blit features: -update to
// regenerate snapshots, -junit/-html to emit reports, -filter to pick
// specific tests, -parallel to set parallelism, and -watch to re-run on
// file changes.
//
// Subcommands:
//
//	blit diff <testname>   show the failure diff for a named test
//	blit review [root]     interactive review of pending .golden.new snapshots
//
// Usage:
//
//	blit [flags] [packages...]
//	blit record <name> -- <command> [args...]
//	blit replay [--speed 1x] <name>
//	blit review [root]
//
// Packages default to "./..." when none are provided. The default reporter
// is the vitest-style runner already wired into the test code.
//
// Examples:
//
//	blit                                   # go test ./...
//	blit -filter TestHarness ./btest/... # run tests matching TestHarness
//	blit -update ./btest/...             # regenerate snapshots
//	blit -junit out/junit.xml -parallel 4  # parallel run + junit output
//	blit -watch                            # re-run on file changes (1s poll)
//	blit diff TestFoo                      # show diff for TestFoo failure
//	blit record dashboard -- ./bin/dashboard
//	blit replay dashboard --speed 2x
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

	"github.com/blitui/blit/btest"
)

func main() {
	// Sub-commands handled before flag parsing so that subcommand flags
	// don't collide with the top-level flag set.
	if len(os.Args) >= 2 {
		switch os.Args[1] {
		case "diff":
			runDiffSubcommand(os.Args[2:])
			return
		case "record":
			os.Exit(runRecord(os.Args[2:]))
		case "replay":
			os.Exit(runReplay(os.Args[2:]))
		case "init":
			os.Exit(runInit(os.Args[2:]))
		case "dev":
			os.Exit(runDev(os.Args[2:]))
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
		case "review":
			os.Exit(runReview(os.Args[2:]))
		case "theme":
			os.Exit(runTheme(os.Args[2:]))
		case "vhs":
			os.Exit(runVHS(os.Args[2:]))
		}
	}

	var (
		filter   = flag.String("filter", "", "run only tests matching regexp (maps to go test -run)")
		update   = flag.Bool("update", false, "regenerate blit snapshots (passes -btest.update to tests)")
		junit    = flag.String("junit", "", "write JUnit XML report to path (informational; tests must use JUnitReporter to populate it)")
		htmlOut  = flag.String("html", "", "write HTML report to path (informational; tests must use HTMLReporter to populate it)")
		parallel = flag.Int("parallel", 0, "maximum number of tests to run in parallel (maps to go test -parallel)")
		watch    = flag.Bool("watch", false, "watch the working tree for changes and re-run on modification")
		verbose  = flag.Bool("v", false, "verbose go test output")
		keep     = flag.Int("keep", defaultKeep, "max history entries to keep (prune older runs)")
		coverage = flag.Bool("coverage", false, "run go test with -coverprofile and display a coverage summary panel")
		timeout  = flag.String("timeout", "", "per-test timeout (e.g., 30s, 2m); passed to go test -timeout")
		jsonOut  = flag.Bool("json", false, "emit test results as JSON lines (uses go test -json)")
		failOnly = flag.Bool("fail", false, "only print failing tests; suppress passing test output")
		smoke    = flag.Bool("smoke", false, "run auto-generated smoke tests (init, render, keys, resize, mouse)")
	)
	flag.Parse()

	packages := flag.Args()
	if len(packages) == 0 {
		packages = []string{"./..."}
	}

	if *jsonOut && *failOnly {
		fmt.Fprintln(os.Stderr, "[blit] --json and --fail cannot be used together (--fail filters plain-text output)")
		os.Exit(1)
	}

	if *smoke {
		os.Exit(runSmoke(packages, *verbose))
	}

	if *coverage {
		os.Exit(runCoverage(packages))
	}

	runOnce := func() int {
		code := runGoTest(runOpts{
			filter:   *filter,
			update:   *update,
			junit:    *junit,
			htmlOut:  *htmlOut,
			parallel: *parallel,
			verbose:  *verbose,
			keep:     *keep,
			timeout:  *timeout,
			jsonOut:  *jsonOut,
			failOnly: *failOnly,
		}, packages)
		if code != 0 && *watch {
			printFailureDiffHints()
		}
		return code
	}

	if !*watch {
		os.Exit(runOnce())
	}

	// Watch mode: interactive TUI with status bar, filter panel, and log viewer.
	if err := RunWatchMode(packages); err != nil {
		fmt.Fprintf(os.Stderr, "[blit] watch mode error: %v\n", err)
		os.Exit(1)
	}
}

// runDiffSubcommand implements `blit diff [testname]`.
// Without a testname it lists all available failure captures.
// With a testname it prints the diff to stdout using DiffViewer's text output.
func runDiffSubcommand(args []string) {
	if len(args) == 0 {
		// List available captures.
		names, err := btest.ListFailureCaptures()
		if err != nil {
			fmt.Fprintf(os.Stderr, "[blit diff] error: %v\n", err)
			os.Exit(1)
		}
		if len(names) == 0 {
			fmt.Println("[blit diff] no failure captures found (run tests first)")
			return
		}
		fmt.Println("Available failure captures:")
		for _, n := range names {
			fmt.Println("  " + n)
		}
		return
	}

	testName := strings.Join(args, " ")
	fc, err := btest.LoadFailureCapture(testName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[blit diff] %v\n", err)
		os.Exit(1)
	}

	dv := btest.NewDiffViewer(fc)
	dv.SetSize(120, 40)
	printDiffViewerModes(dv, fc)
}

// printDiffViewerModes renders all three modes of the DiffViewer to stdout.
func printDiffViewerModes(dv *btest.DiffViewer, fc *btest.FailureCapture) {
	modes := []struct {
		key  string
		mode btest.DiffMode
	}{
		{"s", btest.DiffModeSideBySide},
		{"u", btest.DiffModeUnified},
		{"d", btest.DiffModeCellsOnly},
	}
	// Show side-by-side by default; user can re-run to see other modes.
	_ = modes
	// For one-shot CLI we just render side-by-side then unified then cells.
	for _, m := range modes {
		dv.SetMode(m.mode)
		fmt.Println(dv.View())
		fmt.Println(strings.Repeat("─", 80))
	}
}

// printFailureDiffHints prints a hint after a failed watch-mode run showing
// which test failures have diff captures available.
func printFailureDiffHints() {
	names, err := btest.ListFailureCaptures()
	if err != nil || len(names) == 0 {
		return
	}
	fmt.Fprintln(os.Stderr, "[blit] failure diffs available — view with:")
	for _, n := range names {
		fmt.Fprintf(os.Stderr, "  blit diff %s\n", n)
	}
}

// runOpts holds configuration for a single go test invocation.
type runOpts struct {
	filter   string
	update   bool
	junit    string
	htmlOut  string
	parallel int
	verbose  bool
	keep     int
	timeout  string
	jsonOut  bool
	failOnly bool
}

func runGoTest(opts runOpts, packages []string) int {
	// When --junit or --html is specified, force JSON mode so we can parse
	// the output and generate reports automatically.
	reportMode := opts.junit != "" || opts.htmlOut != ""

	args := []string{"test"}
	if opts.jsonOut || reportMode {
		args = append(args, "-json")
	}
	if opts.verbose {
		args = append(args, "-v")
	}
	if opts.filter != "" {
		args = append(args, "-run", opts.filter)
	}
	if opts.parallel > 0 {
		args = append(args, "-parallel", fmt.Sprintf("%d", opts.parallel))
	}
	if opts.timeout != "" {
		args = append(args, "-timeout", opts.timeout)
	}
	args = append(args, packages...)
	if opts.update {
		args = append(args, "-args", "-btest.update")
	}

	start := time.Now()
	cmd := exec.Command("go", args...)

	if opts.failOnly {
		// Capture output and filter to failures only.
		out, runErr := cmd.CombinedOutput()
		duration := time.Since(start).Seconds()
		exitCode := 0
		if runErr != nil {
			if exit, ok := runErr.(*exec.ExitError); ok {
				exitCode = exit.ExitCode()
			} else {
				fmt.Fprintf(os.Stderr, "[blit] run failed: %v\n", runErr)
				return 1
			}
		}
		printFailuresOnly(string(out))
		writeRunRecord(duration, exitCode, opts.keep, packages)
		return exitCode
	}

	if reportMode || opts.jsonOut {
		// Capture output for parsing and optional display.
		out, runErr := cmd.CombinedOutput()
		duration := time.Since(start).Seconds()
		exitCode := 0
		if runErr != nil {
			if exit, ok := runErr.(*exec.ExitError); ok {
				exitCode = exit.ExitCode()
			} else {
				fmt.Fprintf(os.Stderr, "[blit] run failed: %v\n", runErr)
				return 1
			}
		}

		if opts.jsonOut && !reportMode {
			// User just wants raw JSON output.
			_, _ = os.Stdout.Write(out)
		} else {
			// Parse events and generate reports.
			events := parseTestEvents(strings.NewReader(string(out)))
			report := buildReport(events, strings.Join(packages, " "))
			printTestSummary(report)

			if opts.junit != "" {
				if err := report.WriteJUnit(opts.junit); err != nil {
					fmt.Fprintf(os.Stderr, "[blit] failed to write JUnit report: %v\n", err)
				} else {
					fmt.Fprintf(os.Stderr, "[blit] JUnit report written to %s\n", opts.junit)
				}
			}
			if opts.htmlOut != "" {
				if err := report.WriteHTML(opts.htmlOut); err != nil {
					fmt.Fprintf(os.Stderr, "[blit] failed to write HTML report: %v\n", err)
				} else {
					fmt.Fprintf(os.Stderr, "[blit] HTML report written to %s\n", opts.htmlOut)
				}
			}
		}

		writeRunRecord(duration, exitCode, opts.keep, packages)
		return exitCode
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	runErr := cmd.Run()
	duration := time.Since(start).Seconds()

	exitCode := 0
	if runErr != nil {
		if exit, ok := runErr.(*exec.ExitError); ok {
			exitCode = exit.ExitCode()
		} else {
			fmt.Fprintf(os.Stderr, "[blit] run failed: %v\n", runErr)
			return 1
		}
	}

	writeRunRecord(duration, exitCode, opts.keep, packages)
	return exitCode
}

// printTestSummary prints a concise test summary from a parsed report.
func printTestSummary(report *btest.Report) {
	total, passed, failed, skipped := report.Totals()
	duration := report.TotalDuration()

	if failed > 0 {
		for _, r := range report.Results {
			if !r.Passed && !r.Skipped {
				fmt.Printf("FAIL %s/%s (%.2fs)\n", r.Package, r.Name, r.Duration.Seconds())
				if r.Failure != "" {
					for _, line := range strings.Split(r.Failure, "\n") {
						if strings.TrimSpace(line) != "" {
							fmt.Printf("  %s\n", line)
						}
					}
				}
			}
		}
		fmt.Println()
	}
	fmt.Printf("[blit] %d passed, %d failed, %d skipped (%d total) in %.2fs\n",
		passed, failed, skipped, total, duration.Seconds())
}

// printFailuresOnly filters go test output to show only FAIL lines
// and their associated output, plus the final summary.
func printFailuresOnly(output string) {
	lines := strings.Split(output, "\n")
	inFail := false
	for _, line := range lines {
		switch {
		case strings.HasPrefix(line, "--- FAIL:"):
			inFail = true
			fmt.Println(line)
		case strings.HasPrefix(line, "--- PASS:") || strings.HasPrefix(line, "--- SKIP:"):
			inFail = false
		case strings.HasPrefix(line, "FAIL") || strings.HasPrefix(line, "ok "):
			// Summary lines — always print.
			fmt.Println(line)
			inFail = false
		case inFail:
			fmt.Println(line)
		}
	}
}

func writeRunRecord(duration float64, exitCode, keep int, packages []string) {
	failed := 0
	if exitCode != 0 {
		failed = 1
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
	_ = writeHistory(rec, keep)
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
