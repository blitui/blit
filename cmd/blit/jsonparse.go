package main

import (
	"bufio"
	"encoding/json"
	"io"
	"strings"
	"time"

	"github.com/blitui/blit/btest"
)

// TestEvent represents a single line of go test -json output.
type TestEvent struct {
	Time    time.Time `json:"Time"`
	Action  string    `json:"Action"`
	Package string    `json:"Package"`
	Test    string    `json:"Test"`
	Elapsed float64   `json:"Elapsed"`
	Output  string    `json:"Output"`
}

// parseTestEvents reads go test -json output and returns all events.
func parseTestEvents(r io.Reader) []TestEvent {
	var events []TestEvent
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var ev TestEvent
		if err := json.Unmarshal(line, &ev); err != nil {
			continue
		}
		events = append(events, ev)
	}
	return events
}

// buildReport converts parsed test events into a btest.Report suitable for
// JUnit/HTML output.
func buildReport(events []TestEvent, suite string) *btest.Report {
	report := &btest.Report{
		Suite:     suite,
		StartedAt: time.Now(),
	}

	// Track per-test state.
	type testState struct {
		pkg     string
		output  strings.Builder
		started time.Time
	}
	tests := make(map[string]*testState)

	for _, ev := range events {
		if ev.Test == "" {
			continue
		}
		key := ev.Package + "/" + ev.Test

		switch ev.Action {
		case "run":
			tests[key] = &testState{pkg: ev.Package, started: ev.Time}
		case "output":
			if ts, ok := tests[key]; ok {
				ts.output.WriteString(ev.Output)
			}
		case "pass":
			dur := time.Duration(ev.Elapsed * float64(time.Second))
			report.Results = append(report.Results, btest.TestResult{
				Name:     ev.Test,
				Package:  ev.Package,
				Duration: dur,
				Passed:   true,
			})
			delete(tests, key)
		case "fail":
			dur := time.Duration(ev.Elapsed * float64(time.Second))
			output := ""
			if ts, ok := tests[key]; ok {
				output = ts.output.String()
			}
			report.Results = append(report.Results, btest.TestResult{
				Name:     ev.Test,
				Package:  ev.Package,
				Duration: dur,
				Passed:   false,
				Failure:  strings.TrimSpace(output),
			})
			delete(tests, key)
		case "skip":
			dur := time.Duration(ev.Elapsed * float64(time.Second))
			report.Results = append(report.Results, btest.TestResult{
				Name:     ev.Test,
				Package:  ev.Package,
				Duration: dur,
				Skipped:  true,
			})
			delete(tests, key)
		}
	}

	if len(events) > 0 {
		report.StartedAt = events[0].Time
	}
	return report
}
