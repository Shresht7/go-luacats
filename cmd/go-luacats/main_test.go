package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
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
	if err := run([]string{source, target}, nil, &stdout); err != nil {
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
	if err := run(nil, nil, &stdout); err == nil {
		t.Fatal("expected an error for missing source")
	}
}

func TestRunWritesFileToStdoutByDefault(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "example.go")
	target := filepath.Join(dir, "example.lua")

	if err := os.WriteFile(source, []byte("package example\n"), 0644); err != nil {
		t.Fatal(err)
	}

	var stdout bytes.Buffer
	if err := run([]string{source}, nil, &stdout); err != nil {
		t.Fatal(err)
	}

	got := stdout.String()
	want := "---@meta\n\n"
	if got != want {
		t.Errorf("stdout: got %q, want %q", got, want)
	}

	if _, err := os.Stat(target); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected no output file, got stat error: %v", err)
	}
}

func TestRunWritersDirectoryToStdoutByDefault(t *testing.T) {
	source := t.TempDir()

	inputs := map[string]string{
		"a.go": "package example\ntype Alpha struct{}\n",
		"b.go": "package example\ntype Beta struct{}\n",
	}

	for name, content := range inputs {
		if err := os.WriteFile(filepath.Join(source, name), []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	var stdout bytes.Buffer
	if err := run([]string{source}, nil, &stdout); err != nil {
		t.Fatal(err)
	}

	want := "---@meta\n\n" +
		"---@class Alpha\n" +
		"local Alpha = {}\n\n" +
		"---@class Beta\n" +
		"local Beta = {}\n\n"

	if got := stdout.String(); got != want {
		t.Errorf("stdout: got %q, want %q", got, want)
	}

	for _, name := range []string{"a.lua", "b.lua"} {
		if _, err := os.Stat(filepath.Join(source, name)); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("expected no %s output file, got stat error: %v", name, err)
		}
	}
}

func TestRunRejectsEmptyTarget(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "example.go")
	target := filepath.Join(dir, "example.lua")
	if err := os.WriteFile(source, []byte("package example\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target, []byte("existing output"), 0644); err != nil {
		t.Fatal(err)
	}

	var stdout bytes.Buffer
	if err := run([]string{source, ""}, nil, &stdout); err == nil {
		t.Error("expected an error for an empty target")
	}

	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}

	if string(got) != "existing output" {
		t.Fatalf("target changed: %q", got)
	}
}

func TestRunHelpAndVersion(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "short help",
			args: []string{"-h"},
			want: "Usage: go-luacats <source> [target]",
		},
		{
			name: "long help",
			args: []string{"--help"},
			want: "Usage: go-luacats <source> [target]",
		},
		{
			name: "short version",
			args: []string{"-v"},
			want: "go-luacats " + version,
		},
		{
			name: "long version",
			args: []string{"--version"},
			want: "go-luacats " + version,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var stdout bytes.Buffer
			if err := run(test.args, nil, &stdout); err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(stdout.String(), test.want) {
				t.Fatalf("output %q does not contain %q", stdout.String(), test.want)
			}
		})
	}
}

func TestRunWithDirectoryTarget(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "source")
	target := filepath.Join(dir, "generated")

	if err := os.Mkdir(source, 0755); err != nil {
		t.Fatal(err)
	}

	inputs := map[string]string{
		"first.go":  "package example\ntype First struct{}\n",
		"second.go": "package example\ntype Second struct{}\n",
	}
	for name, content := range inputs {
		if err := os.WriteFile(filepath.Join(source, name), []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	var stdout bytes.Buffer
	if err := run([]string{source, target}, nil, &stdout); err != nil {
		t.Fatal(err)
	}
	if got, want := stdout.String(), "Generated LuaCATS\n"; got != want {
		t.Fatalf("stdout: got %q, want %q", got, want)
	}

	want := map[string]string{
		"first.lua":  "---@meta\n\n---@class First\nlocal First = {}\n\n",
		"second.lua": "---@meta\n\n---@class Second\nlocal Second = {}\n\n",
	}
	for name, expected := range want {
		got, err := os.ReadFile(filepath.Join(target, name))
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != expected {
			t.Errorf("%s: got %q, want %q", name, got, expected)
		}
	}
}

func TestRunWithSelectedTypes(t *testing.T) {
	root := t.TempDir()
	pointDir := filepath.Join(root, "internal", "api")
	widgetDir := filepath.Join(root, "internal", "ui")
	target := filepath.Join(t.TempDir(), "generated")

	for _, dir := range []string{pointDir, widgetDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
	}

	pointSource := "package api\ntype Point struct { X int }\n"
	if err := os.WriteFile(filepath.Join(pointDir, "point.go"), []byte(pointSource), 0644); err != nil {
		t.Fatal(err)
	}

	widgetSource := "package ui\ntype Widget struct { Name string }\n"
	if err := os.WriteFile(filepath.Join(widgetDir, "widget.go"), []byte(widgetSource), 0644); err != nil {
		t.Fatal(err)
	}

	args := []string{
		"--type", "internal/api.Point",
		root, target,
		"--type", "internal/ui.Widget",
	}
	var stdout bytes.Buffer
	if err := run(args, nil, &stdout); err != nil {
		t.Fatal(err)
	}

	wantFiles := map[string]string{
		filepath.Join("internal", "api", "Point.lua"): "---@meta\n\n---@class Point\n---@field X integer\nlocal Point = {}\n\n",
		filepath.Join("internal", "ui", "Widget.lua"): "---@meta\n\n---@class Widget\n---@field Name string\nlocal Widget = {}\n\n",
	}
	for relPath, want := range wantFiles {
		got, err := os.ReadFile(filepath.Join(target, relPath))
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != want {
			t.Errorf("%s: got %q, want %q", relPath, got, want)
		}
	}

}

func TestRunReadsStdinToStdout(t *testing.T) {
	source := strings.NewReader("package example\ntype Point struct { X int }\n")

	var stdout bytes.Buffer
	if err := run([]string{"-"}, source, &stdout); err != nil {
		t.Fatal(err)
	}

	want := "---@meta\n\n---@class Point\n---@field X integer\nlocal Point = {}\n\n"
	if got := stdout.String(); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestRunReadsStdinToFile(t *testing.T) {
	target := filepath.Join(t.TempDir(), "out.lua")
	source := strings.NewReader("package example\ntype Point struct { X int }\n")

	var stdout bytes.Buffer
	if err := run([]string{"-", target}, source, &stdout); err != nil {
		t.Fatal(err)
	}

	got, err_ := os.ReadFile(target)
	if err_ != nil {
		t.Fatal(err_)
	}

	if string(got) != "---@meta\n\n---@class Point\n---@field X integer\nlocal Point = {}\n\n" {
		t.Error("unexpected output")
	}
}
