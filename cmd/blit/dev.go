package main

import (
	"fmt"
	"io/fs"
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

// configExts are file extensions that trigger a config/theme hot-reload
// notification instead of a full rebuild.
var configExts = map[string]bool{
	".yaml": true,
	".yml":  true,
	".json": true,
	".toml": true,
}

// runDev implements the `blit dev` subcommand.
// It builds the target package, runs the binary, and restarts on file changes.
// Config/theme file changes (.yaml, .yml, .json, .toml) are detected but do
// not trigger a rebuild — the app's built-in watchers (WithThemeHotReload,
// Config.WatchFile) handle those at runtime.
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
	fmt.Println("[blit dev] .go changes rebuild+restart | config/theme changes hot-reload in-app")

	// Trap Ctrl+C to clean up.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	var proc *exec.Cmd
	lastGoHash := snapshotTree(".")
	lastCfgHash := snapshotConfigTree(".")

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
			goHash := snapshotTree(".")
			cfgHash := snapshotConfigTree(".")

			goChanged := goHash != lastGoHash
			cfgChanged := cfgHash != lastCfgHash

			if !goChanged && !cfgChanged {
				continue
			}

			lastCfgHash = cfgHash

			// Config/theme files changed but no Go files — no restart needed.
			if cfgChanged && !goChanged {
				fmt.Println("[blit dev] config/theme file changed (hot-reload handled by app)")
				continue
			}

			lastGoHash = goHash

			// Debounce: wait 100ms and re-check.
			time.Sleep(100 * time.Millisecond)
			h2 := snapshotTree(".")
			if h2 != goHash {
				lastGoHash = h2
			}

			fmt.Println("[blit dev] .go change detected, rebuilding...")
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

// snapshotConfigTree returns a coarse hash of config/theme file modification
// times under root. Used to detect changes to .yaml, .yml, .json, and .toml
// files without triggering a full rebuild.
func snapshotConfigTree(root string) string {
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
		ext := strings.ToLower(filepath.Ext(path))
		if !configExts[ext] {
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
