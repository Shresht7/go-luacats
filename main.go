package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
)

func main() {

	// Show help message if needed
	if len(os.Args) < 2 {
		log.Fatal("Usage: go-luacats <path-to-go-file-or-directory>")
	}

	// Extract filepath out of command-line arguments
	source := os.Args[1]

	// Ensure filepath exists
	fileinfo, err := os.Stat(source)
	if err != nil {
		panic(err) // TODO: Improve error handling here
	}

	// Crete a new FileSet for position information. Positions are relative to the fset
	fset := token.NewFileSet()

	// Declare the file as meta
	fmt.Printf("---@meta\n\n")

	if fileinfo.IsDir() {
		processDir(source, fset)
	} else {
		// Parse the Go file (including comments)
		processFile(source, fset)
	}
}

// Processes a single Go file
func processFile(path string, fset *token.FileSet) {
	node, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		log.Fatalf("Failed to parse file: %v", err)
	}
	processDecls(node.Decls)
}

// Processes a directory of Go files
func processDir(path string, fset *token.FileSet) {
	files, _ := os.ReadDir(path)
	for _, f := range files {
		if !f.IsDir() {
			processFile(path+"/"+f.Name(), fset)
		}
	}
}

// Handles the extraction and generation process
func processDecls(decls []ast.Decl) {
	// Iterate through all top-level declarations in the file
	for _, decl := range decls {
		switch d := decl.(type) {

		// Generic Declarations: Types, Structs, Consts, Vars
		case *ast.GenDecl:
			generateStructDocs(d)
		// Function Declarations: (standalone functions and methods)
		case *ast.FuncDecl:
			generateFuncDocs(d)
		}
	}
}
