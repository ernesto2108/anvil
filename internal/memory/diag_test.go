package memory

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func diagEmbed(text string) ([]float32, error) {
	body, _ := json.Marshal(map[string]string{
		"model": "nomic-embed-text",
		"input": text,
	})
	resp, err := http.Post("http://localhost:11434/api/embed", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var r struct {
		Embeddings [][]float64 `json:"embeddings"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, err
	}
	if len(r.Embeddings) == 0 {
		return nil, fmt.Errorf("empty embeddings")
	}
	out := make([]float32, len(r.Embeddings[0]))
	for i, v := range r.Embeddings[0] {
		out[i] = float32(v)
	}
	return out, nil
}

// TestDiag_MemorySimilarity is a one-off diagnostic. Run with:
//
//	go test -run TestDiag_MemorySimilarity -v ./internal/memory/...
//
// Skipped unless ANVIL_DIAG=1 to keep CI clean.
func TestDiag_MemorySimilarity(t *testing.T) {
	if os.Getenv("ANVIL_DIAG") != "1" {
		t.Skip("set ANVIL_DIAG=1 to run")
	}

	home, _ := os.UserHomeDir()
	dbPath := filepath.Join(home, ".anvil", "runs.db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	digests, err := ListDigestsByProject(context.Background(), db, "anvil")
	if err != nil {
		t.Fatalf("list digests: %v", err)
	}
	fmt.Printf("\nFound %d digests for project 'anvil'\n", len(digests))

	queryVec, err := diagEmbed("agregar pruebas unitarias a internal/memory")
	if err != nil {
		t.Fatalf("embed query: %v", err)
	}
	fmt.Printf("Query vec dim: %d\n\n", len(queryVec))

	for _, d := range digests {
		if d.Embedding == nil {
			fmt.Printf("[no embedding] %s\n", d.ID)
			continue
		}
		score := CosineSimilarity(queryVec, d.Embedding)
		fmt.Printf("score=%.4f | dim=%d | %s | %.80s\n", score, len(d.Embedding), d.ID, d.Summary)
	}
}
