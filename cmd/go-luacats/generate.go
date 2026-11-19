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

// Represents a type selector like "pkg.Type"
type typeSelector struct {
	packagePath string
	typeName    string
}

// Parses a type selector like "pkg.Type" into package and type parts
func parseTypeSelector(value string) (typeSelector, error) {
	// Find last dot separating package and type
	dot := strings.LastIndex(value, ".")
	if dot <= 0 || dot == len(value)-1 {
		return typeSelector{}, fmt.Errorf("invalid type selector %q: expected package.Type", value)
	}

	// Convert package path to filesystem path
	packagePath := filepath.FromSlash(value[:dot])
	if !filepath.IsLocal(packagePath) {
		return typeSelector{}, fmt.Errorf("invalid package path in selector %q", value)
	}

	// Return selector with package path and type name
	return typeSelector{packagePath: packagePath, typeName: value[dot+1:]}, nil
}

// Generates LuaCATS for selected types only
func generateSelectedTypes(sourceRoot, target string, values []string) error {
	selectors := make([]typeSelector, 0, len(values))

	// Parse all selectors
	for _, value := range values {
		selector, err := parseTypeSelector(value)
		if err != nil {
			return err
		}
		selectors = append(selectors, selector)
	}

	// Generate each selected type
	for _, selector := range selectors {

		// Determine package directory
		packageDir := filepath.Join(sourceRoot, selector.packagePath)

		// Generate LuaCATS for this type
		var output bytes.Buffer
		if err := luacats.GenerateSymbol(packageDir, selector.typeName, &output); err != nil {
			return fmt.Errorf("generate %s: %w", selectorPath(selector), err)
		}

		// Prepare output path
		outPath := filepath.Join(target, selector.packagePath, selector.typeName+".lua")
		if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
			return fmt.Errorf("create output directory for %s: %w", selectorPath(selector), err)
		}

		// Write output to file
		if err := os.WriteFile(outPath, output.Bytes(), 0644); err != nil {
			return fmt.Errorf("write %s: %w", outPath, err)
		}
	}

	return nil
}

// Returns the filesystem path for a selector (e.g. pkg.Type -> pkg/Type)
func selectorPath(selector typeSelector) string {
	return filepath.Join(selector.packagePath, selector.typeName)
}

// Generates LuaCATS from stdin to a writer
func generateStdinToWriter(src io.Reader, dst io.Writer) error {
	return luacats.Generate("<stdin>", src, dst)
}

// Generates LuaCATS from stdin to a file
func generateStdinToFile(src io.Reader, target string) error {
	if !strings.HasSuffix(target, ".lua") {
		return fmt.Errorf("target must be a .lua file")
	}

	// Generate before opening the target so parse errors don't overwrite it
	var output bytes.Buffer
	if err := luacats.Generate("<stdin>", src, &output); err != nil {
		return err
	}

	if err := os.WriteFile(target, output.Bytes(), 0644); err != nil {
		return fmt.Errorf("write %s: %w", target, err)
	}

	return nil
}
