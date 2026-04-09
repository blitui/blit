package tuikit_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	tuikit "github.com/moneycaringcoder/tuikit-go"
)

func TestThemeHotReload_ParseYAML(t *testing.T) {
	data := []byte(`
positive: "#22c55e"
negative: "#ef4444"
accent: "#3b82f6"
muted: "#6b7280"
text: "#e5e7eb"
text_inverse: "#111827"
cursor: "#38bdf8"
border: "#374151"
flash: "#facc15"
extra:
  brand: "#ff00ff"
`)
	theme, err := tuikit.ParseThemeYAML(data)
	if err != nil {
		t.Fatalf("ParseThemeYAML: %v", err)
	}
	if string(theme.Accent) != "#3b82f6" {
		t.Errorf("Accent = %q, want %q", theme.Accent, "#3b82f6")
	}
	if string(theme.Positive) != "#22c55e" {
		t.Errorf("Positive = %q, want %q", theme.Positive, "#22c55e")
	}
	brand := theme.Color("brand", "")
	if string(brand) != "#ff00ff" {
		t.Errorf("Extra[brand] = %q, want %q", brand, "#ff00ff")
	}
}

func TestThemeHotReload_ParseYAML_Partial(t *testing.T) {
	// Only accent is set; other fields fall back to DefaultTheme.
	data := []byte(`accent: "#aabbcc"`)
	theme, err := tuikit.ParseThemeYAML(data)
	if err != nil {
		t.Fatalf("ParseThemeYAML: %v", err)
	}
	if string(theme.Accent) != "#aabbcc" {
		t.Errorf("Accent = %q, want #aabbcc", theme.Accent)
	}
	def := tuikit.DefaultTheme()
	if theme.Text != def.Text {
		t.Errorf("Text should fall back to default, got %q", theme.Text)
	}
}

func TestThemeHotReload_ParseYAML_InvalidYAML(t *testing.T) {
	// accent mapped to a sequence instead of a string triggers an unmarshal error.
	_, err := tuikit.ParseThemeYAML([]byte("accent: [1, 2, 3]"))
	if err == nil {
		t.Fatal("expected error for invalid YAML, got nil")
	}
}

func TestThemeHotReload_FileWatch(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "theme.yaml")

	initial := []byte(`accent: "#111111"`)
	if err := os.WriteFile(path, initial, 0644); err != nil {
		t.Fatal(err)
	}

	received := make(chan interface{}, 4)
	sender := func(msg interface{}) { received <- msg }

	hr, err := tuikit.NewThemeHotReload(path, sender)
	if err != nil {
		t.Fatalf("NewThemeHotReload: %v", err)
	}
	defer hr.Stop()

	// Write an updated theme file.
	updated := []byte(`accent: "#aabbcc"`)
	if err := os.WriteFile(path, updated, 0644); err != nil {
		t.Fatal(err)
	}

	// Wait for the reload message (debounce is 200 ms, give 2 s total).
	timeout := time.After(2 * time.Second)
	for {
		select {
		case msg := <-received:
			if reload, ok := msg.(tuikit.ThemeHotReloadMsg); ok {
				if string(reload.Theme.Accent) != "#aabbcc" {
					t.Errorf("hot-reload accent = %q, want #aabbcc", reload.Theme.Accent)
				}
				return
			}
		case <-timeout:
			t.Fatal("timed out waiting for hot-reload message")
		}
	}
}

func TestThemeHotReload_InvalidFileEmitsError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "theme.yaml")

	if err := os.WriteFile(path, []byte(`accent: "#111111"`), 0644); err != nil {
		t.Fatal(err)
	}

	received := make(chan interface{}, 4)
	sender := func(msg interface{}) { received <- msg }

	hr, err := tuikit.NewThemeHotReload(path, sender)
	if err != nil {
		t.Fatalf("NewThemeHotReload: %v", err)
	}
	defer hr.Stop()

	// Write YAML that fails unmarshal (sequence into string field).
	if err := os.WriteFile(path, []byte("accent: [1, 2, 3]"), 0644); err != nil {
		t.Fatal(err)
	}

	timeout := time.After(2 * time.Second)
	for {
		select {
		case msg := <-received:
			if _, ok := msg.(tuikit.ThemeHotReloadErrMsg); ok {
				return // got the expected error message
			}
		case <-timeout:
			t.Fatal("timed out waiting for hot-reload error message")
		}
	}
}
