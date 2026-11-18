package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"strings"
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

	// Declare the file as meta
	fmt.Printf("---@meta\n\n")

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
			generateStructDocs(d)
		// Function Declarations: (standalone functions and methods)
		case *ast.FuncDecl:
			generateFuncDocs(d)
		}
	}
}

// Generate LuaCATS docs for methods and functions
func generateFuncDocs(d *ast.FuncDecl) {
	if d.Name.IsExported() {

		// Determine the Receiver struct (e.g. func (b *Buffer) ... -> Buffer)
		recvType := ""
		if d.Recv != nil {
			switch t := d.Recv.List[0].Type.(type) {
			case *ast.StarExpr:
				if ident, ok := t.X.(*ast.Ident); ok {
					recvType = ident.Name
				}
			case *ast.Ident:
				recvType = t.Name
			}
		}

		// Docstrings
		if d.Doc != nil {
			doc := strings.TrimSpace(d.Doc.Text())
			for _, line := range strings.Split(doc, "\n") {
				fmt.Printf("--- %s\n", line)
			}
		}

		// Parameters
		var paramNames []string
		if d.Type.Params != nil {
			for _, param := range d.Type.Params.List {
				names := param.Names
				if len(names) > 0 {
					for _, name := range names {
						paramNames = append(paramNames, name.Name)
						fmt.Printf("---@param %s %s\n", name.Name, goTypeToLua(param.Type))
					}
				}
			}
		}

		// Returns
		if d.Type.Results != nil {
			for _, res := range d.Type.Results.List {
				if len(res.Names) > 0 {
					for _, name := range res.Names {
						fmt.Printf("---@return %s %s\n", goTypeToLua(res.Type), name.Name)
					}
				} else {
					fmt.Printf("---@return %s\n", goTypeToLua(res.Type))
				}
			}
		}

		// Print the Lua Function Definition
		paramsJoined := strings.Join(paramNames, ", ")
		recvPart := ""
		if recvType != "" {
			recvPart = recvType + ":"
		}
		fmt.Printf("function %s%s(%s) end\n\n", recvPart, d.Name.Name, paramsJoined)
	}
}
