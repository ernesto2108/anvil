package frontmatter

import (
	"fmt"
	"strings"
)

type Document struct {
	Fields map[string]string
	Body   string
}

func Parse(content string) Document {
	doc := Document{Fields: make(map[string]string)}

	lines := strings.Split(content, "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		doc.Body = content
		return doc
	}

	closingIdx := -1
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			closingIdx = i
			break
		}
	}

	if closingIdx == -1 {
		doc.Body = content
		return doc
	}

	for _, line := range lines[1:closingIdx] {
		key, val, found := strings.Cut(line, ":")
		if !found {
			continue
		}
		doc.Fields[strings.TrimSpace(key)] = strings.TrimSpace(val)
	}

	if closingIdx+1 < len(lines) {
		body := strings.Join(lines[closingIdx+1:], "\n")
		doc.Body = strings.TrimLeft(body, "\n")
	}

	return doc
}

func Get(content, key string) string {
	doc := Parse(content)
	return doc.Fields[key]
}

func ReplaceField(content, key, oldVal, newVal string) string {
	old := fmt.Sprintf("%s: %s", key, oldVal)
	repl := fmt.Sprintf("%s: %s", key, newVal)
	return strings.Replace(content, old, repl, 1)
}

// HasField returns true if the frontmatter contains the given key with the given value.
func HasField(content, key, value string) bool {
	doc := Parse(content)
	return doc.Fields[key] == value
}

// SetField adds or updates a field in the frontmatter block.
// If the content has no frontmatter, it wraps it with one.
func SetField(content, key, value string) string {
	entry := fmt.Sprintf("%s: %s", key, value)
	lines := strings.Split(content, "\n")

	// No frontmatter — create one
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return fmt.Sprintf("---\n%s\n---\n%s", entry, content)
	}

	// Find closing ---
	closingIdx := -1
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			closingIdx = i
			break
		}
	}
	if closingIdx == -1 {
		return fmt.Sprintf("---\n%s\n---\n%s", entry, content)
	}

	// Check if key already exists — update it
	for i := 1; i < closingIdx; i++ {
		k, _, found := strings.Cut(lines[i], ":")
		if found && strings.TrimSpace(k) == key {
			lines[i] = entry
			return strings.Join(lines, "\n")
		}
	}

	// Key not found — insert before closing ---
	newLines := make([]string, 0, len(lines)+1)
	newLines = append(newLines, lines[:closingIdx]...)
	newLines = append(newLines, entry)
	newLines = append(newLines, lines[closingIdx:]...)
	return strings.Join(newLines, "\n")
}
