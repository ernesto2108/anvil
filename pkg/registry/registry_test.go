package registry_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ernesto2108/anvil/pkg/registry"
)

func sampleIndex() registry.Index {
	return registry.Index{
		Name: "test-registry",
		Entries: []registry.Entry{
			{Name: "code-reviewer", Type: "agent", Description: "Reviews code", Version: "1.0.0", URL: "https://example.com/code-reviewer.md", Author: "alice"},
			{Name: "docker-skill", Type: "skill", Description: "Docker conventions", Version: "1.0.0", URL: "https://example.com/docker-skill.md"},
			{Name: "linter-agent", Type: "agent", Description: "Runs linters", Version: "2.0.0", URL: "https://example.com/linter.md"},
		},
	}
}

func Test_Fetch(t *testing.T) {
	idx := sampleIndex()
	data, _ := json.Marshal(idx)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(data)
	}))
	defer srv.Close()

	result, err := registry.Fetch(srv.URL)
	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}

	if result.Name != "test-registry" {
		t.Errorf("name = %q, want %q", result.Name, "test-registry")
	}
	if len(result.Entries) != 3 {
		t.Errorf("got %d entries, want 3", len(result.Entries))
	}
}

func Test_Fetch_Error(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	_, err := registry.Fetch(srv.URL)
	if err == nil {
		t.Error("expected error for 404 response")
	}
}

func Test_FilterByType(t *testing.T) {
	idx := sampleIndex()

	agents := registry.FilterByType(idx.Entries, "agent")
	if len(agents) != 2 {
		t.Errorf("got %d agents, want 2", len(agents))
	}

	skills := registry.FilterByType(idx.Entries, "skill")
	if len(skills) != 1 {
		t.Errorf("got %d skills, want 1", len(skills))
	}
}

func Test_FindByName(t *testing.T) {
	idx := sampleIndex()

	found := registry.FindByName(idx.Entries, "code-reviewer")
	if found == nil {
		t.Fatal("expected to find code-reviewer")
	}
	if found.Type != "agent" {
		t.Errorf("type = %q, want %q", found.Type, "agent")
	}

	notFound := registry.FindByName(idx.Entries, "nonexistent")
	if notFound != nil {
		t.Error("expected nil for nonexistent entry")
	}
}

func Test_DownloadFile(t *testing.T) {
	content := "---\nname: test-agent\n---\n\nYou are a test agent."
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(content))
	}))
	defer srv.Close()

	data, err := registry.DownloadFile(srv.URL)
	if err != nil {
		t.Fatalf("DownloadFile failed: %v", err)
	}
	if string(data) != content {
		t.Errorf("got %q, want %q", string(data), content)
	}
}
