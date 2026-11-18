package main

import (
	"fmt"
	"go/ast"
	"go/token"
	"io"
	"strings"
)

// Generate LuaCATS docs for Go structs
func generateStructDocs(d *ast.GenDecl, w io.Writer) {
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
							fmt.Fprintf(w, "--- %s\n", line)
						}
					}

					// Create ---@class doc for the struct
					fmt.Fprintf(w, "---@class %s\n", typeSpec.Name.Name)

					for _, field := range structType.Fields.List {
						// A field can have multiple declarations (e.g. X, Y int)
						for _, name := range field.Names {
							if name.IsExported() {
								// Create ---@field doc for each field
								fmt.Fprintf(w, "---@field %s %s\n", name.Name, goTypeToLua(field.Type))
							}
						}
					}
					// Instantiate the local table so that we can attach methods to it
					fmt.Fprintf(w, "local %s = {}\n\n", typeSpec.Name.Name)

				}

			}
		}
	}
}
