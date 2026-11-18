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

// Translates Go AST expressions into LuaCATS type strings
func goTypeToLua(t ast.Expr) string {
	switch t := t.(type) {

	// Identifiers (basic types and custom structs)
	case *ast.Ident:
		switch t.Name {
		case "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64", "rune", "byte":
			return "integer"
		case "float32", "float64":
			return "number"
		case "bool":
			return "boolean"
		case "string":
			return "string"
		case "error":
			return "any" // Lua doesn't have an equivalent type
		default:
			return t.Name // Custom structs (e.g. "Point", "Line", "Rect")
		}

	// Arrays and slices
	case *ast.ArrayType:
		return goTypeToLua(t.Elt) + "[]" // e.g. A []Point becomes Point[] in LuaCATS

	// Maps
	case *ast.MapType:
		return fmt.Sprintf("table<%s,%s>", goTypeToLua(t.Key), goTypeToLua(t.Value))

	// Pointers
	case *ast.StarExpr:
		return goTypeToLua(t.X) // Pointers are just the type itself in LuaCATS (drop the *)

	// Selector expressions (e.g. pkg.Type)
	case *ast.SelectorExpr:
		return goTypeToLua(t.Sel) // Drop the pkg prefix

	// Interface types
	case *ast.InterfaceType:
		return "any"

	// Fallback to "any" type
	default:
		return "any"
	}
}

// Generate LuaCATS docs for Go structs
func generateStructDocs(d *ast.GenDecl) {
	if d.Tok == token.TYPE {
		for _, spec := range d.Specs {
			if typeSpec, ok := spec.(*ast.TypeSpec); ok {

				// Only process exported public structs
				if !typeSpec.Name.IsExported() {
					continue
				}

				if structType, ok := typeSpec.Type.(*ast.StructType); ok {

					// Struct docstrings
					if d.Doc != nil {
						doc := strings.TrimSpace(d.Doc.Text())
						for _, line := range strings.Split(doc, "\n") {
							fmt.Printf("--- %s\n", line)
						}
					}

					// Create ---@class doc for the struct
					fmt.Printf("---@class %s\n", typeSpec.Name.Name)

					for _, field := range structType.Fields.List {
						// A field can have multiple declarations (e.g. X, Y int)
						for _, name := range field.Names {
							if name.IsExported() {
								// Create ---@field doc for each field
								fmt.Printf("---@field %s %s\n", name.Name, goTypeToLua(field.Type))
							}
						}
					}
					// Instantiate the local table so that we can attach methods to it
					fmt.Printf("local %s = {}\n\n", typeSpec.Name.Name)

				}

			}
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
