package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/Shresht7/go-luacats"
)

// Generates LuaCATS definitions for the given Go source file or directory
func Generate(source string, target string) error {

	// Get source fs info. Also ensures if it exists
	info, err := os.Stat(source)
	if err != nil {
		return fmt.Errorf("Failed to stat source: %w", err)
	}

	// Resolve target if not explicitly set
	if target == "" {
		if info.IsDir() {
			target = source
		} else {
			baseName := strings.TrimSuffix(source, ".go")
			target = baseName + ".lua"
		}
	}

	// Collect the Go source files we need to operate on
	var files []string
	if info.IsDir() {
		entries, err := os.ReadDir(source)
		if err != nil {
			return fmt.Errorf("Failed to read directory: %w", err)
		}
		for _, f := range entries {
			if !f.IsDir() && strings.HasSuffix(f.Name(), ".go") {
				files = append(files, filepath.Join(source, f.Name()))
			}
		}

		// Ensure that the target directory exists
		if err := os.MkdirAll(target, 0755); err != nil {
			return fmt.Errorf("Failed to create target directory: %w", err)
		}
	} else {
		// Check target extension in this case
		if !strings.HasSuffix(target, ".lua") {
			return fmt.Errorf("Target must be a .lua file")
		}
		files = append(files, source)
	}

	// Process files
	for _, file := range files {

		// Generate LuaCATS output for this file
		output, err := luacats.GenerateSource(file)
		if err != nil {
			return err
		}

		// Determine output path
		baseName := strings.TrimSuffix(filepath.Base(file), ".go")
		outPath := target
		if info.IsDir() {
			outPath = filepath.Join(target, baseName+".lua")
		}

		// Create output file
		outFile, err := os.Create(outPath)
		if err != nil {
			return fmt.Errorf("create %s: %w", outPath, err)
		}

		// Write output to file, and report errors
		_, writeError := io.Copy(outFile, output)
		closeErr := outFile.Close()
		if writeError != nil {
			return fmt.Errorf("write %s: %w", outPath, writeError)
		}
		if closeErr != nil {
			return fmt.Errorf("close %s: %w", outPath, closeErr)
		}

	}

	return nil
}
