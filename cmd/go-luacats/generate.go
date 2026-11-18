package main

import (
	"fmt"
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

		// Determine target output path
		baseName := strings.TrimSuffix(filepath.Base(file), ".go")
		outPath := target
		if info.IsDir() {
			outPath = filepath.Join(target, baseName+".lua")
		}

		// Generate LuaCATS output to the target file
		if err := luacats.GenerateFileToFile(file, outPath); err != nil {
			return err
		}

	}

	return nil
}
