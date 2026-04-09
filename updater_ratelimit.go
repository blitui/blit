package tuikit

import (
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

// RateLimitError is returned when GitHub's rate limiter rejects a request.
// ResetAt is the wall-clock time at which the rate limit window resets
// (parsed from the X-RateLimit-Reset header); zero if unknown.
type RateLimitError struct {
	StatusCode int
	ResetAt    time.Time
	Body       string
}

func (e *RateLimitError) Error() string {
	if !e.ResetAt.IsZero() {
		return fmt.Sprintf("github rate-limited (status %d, resets %s)",
			e.StatusCode, e.ResetAt.Format(time.RFC3339))
	}
	return fmt.Sprintf("github rate-limited (status %d)", e.StatusCode)
}

// IsRateLimited reports whether err is a RateLimitError.
func IsRateLimited(err error) bool {
	_, ok := err.(*RateLimitError)
	return ok
}

// parseRateLimitReset returns the wall-clock time from the X-RateLimit-Reset
// header, or zero time if absent/unparseable.
func parseRateLimitReset(h http.Header) time.Time {
	v := h.Get("X-RateLimit-Reset")
	if v == "" {
		return time.Time{}
	}
	secs, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return time.Time{}
	}
	return time.Unix(secs, 0)
}

// backoffDuration computes the next sleep for attempt n (zero-indexed).
// Base is 500ms, cap is 30s. Jitter is a random value in [0, base*2^n/2].
func backoffDuration(attempt int) time.Duration {
	const base = 500 * time.Millisecond
	const cap_ = 30 * time.Second
	d := base << attempt
	if d > cap_ {
		d = cap_
	}
	jitter := time.Duration(rand.Int63n(int64(d/2) + 1))
	return d/2 + jitter
}

// fetchWithBackoff issues GET requests with exponential-backoff retries on
// 403 (rate-limited) and 5xx responses. On 403 with X-RateLimit-Reset, it
// sleeps until the reset time (capped at 30s) before retrying. Honors ctx
// via the client's Timeout setting on each attempt.
func fetchWithBackoff(client *http.Client, url string, maxAttempts int) (*http.Response, error) {
	if maxAttempts < 1 {
		maxAttempts = 1
	}
	var lastErr error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		resp, err := client.Get(url)
		if err != nil {
			lastErr = err
			time.Sleep(backoffDuration(attempt))
			continue
		}
		// Rate-limited (GitHub uses 403 with X-RateLimit-Remaining: 0).
		if resp.StatusCode == http.StatusForbidden && resp.Header.Get("X-RateLimit-Remaining") == "0" {
			reset := parseRateLimitReset(resp.Header)
			resp.Body.Close()
			lastErr = &RateLimitError{StatusCode: resp.StatusCode, ResetAt: reset}
			wait := backoffDuration(attempt)
			if !reset.IsZero() {
				until := time.Until(reset)
				if until > 0 && until < 30*time.Second {
					wait = until
				}
			}
			time.Sleep(wait)
			continue
		}
		if resp.StatusCode >= 500 && resp.StatusCode < 600 {
			resp.Body.Close()
			lastErr = fmt.Errorf("server error: %d", resp.StatusCode)
			time.Sleep(backoffDuration(attempt))
			continue
		}
		return resp, nil
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("fetchWithBackoff: exhausted %d attempts", maxAttempts)
	}
	return nil, lastErr
}
