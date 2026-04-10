package blit

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// DefaultConfigPath returns the default configuration file path for an
// application: <UserConfigDir>/<appName>/config.yaml.
// On Linux this is typically ~/.config/<appName>/config.yaml.
func DefaultConfigPath(appName string) (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("config dir: %w", err)
	}
	return filepath.Join(dir, appName, "config.yaml"), nil
}

// EnsureConfigDir creates the configuration directory for appName if it
// does not already exist. Returns the directory path.
func EnsureConfigDir(appName string) (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("config dir: %w", err)
	}
	p := filepath.Join(dir, appName)
	if err := os.MkdirAll(p, 0o755); err != nil {
		return "", fmt.Errorf("create config dir: %w", err)
	}
	return p, nil
}

// LoadYAML reads a YAML file at path and unmarshals it into dst.
// dst must be a pointer to the target struct. If the file does not exist,
// dst is left unchanged and no error is returned, allowing callers to
// rely on struct zero values as defaults.
func LoadYAML(path string, dst any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read config: %w", err)
	}
	if err := yaml.Unmarshal(data, dst); err != nil {
		return fmt.Errorf("parse config: %w", err)
	}
	return nil
}

// SaveYAML marshals src to YAML and writes it atomically to path.
// The file is written to a temporary location first, then renamed
// to avoid partial writes on crash.
func SaveYAML(path string, src any) error {
	data, err := yaml.Marshal(src)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	tmp, err := os.CreateTemp(dir, ".blit-config-*.yaml")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpName := tmp.Name()

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return fmt.Errorf("write config: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("close config: %w", err)
	}

	if err := os.Rename(tmpName, path); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("rename config: %w", err)
	}
	return nil
}
