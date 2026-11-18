package luacats

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os"
)

// Generate reads one Go source file and generates LuaCATS definitions.
// filename is only used to identify the source in parser errors; Generate does not close src or dst
// Writes no output if parsing fails
func Generate(filename string, src io.Reader, dst io.Writer) error {

	// Check for nil source reader
	if src == nil {
		return fmt.Errorf("source reader is nil")
	}

	// Crete a new FileSet for position information. Positions are relative to the fset
	fset := token.NewFileSet()

	// Parse the file into an AST
	file, err := parser.ParseFile(fset, filename, src, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("parse %s: %w", filename, err)
	}

	// Process the AST and generate LuaCATS output
	var output bytes.Buffer
	processDecls(file.Decls, &output)

	// Write output buffer to the destination
	if _, err = io.Copy(dst, &output); err != nil {
		return err
	}

	return nil
}

// Writes LuaCATS definitions for one Go source file
func GenerateToWriter(source string, w io.Writer) error {
	output, err := GenerateSource(source)
	if err != nil {
		return err
	}

	if _, err := io.Copy(w, output); err != nil {
		return fmt.Errorf("write LuaCATS output: %w", err)
	}

	return nil
}

// Generates LuaCATS output for a Go source file into a buffer
func GenerateSource(source string) (*bytes.Buffer, error) {

	// Open the source file for reading
	file, err := os.Open(source)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", source, err)
	}

	// Generate LuaCATS output into buffer
	var output bytes.Buffer
	generateErr := Generate(source, file, &output)
	closeError := file.Close()

	// Check for errors
	if generateErr != nil {
		return nil, generateErr
	}
	if closeError != nil {
		return nil, fmt.Errorf("close %s: %w", source, closeError)
	}

	// Return the buffer if no errors occurred
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
