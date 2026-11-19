package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestRunWithTarget(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "example.go")
	target := filepath.Join(dir, "example.lua")

	if err := os.WriteFile(source, []byte("package example\n"), 0644); err != nil {
		t.Fatal(err)
	}

	var stdout bytes.Buffer
	if err := run([]string{source, target}, &stdout); err != nil {
		t.Fatal(err)
	}

	got := stdout.String()
	want := "Generated LuaCATS\n"
	if got != want {
		t.Errorf("stdout: got %q, want %q", got, want)
	}

	file, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(file) != "---@meta\n\n" {
		t.Fatalf("target: got %q", file)
	}
}

func TestRunRequiresSource(t *testing.T) {
	var stdout bytes.Buffer
	if err := run(nil, &stdout); err == nil {
		t.Fatal("expected an error for missing source")
	}
}
