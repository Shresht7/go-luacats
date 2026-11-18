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
		log.Fatal("Usage: go-luacats <path-to-go-file>")
	}

	// Extract filepath out of command-line arguments
	sourceFile := os.Args[1]

	// Ensure filepath exists
	_, err := os.Stat(sourceFile)
	if err != nil {
		panic(err) // TODO: Improve error handling here
	}

	// Crete a new FileSet for position information. Positions are relative to the fset
	fset := token.NewFileSet()

	// Parse the Go file (including comments)
	node, err := parser.ParseFile(fset, sourceFile, nil, parser.ParseComments)
	if err != nil {
		log.Fatalf("Failed to parse file: %v", err)
	}

	// Iterate through all top-level declarations in the file
	for _, decl := range node.Decls {
		switch d := decl.(type) {

		// Generic Declarations: Types, Structs, Consts, Vars
		case *ast.GenDecl:
			if d.Tok == token.TYPE {
				for _, spec := range d.Specs {
					if typeSpec, ok := spec.(*ast.TypeSpec); ok {
						fmt.Printf("TYPE: %s\n", typeSpec.Name.Name)
					}
				}
			}
		// Function Declarations: (standalone functions and methods)
		case *ast.FuncDecl:
			if d.Recv != nil {
				// If it has a receiver, then it is a method attached to a struct
				fmt.Printf("METHOD: %s\n", d.Name.Name)
			} else {
				// Otherwise, if it has no receiver, it is a plain old function
				fmt.Printf("FUNC: %s\n", d.Name.Name)
			}

		}
	}
}
