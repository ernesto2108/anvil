package registry

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Entry represents a single agent or skill available in a remote registry.
type Entry struct {
	Name        string `json:"name"`
	Type        string `json:"type"` // "agent" or "skill"
	Description string `json:"description"`
	Version     string `json:"version"`
	URL         string `json:"url"` // raw URL to download the file or directory
	Author      string `json:"author,omitempty"`
}

// Index represents the full registry index fetched from a remote URL.
type Index struct {
	Name    string  `json:"name"`
	Entries []Entry `json:"entries"`
}

// Fetch downloads and parses a registry index from the given URL.
func Fetch(url string) (*Index, error) {
	client := &http.Client{Timeout: 15 * time.Second}

	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetch registry: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("registry returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read registry body: %w", err)
	}

	var idx Index
	if err := json.Unmarshal(body, &idx); err != nil {
		return nil, fmt.Errorf("parse registry JSON: %w", err)
	}

	return &idx, nil
}

// DownloadFile fetches a single file from a URL and returns its content.
func DownloadFile(url string) ([]byte, error) {
	client := &http.Client{Timeout: 30 * time.Second}

	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download returned status %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

// FilterByType returns entries matching the given type ("agent" or "skill").
func FilterByType(entries []Entry, typ string) []Entry {
	var result []Entry
	for _, e := range entries {
		if e.Type == typ {
			result = append(result, e)
		}
	}
	return result
}

// FindByName returns the first entry matching the given name, or nil.
func FindByName(entries []Entry, name string) *Entry {
	for _, e := range entries {
		if e.Name == name {
			return &e
		}
	}
	return nil
}
