package luacats

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateToWriter(t *testing.T) {
	source := filepath.Join(t.TempDir(), "point.go")
	goSource := `package example

type Point struct {
	X int
	Y int
}
`

	if err := os.WriteFile(source, []byte(goSource), 0644); err != nil {
		t.Fatal(err)
	}

	var output bytes.Buffer
	if err := GenerateToWriter(source, &output); err != nil {
		t.Fatal(err)
	}

	want := "---@meta\n\n" +
		"---@class Point\n" +
		"---@field X integer\n" +
		"---@field Y integer\n" +
		"local Point = {}\n\n"

	if output.String() != want {
		t.Fatalf("generated output:\n%s\nwant:\n%s", output.String(), want)
	}
}

type failingWriter struct {
	err error
}

func (w failingWriter) Write(_ []byte) (int, error) {
	return 0, w.err
}

func TestGenerateToWriterReturnsWriteError(t *testing.T) {
	source := filepath.Join(t.TempDir(), "point.go")
	if err := os.WriteFile(source, []byte("package example\n"), 0644); err != nil {
		t.Fatal(err)
	}

	writeErr := errors.New("write failed")
	err := GenerateToWriter(source, failingWriter{err: writeErr})
	if !errors.Is(err, writeErr) {
		t.Fatalf("got error %v, want %v", err, writeErr)
	}
}
