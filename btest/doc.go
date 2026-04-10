// Package btest provides a test toolkit for TUI applications built with
// Bubble Tea. It offers a virtual terminal, assertion helpers, snapshot
// testing, session recording, fuzz testing, mutation testing, and
// accessibility checks — everything needed to thoroughly test interactive
// terminal programs without a real TTY.
//
// # Quick Start
//
// Create a [TestModel] to drive your tea.Model synchronously:
//
//	tm := btest.NewTestModel(t, myModel, 80, 24)
//	tm.SendKey("enter")
//	btest.AssertContains(t, tm.Screen(), "Welcome")
//
// # Core Concepts
//
//   - [TestModel] wraps a tea.Model and processes messages synchronously.
//     Use SendKey, Type, SendResize, and SendMsg to drive the model.
//   - [Screen] is a virtual terminal that decodes ANSI output into a grid
//     of cells with text and style attributes.
//   - Assertion functions (AssertContains, AssertRowEquals, AssertStyleAt, etc.)
//     check screen content and fail the test with descriptive messages.
//
// # Snapshot Testing
//
// [AssertSnapshot] and [AssertGolden] compare screen content against stored
// files. On first run (or with -btest.update), the file is created. On
// subsequent runs, mismatches fail the test with a diff.
//
//	btest.AssertSnapshot(t, tm.Screen(), "login-form")
//
// # Session Recording
//
// [SessionRecorder] captures a sequence of inputs and screens that can be
// saved to .tuisess files and replayed with [Replay] for regression testing.
//
//	rec := btest.NewSessionRecorder(tm)
//	rec.Key("down").Key("enter").Type("hello")
//	rec.Save("testdata/sessions/flow.tuisess")
//
// # Fuzz Testing
//
// [Fuzz] and [FuzzAndFail] send random inputs to a model to find panics.
// Configure weights for different event types via [FuzzConfig].
//
//	btest.FuzzAndFail(t, myModel, btest.FuzzConfig{
//	    Iterations: 1000,
//	    Seed:       42,
//	})
//
// # Mutation Testing
//
// [MutationTest] verifies test suite quality by injecting behavioral mutations
// into a model and checking that tests detect them.
//
//	report := btest.MutationTest(t, factory, btest.MutationConfig{
//	    Test: func(t testing.TB, m tea.Model) { /* assertions */ },
//	})
//	btest.AssertMutationScore(t, report, 80.0)
//
// # Accessibility
//
// Contrast checking ([CheckContrast], [AssertContrast]), keyboard navigation
// verification ([CheckKeyNav], [AssertKeyNavCycle]), and color blindness
// simulation ([SimulateColorBlind], [AssertDistinguishable]) help ensure
// your TUI is accessible.
//
// # Smoke Testing
//
// [SmokeTest] runs a battery of automated checks (init/render, key handling,
// resize, mouse) against any tea.Model to catch common issues.
//
//	btest.SmokeTest(t, myModel, nil)
//
// # Harness
//
// [Harness] provides a fluent API for concise test scripts:
//
//	btest.NewHarness(t, model, 80, 24).
//	    Keys("down", "down", "enter").
//	    Expect("Selected").
//	    Done()
package btest
