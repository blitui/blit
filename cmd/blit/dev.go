package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

// devOpts holds options for the dev server.
type devOpts struct {
	pkg     string // package to build (default: auto-detect)
	binName string // output binary name
}

// runDev implements the `blit dev` subcommand.
// It builds the target package, runs the binary, and restarts on file changes.
func runDev(args []string) int {
	opts := devOpts{}

	// Parse target package from args or auto-detect.
	if len(args) > 0 {
		opts.pkg = args[0]
	} else {
		opts.pkg = detectMainPackage()
		if opts.pkg == "" {
			fmt.Fprintln(os.Stderr, "[blit dev] no main package found. Usage: blit dev [package]")
			fmt.Fprintln(os.Stderr, "  e.g. blit dev ./cmd/myapp")
			return 1
		}
	}

	opts.binName = filepath.Join(os.TempDir(), "blit-dev-"+filepath.Base(opts.pkg))

	fmt.Printf("[blit dev] watching %s\n", opts.pkg)

	// Trap Ctrl+C to clean up.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	var proc *exec.Cmd
	lastHash := snapshotTree(".")

	// Initial build + run.
	var err error
	proc, err = buildAndRun(opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[blit dev] build failed:\n%s\n", err)
	}

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-sigCh:
			fmt.Println("\n[blit dev] shutting down")
			stopProcess(proc)
			_ = os.Remove(opts.binName)
			return 0

		case <-ticker.C:
			h := snapshotTree(".")
			if h == lastHash {
				continue
			}
			lastHash = h

			// Debounce: wait 100ms and re-check.
			time.Sleep(100 * time.Millisecond)
			h2 := snapshotTree(".")
			if h2 != h {
				lastHash = h2
			}

			fmt.Println("[blit dev] change detected, rebuilding...")
			stopProcess(proc)

			proc, err = buildAndRun(opts)
			if err != nil {
				fmt.Fprintf(os.Stderr, "[blit dev] build failed:\n%s\n", err)
				fmt.Fprintln(os.Stderr, "[blit dev] waiting for changes...")
				proc = nil
			}
		}
	}
}

// buildAndRun compiles the package and starts the resulting binary.
// Returns the running process or a build error.
func buildAndRun(opts devOpts) (*exec.Cmd, error) {
	start := time.Now()

	// Build
	build := exec.Command("go", "build", "-o", opts.binName, opts.pkg)
	build.Env = append(os.Environ(), "CGO_ENABLED=0")
	out, err := build.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("%s", strings.TrimSpace(string(out)))
	}

	elapsed := time.Since(start)
	fmt.Printf("[blit dev] built in %s\n", elapsed.Round(time.Millisecond))

	// Run
	proc := exec.Command(opts.binName)
	proc.Stdout = os.Stdout
	proc.Stderr = os.Stderr
	proc.Stdin = os.Stdin
	if err := proc.Start(); err != nil {
		return nil, fmt.Errorf("start: %w", err)
	}

	fmt.Printf("[blit dev] running %s (pid %d)\n", opts.pkg, proc.Process.Pid)

	// Wait in background so we can detect when the process exits on its own.
	go func() {
		_ = proc.Wait()
	}()

	return proc, nil
}

// stopProcess terminates a running process gracefully.
func stopProcess(proc *exec.Cmd) {
	if proc == nil || proc.Process == nil {
		return
	}
	// Try interrupt first, then kill after timeout.
	_ = proc.Process.Signal(os.Interrupt)
	done := make(chan struct{})
	go func() {
		_ = proc.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		_ = proc.Process.Kill()
		<-done
	}
}

// detectMainPackage looks for a cmd/ directory with a main package.
func detectMainPackage() string {
	entries, err := os.ReadDir("cmd")
	if err != nil {
		// No cmd/ directory — check if current dir has a main.go.
		if _, err := os.Stat("main.go"); err == nil {
			return "."
		}
		return ""
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		mainPath := filepath.Join("cmd", e.Name(), "main.go")
		if _, err := os.Stat(mainPath); err == nil {
			return "./" + filepath.ToSlash(filepath.Join("cmd", e.Name()))
		}
	}
	return ""
}
