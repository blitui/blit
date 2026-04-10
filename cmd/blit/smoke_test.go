package main

import "testing"

func TestRunSmokeCompileCheck(t *testing.T) {
	// Smoke-check a lightweight internal package (not "." which would
	// recursively run all cmd/blit tests and timeout).
	code := runSmokeCompileCheck("../../internal/fuzzy", false)
	if code != 0 {
		t.Errorf("expected exit 0 for internal/fuzzy package, got %d", code)
	}
}
