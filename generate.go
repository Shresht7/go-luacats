package luacats

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Generate parses one Go source file from `src` and writes its LuaCATS definitions to `dst`.
// filename is used for parser diagnostics only. Generate does not close `src` or `dst`.
// It writes nothing if parsing fails, a write error may leave partial output.
func Generate(filename string, src io.Reader, dst io.Writer) error {

	// Check for nil source reader
	if src == nil {
		return fmt.Errorf("source reader is nil")
	}

	// Crete a new FileSet for position information. Positions are relative to the fset
	fset := token.NewFileSet()

	// Parse the file into an AST
	file, err := parser.ParseFile(fset, filename, src, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("parse %s: %w", filename, err)
	}

	// Attach comments to fields if missing
	attachFieldComments(fset, file)

	// Process the AST and generate LuaCATS output
	var output bytes.Buffer
	processDecls(file.Decls, &output)

	// Write output buffer to the destination
	if _, err = io.Copy(dst, &output); err != nil {
		return err
	}

	return nil
}

// GenerateFromFile opens source, generates its LuaCATS definitions, and writes them to dst.
// It closes the source file but does *not* close dst.
func GenerateFromFile(source string, dst io.Writer) error {

	// Open the source file for reading
	file, err := os.Open(source)
	if err != nil {
		return fmt.Errorf("open %s: %w", source, err)
	}

	// Delegate to Generate the LuaCATS output in the dst writer
	generateErr := Generate(source, file, dst)
	closeErr := file.Close()

	// Check for errors from generate and close
	if generateErr != nil {
		return generateErr
	}
	if closeErr != nil {
		return fmt.Errorf("close %s: %w", source, closeErr)
	}

	return nil
}

// GenerateFileFromFile reads source file and writes its LuaCATS definitions to the target file.
// It overwrites an existing target after successfully parsing the source file and rejects
// targets that refer to the same file as source. A write error may leave partial output.
func GenerateFileFromFile(source, target string) (err error) {

	// Open the source file for reading
	input, err := os.Open(source)
	if err != nil {
		return fmt.Errorf("open %s: %w", source, err)
	}

	// Defer closing the input file
	defer func() {
		if closeErr := input.Close(); err == nil && closeErr != nil {
			err = fmt.Errorf("close %s: %w", source, closeErr)
		}
	}()

	// Get source info for later comparison
	sourceInfo, err := input.Stat()
	if err != nil {
		return fmt.Errorf("stat %s: %w", source, err)
	}

	// Generate LuaCATS definitions into a buffer
	var output bytes.Buffer
	if err := Generate(source, input, &output); err != nil {
		return err
	}

	// Open target without truncating so we can check the actual destination file
	dest, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return fmt.Errorf("open %s: %w", target, err)
	}

	// Defer closing the target file
	defer func() {
		if closeErr := dest.Close(); err == nil && closeErr != nil {
			err = fmt.Errorf("close %s: %w", target, closeErr)
		}
	}()

	// Get target info for comparison
	destInfo, err := dest.Stat()
	if err != nil {
		return fmt.Errorf("stat %s: %w", target, err)
	}

	// If source and target are the same, error out
	if os.SameFile(sourceInfo, destInfo) {
		return fmt.Errorf("source and target are the same file!")
	}

	// Truncate the target before writing
	if err := dest.Truncate(0); err != nil {
		return fmt.Errorf("truncate %s: %w", target, err)
	}

	// Write buffer to target file
	if _, err := io.Copy(dest, &output); err != nil {
		return fmt.Errorf("write %s: %w", target, err)
	}

	// All done
	return nil
}

// GenerateType writes LuaCATS definitions for an exported struct and its
// exported methods from a Go package directory. It does not close dst.
func GenerateType(packageDir, typeName string, dst io.Writer) error {

	// Check package directory exists and is a directory
	info, err := os.Stat(packageDir)
	if err != nil {
		return fmt.Errorf("stat package directory %s: %w", packageDir, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("%s is not a package directory", packageDir)
	}

	// Read all Go files in the package
	entries, err := os.ReadDir(packageDir)
	if err != nil {
		return fmt.Errorf("read package directory %s: %w", packageDir, err)
	}

	// Parse all Go files for the type and its methods
	fset := token.NewFileSet()
	var packageName string
	var selectedDecl *ast.GenDecl
	var methods []*ast.FuncDecl
	for _, entry := range entries {
		name := entry.Name()

		// Skip non-Go files and tests
		if entry.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}

		// Respect build constraints and the current GOOS/GOARCH
		matches, err := build.Default.MatchFile(packageDir, name)
		if err != nil {
			return fmt.Errorf("check build constraints for %s: %w", name, err)
		}
		if !matches {
			continue
		}

		// Parse the Go file
		path := filepath.Join(packageDir, name)
		file, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			return fmt.Errorf("parse %s: %w", path, err)
		}

		// Check package name
		if packageName == "" {
			packageName = file.Name.Name
		} else if packageName != file.Name.Name {
			return fmt.Errorf("package mismatch in %s: got %s, want %s", path, file.Name.Name, packageName)
		}

		// Attach comments to struct fields
		attachFieldComments(fset, file)

		// Collect type and methods
		for _, decl := range file.Decls {
			switch d := decl.(type) {

			// Handle type declarations
			case *ast.GenDecl:
				if d.Tok != token.TYPE {
					continue
				}
				for _, spec := range d.Specs {
					typeSpec, ok := spec.(*ast.TypeSpec)

					// Skip if not the type we're looking for
					if !ok || typeSpec.Name.Name != typeName {
						continue
					}

					// Must be exported
					if !typeSpec.Name.IsExported() {
						return fmt.Errorf("type %s is not exported", typeName)
					}

					// Must be a struct
					if _, ok := typeSpec.Type.(*ast.StructType); !ok {
						return fmt.Errorf("type %s is not a struct", typeName)
					}

					// Must not be declared twice
					if selectedDecl != nil {
						return fmt.Errorf("type %s is declared more than once in %s", typeName, packageDir)
					}

					// Use doc from type or declaration
					doc := d.Doc
					if typeSpec.Doc != nil {
						doc = typeSpec.Doc
					}

					// Save the selected declaration
					selectedDecl = &ast.GenDecl{
						Doc:   doc,
						Tok:   d.Tok,
						Specs: []ast.Spec{typeSpec},
					}
				}

			case *ast.FuncDecl:
				// Collect methods for the type
				if receiverTypeName(d) == typeName {
					methods = append(methods, d)
				}
			}
		}
	}

	// If no struct found
	if selectedDecl == nil {
		return fmt.Errorf("exported struct %s not found in %s", typeName, packageDir)
	}

	// Consolidate decls and methods into one list
	decls := []ast.Decl{selectedDecl}
	for _, method := range methods {
		decls = append(decls, method)
	}

	// Generate LuaCATS output
	var output bytes.Buffer
	processDecls(decls, &output)

	// Write output to destination
	if _, err := io.Copy(dst, &output); err != nil {
		return fmt.Errorf("write LuaCATS output: %w", err)
	}

	// Done successfully
	return nil
}

// Returns the type name of the receiver for a method, if any
func receiverTypeName(decl *ast.FuncDecl) string {

	// No receiver
	if decl.Recv == nil || len(decl.Recv.List) == 0 {
		return ""
	}

	// Get the receiver type
	receiver := decl.Recv.List[0].Type
	if pointer, ok := receiver.(*ast.StarExpr); ok {
		receiver = pointer.X
	}

	// Handle pointer receivers
	switch receiver := receiver.(type) {
	case *ast.Ident:
		return receiver.Name
	case *ast.IndexExpr:
		if name, ok := receiver.X.(*ast.Ident); ok {
			return name.Name
		}
	case *ast.IndexListExpr:
		if name, ok := receiver.X.(*ast.Ident); ok {
			return name.Name
		}
	}

	return ""
}

// Handles the extraction and generation process
func processDecls(decls []ast.Decl, w io.Writer) {
	fmt.Fprintf(w, "---@meta\n\n")

	// Iterate through all top-level declarations in the file
	for _, decl := range decls {
		switch d := decl.(type) {

		// Generic Declarations: Types, Structs, Consts, Vars
		case *ast.GenDecl:
			generateStructDocs(d, w)
		// Function Declarations: (standalone functions and methods)
		case *ast.FuncDecl:
			generateFuncDocs(d, w)
		}
	}
}
