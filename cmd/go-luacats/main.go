package main

import (
	"fmt"
	"io"
	"log"
	"os"
)

// the main entrypoint of the cli
func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		log.Fatal(err)
	}
}

// run executes the generation with given args and writes to stdout
func run(args []string, stdout io.Writer) error {

	// Check for required args
	if len(args) < 1 {
		return fmt.Errorf("Usage: go-luacats <path-to-go-file-or-directory>")
	}

	// Extract the source file or directory from the args
	source := args[0]

	// If only source is given, generate to stdout
	if len(args) == 1 {
		return GenerateStdout(source, stdout)
	}

	// Extract target if given
	target := ""
	if len(args) > 1 {
		target = args[1]
	}

	// Generate LuaCATS files
	if err := Generate(source, target); err != nil {
		return err
	}

	// Write output message
	_, err := fmt.Fprintln(stdout, "Generated LuaCATS")
	return err
}
