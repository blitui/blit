package blit

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestUpdateMode_String(t *testing.T) {
	cases := map[UpdateMode]string{
		UpdateNotify:   "notify",
		UpdateBlocking: "blocking",
		UpdateSilent:   "silent",
		UpdateForced:   "forced",
		UpdateDryRun:   "dryrun",
	}
	for m, want := range cases {
		if got := m.String(); got != want {
			t.Errorf("UpdateMode(%d).String() = %q, want %q", m, got, want)
		}
	}
	if UpdateMode(99).String() != "unknown" {
		t.Error("unknown UpdateMode should stringify as 'unknown'")
	}
}

// failingServer returns 500 on every request. If any update function hits
// it, the test fails — used to prove the kill switch short-circuits.
func failingServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("unexpected HTTP request to %s while updates should be disabled", r.URL.Path)
		w.WriteHeader(http.StatusInternalServerError)
	}))
}

func TestCheckForUpdate_EnvKillSwitch(t *testing.T) {
	srv := failingServer(t)
	defer srv.Close()

	t.Setenv(EnvDisableUpdate, "1")
	cfg := UpdateConfig{
		Owner:      "octocat",
		Repo:       "hello",
		BinaryName: "hello",
		Version:    "v1.0.0",
		APIBaseURL: srv.URL,
	}
	res, err := CheckForUpdate(cfg)
	if err != nil {
		t.Fatalf("CheckForUpdate err=%v", err)
	}
	if res.Available {
		t.Error("Available should be false when kill switch is set")
	}
	if res.CurrentVersion != "v1.0.0" {
		t.Errorf("CurrentVersion = %q, want v1.0.0", res.CurrentVersion)
	}
}

func TestCheckForUpdate_ConfigDisabled(t *testing.T) {
	srv := failingServer(t)
	defer srv.Close()
	os.Unsetenv(EnvDisableUpdate)

	cfg := UpdateConfig{
		Owner:      "octocat",
		Repo:       "hello",
		BinaryName: "hello",
		Version:    "v1.0.0",
		APIBaseURL: srv.URL,
		Disabled:   true,
	}
	if _, err := CheckForUpdate(cfg); err != nil {
		t.Fatalf("CheckForUpdate err=%v", err)
	}
}

func TestSelfUpdate_EnvKillSwitchNoNetwork(t *testing.T) {
	srv := failingServer(t)
	defer srv.Close()
	t.Setenv(EnvDisableUpdate, "true")

	cfg := UpdateConfig{
		Owner:      "octocat",
		Repo:       "hello",
		BinaryName: "hello",
		Version:    "v1.0.0",
		APIBaseURL: srv.URL,
	}
	if err := SelfUpdate(cfg); err != nil {
		t.Errorf("SelfUpdate with kill switch should return nil, got %v", err)
	}
}

func TestUpdateDisabled_CaseInsensitive(t *testing.T) {
	cases := []string{"1", "true", "TRUE", "Yes", "on"}
	for _, v := range cases {
		t.Run(v, func(t *testing.T) {
			t.Setenv(EnvDisableUpdate, v)
			if !updateDisabled(UpdateConfig{}) {
				t.Errorf("value %q should disable", v)
			}
		})
	}
}

func TestUpdateDisabled_EmptyAllowsUpdate(t *testing.T) {
	os.Unsetenv(EnvDisableUpdate)
	if updateDisabled(UpdateConfig{}) {
		t.Error("unset env var should not disable")
	}
	t.Setenv(EnvDisableUpdate, "0")
	if updateDisabled(UpdateConfig{}) {
		t.Error("value '0' should not disable")
	}
}

func TestSelfUpdate_OnBeforeUpdateAborts(t *testing.T) {
	os.Unsetenv(EnvDisableUpdate)
	called := false
	cfg := UpdateConfig{
		Owner:      "octocat",
		Repo:       "hello",
		BinaryName: "hello",
		Version:    "v1.0.0",
		OnBeforeUpdate: func() error {
			called = true
			return errSentinel
		},
		OnUpdateError: func(err error) {
			if err != errSentinel {
				t.Errorf("OnUpdateError got %v, want sentinel", err)
			}
		},
	}
	err := SelfUpdate(cfg)
	if !called {
		t.Error("OnBeforeUpdate was not called")
	}
	if err == nil || !strings.Contains(err.Error(), "OnBeforeUpdate aborted") {
		t.Errorf("expected OnBeforeUpdate abort error, got %v", err)
	}
}

var errSentinel = &sentinelError{"sentinel"}

type sentinelError struct{ s string }

func (e *sentinelError) Error() string { return e.s }

func TestSelfUpdate_DryRunNoWrites(t *testing.T) {
	os.Unsetenv(EnvDisableUpdate)
	// Mock server returning a valid-looking release.
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/octocat/hello/releases/latest", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"tag_name": "v2.0.0",
			"html_url": "https://example.com/release",
			"body": "notes",
			"assets": []
		}`))
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	cfg := UpdateConfig{
		Owner:      "octocat",
		Repo:       "hello",
		BinaryName: "hello",
		Version:    "v1.0.0",
		Mode:       UpdateDryRun,
		APIBaseURL: srv.URL,
	}
	if err := SelfUpdate(cfg); err != nil {
		t.Errorf("dry-run SelfUpdate err=%v", err)
	}
}
