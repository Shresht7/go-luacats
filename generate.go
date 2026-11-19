package luacats

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os"
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
