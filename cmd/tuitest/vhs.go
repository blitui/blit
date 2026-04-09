package main

// vhs.go implements the `tuitest vhs <session>.tuisess` subcommand.
//
// Without --render it prints the generated VHS tape script to stdout.
// With --render <out.gif> it shells out to the `vhs` binary and produces a GIF.
//
// Usage:
//
//	tuitest vhs <session>.tuisess               # print tape to stdout
//	tuitest vhs <session>.tuisess -o out.gif    # render GIF via vhs

import (
	"flag"
	"fmt"
	"os"
	"os/exec"

	"github.com/moneycaringcoder/tuikit-go/internal/tape"
	"github.com/moneycaringcoder/tuikit-go/tuitest"
)

// runVHS is the entry point for `tuitest vhs`.
func runVHS(args []string) int {
	fs := flag.NewFlagSet("vhs", flag.ContinueOnError)
	outGIF := fs.String("o", "", "render GIF to this path via the vhs binary")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: tuitest vhs <session>.tuisess [-o out.gif]")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Without -o, the tape script is printed to stdout.")
		fmt.Fprintln(os.Stderr, "With -o, the vhs binary is invoked to produce a GIF.")
		fmt.Fprintln(os.Stderr, "Install vhs: https://github.com/charmbracelet/vhs#installation")
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		return 1
	}
	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "tuitest vhs: <session>.tuisess argument required")
		fs.Usage()
		return 1
	}

	sessPath := fs.Arg(0)
	sess, err := tuitest.LoadSession(sessPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "tuitest vhs: %v\n", err)
		return 1
	}

	script := tape.Generate(sess)

	if *outGIF == "" {
		// No render requested — just print the tape script.
		fmt.Print(script)
		return 0
	}

	return renderGIF(script, *outGIF)
}

// renderGIF writes the tape script to a temp file and shells out to `vhs`.
func renderGIF(script, outPath string) int {
	// Check that the vhs binary is available before writing the temp file.
	vhsBin, err := exec.LookPath("vhs")
	if err != nil {
		fmt.Fprintln(os.Stderr, "tuitest vhs: 'vhs' binary not found in PATH")
		fmt.Fprintln(os.Stderr, "Install it from: https://github.com/charmbracelet/vhs#installation")
		fmt.Fprintln(os.Stderr, "  go install github.com/charmbracelet/vhs@latest")
		fmt.Fprintln(os.Stderr, "  brew install vhs")
		return 1
	}

	// Write tape script to a temporary file.
	tmp, err := os.CreateTemp("", "tuitest-*.tape")
	if err != nil {
		fmt.Fprintf(os.Stderr, "tuitest vhs: create temp tape: %v\n", err)
		return 1
	}
	defer os.Remove(tmp.Name())

	if _, err := tmp.WriteString(script); err != nil {
		tmp.Close()
		fmt.Fprintf(os.Stderr, "tuitest vhs: write tape: %v\n", err)
		return 1
	}
	tmp.Close()

	// Append --output flag so vhs writes to the requested path.
	cmd := exec.Command(vhsBin, tmp.Name(), "--output", outPath) //nolint:gosec
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if runErr := cmd.Run(); runErr != nil {
		fmt.Fprintf(os.Stderr, "tuitest vhs: vhs exited with error: %v\n", runErr)
		return 1
	}

	fmt.Fprintf(os.Stderr, "[tuitest] rendered %s\n", outPath)
	return 0
}
