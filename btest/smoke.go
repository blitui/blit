package btest

import (
	"fmt"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// SmokeResult holds the outcome of a single smoke test check.
type SmokeResult struct {
	Name   string
	Passed bool
	Detail string
}

// SmokeReport holds the results of a full smoke test run.
type SmokeReport struct {
	Results []SmokeResult
	Passed  int
	Failed  int
}

// SmokeOpts configures the smoke test suite.
type SmokeOpts struct {
	// Sizes to test during resize checks. Defaults to standard sizes if nil.
	Sizes []SmokeSize
	// Keys to test during key dispatch checks. Defaults to common navigation keys if nil.
	Keys []string
}

// SmokeSize defines a terminal size for resize testing.
type SmokeSize struct {
	Cols  int
	Lines int
}

var defaultSizes = []SmokeSize{
	{80, 24},
	{120, 40},
	{40, 10},
}

var defaultKeys = []string{
	"up", "down", "left", "right",
	"enter", "esc", "tab",
	"home", "end",
	"g", "G", "q",
}

// Smoke runs a standard smoke test suite against a tea.Model. It verifies
// that the model can render, handle key events, handle resize events, and
// handle mouse events without panicking. Returns a SmokeReport with results.
//
// Use this for quick validation that a model is wired up correctly.
func Smoke(t testing.TB, model tea.Model, opts *SmokeOpts) *SmokeReport {
	t.Helper()

	if opts == nil {
		opts = &SmokeOpts{}
	}
	sizes := opts.Sizes
	if sizes == nil {
		sizes = defaultSizes
	}
	keys := opts.Keys
	if keys == nil {
		keys = defaultKeys
	}

	report := &SmokeReport{}

	// 1. Init + Render test
	report.add(smokeInitRender(t, model))

	// 2. Resize tests
	for _, sz := range sizes {
		report.add(smokeResize(t, model, sz))
	}

	// 3. Key dispatch tests
	report.add(smokeKeys(t, model, keys))

	// 4. Mouse tests
	report.add(smokeMouse(t, model))

	return report
}

// SmokeTest is a convenience wrapper that calls Smoke and fails the test if
// any check fails.
func SmokeTest(t testing.TB, model tea.Model, opts *SmokeOpts) {
	t.Helper()
	report := Smoke(t, model, opts)
	if report.Failed > 0 {
		for _, r := range report.Results {
			if !r.Passed {
				t.Errorf("FAIL: %s — %s", r.Name, r.Detail)
			}
		}
	}
}

func (r *SmokeReport) add(result SmokeResult) {
	r.Results = append(r.Results, result)
	if result.Passed {
		r.Passed++
	} else {
		r.Failed++
	}
}

func smokeInitRender(t testing.TB, model tea.Model) SmokeResult {
	t.Helper()
	name := "init+render"

	defer func() {
		if r := recover(); r != nil {
			// Will be caught by the result below
		}
	}()

	tm := NewTestModel(t, model, 80, 24)
	s := tm.Screen()
	if s.IsEmpty() {
		return SmokeResult{Name: name, Passed: false, Detail: "View() returned empty screen"}
	}
	return SmokeResult{Name: name, Passed: true, Detail: "rendered successfully"}
}

func smokeResize(t testing.TB, model tea.Model, sz SmokeSize) SmokeResult {
	t.Helper()
	name := fmt.Sprintf("resize-%dx%d", sz.Cols, sz.Lines)

	recovered := false
	func() {
		defer func() {
			if r := recover(); r != nil {
				recovered = true
			}
		}()
		tm := NewTestModel(t, model, 80, 24)
		tm.SendResize(sz.Cols, sz.Lines)
		_ = tm.Screen()
	}()

	if recovered {
		return SmokeResult{Name: name, Passed: false, Detail: "panicked during resize"}
	}
	return SmokeResult{Name: name, Passed: true, Detail: "no panic"}
}

func smokeKeys(t testing.TB, model tea.Model, keys []string) SmokeResult {
	t.Helper()
	name := "key-dispatch"

	recovered := false
	panicKey := ""
	func() {
		defer func() {
			if r := recover(); r != nil {
				recovered = true
			}
		}()
		tm := NewTestModel(t, model, 80, 24)
		for _, k := range keys {
			panicKey = k
			tm.SendKey(k)
		}
		_ = tm.Screen()
	}()

	if recovered {
		return SmokeResult{Name: name, Passed: false, Detail: fmt.Sprintf("panicked on key %q", panicKey)}
	}
	return SmokeResult{Name: name, Passed: true, Detail: fmt.Sprintf("%d keys dispatched", len(keys))}
}

func smokeMouse(t testing.TB, model tea.Model) SmokeResult {
	t.Helper()
	name := "mouse"

	recovered := false
	func() {
		defer func() {
			if r := recover(); r != nil {
				recovered = true
			}
		}()
		tm := NewTestModel(t, model, 80, 24)
		// Click at center
		tm.SendMouse(40, 12, tea.MouseButtonLeft)
		// Scroll down
		tm.SendMouse(40, 12, tea.MouseButtonWheelDown)
		// Scroll up
		tm.SendMouse(40, 12, tea.MouseButtonWheelUp)
		_ = tm.Screen()
	}()

	if recovered {
		return SmokeResult{Name: name, Passed: false, Detail: "panicked during mouse events"}
	}
	return SmokeResult{Name: name, Passed: true, Detail: "click + scroll handled"}
}
