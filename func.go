package luacats

import (
	"fmt"
	"go/ast"
	"io"
	"strings"
)

// Generate LuaCATS docs for methods and functions
func generateFuncDocs(d *ast.FuncDecl, w io.Writer) {
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
				fmt.Fprintf(w, "--- %s\n", line)
			}
		}

		// Parameters
		var paramNames []string
		paramIndex := 0
		if d.Type.Params != nil {
			for _, param := range d.Type.Params.List {
				desc := fieldDescription(param)
				luaType := goTypeToLua(param.Type)
				names := param.Names
				if len(names) > 0 {
					for _, name := range names {
						paramIndex++
						paramNames = append(paramNames, name.Name)
						writeTypeAnnotation(w, "---@param", name.Name, luaType, desc)
					}
				} else {
					paramIndex++
					name := fmt.Sprintf("arg%d", paramIndex)
					paramNames = append(paramNames, name)
					writeTypeAnnotation(w, "---@param", name, luaType, desc)
				}
			}
		}

		// Returns
		if d.Type.Results != nil {
			for _, res := range d.Type.Results.List {
				if len(res.Names) > 0 {
					for _, name := range res.Names {
						fmt.Fprintf(w, "---@return %s %s\n", goTypeToLua(res.Type), name.Name)
					}
				} else {
					fmt.Fprintf(w, "---@return %s\n", goTypeToLua(res.Type))
				}
			}
		}

		// Print the Lua Function Definition
		paramsJoined := strings.Join(paramNames, ", ")
		recvPart := ""
		if recvType != "" {
			recvPart = recvType + ":"
		}
		fmt.Fprintf(w, "function %s%s(%s) end\n\n", recvPart, d.Name.Name, paramsJoined)
	}
}
