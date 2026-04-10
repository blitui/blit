package btest_test

import (
	"encoding/xml"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/blitui/blit/btest"
)

func sampleReport() *btest.Report {
	return &btest.Report{
		Suite:     "sample",
		StartedAt: time.Date(2026, 4, 9, 12, 0, 0, 0, time.UTC),
		Results: []btest.TestResult{
			{Name: "TestA", Package: "pkg", Duration: 12 * time.Millisecond, Passed: true},
			{Name: "TestB", Package: "pkg", Duration: 34 * time.Millisecond, Passed: false, Failure: "expected X got Y"},
			{Name: "TestC", Package: "pkg", Duration: 5 * time.Millisecond, Skipped: true},
		},
	}
}

func TestReport_Totals(t *testing.T) {
	r := sampleReport()
	total, passed, failed, skipped := r.Totals()
	if total != 3 || passed != 1 || failed != 1 || skipped != 1 {
		t.Errorf("totals = (%d,%d,%d,%d)", total, passed, failed, skipped)
	}
}

func TestReport_JUnit(t *testing.T) {
	r := sampleReport()
	path := filepath.Join(t.TempDir(), "junit.xml")
	if err := r.WriteJUnit(path); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	// Parse with stdlib to verify it is well-formed XML.
	var suite struct {
		XMLName  xml.Name `xml:"testsuite"`
		Name     string   `xml:"name,attr"`
		Tests    int      `xml:"tests,attr"`
		Failures int      `xml:"failures,attr"`
		Skipped  int      `xml:"skipped,attr"`
	}
	if err := xml.Unmarshal(data, &suite); err != nil {
		t.Fatalf("junit parse: %v", err)
	}
	if suite.Tests != 3 || suite.Failures != 1 || suite.Skipped != 1 {
		t.Errorf("junit totals wrong: %+v", suite)
	}
	if !strings.Contains(string(data), "expected X got Y") {
		t.Error("failure body missing from junit")
	}
}

func TestReport_HTML(t *testing.T) {
	r := sampleReport()
	path := filepath.Join(t.TempDir(), "report.html")
	if err := r.WriteHTML(path); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	s := string(data)
	for _, want := range []string{"<!doctype html>", "1 passed", "1 failed", "1 skipped", "expected X got Y"} {
		if !strings.Contains(s, want) {
			t.Errorf("html missing %q", want)
		}
	}
}

func TestReport_WriteJUnit_CreatesDirectory(t *testing.T) {
	r := sampleReport()
	path := filepath.Join(t.TempDir(), "nested", "dir", "junit.xml")
	if err := r.WriteJUnit(path); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Error(err)
	}
}
