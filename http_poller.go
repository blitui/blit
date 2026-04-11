package blit

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// HTTPResourceOpts configures an HTTPResource for API polling with ETag
// caching, rate-limit awareness, and fallback.
type HTTPResourceOpts struct {
	// Name identifies this resource in DevConsole and log messages.
	Name string

	// BuildURL returns the URL to fetch for the given page number (1-indexed).
	BuildURL func(page int) string

	// Parse decodes the response body into an arbitrary result.
	// The returned value is typically wrapped in a tea.Msg by the caller.
	Parse func(body []byte) (any, error)

	// ExtraHeaders returns additional HTTP headers for each request.
	// Use this for authentication tokens.
	ExtraHeaders func() map[string]string

	// Pages is the number of pages to fetch per poll cycle. Default: 1.
	Pages int

	// PageSize is the per_page query parameter. Default: 30.
	PageSize int

	// Parallel fetches pages concurrently when Pages > 1. Default: true.
	Parallel bool

	// CacheResponses enables caching the last successful response per URL
	// and falling back to it on failure. Default: true.
	CacheResponses bool

	// OnRateLimit is called when rate limit info is parsed from response
	// headers. Optional.
	OnRateLimit func(remaining, limit int)
}

// HTTPResource manages API polling with ETag caching, rate-limit parsing,
// and response fallback. It integrates with blit.Poller for timing and
// StatsCollector for observability.
//
// Create one with NewHTTPResource, then call PollCmd to get a tea.Cmd
// suitable for use with Poller.
type HTTPResource struct {
	opts  HTTPResourceOpts
	stats *StatsCollector

	mu     sync.Mutex
	etags  map[string]string // url -> etag
	cache  map[string][]byte // url -> last response body
	client *http.Client
}

// pageResult holds the outcome of a single page fetch.
type pageResult struct {
	page   int
	data   any
	raw    []byte
	notMod bool
	err    error
	rateRL int
	rateLM int
}

// HTTPResourceStats returns the stats snapshot for this resource.
func (r *HTTPResource) Stats() StatsSnapshot {
	if r.stats != nil {
		return r.stats.Snapshot()
	}
	return StatsSnapshot{}
}

// NewHTTPResource creates a new HTTPResource with the given options.
func NewHTTPResource(opts HTTPResourceOpts) *HTTPResource {
	if opts.Pages < 1 {
		opts.Pages = 1
	}
	if opts.PageSize < 1 {
		opts.PageSize = 30
	}
	if opts.CacheResponses {
		// default true
	} else {
		opts.CacheResponses = true // default
	}
	r := &HTTPResource{
		opts:   opts,
		etags:  make(map[string]string),
		cache:  make(map[string][]byte),
		client: &http.Client{Timeout: 30 * time.Second},
	}
	return r
}

// SetStatsCollector wires a StatsCollector so that fetch results are
// automatically tracked. Optional but recommended.
func (r *HTTPResource) SetStatsCollector(sc *StatsCollector) {
	r.stats = sc
}

// PollCmd returns a tea.Cmd that fetches all configured pages and returns
// an HTTPResultMsg. Use this as the Poller's fetch function.
func (r *HTTPResource) PollCmd() Cmd {
	return func() Msg {
		return r.fetch()
	}
}

// fetch executes all page fetches and returns an HTTPResultMsg.
func (r *HTTPResource) fetch() HTTPResultMsg {
	pages := r.opts.Pages

	results := make([]pageResult, pages)

	if r.opts.Parallel && pages > 1 {
		var wg sync.WaitGroup
		for p := 1; p <= pages; p++ {
			wg.Add(1)
			go func(page int) {
				defer wg.Done()
				results[page-1] = r.fetchPage(page)
			}(p)
		}
		wg.Wait()
	} else {
		for p := 1; p <= pages; p++ {
			results[p-1] = r.fetchPage(p)
		}
	}

	msg := HTTPResultMsg{Name: r.opts.Name}
	var lastRL, lastLM int
	for _, res := range results {
		if res.err != nil {
			msg.Errors = append(msg.Errors, fmt.Sprintf("page %d: %v", res.page, res.err))
			continue
		}
		if res.notMod {
			msg.NotModifiedCount++
			continue
		}
		if res.data != nil {
			msg.Results = append(msg.Results, res.data)
		}
		if res.rateLM > 0 {
			lastRL = res.rateRL
			lastLM = res.rateLM
		}
	}

	// Rate limit
	if lastLM > 0 {
		msg.RateRemaining = lastRL
		msg.RateLimit = lastLM
		if r.stats != nil {
			r.stats.SetRateLimit(lastRL, lastLM)
		}
		if r.opts.OnRateLimit != nil {
			r.opts.OnRateLimit(lastRL, lastLM)
		}
	}

	return msg
}

// fetchPage fetches a single page with ETag support and caching.
func (r *HTTPResource) fetchPage(page int) pageResult {
	url := r.opts.BuildURL(page)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return pageResult{page: page, err: err}
	}

	// Extra headers
	if r.opts.ExtraHeaders != nil {
		for k, v := range r.opts.ExtraHeaders() {
			req.Header.Set(k, v)
		}
	}

	// ETag
	r.mu.Lock()
	etag := r.etags[url]
	r.mu.Unlock()
	if etag != "" {
		req.Header.Set("If-None-Match", etag)
	}

	resp, err := r.client.Do(req)
	if err != nil {
		// Fallback to cache
		if r.opts.CacheResponses {
			r.mu.Lock()
			cached, ok := r.cache[url]
			r.mu.Unlock()
			if ok && r.opts.Parse != nil {
				data, parseErr := r.opts.Parse(cached)
				if parseErr == nil {
					if r.stats != nil {
						r.stats.RecordCached(r.opts.Name, 0)
					}
					return pageResult{page: page, data: data}
				}
			}
		}
		if r.stats != nil {
			r.stats.RecordFailure(r.opts.Name, err)
		}
		return pageResult{page: page, err: err}
	}
	defer func() { _ = resp.Body.Close() }()

	// Parse rate limit from headers
	var rateRL, rateLM int
	if v := resp.Header.Get("X-RateLimit-Remaining"); v != "" {
		rateRL, _ = strconv.Atoi(v)
	}
	if v := resp.Header.Get("X-RateLimit-Limit"); v != "" {
		rateLM, _ = strconv.Atoi(v)
	}

	// 304 Not Modified
	if resp.StatusCode == http.StatusNotModified {
		if r.stats != nil {
			r.stats.RecordCached(r.opts.Name, 0)
		}
		return pageResult{page: page, notMod: true, rateRL: rateRL, rateLM: rateLM}
	}

	// Read body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		if r.stats != nil {
			r.stats.RecordFailure(r.opts.Name, err)
		}
		return pageResult{page: page, err: err}
	}

	// Store ETag
	if newEtag := resp.Header.Get("ETag"); newEtag != "" {
		r.mu.Lock()
		r.etags[url] = newEtag
		r.mu.Unlock()
	}

	// Cache response
	if r.opts.CacheResponses {
		r.mu.Lock()
		r.cache[url] = body
		r.mu.Unlock()
	}

	// Parse
	if r.opts.Parse != nil {
		data, parseErr := r.opts.Parse(body)
		if parseErr != nil {
			if r.stats != nil {
				r.stats.RecordFailure(r.opts.Name, parseErr)
			}
			return pageResult{page: page, err: parseErr}
		}
		if r.stats != nil {
			count := 0
			// Try to estimate item count from slice-type results
			if slice, ok := data.([]any); ok {
				count = len(slice)
			}
			r.stats.RecordSuccess(r.opts.Name, count)
		}
		return pageResult{page: page, data: data, rateRL: rateRL, rateLM: rateLM}
	}

	if r.stats != nil {
		r.stats.RecordSuccess(r.opts.Name, 0)
	}
	return pageResult{page: page, raw: body, rateRL: rateRL, rateLM: rateLM}
}

// HTTPResultMsg is the tea.Msg returned by HTTPResource.PollCmd.
type HTTPResultMsg struct {
	Name             string
	Results          []any
	Errors           []string
	NotModifiedCount int
	RateRemaining    int
	RateLimit        int
}

// HasData reports whether any page returned fresh data (not all 304s).
func (m HTTPResultMsg) HasData() bool {
	return len(m.Results) > 0
}

// IsAllNotModified reports whether all pages returned 304 Not Modified.
func (m HTTPResultMsg) IsAllNotModified() bool {
	return m.NotModifiedCount > 0 && len(m.Results) == 0 && len(m.Errors) == 0
}

// --- DevConsole integration ---------------------------------------------------

// DebugProvider returns a DebugDataProvider for this HTTPResource.
func (r *HTTPResource) DebugProvider() DebugDataProvider {
	return &httpResourceDebugProvider{r: r}
}

type httpResourceDebugProvider struct {
	r *HTTPResource
}

func (d *httpResourceDebugProvider) Name() string {
	if d.r.opts.Name != "" {
		return d.r.opts.Name + " HTTP"
	}
	return "HTTP Resource"
}

func (d *httpResourceDebugProvider) View(width, height int, theme Theme) string {
	if d.r.stats != nil {
		return d.r.stats.View(width, height, theme)
	}
	return "  (no stats collector wired)"
}

func (d *httpResourceDebugProvider) Data() map[string]any {
	return d.r.stats.Data()
}

// --- Helpers for common JSON API patterns --------------------------------------

// ParseJSONSlice is a convenience Parse function for APIs that return a
// JSON array of type T.
func ParseJSONSlice[T any]() func(body []byte) (any, error) {
	return func(body []byte) (any, error) {
		var items []T
		if err := json.Unmarshal(body, &items); err != nil {
			return nil, fmt.Errorf("json parse: %w", err)
		}
		return items, nil
	}
}

// ParseJSONObject is a convenience Parse function for APIs that return a
// single JSON object of type T.
func ParseJSONObject[T any]() func(body []byte) (any, error) {
	return func(body []byte) (any, error) {
		var obj T
		if err := json.Unmarshal(body, &obj); err != nil {
			return nil, fmt.Errorf("json parse: %w", err)
		}
		return obj, nil
	}
}

// GitHubAPIURL builds a GitHub API URL for the given endpoint with
// per_page and page parameters. This is the most common use case
// for TUI apps that poll GitHub.
func GitHubAPIURL(endpoint string, perPage, page int) string {
	sep := "?"
	if strings.Contains(endpoint, "?") {
		sep = "&"
	}
	return fmt.Sprintf("https://api.github.com/%s%sper_page=%d&page=%d", endpoint, sep, perPage, page)
}
