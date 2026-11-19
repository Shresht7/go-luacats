package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestCollectGoFilesFromFile(t *testing.T) {
	source := filepath.Join(t.TempDir(), "example.go")
	if err := os.WriteFile(source, []byte("package example\n"), 0644); err != nil {
		t.Fatal(err)
	}

	files, isDir, err := collectGoFiles(source)
	if err != nil {
		t.Fatal(err)
	}

	if isDir {
		t.Fatal("isDir = true, want false")
	}

	if want := []string{source}; !reflect.DeepEqual(files, want) {
		t.Fatalf("files = %v, want %v", files, want)
	}
}

func TestCollectGoFilesFromDirectory(t *testing.T) {
	source := t.TempDir()
	nested := filepath.Join(source, "nested")
	if err := os.Mkdir(nested, 0755); err != nil {
		t.Fatal(err)
	}

	for name, content := range map[string]string{
		"a.go":                               "package example\n",
		"z.go":                               "package example\n",
		"notes.txt":                          "not Go source",
		filepath.Join("nested", "nested.go"): "package example\n",
	} {
		path := filepath.Join(source, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	files, isDir, err := collectGoFiles(source)
	if err != nil {
		t.Fatal(err)
	}

	if !isDir {
		t.Fatal("isDir = false, want true")
	}

	want := []string{
		filepath.Join(source, "a.go"),
		filepath.Join(source, "z.go"),
	}
	if !reflect.DeepEqual(files, want) {
		t.Fatalf("files = %v, want %v", files, want)
	}
}

func TestCollectGoFilesMissingSource(t *testing.T) {
	source := filepath.Join(t.TempDir(), "missing.go")

	files, isDir, err := collectGoFiles(source)
	if err == nil {
		t.Fatal("expected an error for missing source")
	}

	if files != nil {
		t.Fatalf("files should be nil, got %v", files)
	}

	if isDir {
		t.Fatal("isDir = true, want false")
	}
}
