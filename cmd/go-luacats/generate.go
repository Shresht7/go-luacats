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
func generateFile(source string, target string) error {

	// Collect Go files from source
	files, isDir, err := collectGoFiles(source)
	if err != nil {
		return err
	}

	// Check target validity
	if target == "" {
		return fmt.Errorf("target must not be empty")
	}

	if isDir {
		if err := os.MkdirAll(target, 0755); err != nil {
			return fmt.Errorf("create target directory %s: %w", target, err)
		}
	} else if !strings.HasSuffix(target, ".lua") {
		return fmt.Errorf("target must be a .lua file")
	}

	// Process each file
	for _, file := range files {
		baseName := strings.TrimSuffix(filepath.Base(file), ".go")
		outPath := target
		if isDir {
			outPath = filepath.Join(target, baseName+".lua")
		}

		// Generate LuaCATS for this file
		if err := luacats.GenerateFileFromFile(file, outPath); err != nil {
			return err
		}
	}

	return nil
}

// Generates LuaCATS output to stdout for a source file or directory
func generateStdout(source string, dst io.Writer) error {

	// Collect Go files from source
	files, isDir, err := collectGoFiles(source)
	if err != nil {
		return err
	}

	// If this is a single file, delegate to GenerateFromFile
	if !isDir {
		return luacats.GenerateFromFile(files[0], dst)
	}

	// Construct a single output
	const header = "---@meta\n\n"
	var output bytes.Buffer
	output.WriteString(header)

	// Generate for each file and strip redundant header
	for _, file := range files {
		var fileOutput bytes.Buffer
		if err := luacats.GenerateFromFile(file, &fileOutput); err != nil {
			return err
		}

		body, ok := bytes.CutPrefix(fileOutput.Bytes(), []byte(header))
		if !ok {
			return fmt.Errorf("Unexpected LuaCATS header for %s", file)
		}
		output.Write(body)
	}

	// Write final output to dst
	_, err = io.Copy(dst, &output)
	return err
}
