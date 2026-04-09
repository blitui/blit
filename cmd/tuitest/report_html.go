package main

import (
	"fmt"
	"html"
	"os"
	"strings"
	"time"
)

const reportHTMLTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>tuitest Report</title>
<style>
  body { font-family: system-ui, -apple-system, sans-serif; margin: 2rem; background: #0d1117; color: #e6edf3; }
  h1 { color: #58a6ff; margin-bottom: 0.25rem; }
  .meta { color: #8b949e; font-size: 0.875rem; margin-bottom: 2rem; }
  .summary { display: flex; gap: 1.5rem; margin-bottom: 2rem; }
  .stat { background: #161b22; border: 1px solid #30363d; border-radius: 8px; padding: 1rem 1.5rem; text-align: center; min-width: 100px; }
  .stat-label { font-size: 0.75rem; color: #8b949e; text-transform: uppercase; letter-spacing: 0.05em; }
  .stat-value { font-size: 2rem; font-weight: 700; margin-top: 0.25rem; }
  .pass { color: #3fb950; }
  .fail { color: #f85149; }
  .neutral { color: #58a6ff; }
  .sparkline { font-family: monospace; font-size: 1.25rem; letter-spacing: 0.05em; margin-bottom: 2rem; }
  .sparkline-label { font-size: 0.75rem; color: #8b949e; margin-bottom: 0.25rem; }
  table { width: 100%%; border-collapse: collapse; background: #161b22; border-radius: 8px; overflow: hidden; }
  th { background: #21262d; color: #8b949e; font-size: 0.75rem; text-transform: uppercase; letter-spacing: 0.05em; padding: 0.75rem 1rem; text-align: left; }
  td { padding: 0.75rem 1rem; border-top: 1px solid #21262d; font-size: 0.875rem; }
  tr:hover td { background: #1c2128; }
  .badge { display: inline-block; padding: 0.125rem 0.5rem; border-radius: 9999px; font-size: 0.75rem; font-weight: 600; }
  .badge-pass { background: #0d4a1d; color: #3fb950; }
  .badge-fail { background: #4a0d0d; color: #f85149; }
</style>
</head>
<body>
<h1>tuitest Report</h1>
<div class="meta">Generated %s</div>

<div class="summary">
  <div class="stat"><div class="stat-label">Total Runs</div><div class="stat-value neutral">%d</div></div>
  <div class="stat"><div class="stat-label">Total Passed</div><div class="stat-value pass">%d</div></div>
  <div class="stat"><div class="stat-label">Total Failed</div><div class="stat-value fail">%d</div></div>
  <div class="stat"><div class="stat-label">Pass Rate</div><div class="stat-value %s">%s</div></div>
</div>

<div class="sparkline-label">Pass rate — last %d runs (oldest → newest)</div>
<div class="sparkline">%s</div>

<table>
<thead>
<tr>
  <th>Run At</th>
  <th>Status</th>
  <th>Passed</th>
  <th>Failed</th>
  <th>Total</th>
  <th>Duration</th>
</tr>
</thead>
<tbody>
%s
</tbody>
</table>
</body>
</html>
`

// cmdReport generates report.html from history and writes it to outPath.
func cmdReport(outPath string) int {
	if outPath == "" {
		outPath = "report.html"
	}

	records, err := loadHistory()
	if err != nil {
		fmt.Fprintf(os.Stderr, "[tuitest report] error loading history: %v\n", err)
		return 1
	}
	if len(records) == 0 {
		fmt.Fprintln(os.Stderr, "[tuitest report] no history found — run some tests first")
		return 1
	}

	// Aggregate totals.
	totalPassed, totalFailed := 0, 0
	for _, r := range records {
		totalPassed += r.Passed
		totalFailed += r.Failed
	}
	totalTests := totalPassed + totalFailed
	passRateStr := "N/A"
	passRateClass := "neutral"
	if totalTests > 0 {
		rate := float64(totalPassed) / float64(totalTests) * 100
		passRateStr = fmt.Sprintf("%.1f%%", rate)
		if rate >= 80 {
			passRateClass = "pass"
		} else {
			passRateClass = "fail"
		}
	}

	spark := sparkline(records, sparklineWidth)

	// Build table rows (newest first).
	var rows strings.Builder
	for _, r := range records {
		status := "pass"
		badgeClass := "badge-pass"
		if r.Failed > 0 {
			status = "fail"
			badgeClass = "badge-fail"
		}
		rows.WriteString(fmt.Sprintf(
			"<tr><td>%s</td><td><span class=\"badge %s\">%s</span></td><td>%d</td><td>%d</td><td>%d</td><td>%.2fs</td></tr>\n",
			html.EscapeString(r.RunAt.Format("2006-01-02 15:04:05")),
			badgeClass, status,
			r.Passed, r.Failed, r.Total,
			r.Duration,
		))
	}

	content := fmt.Sprintf(reportHTMLTemplate,
		html.EscapeString(time.Now().Format("2006-01-02 15:04:05")),
		len(records),
		totalPassed, totalFailed,
		passRateClass, html.EscapeString(passRateStr),
		min(sparklineWidth, len(records)),
		html.EscapeString(spark),
		rows.String(),
	)

	if err := os.WriteFile(outPath, []byte(content), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "[tuitest report] write failed: %v\n", err)
		return 1
	}
	fmt.Printf("[tuitest report] wrote %s (%d runs)\n", outPath, len(records))
	return 0
}
