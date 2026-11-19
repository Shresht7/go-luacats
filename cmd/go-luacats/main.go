package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

var version = "0.2.0"

// the main entrypoint of the cli
func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		log.Fatal(err)
	}
}

// run executes the generation with given args and writes to stdout
func run(args []string, stdout io.Writer) error {

	// Require at least one argument
	if len(args) == 0 {
		return fmt.Errorf("%s", getUsage())
	}

	// Show help if requested
	if args[0] == "-h" || args[0] == "--help" {
		return showHelp(stdout)
	}

	// Show version if requested
	if args[0] == "-v" || args[0] == "--version" {
		return showVersion(stdout)
	}

	// Parse args into positionals and type selectors
	positionals, selectors, err := parseArgs(args)
	if err != nil {
		return err
	}

	// Require at least one positional
	if len(positionals) == 0 {
		return fmt.Errorf("%s", getUsage())
	}

	// Extract the source file or directory from the args
	source := positionals[0]

	// Extract target if given
	target := ""
	if len(positionals) > 1 {
		target = positionals[1]
	}

	// If selectors are given, generate selected types only
	if len(selectors) > 0 {
		if target == "" {
			return fmt.Errorf("a target directory is required with --type")
		}
		if err := generateSelectedTypes(source, target, selectors); err != nil {
			return err
		}
		_, err := fmt.Fprintln(stdout, "Generated LuaCATS")
		return err
	}

	// If only source is given, generate to stdout
	if len(positionals) == 1 {
		return generateStdout(source, stdout)
	}

	// Generate LuaCATS files
	if err := generateFile(source, target); err != nil {
		return err
	}

	// Write output message
	_, err = fmt.Fprintln(stdout, "Generated LuaCATS")
	return err
}

// parseArgs splits args into positionals and selectors from flags
func parseArgs(args []string) (positionals, selectors []string, err error) {
	for i := 0; i < len(args); i++ {
		arg := args[i]

		if arg == "--" {
			positionals = append(positionals, args[i+1:]...)
			break
		}

		switch {
		case arg == "--type" || arg == "-type" || arg == "-t":
			if i+1 >= len(args) || args[i+1] == "" {
				return nil, nil, fmt.Errorf("%s requires a package.Type selector", arg)
			}
			selectors = append(selectors, args[i+1])
			i++
		case strings.HasPrefix(arg, "--type="):
			selectors = append(selectors, arg[len("--type="):])
		case strings.HasPrefix(arg, "-type="):
			selectors = append(selectors, arg[len("-type="):])
		case strings.HasPrefix(arg, "-t="):
			selectors = append(selectors, arg[len("-t="):])
		default:
			positionals = append(positionals, arg)
		}
	}

	return positionals, selectors, nil
}

// returns usage string
func getUsage() string {
	return "Usage: go-luacats <source> [target] [--type package.Type ...]"
}

// prints showHelp message to the writer
func showHelp(w io.Writer) error {
	_, err := fmt.Fprintf(w, `%s

Generate LuaCATS definitions from a Go source file or directory.

With no target, output is written to stdout.
With a target, a source file generates one Lua file; a directory generates one Lua file per Go file.

With --type, source is a Go module root and target is a required output directory.
Repeat --type to generate multiple selected structs and their exported methods.

Caution: Existing output files are overwritten.
`, getUsage())
	return err
}

// prints showVersion info
func showVersion(w io.Writer) error {
	_, err := fmt.Fprintf(w, "go-luacats %s\n", version)
	return err
}
