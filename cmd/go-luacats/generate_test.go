package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateFile(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "point.go")
	target := filepath.Join(dir, "point.lua")

	goSource := `package main

type Point struct{
	X int
	Y int
}
`

	if err := os.WriteFile(source, []byte(goSource), 0644); err != nil {
		t.Fatal(err)
	}

	if err := generateFile(source, target); err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}

	want := "---@meta\n\n" +
		"---@class Point\n" +
		"---@field X integer\n" +
		"---@field Y integer\n" +
		"local Point = {}\n\n"

	if string(got) != want {
		t.Fatalf("generated output:\n%swant:\n%s", got, want)
	}
}

func TestGeneratePreservesOutputOnParseError(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "bad.go")
	target := filepath.Join(dir, "bad.lua")

	if err := os.WriteFile(source, []byte("package example\ntype Broken struct {\n"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(target, []byte("existing output"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := generateFile(source, target); err == nil {
		t.Fatal("expected a parse error")
	}

	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}

	if string(got) != "existing output" {
		t.Fatalf("output was changed: %q", got)
	}
}

func TestGenerateDirectory(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "source")
	target := filepath.Join(dir, "output")

	if err := os.Mkdir(source, 0755); err != nil {
		t.Fatal(err)
	}

	inputs := map[string]string{
		"first.go":  "package example\ntype First struct {}\n",
		"second.go": "package example\ntype Second struct {}\n",
		"notes.txt": "Not Go Source Code",
	}

	for name, content := range inputs {
		if err := os.WriteFile(filepath.Join(source, name), []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	if err := generateFile(source, target); err != nil {
		t.Fatal(err)
	}

	want := map[string]string{
		"first.lua":  "---@meta\n\n---@class First\nlocal First = {}\n\n",
		"second.lua": "---@meta\n\n---@class Second\nlocal Second = {}\n\n",
	}

	entries, err := os.ReadDir(target)
	if err != nil {
		t.Fatal(err)
	}

	if len(entries) != len(want) {
		t.Fatalf("got %d output files, want %d", len(entries), len(want))
	}

	for name, expected := range want {
		got, err := os.ReadFile(filepath.Join(target, name))
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != expected {
			t.Fatalf("%s: got %q, want %q", name, got, expected)
		}
	}
}
