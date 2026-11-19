package luacats

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerate(t *testing.T) {
	source := strings.NewReader(`package example

type Point struct {
	X int
	Y int
}
`)

	var output bytes.Buffer
	if err := Generate("point.go", source, &output); err != nil {
		t.Fatal(err)
	}

	want := "---@meta\n\n" +
		"---@class Point\n" +
		"---@field X integer\n" +
		"---@field Y integer\n" +
		"local Point = {}\n\n"

	if output.String() != want {
		t.Fatalf("got %q, want %q", output.String(), want)
	}
}

// Writer that always fails
type failingWriter struct {
	err error
}

// Implements io.Writer but always fails
func (w failingWriter) Write(_ []byte) (int, error) {
	return 0, w.err
}

func TestGenerateReturnsWriteError(t *testing.T) {
	source := strings.NewReader("package example\n")
	writeErr := errors.New("write failed")
	err := Generate("example.go", source, failingWriter{err: writeErr})
	if !errors.Is(err, writeErr) {
		t.Fatalf("got error %v, want %v", err, writeErr)
	}
}

func TestGenerateDoesNotWriteOnParseError(t *testing.T) {
	source := strings.NewReader("package example\ntype Broken struct{\n")
	var output bytes.Buffer
	if err := Generate("broken.go", source, &output); err == nil {
		t.Fatal("expected parse error, got nil")
	}
	if output.Len() != 0 {
		t.Fatalf("output should be empty: %q", output.String())
	}
}

func TestGenerateFile(t *testing.T) {
	source := filepath.Join(t.TempDir(), "example.go")
	if err := os.WriteFile(source, []byte("package example\n"), 0644); err != nil {
		t.Fatal(err)
	}

	var output bytes.Buffer
	if err := GenerateFromFile(source, &output); err != nil {
		t.Fatal(err)
	}

	got := output.String()
	want := "---@meta\n\n"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestGenerateFileReturnsOpenError(t *testing.T) {
	source := filepath.Join(t.TempDir(), "missing.go")

	err := GenerateFromFile(source, &bytes.Buffer{})
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatal(err)
	}
}

func TestGenerateFileFromFile(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "example.go")
	target := filepath.Join(dir, "example.lua")
	if err := os.WriteFile(source, []byte("package example\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target, []byte("old output"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := GenerateFileFromFile(source, target); err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}

	if string(got) != "---@meta\n\n" {
		t.Fatalf("got %q, want LuaCATS output", string(got))
	}

}

func TestGenerateFileFromFilePreservesTargetOnParseError(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "bad.go")
	target := filepath.Join(dir, "bad.lua")
	if err := os.WriteFile(source, []byte("package example\ntype Broken struct {\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target, []byte("existing output"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := GenerateFileFromFile(source, target); err == nil {
		t.Fatal("expected a parse error")
	}

	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}

	if string(got) != "existing output" {
		t.Fatalf("target changed: %q", string(got))
	}
}

func TestGenerateFileFromFileRejectsSourceAsTarget(t *testing.T) {
	source := filepath.Join(t.TempDir(), "example.go")
	original := "package example\n"

	if err := os.WriteFile(source, []byte(original), 0644); err != nil {
		t.Fatal(err)
	}

	if err := GenerateFileFromFile(source, source); err == nil {
		t.Fatal("expected an error for identical source and target")
	}

	got, err := os.ReadFile(source)
	if err != nil {
		t.Fatal(err)
	}

	if string(got) != original {
		t.Fatalf("soruce changed: %q", string(got))
	}
}

func TestGenerateFileFromFileRejectsHardLink(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "example.go")
	target := filepath.Join(dir, "linked.lua")
	original := "package example\n"

	if err := os.WriteFile(source, []byte(original), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Link(source, target); err != nil {
		t.Skipf("hard links unavailable: %v", err)
	}

	if err := GenerateFileFromFile(source, target); err == nil {
		t.Fatal("expected an error for a target linked to the source")
	}

	got, err := os.ReadFile(source)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != original {
		t.Fatalf("source changed: %q", got)
	}
}

func TestGenerateStructAnnotations(t *testing.T) {
	source := strings.NewReader(`package example

// Point represents a position.
type Point struct {
	X, Y int
	hidden string
}

type Alias string
`)

	var output bytes.Buffer
	if err := Generate("point.go", source, &output); err != nil {
		t.Fatal(err)
	}

	want := "---@meta\n\n" +
		"--- Point represents a position.\n" +
		"---@class Point\n" +
		"---@field X integer\n" +
		"---@field Y integer\n" +
		"local Point = {}\n\n"

	if got := output.String(); got != want {
		t.Fatalf("generated output:\n%s\nwant:\n%s", got, want)
	}
}

func TestGenerateFunctionAnnotations(t *testing.T) {
	source := strings.NewReader(`package example

// Adds two values.
func Add(a, b int) int {
	return a + b
}

// Moves the point.
func (p *Point) Move(dx int) (ok bool) {
	return true
}

func hidden() {}
`)

	var output bytes.Buffer
	if err := Generate("point.go", source, &output); err != nil {
		t.Fatal(err)
	}

	want := "---@meta\n\n" +
		"--- Adds two values.\n" +
		"---@param a integer\n" +
		"---@param b integer\n" +
		"---@return integer\n" +
		"function Add(a, b) end\n\n" +
		"--- Moves the point.\n" +
		"---@param dx integer\n" +
		"---@return boolean ok\n" +
		"function Point:Move(dx) end\n\n"

	if got := output.String(); got != want {
		t.Fatalf("generated output:\n%s\nwant:\n%s", got, want)
	}
}

func TestGenerateUnnamedParameters(t *testing.T) {
	source := strings.NewReader(`package example

// Accepts values without Go parameter names.
func Consume(int, string) {}
`)

	var output bytes.Buffer
	if err := Generate("example.go", source, &output); err != nil {
		t.Fatal(err)
	}

	want := "---@meta\n\n" +
		"--- Accepts values without Go parameter names.\n" +
		"---@param arg1 integer\n" +
		"---@param arg2 string\n" +
		"function Consume(arg1, arg2) end\n\n"

	if got := output.String(); got != want {
		t.Fatalf("generated output:\n%s\nwant:\n%s", got, want)
	}
}

func TestGenerateFieldAndParameterDescriptions(t *testing.T) {
	source := strings.NewReader(`package example

type Point struct {
	// Horizontal coordinate.
	X int
}

func Translate(
	// Horizontal distance.
	dx int,
) {}
`)

	var output bytes.Buffer
	if err := Generate("point.go", source, &output); err != nil {
		t.Fatal(err)
	}

	want := "---@meta\n\n" +
		"---@class Point\n" +
		"---@field X integer Horizontal coordinate.\n" +
		"local Point = {}\n\n" +
		"---@param dx integer Horizontal distance.\n" +
		"function Translate(dx) end\n\n"

	if got := output.String(); got != want {
		t.Fatalf("generated output:\n%s\nwant:\n%s", got, want)
	}
}
