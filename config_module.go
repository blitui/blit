package blit

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/fsnotify/fsnotify"
)

// Config manages loading, saving, and editing application configuration.
// T must be a struct type with exported fields. Fields are discovered via
// reflection and configured with `blit:"..."` struct tags.
type Config[T any] struct {
	// Value holds the current configuration state.
	Value T

	// path is the config file location.
	path string

	// dirty tracks unsaved changes.
	dirty bool

	// appName for XDG path resolution.
	appName string

	// signal is lazily created by AsSignal and emits on config changes.
	signal *Signal[T]
}

// fieldMeta holds parsed struct tag metadata for a single config field.
type fieldMeta struct {
	Name     string   // Go field name
	Label    string   // display label
	Group    string   // section heading
	Hint     string   // help text
	Default  string   // default value as string
	Min      int      // numeric minimum (0 means unset)
	Max      int      // numeric maximum (0 means unset)
	Options  []string // enum values
	ReadOnly bool
}

// ConfigOption configures LoadConfig behavior.
type ConfigOption func(*configOpts)

type configOpts struct {
	path string
}

// WithConfigPath overrides the default config file path.
func WithConfigPath(path string) ConfigOption {
	return func(o *configOpts) {
		o.path = path
	}
}

// LoadConfig loads configuration for the named app from the platform-appropriate
// config directory. Missing files use struct defaults. Fields with `default` tags
// are populated if the loaded value is zero.
//
// Config file paths follow os.UserConfigDir():
//   - Linux:   ~/.config/<appName>/config.yaml
//   - macOS:   ~/Library/Application Support/<appName>/config.yaml
//   - Windows: %APPDATA%\<appName>\config.yaml
func LoadConfig[T any](appName string, opts ...ConfigOption) (*Config[T], error) {
	o := &configOpts{}
	for _, fn := range opts {
		fn(o)
	}

	path := o.path
	if path == "" {
		var err error
		path, err = DefaultConfigPath(appName)
		if err != nil {
			return nil, fmt.Errorf("config path: %w", err)
		}
	}

	c := &Config[T]{
		path:    path,
		appName: appName,
	}

	if err := LoadYAML(path, &c.Value); err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	applyDefaults(&c.Value)
	applyEnvOverrides(appName, &c.Value)

	return c, nil
}

// Save persists the current config to disk as YAML.
func (c *Config[T]) Save() error {
	if err := SaveYAML(c.path, c.Value); err != nil {
		return err
	}
	c.dirty = false
	return nil
}

// SetValue updates the config via a mutator function and marks it dirty.
func (c *Config[T]) SetValue(fn func(*T)) error {
	fn(&c.Value)
	c.dirty = true
	if c.signal != nil {
		c.signal.Set(c.Value)
	}
	return nil
}

// Path returns the config file path.
func (c *Config[T]) Path() string {
	return c.path
}

// IsDirty returns whether unsaved changes exist.
func (c *Config[T]) IsDirty() bool {
	return c.dirty
}

// AsSignal returns a reactive Signal that tracks the config value.
// The signal is created lazily on first call and emits whenever the config
// changes via SetValue or WatchFile reload. Components can Subscribe to the
// returned signal for reactive updates.
func (c *Config[T]) AsSignal() *Signal[T] {
	if c.signal == nil {
		c.signal = NewSignal(c.Value)
	}
	return c.signal
}

// Defaults returns a new T with all `default` tags applied.
func (c *Config[T]) Defaults() T {
	var v T
	applyDefaults(&v)
	return v
}

// Reset restores all fields to their default values and saves.
func (c *Config[T]) Reset() error {
	c.Value = c.Defaults()
	c.dirty = true
	return c.Save()
}

// WatchFile watches the config file for changes and reloads automatically.
// On each detected change, the YAML is re-read and Value is updated in place.
// The onChange callback (if non-nil) is invoked after a successful reload.
// WatchFile blocks until ctx is cancelled; run it in a goroutine.
//
// File events are debounced by 200ms to avoid flicker from editors that save
// in multiple steps (write tmp + rename).
func (c *Config[T]) WatchFile(ctx context.Context, onChange func(T)) error {
	if c.path == "" {
		return fmt.Errorf("config: no file path to watch")
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("config watch: %w", err)
	}
	defer func() { _ = watcher.Close() }()

	if err := watcher.Add(c.path); err != nil {
		return fmt.Errorf("config watch add: %w", err)
	}

	var mu sync.Mutex
	var debounce *time.Timer

	reload := func() {
		var v T
		if err := LoadYAML(c.path, &v); err != nil {
			return // silently ignore reload errors
		}
		applyDefaults(&v)
		applyEnvOverrides(c.appName, &v)
		c.Value = v
		c.dirty = false
		if c.signal != nil {
			c.signal.Set(v)
		}
		if onChange != nil {
			onChange(v)
		}
	}

	for {
		select {
		case <-ctx.Done():
			mu.Lock()
			if debounce != nil {
				debounce.Stop()
			}
			mu.Unlock()
			return nil
		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}
			if event.Op&(fsnotify.Write|fsnotify.Create) == 0 {
				continue
			}
			mu.Lock()
			if debounce != nil {
				debounce.Stop()
			}
			debounce = time.AfterFunc(200*time.Millisecond, reload)
			mu.Unlock()
		case _, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			// Silently ignore watcher errors; keep watching.
		}
	}
}

// Editor returns a ConfigEditor auto-generated from the struct tags.
// Each exported field with a `blit` tag becomes a ConfigField entry.
func (c *Config[T]) Editor() *ConfigEditor {
	metas := parseStructMetas(reflect.TypeOf(c.Value))
	fields := make([]ConfigField, 0, len(metas))

	for _, m := range metas {
		m := m // capture
		if m.ReadOnly {
			fields = append(fields, ConfigField{
				Label: m.Label,
				Group: m.Group,
				Hint:  m.Hint,
				Get:   func() string { return getFieldString(&c.Value, m.Name) },
			})
			continue
		}

		fields = append(fields, ConfigField{
			Label: m.Label,
			Group: m.Group,
			Hint:  m.Hint,
			Get:   func() string { return getFieldString(&c.Value, m.Name) },
			Set: func(s string) error {
				if err := setFieldString(&c.Value, m.Name, s); err != nil {
					return err
				}
				c.dirty = true
				return c.Save()
			},
			Validate: func(s string) error {
				return validateField(&c.Value, m.Name, s, m)
			},
		})
	}

	return NewConfigEditor(fields)
}

// parseBlitTag parses a blit:"..." struct tag into a fieldMeta.
func parseBlitTag(tag string) fieldMeta {
	var m fieldMeta
	if tag == "" {
		return m
	}

	pairs := splitTag(tag)
	for _, pair := range pairs {
		k, v, _ := strings.Cut(pair, "=")
		k = strings.TrimSpace(k)
		v = strings.TrimSpace(v)
		switch k {
		case "label":
			m.Label = v
		case "group":
			m.Group = v
		case "hint":
			m.Hint = v
		case "default":
			m.Default = v
		case "min":
			m.Min, _ = strconv.Atoi(v)
		case "max":
			m.Max, _ = strconv.Atoi(v)
		case "options":
			m.Options = strings.Split(v, "|")
		case "readonly":
			m.ReadOnly = true
		}
	}
	return m
}

// splitTag splits a blit tag value by commas, but respects pipes within option values.
func splitTag(tag string) []string {
	var parts []string
	var current strings.Builder
	for i := 0; i < len(tag); i++ {
		ch := tag[i]
		if ch == ',' {
			// Check if we're inside an options value by looking ahead
			// Simple approach: split on commas that are followed by a known key=
			part := strings.TrimSpace(current.String())
			if part != "" {
				parts = append(parts, part)
			}
			current.Reset()
		} else {
			current.WriteByte(ch)
		}
	}
	if s := strings.TrimSpace(current.String()); s != "" {
		parts = append(parts, s)
	}
	return parts
}

// parseStructMetas extracts fieldMeta for all exported fields with blit tags.
func parseStructMetas(t reflect.Type) []fieldMeta {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	var metas []fieldMeta
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if !f.IsExported() {
			continue
		}
		tag, ok := f.Tag.Lookup("blit")
		if !ok {
			continue
		}
		m := parseBlitTag(tag)
		m.Name = f.Name
		if m.Label == "" {
			m.Label = f.Name
		}
		metas = append(metas, m)
	}
	return metas
}

// applyDefaults sets zero-valued fields to their tag defaults.
func applyDefaults(v any) {
	rv := reflect.ValueOf(v).Elem()
	rt := rv.Type()
	for i := 0; i < rt.NumField(); i++ {
		f := rt.Field(i)
		if !f.IsExported() {
			continue
		}
		tag, ok := f.Tag.Lookup("blit")
		if !ok {
			continue
		}
		m := parseBlitTag(tag)
		if m.Default == "" {
			continue
		}
		fv := rv.Field(i)
		if !fv.IsZero() {
			continue
		}
		_ = setReflectValue(fv, m.Default)
	}
}

// applyEnvOverrides checks environment variables of the form APPNAME_FIELDNAME
// and overrides the corresponding struct fields. Field names are converted from
// CamelCase to UPPER_SNAKE_CASE (e.g., MaxRetries → MAX_RETRIES).
func applyEnvOverrides(appName string, v any) {
	prefix := camelToUpperSnake(appName) + "_"
	rv := reflect.ValueOf(v).Elem()
	rt := rv.Type()
	for i := 0; i < rt.NumField(); i++ {
		f := rt.Field(i)
		if !f.IsExported() {
			continue
		}
		envKey := prefix + camelToUpperSnake(f.Name)
		val, ok := os.LookupEnv(envKey)
		if !ok {
			continue
		}
		_ = setReflectValue(rv.Field(i), val)
	}
}

// camelToUpperSnake converts a CamelCase string to UPPER_SNAKE_CASE.
// e.g., "MaxRetries" → "MAX_RETRIES", "appName" → "APP_NAME".
func camelToUpperSnake(s string) string {
	var b strings.Builder
	for i, r := range s {
		if unicode.IsUpper(r) && i > 0 {
			prev := rune(s[i-1])
			if unicode.IsLower(prev) || unicode.IsDigit(prev) {
				b.WriteByte('_')
			}
		}
		b.WriteRune(unicode.ToUpper(r))
	}
	return b.String()
}

// getFieldString reads a struct field and returns its string representation.
func getFieldString(v any, name string) string {
	rv := reflect.ValueOf(v).Elem()
	fv := rv.FieldByName(name)
	if !fv.IsValid() {
		return ""
	}
	switch fv.Kind() {
	case reflect.Int, reflect.Int64:
		return strconv.FormatInt(fv.Int(), 10)
	case reflect.Bool:
		return strconv.FormatBool(fv.Bool())
	case reflect.String:
		return fv.String()
	case reflect.Slice:
		if fv.Type().Elem().Kind() == reflect.String {
			parts := make([]string, fv.Len())
			for i := 0; i < fv.Len(); i++ {
				parts[i] = fv.Index(i).String()
			}
			return strings.Join(parts, ",")
		}
	}
	return fmt.Sprintf("%v", fv.Interface())
}

// setFieldString parses a string and sets the named field.
func setFieldString(v any, name string, s string) error {
	rv := reflect.ValueOf(v).Elem()
	fv := rv.FieldByName(name)
	if !fv.IsValid() {
		return fmt.Errorf("unknown field: %s", name)
	}
	return setReflectValue(fv, s)
}

// setReflectValue sets a reflect.Value from a string.
func setReflectValue(fv reflect.Value, s string) error {
	switch fv.Kind() {
	case reflect.Int, reflect.Int64:
		n, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid integer: %w", err)
		}
		fv.SetInt(n)
	case reflect.Bool:
		b, err := strconv.ParseBool(s)
		if err != nil {
			return fmt.Errorf("invalid boolean: %w", err)
		}
		fv.SetBool(b)
	case reflect.String:
		fv.SetString(s)
	case reflect.Slice:
		if fv.Type().Elem().Kind() == reflect.String {
			if s == "" {
				fv.Set(reflect.MakeSlice(fv.Type(), 0, 0))
			} else {
				parts := strings.Split(s, ",")
				slice := reflect.MakeSlice(fv.Type(), len(parts), len(parts))
				for i, p := range parts {
					slice.Index(i).SetString(strings.TrimSpace(p))
				}
				fv.Set(slice)
			}
		} else {
			return fmt.Errorf("unsupported slice element type: %s", fv.Type().Elem().Kind())
		}
	default:
		return fmt.Errorf("unsupported field type: %s", fv.Kind())
	}
	return nil
}

// validateField checks a string value against the field's constraints.
func validateField(v any, name string, s string, m fieldMeta) error {
	rv := reflect.ValueOf(v).Elem()
	fv := rv.FieldByName(name)
	if !fv.IsValid() {
		return fmt.Errorf("unknown field: %s", name)
	}

	switch fv.Kind() {
	case reflect.Int, reflect.Int64:
		n, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid integer: %w", err)
		}
		if m.Min != 0 && int(n) < m.Min {
			return fmt.Errorf("value %d is below minimum %d", n, m.Min)
		}
		if m.Max != 0 && int(n) > m.Max {
			return fmt.Errorf("value %d exceeds maximum %d", n, m.Max)
		}
	case reflect.Bool:
		if _, err := strconv.ParseBool(s); err != nil {
			return fmt.Errorf("invalid boolean: %w", err)
		}
	case reflect.String:
		if len(m.Options) > 0 {
			found := false
			for _, opt := range m.Options {
				if s == opt {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("value %q not in allowed options: %s", s, strings.Join(m.Options, ", "))
			}
		}
	}
	return nil
}
