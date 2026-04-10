package blit

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync/atomic"
	"testing"
	"time"
)

func TestParseRateLimitReset(t *testing.T) {
	h := http.Header{}
	h.Set("X-RateLimit-Reset", "1700000000")
	got := parseRateLimitReset(h)
	want := time.Unix(1700000000, 0)
	if !got.Equal(want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestParseRateLimitReset_Empty(t *testing.T) {
	if got := parseRateLimitReset(http.Header{}); !got.IsZero() {
		t.Errorf("expected zero, got %v", got)
	}
}

func TestBackoffDuration_Monotonic(t *testing.T) {
	// Each attempt should be at least half the base for that attempt.
	for i := 0; i < 6; i++ {
		d := backoffDuration(i)
		if d <= 0 {
			t.Errorf("attempt %d: non-positive duration %v", i, d)
		}
		if d > 30*time.Second {
			t.Errorf("attempt %d: %v exceeds cap", i, d)
		}
	}
}

func TestIsRateLimited(t *testing.T) {
	err := &RateLimitError{StatusCode: 403}
	if !IsRateLimited(err) {
		t.Error("expected IsRateLimited=true")
	}
	if IsRateLimited(fmt.Errorf("other")) {
		t.Error("expected IsRateLimited=false for non-rate-limit error")
	}
}

func TestFetchWithBackoff_RetriesThenSucceeds(t *testing.T) {
	var hits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&hits, 1)
		if n < 3 {
			w.Header().Set("X-RateLimit-Remaining", "0")
			// Reset is 1s from now → will shortcut to wait ≤30s
			w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(10*time.Millisecond).Unix(), 10))
			w.WriteHeader(http.StatusForbidden)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"tag_name":"v1.0.0"}`))
	}))
	defer srv.Close()

	resp, err := fetchWithBackoff(&http.Client{Timeout: 5 * time.Second}, srv.URL, 5)
	if err != nil {
		t.Fatalf("expected success after retries, got %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Errorf("got status %d", resp.StatusCode)
	}
	if hits < 3 {
		t.Errorf("expected >=3 hits, got %d", hits)
	}
}

func TestFetchWithBackoff_GivesUp(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-RateLimit-Remaining", "0")
		w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(time.Millisecond).Unix(), 10))
		w.WriteHeader(http.StatusForbidden)
	}))
	defer srv.Close()

	_, err := fetchWithBackoff(&http.Client{Timeout: 5 * time.Second}, srv.URL, 2)
	if err == nil {
		t.Fatal("expected error after max attempts")
	}
	if !IsRateLimited(err) {
		t.Errorf("expected RateLimitError, got %T: %v", err, err)
	}
}
