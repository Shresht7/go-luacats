package main

import (
	"bytes"
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

	// Require an explicit target for file output
	if target == "" {
		return fmt.Errorf("target must not be empty")
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
		if err := luacats.GenerateFileFromFile(file, outPath); err != nil {
			return err
		}

	}

	return nil
}

// Generates LuaCATS output to stdout for a source file or directory
func GenerateStdout(source string, dst io.Writer) error {

	// Get info about the source file or dir
	info, err := os.Stat(source)
	if err != nil {
		return fmt.Errorf("stat %s: %w", source, err)
	}

	// If this is a single file, delegate to GenerateFromFile
	if !info.IsDir() {
		return luacats.GenerateFromFile(source, dst)
	}

	// Read the directory of source files
	entries, err := os.ReadDir(source)
	if err != nil {
		return fmt.Errorf("read directory %s: %w", source, err)
	}

	// Construct a single output
	const header = "---@meta\n\n"
	var output bytes.Buffer
	output.WriteString(header)

	// Process each Go file in the directory
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") {
			continue
		}

		path := filepath.Join(source, entry.Name())
		var fileOutput bytes.Buffer
		if err := luacats.GenerateFromFile(path, &fileOutput); err != nil {
			return err
		}

		// Remove the individual `---@meta` headers
		body, ok := bytes.CutPrefix(fileOutput.Bytes(), []byte(header))
		if !ok {
			return fmt.Errorf("unexpected LuaCATS header for %s", path)
		}

		output.Write(body)
	}

	// Write combined output to dst
	_, err = io.Copy(dst, &output)
	return err
}
