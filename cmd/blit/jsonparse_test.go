package main

import (
	"strings"
	"testing"
)

func TestParseTestEvents(t *testing.T) {
	input := `{"Time":"2026-04-10T12:00:00Z","Action":"run","Package":"example.com/foo","Test":"TestHello"}
{"Time":"2026-04-10T12:00:00Z","Action":"output","Package":"example.com/foo","Test":"TestHello","Output":"=== RUN   TestHello\n"}
{"Time":"2026-04-10T12:00:01Z","Action":"output","Package":"example.com/foo","Test":"TestHello","Output":"--- PASS: TestHello (0.50s)\n"}
{"Time":"2026-04-10T12:00:01Z","Action":"pass","Package":"example.com/foo","Test":"TestHello","Elapsed":0.5}
{"Time":"2026-04-10T12:00:01Z","Action":"run","Package":"example.com/foo","Test":"TestFail"}
{"Time":"2026-04-10T12:00:01Z","Action":"output","Package":"example.com/foo","Test":"TestFail","Output":"    fail_test.go:10: expected 1, got 2\n"}
{"Time":"2026-04-10T12:00:02Z","Action":"fail","Package":"example.com/foo","Test":"TestFail","Elapsed":0.3}
{"Time":"2026-04-10T12:00:02Z","Action":"run","Package":"example.com/foo","Test":"TestSkip"}
{"Time":"2026-04-10T12:00:02Z","Action":"skip","Package":"example.com/foo","Test":"TestSkip","Elapsed":0.0}
`
	events := parseTestEvents(strings.NewReader(input))
	if len(events) != 9 {
		t.Fatalf("expected 9 events, got %d", len(events))
	}

	report := buildReport(events, "example.com/foo")
	total, passed, failed, skipped := report.Totals()
	if total != 3 {
		t.Errorf("expected 3 total, got %d", total)
	}
	if passed != 1 {
		t.Errorf("expected 1 passed, got %d", passed)
	}
	if failed != 1 {
		t.Errorf("expected 1 failed, got %d", failed)
	}
	if skipped != 1 {
		t.Errorf("expected 1 skipped, got %d", skipped)
	}

	// Check the failed test has output captured.
	for _, r := range report.Results {
		if r.Name == "TestFail" {
			if !strings.Contains(r.Failure, "expected 1, got 2") {
				t.Errorf("expected failure output, got %q", r.Failure)
			}
		}
	}
}

func TestParseTestEvents_Empty(t *testing.T) {
	events := parseTestEvents(strings.NewReader(""))
	if len(events) != 0 {
		t.Errorf("expected 0 events, got %d", len(events))
	}

	report := buildReport(events, "empty")
	total, _, _, _ := report.Totals()
	if total != 0 {
		t.Errorf("expected 0 total, got %d", total)
	}
}

func TestParseTestEvents_InvalidJSON(t *testing.T) {
	input := "not json\n{\"Action\":\"run\",\"Package\":\"p\",\"Test\":\"T\"}\ngarbage\n"
	events := parseTestEvents(strings.NewReader(input))
	if len(events) != 1 {
		t.Errorf("expected 1 valid event, got %d", len(events))
	}
}
