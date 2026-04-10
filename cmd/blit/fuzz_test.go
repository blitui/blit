package main

import (
	"testing"
)

func TestParseFuzzFlags_Defaults(t *testing.T) {
	cfg, pkgs := parseFuzzFlags([]string{})
	if cfg.pattern != "TestFuzz" {
		t.Errorf("pattern = %q, want TestFuzz", cfg.pattern)
	}
	if cfg.native {
		t.Error("native should default to false")
	}
	if cfg.fuzzTime != "10s" {
		t.Errorf("fuzzTime = %q, want 10s", cfg.fuzzTime)
	}
	if cfg.verbose {
		t.Error("verbose should default to false")
	}
	if cfg.seed != 0 {
		t.Errorf("seed = %d, want 0", cfg.seed)
	}
	if cfg.iterations != 0 {
		t.Errorf("iterations = %d, want 0", cfg.iterations)
	}
	if len(pkgs) != 0 {
		t.Errorf("packages = %v, want empty", pkgs)
	}
}

func TestParseFuzzFlags_WithArgs(t *testing.T) {
	cfg, pkgs := parseFuzzFlags([]string{
		"-pattern", "TestMyFuzz",
		"-native",
		"-fuzz-time", "30s",
		"-v",
		"-seed", "42",
		"-iterations", "5000",
		"./btest/...",
	})
	if cfg.pattern != "TestMyFuzz" {
		t.Errorf("pattern = %q, want TestMyFuzz", cfg.pattern)
	}
	if !cfg.native {
		t.Error("native should be true")
	}
	if cfg.fuzzTime != "30s" {
		t.Errorf("fuzzTime = %q, want 30s", cfg.fuzzTime)
	}
	if !cfg.verbose {
		t.Error("verbose should be true")
	}
	if cfg.seed != 42 {
		t.Errorf("seed = %d, want 42", cfg.seed)
	}
	if cfg.iterations != 5000 {
		t.Errorf("iterations = %d, want 5000", cfg.iterations)
	}
	if len(pkgs) != 1 || pkgs[0] != "./btest/..." {
		t.Errorf("packages = %v, want [./btest/...]", pkgs)
	}
}

func TestParseFuzzFlags_PackagesOnly(t *testing.T) {
	cfg, pkgs := parseFuzzFlags([]string{"./cmd/blit/", "./btest/"})
	if cfg.pattern != "TestFuzz" {
		t.Errorf("pattern = %q, want TestFuzz", cfg.pattern)
	}
	if len(pkgs) != 2 {
		t.Errorf("packages = %v, want 2 entries", pkgs)
	}
}
