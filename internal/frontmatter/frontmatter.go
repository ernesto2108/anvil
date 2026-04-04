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
