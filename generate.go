package luacats

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"strings"
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

	// Crete a new FileSet for position information. Positions are relative to the fset
	fset := token.NewFileSet()

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
		astFile, err := parser.ParseFile(fset, file, nil, parser.ParseComments)
		if err != nil {
			return fmt.Errorf("Failed to parse %s: %w", file, err)
		}

		var w io.Writer
		var outFile *os.File

		baseName := strings.TrimSuffix(filepath.Base(file), ".go")
		outPath := target
		if info.IsDir() {
			outPath = filepath.Join(target, baseName+".lua")
		}

		outFile, err = os.Create(outPath)
		if err != nil {
			return err
		}
		w = outFile

		processDecls(astFile.Decls, w)

		if outFile != nil {
			outFile.Close()
		}
	}

	return nil
}

// Writes LuaCATS definitions for one Go source file
func GenerateToWriter(source string, w io.Writer) error {
	output, err := generateSource(source)
	if err != nil {
		return err
	}

	if _, err := io.Copy(w, output); err != nil {
		return fmt.Errorf("write LuaCATS output: %w", err)
	}

	return nil
}

// Generates LuaCATS output for a Go source file into a buffer
func generateSource(source string) (*bytes.Buffer, error) {
	// Crete a new FileSet for position information. Positions are relative to the fset
	fset := token.NewFileSet()

	// Parse the Go source file into an AST
	file, err := parser.ParseFile(fset, source, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", source, err)
	}

	// Process the AST and generate LuaCATS output
	var output bytes.Buffer
	processDecls(file.Decls, &output)

	// Return the generated output buffer
	return &output, nil
}

// Handles the extraction and generation process
func processDecls(decls []ast.Decl, w io.Writer) {
	fmt.Fprintf(w, "---@meta\n\n")

	// Iterate through all top-level declarations in the file
	for _, decl := range decls {
		switch d := decl.(type) {

		// Generic Declarations: Types, Structs, Consts, Vars
		case *ast.GenDecl:
			generateStructDocs(d, w)
		// Function Declarations: (standalone functions and methods)
		case *ast.FuncDecl:
			generateFuncDocs(d, w)
		}
	}
}
