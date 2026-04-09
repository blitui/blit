package tuitest

import (
	"encoding/xml"
	"fmt"
	"html"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// JUnitReporter registers a t.Cleanup hook that writes the report as JUnit
// XML to path when the test (or parent TestMain) finishes. The caller owns
// the Report and is responsible for populating Results before teardown.
func JUnitReporter(t testing.TB, report *Report, path string) {
	t.Helper()
	t.Cleanup(func() {
		if err := report.WriteJUnit(path); err != nil {
			t.Errorf("JUnitReporter: %v", err)
		}
	})
}

// HTMLReporter registers a t.Cleanup hook that writes the report as HTML
// to path when the test finishes.
func HTMLReporter(t testing.TB, report *Report, path string) {
	t.Helper()
	t.Cleanup(func() {
		if err := report.WriteHTML(path); err != nil {
			t.Errorf("HTMLReporter: %v", err)
		}
	})
}

// TestResult describes one test outcome used by the reporters.
type TestResult struct {
	Name     string
	Package  string
	Duration time.Duration
	Passed   bool
	Failure  string
	Skipped  bool
	Before   string // optional screen before failure
	After    string // optional screen after failure
}

// Report is a collection of test results plus suite metadata.
type Report struct {
	Suite     string
	StartedAt time.Time
	Results   []TestResult
}

// Totals returns (total, passed, failed, skipped).
func (r *Report) Totals() (total, passed, failed, skipped int) {
	for _, res := range r.Results {
		total++
		switch {
		case res.Skipped:
			skipped++
		case res.Passed:
			passed++
		default:
			failed++
		}
	}
	return
}

// TotalDuration sums the duration of all results.
func (r *Report) TotalDuration() time.Duration {
	var d time.Duration
	for _, res := range r.Results {
		d += res.Duration
	}
	return d
}

// --- JUnit XML ----------------------------------------------------------

type junitTestSuite struct {
	XMLName   xml.Name        `xml:"testsuite"`
	Name      string          `xml:"name,attr"`
	Tests     int             `xml:"tests,attr"`
	Failures  int             `xml:"failures,attr"`
	Skipped   int             `xml:"skipped,attr"`
	Time      string          `xml:"time,attr"`
	Timestamp string          `xml:"timestamp,attr"`
	Cases     []junitTestCase `xml:"testcase"`
}

type junitTestCase struct {
	XMLName   xml.Name      `xml:"testcase"`
	Name      string        `xml:"name,attr"`
	Classname string        `xml:"classname,attr"`
	Time      string        `xml:"time,attr"`
	Failure   *junitFailure `xml:"failure,omitempty"`
	Skipped   *junitSkipped `xml:"skipped,omitempty"`
}

type junitFailure struct {
	Message string `xml:"message,attr"`
	Body    string `xml:",chardata"`
}

type junitSkipped struct{}

// WriteJUnit writes the report as a JUnit XML file at path. Parent
// directories are created as needed.
func (r *Report) WriteJUnit(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", filepath.Dir(path), err)
	}
	total, _, failures, skipped := r.Totals()
	suite := junitTestSuite{
		Name:      r.Suite,
		Tests:     total,
		Failures:  failures,
		Skipped:   skipped,
		Time:      fmt.Sprintf("%.3f", r.TotalDuration().Seconds()),
		Timestamp: r.StartedAt.UTC().Format(time.RFC3339),
	}
	for _, res := range r.Results {
		tc := junitTestCase{
			Name:      res.Name,
			Classname: res.Package,
			Time:      fmt.Sprintf("%.3f", res.Duration.Seconds()),
		}
		if res.Skipped {
			tc.Skipped = &junitSkipped{}
		} else if !res.Passed {
			tc.Failure = &junitFailure{Message: res.Failure, Body: res.Failure}
		}
		suite.Cases = append(suite.Cases, tc)
	}
	data, err := xml.MarshalIndent(suite, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal junit: %w", err)
	}
	// Prepend the XML header.
	out := append([]byte(xml.Header), data...)
	return os.WriteFile(path, out, 0o644)
}

// --- HTML -------------------------------------------------------------

// WriteHTML writes the report as a self-contained HTML page.
func (r *Report) WriteHTML(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", filepath.Dir(path), err)
	}
	total, passed, failures, skipped := r.Totals()
	var b strings.Builder
	fmt.Fprintf(&b, `<!doctype html>
<html><head><meta charset="utf-8"><title>tuitest report — %s</title>
<style>
body { font: 14px/1.4 -apple-system,BlinkMacSystemFont,sans-serif; margin: 2em; color: #222; }
h1 { margin-top: 0; }
.summary { display: flex; gap: 1em; margin-bottom: 1em; }
.summary div { padding: .5em 1em; border-radius: 6px; }
.pass { background: #d9f5d3; }
.fail { background: #f7d1d1; }
.skip { background: #eee; }
table { border-collapse: collapse; width: 100%%; }
th, td { text-align: left; padding: .4em .6em; border-bottom: 1px solid #eee; }
tr.failed td { background: #fdecec; }
tr.skipped td { background: #f5f5f5; }
pre { background: #111; color: #e0e0e0; padding: .8em; border-radius: 6px; overflow-x: auto; }
details { margin-top: .4em; }
</style></head><body>
<h1>%s</h1>
<p>Started %s — duration %.3fs</p>
<div class="summary">
  <div class="pass">%d passed</div>
  <div class="fail">%d failed</div>
  <div class="skip">%d skipped</div>
  <div>%d total</div>
</div>
<table><thead><tr><th>Test</th><th>Package</th><th>Duration</th><th>Result</th></tr></thead><tbody>
`, html.EscapeString(r.Suite), html.EscapeString(r.Suite),
		r.StartedAt.UTC().Format(time.RFC3339),
		r.TotalDuration().Seconds(),
		passed, failures, skipped, total,
	)
	for _, res := range r.Results {
		row := ""
		result := "pass"
		if res.Skipped {
			row = "skipped"
			result = "skipped"
		} else if !res.Passed {
			row = "failed"
			result = "fail"
		}
		fmt.Fprintf(&b, `<tr class="%s"><td>%s</td><td>%s</td><td>%.3fs</td><td>%s</td></tr>`,
			row, html.EscapeString(res.Name), html.EscapeString(res.Package),
			res.Duration.Seconds(), result,
		)
		if !res.Passed && !res.Skipped {
			fmt.Fprintf(&b, `<tr><td colspan="4"><details open><summary>Failure</summary><pre>%s</pre>`,
				html.EscapeString(res.Failure))
			if res.Before != "" {
				fmt.Fprintf(&b, `<strong>Before:</strong><pre>%s</pre>`, html.EscapeString(res.Before))
			}
			if res.After != "" {
				fmt.Fprintf(&b, `<strong>After:</strong><pre>%s</pre>`, html.EscapeString(res.After))
			}
			b.WriteString(`</details></td></tr>`)
		}
	}
	b.WriteString(`</tbody></table></body></html>`)
	return os.WriteFile(path, []byte(b.String()), 0o644)
}
