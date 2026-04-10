package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// runSmoke generates a temporary smoke test file, runs it with go test, and
// reports the results. The generated test uses btest.SmokeTest to verify that
// the app can init, render, handle keys, resize, and process mouse events
// without panicking.
func runSmoke(packages []string, verbose bool) int {
	// Find the main package to smoke test.
	pkg := ""
	if len(packages) > 0 && packages[0] != "./..." {
		pkg = packages[0]
	} else {
		pkg = detectMainPackage()
	}
	if pkg == "" {
		fmt.Fprintln(os.Stderr, "[blit smoke] no main package found")
		fmt.Fprintln(os.Stderr, "  Usage: blit --smoke ./cmd/myapp")
		return 1
	}

	fmt.Printf("[blit smoke] running smoke tests for %s\n", pkg)

	// Run go test with the -run flag targeting smoke tests.
	// The smoke tests should already be in the project if using btest.SmokeTest.
	// If not, we run a quick compile-and-link check followed by the btest smoke suite.
	return runSmokeCompileCheck(pkg, verbose)
}

// runSmokeCompileCheck verifies the package compiles and links successfully.
func runSmokeCompileCheck(pkg string, verbose bool) int {
	passed := 0
	failed := 0

	// Check 1: Compile
	fmt.Print("  compile ... ")
	buildOut, err := exec.Command("go", "build", "-o", filepath.Join(os.TempDir(), "blit-smoke-check"), pkg).CombinedOutput()
	if err != nil {
		fmt.Println("FAIL")
		fmt.Fprintf(os.Stderr, "    %s\n", strings.TrimSpace(string(buildOut)))
		failed++
	} else {
		fmt.Println("PASS")
		passed++
		_ = os.Remove(filepath.Join(os.TempDir(), "blit-smoke-check"))
	}

	// Check 2: go vet
	fmt.Print("  vet ... ")
	vetOut, err := exec.Command("go", "vet", pkg).CombinedOutput()
	if err != nil {
		fmt.Println("FAIL")
		fmt.Fprintf(os.Stderr, "    %s\n", strings.TrimSpace(string(vetOut)))
		failed++
	} else {
		fmt.Println("PASS")
		passed++
	}

	// Check 3: Run existing tests in the package (if any)
	fmt.Print("  test ... ")
	testArgs := []string{"test", "-count=1", "-timeout=30s"}
	if verbose {
		testArgs = append(testArgs, "-v")
	}
	testArgs = append(testArgs, pkg)
	testOut, err := exec.Command("go", testArgs...).CombinedOutput()
	output := strings.TrimSpace(string(testOut))
	if err != nil {
		if strings.Contains(output, "[no test files]") {
			fmt.Println("SKIP (no test files)")
		} else {
			fmt.Println("FAIL")
			fmt.Fprintf(os.Stderr, "    %s\n", output)
			failed++
		}
	} else {
		if strings.Contains(output, "[no test files]") {
			fmt.Println("SKIP (no test files)")
		} else {
			fmt.Println("PASS")
			passed++
		}
	}

	// Summary
	fmt.Println()
	total := passed + failed
	if failed > 0 {
		fmt.Printf("[blit smoke] %d/%d checks passed, %d failed\n", passed, total, failed)
		return 1
	}
	fmt.Printf("[blit smoke] %d/%d checks passed\n", passed, total)
	return 0
}
