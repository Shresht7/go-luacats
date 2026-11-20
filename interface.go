package luacats

import (
	"fmt"
	"go/ast"
	"io"
	"strings"
)

// generateInterfaceDocs emits an interface and its directly delcared methods
func generateInterfaceDocs(name string, doc *ast.CommentGroup, iface *ast.InterfaceType, w io.Writer) error {
	if doc != nil {
		for _, line := range strings.Split(strings.TrimSpace(doc.Text()), "\n") {
			if _, err := fmt.Fprintf(w, "--- %s\n", line); err != nil {
				return err
			}
		}
	}

	if _, err := fmt.Fprintf(w, "---@class %s\nlocal %s = {}\n\n", name, name); err != nil {
		return err
	}

	for _, field := range iface.Methods.List {
		methodType, ok := field.Type.(*ast.FuncType)
		if !ok {
			continue // TODO: Embedded interfaces are not handled yet
		}

		for _, methodName := range field.Names {
			if !methodName.IsExported() {
				continue
			}

			method := &ast.FuncDecl{
				Doc:  field.Doc,
				Name: methodName,
				Type: methodType,
				Recv: &ast.FieldList{
					List: []*ast.Field{{Type: ast.NewIdent(name)}},
				},
			}

			if err := generateFuncDocs(method, w); err != nil {
				return err
			}
		}
	}

	return nil
}
