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

	if err := Generate(source, target); err != nil {
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
