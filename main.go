package main

import (
	"fmt"
	"log"
	"go/parser"
	"go/token"
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
		panic(err)	// TODO: Improve error handling here
	}

	// Crete a new FileSet for position information. Positions are relative to the fset
	fset := token.NewFileSet()

	// Parse the Go file (including comments)
	node, err := parser.ParseFile(fset, sourceFile, nil, parser.ParseComments)
	if err != nil {
		log.Fatalf("Failed to parse file: %v",err)
	}

	// Iterate through all top-level declarations in the file
	for _, decl := range node.Decls {
		fmt.Println(decl)
	}
}
