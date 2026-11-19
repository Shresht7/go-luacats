package luacats

import (
	"fmt"
	"go/ast"
	"io"
	"strings"
)

// Generate LuaCATS docs for methods and functions
func generateFuncDocs(d *ast.FuncDecl, w io.Writer) error {
	if !d.Name.IsExported() {
		return nil
	}

	// Determine the Receiver struct (e.g. func (b *Buffer) ... -> Buffer)
	recvType := receiverTypeName(d)

	// Docstrings
	if d.Doc != nil {
		doc := strings.TrimSpace(d.Doc.Text())
		for _, line := range strings.Split(doc, "\n") {
			if _, err := fmt.Fprintf(w, "--- %s\n", line); err != nil {
				return err
			}
		}
	}

	// Parameters
	var paramNames []string
	paramIndex := 0
	if d.Type.Params != nil {
		for _, param := range d.Type.Params.List {
			desc := fieldDescription(param)
			luaType := goTypeToLua(param.Type)
			if len(param.Names) > 0 {
				for _, name := range param.Names {
					paramIndex++
					paramNames = append(paramNames, name.Name)
					if err := writeTypeAnnotation(w, "---@param", name.Name, luaType, desc); err != nil {
						return err
					}
				}
			} else {
				paramIndex++
				name := fmt.Sprintf("arg%d", paramIndex)
				paramNames = append(paramNames, name)
				if err := writeTypeAnnotation(w, "---@param", name, luaType, desc); err != nil {
					return err
				}
			}
		}
	}

	// Returns
	if d.Type.Results != nil {
		for _, res := range d.Type.Results.List {
			if len(res.Names) > 0 {
				for _, name := range res.Names {
					if _, err := fmt.Fprintf(w, "---@return %s %s\n", goTypeToLua(res.Type), name.Name); err != nil {
						return err
					}
				}
			} else if _, err := fmt.Fprintf(w, "---@return %s\n", goTypeToLua(res.Type)); err != nil {
				return err
			}
		}
	}

	// Print the Lua Function Definition
	paramsJoined := strings.Join(paramNames, ", ")
	recvPart := ""
	if recvType != "" {
		recvPart = recvType + ":"
	}
	_, err := fmt.Fprintf(w, "function %s%s(%s) end\n\n", recvPart, d.Name.Name, paramsJoined)
	return err
}
