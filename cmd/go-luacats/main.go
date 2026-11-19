package main

import (
	"fmt"
	"io"
	"log"
	"os"
)

var version = "0.1.0"

// the main entrypoint of the cli
func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		log.Fatal(err)
	}
}

// run executes the generation with given args and writes to stdout
func run(args []string, stdout io.Writer) error {

	// Show help if requested
	if len(args) > 0 && (args[0] == "-h" || args[0] == "--help") {
		return showHelp(stdout)
	}

	// Show version if requested
	if len(args) > 0 && (args[0] == "-v" || args[0] == "--version") {
		return showVersion(stdout)
	}

	// Require at least one argument
	if len(args) == 0 {
		return fmt.Errorf("Usage: go-luacats <source> [target]")
	}

	// Extract the source file or directory from the args
	source := args[0]

	// If only source is given, generate to stdout
	if len(args) == 1 {
		return generateStdout(source, stdout)
	}

	// Extract target if given
	target := ""
	if len(args) > 1 {
		target = args[1]
	}

	// Generate LuaCATS files
	if err := generateFile(source, target); err != nil {
		return err
	}

	// Write output message
	_, err := fmt.Fprintln(stdout, "Generated LuaCATS")
	return err
}

// prints showHelp message to the writer
func showHelp(w io.Writer) error {
	_, err := fmt.Fprintf(w, `Usage: go-luacats <source> [target]

Generate LuaCATS definitions from a Go source file or directory.

With no target, output is written to stdout.
With a target, a source file generates one Lua file; a directory generates one Lua file per Go file.

Caution: Existing output files are overwritten.
`)
	return err
}

// prints showVersion info
func showVersion(w io.Writer) error {
	_, err := fmt.Fprintf(w, "go-luacats %s\n", version)
	return err
}
