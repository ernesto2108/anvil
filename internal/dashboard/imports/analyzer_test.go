package imports_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ernesto2108/anvil/internal/dashboard/imports"
)

// writeFile creates a file at dir/relPath with the given content.
func writeFile(t *testing.T, dir, relPath, content string) {
	t.Helper()
	full := filepath.Join(dir, relPath)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(full, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile %s: %v", relPath, err)
	}
}

// TestGoImportBlock verifies that a Go file with an import block detects
// edges to touched files and ignores external/stdlib imports.
func TestGoImportBlock(t *testing.T) {
	dir := t.TempDir()

	writeFile(t, dir, "go.mod", "module github.com/org/repo\n\ngo 1.21\n")
	writeFile(t, dir, "internal/handler/metrics.go", `package handler

import (
	"fmt"
	"github.com/org/repo/internal/service"
	"github.com/org/repo/internal/model"
)

func Hello() { fmt.Println("hi") }
`)
	writeFile(t, dir, "internal/service/metrics.go", "package service\n")
	writeFile(t, dir, "internal/model/run.go", "package model\n")

	touched := map[string]bool{
		"internal/service/metrics.go": true,
		"internal/model/run.go":       true,
	}

	a := imports.New(dir)
	edges, err := a.Analyze("internal/handler/metrics.go", touched)
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}
	if len(edges) != 2 {
		t.Fatalf("expected 2 edges, got %d: %+v", len(edges), edges)
	}
	for _, e := range edges {
		if e.SourcePath != "internal/handler/metrics.go" {
			t.Errorf("wrong SourcePath: %q", e.SourcePath)
		}
		if e.Relation != "imports" {
			t.Errorf("wrong Relation: %q", e.Relation)
		}
		if !touched[e.TargetPath] {
			t.Errorf("unexpected TargetPath: %q", e.TargetPath)
		}
	}
}

// TestGoMissingGoMod verifies that a Go file analyzed without go.mod returns nil, nil.
func TestGoMissingGoMod(t *testing.T) {
	dir := t.TempDir()
	// No go.mod written — only the source file
	writeFile(t, dir, "main.go", `package main

import "github.com/org/repo/internal/foo"

func main() { _ = foo.Bar }
`)

	a := imports.New(dir)
	edges, err := a.Analyze("main.go", map[string]bool{"internal/foo/foo.go": true})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if edges != nil {
		t.Fatalf("expected nil edges, got %+v", edges)
	}
}

// TestTypeScriptRelativeImport verifies that a TS file with a relative import
// returns an edge to the touched file.
func TestTypeScriptRelativeImport(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "src/components/Button.tsx", `import React from 'react'
import { useUser } from './useUser'

export const Button = () => null
`)
	writeFile(t, dir, "src/components/useUser.ts", "export const useUser = () => null\n")

	touched := map[string]bool{
		"src/components/useUser.ts": true,
	}

	a := imports.New(dir)
	edges, err := a.Analyze("src/components/Button.tsx", touched)
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}
	if len(edges) != 1 {
		t.Fatalf("expected 1 edge, got %d: %+v", len(edges), edges)
	}
	if edges[0].TargetPath != "src/components/useUser.ts" {
		t.Errorf("wrong TargetPath: %q", edges[0].TargetPath)
	}
}

// TestTypeScriptPackageImportIgnored verifies that a package import like
// 'import React from "react"' is not included in edges.
func TestTypeScriptPackageImportIgnored(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "src/App.tsx", `import React from 'react'
import { useState } from 'react'
import { Button } from 'lucide-react'
`)

	a := imports.New(dir)
	edges, err := a.Analyze("src/App.tsx", map[string]bool{"react": true, "lucide-react": true})
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}
	if len(edges) != 0 {
		t.Fatalf("expected 0 edges, got %d: %+v", len(edges), edges)
	}
}

// TestRustCrateImport verifies that a Rust file with "use crate::" detects
// an edge to the corresponding touched file.
func TestRustCrateImport(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "src/main.rs", `use crate::handler::metrics;

fn main() {}
`)
	writeFile(t, dir, "src/handler/metrics.rs", "pub fn run() {}\n")

	touched := map[string]bool{
		"src/handler/metrics.rs": true,
	}

	a := imports.New(dir)
	edges, err := a.Analyze("src/main.rs", touched)
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}
	if len(edges) != 1 {
		t.Fatalf("expected 1 edge, got %d: %+v", len(edges), edges)
	}
	if edges[0].TargetPath != "src/handler/metrics.rs" {
		t.Errorf("wrong TargetPath: %q", edges[0].TargetPath)
	}
}

// TestPythonRelativeImport verifies that a Python file with a relative import
// returns an edge to the touched file.
func TestPythonRelativeImport(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "app/handlers/metrics.py", `from .models import Run

def handle(): pass
`)
	writeFile(t, dir, "app/handlers/models.py", "class Run: pass\n")

	touched := map[string]bool{
		"app/handlers/models.py": true,
	}

	a := imports.New(dir)
	edges, err := a.Analyze("app/handlers/metrics.py", touched)
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}
	if len(edges) != 1 {
		t.Fatalf("expected 1 edge, got %d: %+v", len(edges), edges)
	}
	if edges[0].TargetPath != "app/handlers/models.py" {
		t.Errorf("wrong TargetPath: %q", edges[0].TargetPath)
	}
}

// TestUnsupportedExtension verifies that a .yaml file returns nil, nil.
func TestUnsupportedExtension(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "config.yaml", "key: value\n")

	a := imports.New(dir)
	edges, err := a.Analyze("config.yaml", map[string]bool{"other.yaml": true})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if edges != nil {
		t.Fatalf("expected nil edges, got %+v", edges)
	}
}

// TestFileNotFound verifies that analyzing a missing file returns nil, nil.
func TestFileNotFound(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "go.mod", "module github.com/org/repo\n")

	a := imports.New(dir)
	edges, err := a.Analyze("nonexistent.go", map[string]bool{})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if edges != nil {
		t.Fatalf("expected nil edges, got %+v", edges)
	}
}

// TestImportTargetNotInTouchedPaths verifies that imports whose target is not
// in touchedPaths are not included in the returned edges.
func TestImportTargetNotInTouchedPaths(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "go.mod", "module github.com/org/repo\n\ngo 1.21\n")
	writeFile(t, dir, "cmd/main.go", `package main

import (
	"github.com/org/repo/internal/service"
)

func main() {}
`)
	writeFile(t, dir, "internal/service/svc.go", "package service\n")

	// Note: internal/service/svc.go is NOT in touchedPaths
	a := imports.New(dir)
	edges, err := a.Analyze("cmd/main.go", map[string]bool{})
	if err != nil {
		t.Fatalf("Analyze: %v", err)
	}
	if len(edges) != 0 {
		t.Fatalf("expected 0 edges, got %d: %+v", len(edges), edges)
	}
}
