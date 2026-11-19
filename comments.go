package luacats

import (
	"fmt"
	"go/ast"
	"go/token"
	"io"
	"strings"
)

// Get field description from doc comments
func fieldDescription(field *ast.Field) string {
	if field.Doc == nil {
		return ""
	}
	return strings.Join(strings.Fields(field.Doc.Text()), " ")
}

// Write a type annotation line
func writeTypeAnnotation(w io.Writer, tag, name, luaType, desc string) {
	fmt.Fprintf(w, "%s %s %s", tag, name, luaType)
	if desc != "" {
		fmt.Fprintf(w, " %s", desc)
	}
	fmt.Fprintln(w)
}

// Attach doc comments to fields if missing
func attachFieldComments(fset *token.FileSet, file *ast.File) {
	commentMap := ast.NewCommentMap(fset, file, file.Comments)

	for node, groups := range commentMap {
		field, ok := node.(*ast.Field)
		if !ok || field.Doc != nil {
			continue
		}
		var comments []*ast.Comment
		for _, group := range groups {
			comments = append(comments, group.List...)
		}

		if len(comments) > 0 {
			field.Doc = &ast.CommentGroup{List: comments}
		}
	}
}
