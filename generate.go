package luacats

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
)

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
