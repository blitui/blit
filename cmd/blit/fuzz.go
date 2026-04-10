package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// runFuzz implements `blit --fuzz` to discover and run fuzz tests.
//
// It invokes `go test` with `-run` targeting test functions matching the
// fuzz pattern (default: TestFuzz). For Go-native fuzz targets (FuzzXxx),
// it uses `-fuzz` with the configured duration.
//
// Usage examples:
//
//	blit --fuzz                          # run TestFuzz.* in ./...
//	blit --fuzz ./btest/...              # run in specific packages
//	blit --fuzz -fuzz-time 30s ./...     # Go-native fuzz for 30s
func runFuzz(packages []string, cfg fuzzCLIConfig) int {
	if len(packages) == 0 {
		packages = []string{"./..."}
	}

	fmt.Println("[blit fuzz] discovering fuzz tests...")

	if cfg.native {
		return runNativeFuzz(packages, cfg)
	}
	return runBtestFuzz(packages, cfg)
}

// fuzzCLIConfig holds flags for the fuzz subcommand.
type fuzzCLIConfig struct {
	// pattern is the test function name pattern (default: TestFuzz).
	pattern string
	// native enables Go's built-in fuzzing (-fuzz flag).
	native bool
	// fuzzTime is the duration for Go-native fuzzing (default: 10s).
	fuzzTime string
	// verbose passes -v to go test.
	verbose bool
	// seed overrides the random seed via -args -btest.fuzz.seed.
	seed int64
	// iterations overrides iteration count via -args -btest.fuzz.iterations.
	iterations int
}

// parseFuzzFlags parses fuzz-specific flags and returns the config plus
// remaining positional arguments (packages).
func parseFuzzFlags(args []string) (fuzzCLIConfig, []string) {
	fs := flag.NewFlagSet("fuzz", flag.ContinueOnError)
	cfg := fuzzCLIConfig{}
	fs.StringVar(&cfg.pattern, "pattern", "TestFuzz", "test function pattern to match")
	fs.BoolVar(&cfg.native, "native", false, "use Go's built-in fuzzing (-fuzz)")
	fs.StringVar(&cfg.fuzzTime, "fuzz-time", "10s", "duration for Go-native fuzzing")
	fs.BoolVar(&cfg.verbose, "v", false, "verbose output")
	fs.Int64Var(&cfg.seed, "seed", 0, "random seed for btest fuzz (0 = random)")
	fs.IntVar(&cfg.iterations, "iterations", 0, "iteration count override (0 = use default)")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: blit fuzz [flags] [packages...]")
		fs.PrintDefaults()
	}
	_ = fs.Parse(args)
	return cfg, fs.Args()
}

// runBtestFuzz runs btest-style fuzz tests (TestFuzz* functions).
func runBtestFuzz(packages []string, cfg fuzzCLIConfig) int {
	args := []string{"test", "-count=1", "-timeout=120s", "-run", cfg.pattern}
	if cfg.verbose {
		args = append(args, "-v")
	}
	args = append(args, packages...)

	// Pass fuzz configuration as test args.
	var testArgs []string
	if cfg.seed != 0 {
		testArgs = append(testArgs, fmt.Sprintf("-btest.fuzz.seed=%d", cfg.seed))
	}
	if cfg.iterations > 0 {
		testArgs = append(testArgs, fmt.Sprintf("-btest.fuzz.iterations=%d", cfg.iterations))
	}
	if len(testArgs) > 0 {
		args = append(args, "-args")
		args = append(args, testArgs...)
	}

	fmt.Printf("[blit fuzz] go %s\n", strings.Join(args, " "))
	start := time.Now()

	cmd := exec.Command("go", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()

	elapsed := time.Since(start).Round(time.Millisecond)
	fmt.Printf("\n[blit fuzz] completed in %s\n", elapsed)

	if err != nil {
		if exit, ok := err.(*exec.ExitError); ok {
			return exit.ExitCode()
		}
		fmt.Fprintf(os.Stderr, "[blit fuzz] error: %v\n", err)
		return 1
	}
	return 0
}

// runNativeFuzz runs Go's built-in fuzzing via `go test -fuzz`.
func runNativeFuzz(packages []string, cfg fuzzCLIConfig) int {
	// Go-native fuzzing only supports a single package.
	if len(packages) > 1 || (len(packages) == 1 && packages[0] == "./...") {
		fmt.Fprintln(os.Stderr, "[blit fuzz] Go-native fuzzing requires a single package (not ./...)")
		fmt.Fprintln(os.Stderr, "  Usage: blit --fuzz -native ./btest/")
		return 1
	}

	fuzzDuration, err := time.ParseDuration(cfg.fuzzTime)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[blit fuzz] invalid -fuzz-time %q: %v\n", cfg.fuzzTime, err)
		return 1
	}

	args := []string{"test", "-fuzz", cfg.pattern, "-fuzztime", fuzzDuration.String()}
	if cfg.verbose {
		args = append(args, "-v")
	}
	args = append(args, packages...)

	fmt.Printf("[blit fuzz] go %s\n", strings.Join(args, " "))
	start := time.Now()

	cmd := exec.Command("go", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	runErr := cmd.Run()

	elapsed := time.Since(start).Round(time.Millisecond)
	fmt.Printf("\n[blit fuzz] completed in %s\n", elapsed)

	if runErr != nil {
		if exit, ok := runErr.(*exec.ExitError); ok {
			return exit.ExitCode()
		}
		fmt.Fprintf(os.Stderr, "[blit fuzz] error: %v\n", runErr)
		return 1
	}
	return 0
}
