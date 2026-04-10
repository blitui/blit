package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/blitui/blit/internal/scaffold"
)

// runGen implements the `blit gen <pkg>` subcommand.
// It scaffolds a _test.go file with btest.NewTestModel boilerplate.
func runGen(args []string) int {
	fs := flag.NewFlagSet("gen", flag.ContinueOnError)
	component := fs.String("component", "", "customize the component type name in the boilerplate")
	force := fs.Bool("force", false, "overwrite an existing file")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: blit gen [flags] <pkg>")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Scaffold a _test.go file with btest.NewTestModel boilerplate.")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Flags:")
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		return 2
	}

	pkgArgs := fs.Args()
	if len(pkgArgs) != 1 {
		fs.Usage()
		return 2
	}
	pkgPath := pkgArgs[0]

	result, err := scaffold.Generate(scaffold.Options{
		PkgPath:   pkgPath,
		Component: *component,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "[blit gen] generate: %v\n", err)
		return 1
	}

	// Resolve the output directory: strip leading ./ and use the path as-is
	// when it looks like a local directory, otherwise write next to cwd.
	outDir := pkgPath
	if info, statErr := os.Stat(outDir); statErr != nil || !info.IsDir() {
		outDir = "."
	}

	outPath := filepath.Join(outDir, result.FileName)

	if !*force {
		if _, statErr := os.Stat(outPath); statErr == nil {
			fmt.Fprintf(os.Stderr, "[blit gen] %s already exists; use --force to overwrite\n", outPath)
			return 1
		}
	}

	if err := os.WriteFile(outPath, result.Content, 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "[blit gen] write %s: %v\n", outPath, err)
		return 1
	}

	fmt.Printf("[blit gen] wrote %s\n", outPath)
	return 0
}
