package main

import "testing"

func TestCheckGo(t *testing.T) {
	detail, ok := checkGo()
	if !ok {
		t.Skipf("go not available: %s", detail)
	}
	if detail == "" {
		t.Error("expected non-empty detail")
	}
}

func TestCheckGoMod(t *testing.T) {
	_, _ = checkGoMod()
}

func TestCheckGit(t *testing.T) {
	detail, ok := checkGit()
	if !ok {
		t.Skipf("git not available: %s", detail)
	}
	if detail == "" {
		t.Error("expected non-empty detail")
	}
}

func TestCheckTerminal(t *testing.T) {
	detail, _ := checkTerminal()
	if detail == "" {
		t.Error("expected non-empty detail")
	}
}

func TestRunDoctor(t *testing.T) {
	_ = runDoctor(nil)
}
