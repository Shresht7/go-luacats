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
	if err := run(os.Args[1:], os.Stdin, os.Stdout); err != nil {
		log.Fatal(err)
	}
}

// run executes the generation with given args and writes to stdout
// can read from stdin and write to stdout
func run(args []string, stdin io.Reader, stdout io.Writer) error {

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

	// Handle stdin as source
	if source == "-" {
		if len(selectors) > 0 {
			return fmt.Errorf("cannot use --symbol with stdin")
		}
		if target == "" {
			return generateStdinToWriter(stdin, stdout)
		}
		if err := generateStdinToFile(stdin, target); err != nil {
			return err
		}
		_, err := fmt.Fprintln(stdout, "Generated LuaCATS")
		return err
	}

	// If selectors are given, generate selected types only
	if len(selectors) > 0 {
		if target == "" {
			return fmt.Errorf("a target directory is required with --symbol")
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
		case arg == "--symbol" || arg == "-symbol" || arg == "-s":
			if i+1 >= len(args) || args[i+1] == "" {
				return nil, nil, fmt.Errorf("%s requires a package.Symbol selector", arg)
			}
			selectors = append(selectors, args[i+1])
			i++
		case strings.HasPrefix(arg, "--symbol="),
			strings.HasPrefix(arg, "-symbol="),
			strings.HasPrefix(arg, "-s="):
			_, value, _ := strings.Cut(arg, "=")
			if value == "" {
				return nil, nil, fmt.Errorf("%s requires a package.Symbol selector", arg)
			}
			selectors = append(selectors, value)
		default:
			// A lone "-" remains the stdin source. "--" is handled above the switch
			if arg != "-" && strings.HasPrefix(arg, "-") {
				return nil, nil, fmt.Errorf("unknown option %q", arg)
			}
			positionals = append(positionals, arg)
		}
	}

	return positionals, selectors, nil
}

// returns usage string
func getUsage() string {
	return "Usage: go-luacats <source> [target] [--symbol package.Symbol ...]"
}

// prints showHelp message to the writer
func showHelp(w io.Writer) error {
	_, err := fmt.Fprintf(w, `%s

Generate LuaCATS definitions from a Go source file or directory.

Use - as source to read Go source from stdin. Useful when combined with go doc -src

With no target, output is written to stdout.
With a target, a source file generates one Lua file; a directory generates one Lua file per Go file.

With --symbol, source is a Go module root and target is a required output directory.
Repeat --symbol to generate multiple selected structs and their exported methods.

Caution: Existing output files are overwritten.
`, getUsage())
	return err
}

// prints showVersion info
func showVersion(w io.Writer) error {
	_, err := fmt.Fprintf(w, "go-luacats %s\n", version)
	return err
}
