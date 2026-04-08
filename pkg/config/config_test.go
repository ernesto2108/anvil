package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ernesto2108/anvil/pkg/config"
)

func setupTestRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	manifest := `targets:
  claude:
    enabled: true
    path: ~/.claude
  opencode:
    enabled: true
    path: ~/.config/opencode
  gemini:
    enabled: false
    path: ~/.gemini
  codex:
    enabled: true
    path: ~/.codex
  cursor:
    enabled: false

components:
  agents:
    tag: "HEAD"
  skills:
    tag: "HEAD"

ignore:
  - settings.json
`
	cfg := `provider: claude

providers:
  claude:
    high: opus
    medium: sonnet
    low: haiku
  gemini:
    high: gemini-2.5-pro
    medium: gemini-2.5-flash
    low: gemini-2.5-flash-lite

permissions:
  claude:
    read: Glob, Grep, LS, Read
    write: Glob, Grep, LS, Read, Write, Edit
    execute: Glob, Grep, LS, Read, Write, Edit, Bash
  opencode:
    read: read
    write: write
    execute: execute
`

	os.WriteFile(filepath.Join(dir, "forge.yaml"), []byte(manifest), 0o644)
	os.WriteFile(filepath.Join(dir, "forge.config.yaml"), []byte(cfg), 0o644)
	return dir
}

func Test_Load(t *testing.T) {
	dir := setupTestRepo(t)

	app, err := config.Load(dir, "forge")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if app.Name != "forge" {
		t.Errorf("Name = %q, want %q", app.Name, "forge")
	}
	if app.RepoDir != dir {
		t.Errorf("RepoDir = %q, want %q", app.RepoDir, dir)
	}
}

func Test_TargetEnabled(t *testing.T) {
	dir := setupTestRepo(t)
	app, _ := config.Load(dir, "forge")

	tests := []struct {
		target string
		want   bool
	}{
		{"claude", true},
		{"opencode", true},
		{"gemini", false},
		{"codex", true},
		{"cursor", false},
		{"nonexistent", false},
	}

	for _, tt := range tests {
		t.Run(tt.target, func(t *testing.T) {
			got := app.TargetEnabled(tt.target)
			if got != tt.want {
				t.Errorf("TargetEnabled(%q) = %v, want %v", tt.target, got, tt.want)
			}
		})
	}
}

func Test_ResolveTier(t *testing.T) {
	dir := setupTestRepo(t)
	app, _ := config.Load(dir, "forge")

	tests := []struct {
		name     string
		tier     string
		provider string
		want     string
		wantErr  bool
	}{
		{"high claude", "high", "claude", "opus", false},
		{"medium claude", "medium", "claude", "sonnet", false},
		{"low claude", "low", "claude", "haiku", false},
		{"high gemini", "high", "gemini", "gemini-2.5-pro", false},
		{"passthrough", "custom-model", "claude", "custom-model", false},
		{"unknown provider", "high", "unknown", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := app.ResolveTier(tt.tier, tt.provider)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func Test_ResolvePermission(t *testing.T) {
	dir := setupTestRepo(t)
	app, _ := config.Load(dir, "forge")

	tests := []struct {
		name string
		perm string
		tool string
		want string
	}{
		{"claude execute", "execute", "claude", "Glob, Grep, LS, Read, Write, Edit, Bash"},
		{"claude read", "read", "claude", "Glob, Grep, LS, Read"},
		{"opencode execute", "execute", "opencode", "execute"},
		{"unknown tool", "read", "unknown", ""},
		{"unknown perm", "admin", "claude", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := app.ResolvePermission(tt.perm, tt.tool)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func Test_ActiveProvider(t *testing.T) {
	dir := setupTestRepo(t)
	app, _ := config.Load(dir, "forge")

	if got := app.ActiveProvider(); got != "claude" {
		t.Errorf("ActiveProvider() = %q, want %q", got, "claude")
	}
}

func Test_SetTargetEnabled(t *testing.T) {
	dir := setupTestRepo(t)
	app, _ := config.Load(dir, "forge")

	if !app.TargetEnabled("claude") {
		t.Fatal("claude should be enabled initially")
	}

	if err := app.SetTargetEnabled("claude", false); err != nil {
		t.Fatalf("SetTargetEnabled: %v", err)
	}

	if app.TargetEnabled("claude") {
		t.Error("claude should be disabled after SetTargetEnabled(false)")
	}

	// Reload from disk to verify persistence
	app2, _ := config.Load(dir, "forge")
	if app2.TargetEnabled("claude") {
		t.Error("change not persisted to disk")
	}
}

func Test_IsComponentAllowed(t *testing.T) {
	tests := []struct {
		name    string
		item    string
		include []string
		exclude []string
		want    bool
	}{
		{"no filter", "developer", nil, nil, true},
		{"include match", "developer", []string{"developer", "tester"}, nil, true},
		{"include no match", "pm", []string{"developer", "tester"}, nil, false},
		{"exclude match", "pm", nil, []string{"pm", "reporter"}, false},
		{"exclude no match", "developer", nil, []string{"pm", "reporter"}, true},
		{"include beats exclude", "developer", []string{"developer"}, []string{"developer"}, true},
		{"with .md extension", "developer.md", []string{"developer"}, nil, true},
		{"exclude with .md", "pm.md", nil, []string{"pm"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := config.IsComponentAllowed(tt.item, tt.include, tt.exclude)
			if got != tt.want {
				t.Errorf("IsComponentAllowed(%q) = %v, want %v", tt.item, got, tt.want)
			}
		})
	}
}

func Test_SetProvider(t *testing.T) {
	dir := setupTestRepo(t)
	app, _ := config.Load(dir, "forge")

	if err := app.SetProvider("gemini"); err != nil {
		t.Fatalf("SetProvider: %v", err)
	}

	if got := app.ActiveProvider(); got != "gemini" {
		t.Errorf("ActiveProvider() = %q, want %q", got, "gemini")
	}

	// Reload from disk
	app2, _ := config.Load(dir, "forge")
	if got := app2.ActiveProvider(); got != "gemini" {
		t.Errorf("persisted provider = %q, want %q", got, "gemini")
	}
}
