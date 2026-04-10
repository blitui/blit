package tuikit

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	gopkg_yaml "gopkg.in/yaml.v3"
)

// ThemeHotReload watches a YAML theme file and pushes updates to an App via
// the Bubble Tea program sender. It debounces rapid filesystem events by
// 200 ms to avoid flicker on editors that save in multiple steps.
// Most callers should use WithThemeHotReload; this type is exposed for testing.
type ThemeHotReload struct {
	path    string
	sender  func(msg interface{})
	watcher *fsnotify.Watcher
	mu      sync.Mutex
	timer   *time.Timer
	stopped chan struct{}
}

// ThemeHotReloadMsg is a Bubble Tea message used to deliver a newly parsed
// theme to the App on the UI goroutine after a successful file reload.
type ThemeHotReloadMsg struct {
	Theme Theme
}

// ThemeHotReloadErrMsg is sent when the YAML file cannot be re-parsed.
// The App shows it as an error toast so the user sees it inline.
type ThemeHotReloadErrMsg struct {
	Err error
}

// themeYAML is the schema for the watched YAML file. All fields are optional;
// missing keys fall back to DefaultTheme values.
type themeYAML struct {
	Positive    string            `yaml:"positive"`
	Negative    string            `yaml:"negative"`
	Accent      string            `yaml:"accent"`
	Muted       string            `yaml:"muted"`
	Text        string            `yaml:"text"`
	TextInverse string            `yaml:"text_inverse"`
	Cursor      string            `yaml:"cursor"`
	Border      string            `yaml:"border"`
	Flash       string            `yaml:"flash"`
	Extra       map[string]string `yaml:"extra"`
}

// ParseThemeYAML reads a YAML-encoded theme definition and returns a Theme.
// Missing keys fall back to DefaultTheme values. Extra keys are placed in
// Theme.Extra.
func ParseThemeYAML(data []byte) (Theme, error) {
	var y themeYAML
	if err := gopkg_yaml.Unmarshal(data, &y); err != nil {
		return Theme{}, fmt.Errorf("theme yaml: %w", err)
	}
	m := map[string]string{}
	if y.Positive != "" {
		m["positive"] = y.Positive
	}
	if y.Negative != "" {
		m["negative"] = y.Negative
	}
	if y.Accent != "" {
		m["accent"] = y.Accent
	}
	if y.Muted != "" {
		m["muted"] = y.Muted
	}
	if y.Text != "" {
		m["text"] = y.Text
	}
	if y.TextInverse != "" {
		m["text_inverse"] = y.TextInverse
	}
	if y.Cursor != "" {
		m["cursor"] = y.Cursor
	}
	if y.Border != "" {
		m["border"] = y.Border
	}
	if y.Flash != "" {
		m["flash"] = y.Flash
	}
	for k, v := range y.Extra {
		m[k] = v
	}
	return ThemeFromMap(m), nil
}

// NewThemeHotReload starts a filesystem watcher on path and calls sender with
// ThemeHotReloadMsg or ThemeHotReloadErrMsg on every change. Call Stop when done.
// This is exported so tests and custom runners can use it directly; most callers
// should use WithThemeHotReload instead.
func NewThemeHotReload(path string, sender func(msg interface{})) (*ThemeHotReload, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("theme hot-reload: create watcher: %w", err)
	}
	if err := w.Add(path); err != nil {
		_ = w.Close()
		return nil, fmt.Errorf("theme hot-reload: watch %q: %w", path, err)
	}
	hr := &ThemeHotReload{
		path:    path,
		sender:  sender,
		watcher: w,
		stopped: make(chan struct{}),
	}
	go hr.run()
	return hr, nil
}

// run is the background goroutine that processes fsnotify events.
func (hr *ThemeHotReload) run() {
	defer close(hr.stopped)
	for {
		select {
		case event, ok := <-hr.watcher.Events:
			if !ok {
				return
			}
			if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) {
				hr.scheduleReload()
			}
		case err, ok := <-hr.watcher.Errors:
			if !ok {
				return
			}
			if err != nil {
				hr.sender(ThemeHotReloadErrMsg{Err: fmt.Errorf("theme hot-reload watcher: %w", err)})
			}
		}
	}
}

const hotReloadDebounce = 200 * time.Millisecond

// scheduleReload debounces rapid filesystem events by 200 ms.
func (hr *ThemeHotReload) scheduleReload() {
	hr.mu.Lock()
	defer hr.mu.Unlock()
	if hr.timer != nil {
		hr.timer.Reset(hotReloadDebounce)
		return
	}
	hr.timer = time.AfterFunc(hotReloadDebounce, func() {
		hr.mu.Lock()
		hr.timer = nil
		hr.mu.Unlock()
		hr.reload()
	})
}

// reload reads the file and sends the result to the UI goroutine.
func (hr *ThemeHotReload) reload() {
	data, err := os.ReadFile(hr.path)
	if err != nil {
		hr.sender(ThemeHotReloadErrMsg{Err: fmt.Errorf("theme hot-reload read %q: %w", hr.path, err)})
		return
	}
	t, err := ParseThemeYAML(data)
	if err != nil {
		hr.sender(ThemeHotReloadErrMsg{Err: err})
		return
	}
	hr.sender(ThemeHotReloadMsg{Theme: t})
}

// Stop shuts down the watcher and waits for the goroutine to exit.
func (hr *ThemeHotReload) Stop() {
	_ = hr.watcher.Close()
	<-hr.stopped
}

// --- App integration ---

// WithThemeHotReload returns an AppOption that watches the YAML file at path
// and live-reloads the theme whenever the file changes. The reload is debounced
// by 200 ms. If the file cannot be parsed the current theme is kept and an
// error toast is shown.
func WithThemeHotReload(path string) Option {
	return func(a *appModel) {
		a.hotReloadPath = path
	}
}
