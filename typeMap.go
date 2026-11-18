package luacats

import (
	"fmt"
	"go/ast"
)

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
