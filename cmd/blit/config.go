package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// runConfig implements the `blit config` subcommand.
// It provides a CLI interface for reading and writing application configuration.
//
// Usage:
//
//	blit config                     show config file path
//	blit config get <field>         print the value of a field
//	blit config set <field> <val>   update a field and save
//	blit config list                print all fields and values
//	blit config edit                open config in $EDITOR
//	blit config path                print the config file path
func runConfig(args []string) int {
	fs := flag.NewFlagSet("config", flag.ExitOnError)
	appName := fs.String("app", "", "application name (defaults to current directory name)")
	_ = fs.Parse(args)

	remaining := fs.Args()

	if *appName == "" {
		*appName = detectAppName()
	}

	if len(remaining) == 0 {
		return configPath(*appName)
	}

	switch remaining[0] {
	case "path":
		return configPath(*appName)
	case "edit":
		return configEdit(*appName)
	case "get":
		if len(remaining) < 2 {
			fmt.Fprintln(os.Stderr, "Usage: blit config get <field>")
			return 1
		}
		return configGet(*appName, remaining[1])
	case "set":
		if len(remaining) < 3 {
			fmt.Fprintln(os.Stderr, "Usage: blit config set <field> <value>")
			return 1
		}
		return configSet(*appName, remaining[1], remaining[2])
	case "list":
		return configList(*appName)
	default:
		fmt.Fprintf(os.Stderr, "[blit config] unknown subcommand %q\n", remaining[0])
		fmt.Fprintln(os.Stderr, "Available: path, edit, get, set, list")
		return 1
	}
}

func configPath(appName string) int {
	dir, err := os.UserConfigDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "[blit config] %v\n", err)
		return 1
	}
	path := fmt.Sprintf("%s/%s/config.yaml", dir, appName)
	fmt.Println(path)
	return 0
}

func configEdit(appName string) int {
	dir, err := os.UserConfigDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "[blit config] %v\n", err)
		return 1
	}
	path := fmt.Sprintf("%s/%s/config.yaml", dir, appName)

	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = os.Getenv("VISUAL")
	}
	if editor == "" {
		fmt.Fprintln(os.Stderr, "[blit config] $EDITOR is not set")
		return 1
	}

	cmd := exec.Command(editor, path)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "[blit config] editor failed: %v\n", err)
		return 1
	}
	return 0
}

func configGet(appName, field string) int {
	dir, err := os.UserConfigDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "[blit config] %v\n", err)
		return 1
	}
	path := fmt.Sprintf("%s/%s/config.yaml", dir, appName)

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "[blit config] no config file at %s\n", path)
			return 1
		}
		fmt.Fprintf(os.Stderr, "[blit config] %v\n", err)
		return 1
	}

	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, field+":") {
			val := strings.TrimPrefix(line, field+":")
			fmt.Println(strings.TrimSpace(val))
			return 0
		}
	}
	fmt.Fprintf(os.Stderr, "[blit config] field %q not found\n", field)
	return 1
}

func configSet(appName, field, value string) int {
	dir, err := os.UserConfigDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "[blit config] %v\n", err)
		return 1
	}
	path := fmt.Sprintf("%s/%s/config.yaml", dir, appName)

	var lines []string
	data, err := os.ReadFile(path)
	if err == nil {
		lines = strings.Split(string(data), "\n")
	}

	found := false
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, field+":") {
			lines[i] = field + ": " + value
			found = true
			break
		}
	}
	if !found {
		lines = append(lines, field+": "+value)
	}

	configDir := fmt.Sprintf("%s/%s", dir, appName)
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "[blit config] %v\n", err)
		return 1
	}

	output := strings.Join(lines, "\n")
	if !strings.HasSuffix(output, "\n") {
		output += "\n"
	}
	if err := os.WriteFile(path, []byte(output), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "[blit config] %v\n", err)
		return 1
	}

	fmt.Printf("%s: %s\n", field, value)
	return 0
}

func configList(appName string) int {
	dir, err := os.UserConfigDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "[blit config] %v\n", err)
		return 1
	}
	path := fmt.Sprintf("%s/%s/config.yaml", dir, appName)

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("[blit config] no config file at %s\n", path)
			return 0
		}
		fmt.Fprintf(os.Stderr, "[blit config] %v\n", err)
		return 1
	}

	fmt.Print(string(data))
	return 0
}

func detectAppName() string {
	dir, err := os.Getwd()
	if err != nil {
		return "app"
	}
	parts := strings.Split(strings.ReplaceAll(dir, "\\", "/"), "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return "app"
}
