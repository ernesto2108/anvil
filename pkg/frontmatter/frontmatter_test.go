package frontmatter_test

import (
	"testing"

	"github.com/ernesto2108/anvil/pkg/frontmatter"
)

func Test_Parse(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantFields map[string]string
		wantBody   string
	}{
		{
			name:       "full document",
			input:      "---\nmodel: high\npermission: execute\n---\n\nBody content here.",
			wantFields: map[string]string{"model": "high", "permission": "execute"},
			wantBody:   "Body content here.",
		},
		{
			name:       "no frontmatter",
			input:      "Just plain text",
			wantFields: map[string]string{},
			wantBody:   "Just plain text",
		},
		{
			name:       "empty body",
			input:      "---\nkey: value\n---\n",
			wantFields: map[string]string{"key": "value"},
			wantBody:   "",
		},
		{
			name:       "unclosed frontmatter",
			input:      "---\nkey: value\nno closing",
			wantFields: map[string]string{},
			wantBody:   "---\nkey: value\nno closing",
		},
		{
			name:       "empty input",
			input:      "",
			wantFields: map[string]string{},
			wantBody:   "",
		},
		{
			name:       "value with spaces",
			input:      "---\ndescription: A long description here\n---\n\nBody",
			wantFields: map[string]string{"description": "A long description here"},
			wantBody:   "Body",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := frontmatter.Parse(tt.input)

			for k, want := range tt.wantFields {
				got, ok := doc.Fields[k]
				if !ok {
					t.Errorf("missing field %q", k)
					continue
				}
				if got != want {
					t.Errorf("field %q = %q, want %q", k, got, want)
				}
			}

			if len(doc.Fields) != len(tt.wantFields) {
				t.Errorf("got %d fields, want %d", len(doc.Fields), len(tt.wantFields))
			}

			if doc.Body != tt.wantBody {
				t.Errorf("body = %q, want %q", doc.Body, tt.wantBody)
			}
		})
	}
}

func Test_Get(t *testing.T) {
	content := "---\nmodel: high\npermission: write\n---\n\nBody"

	tests := []struct {
		name string
		key  string
		want string
	}{
		{name: "existing key", key: "model", want: "high"},
		{name: "another key", key: "permission", want: "write"},
		{name: "missing key", key: "nonexistent", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := frontmatter.Get(content, tt.key)
			if got != tt.want {
				t.Errorf("Get(%q) = %q, want %q", tt.key, got, tt.want)
			}
		})
	}
}

func Test_ReplaceField(t *testing.T) {
	tests := []struct {
		name    string
		content string
		key     string
		oldVal  string
		newVal  string
		want    string
	}{
		{
			name:    "replace model tier",
			content: "---\nmodel: high\n---\n",
			key:     "model",
			oldVal:  "high",
			newVal:  "opus",
			want:    "---\nmodel: opus\n---\n",
		},
		{
			name:    "no match leaves unchanged",
			content: "---\nmodel: high\n---\n",
			key:     "model",
			oldVal:  "low",
			newVal:  "haiku",
			want:    "---\nmodel: high\n---\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := frontmatter.ReplaceField(tt.content, tt.key, tt.oldVal, tt.newVal)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func Test_HasField(t *testing.T) {
	content := "---\nmanaged-by: anvil\nmodel: high\n---\n\nBody"

	if !frontmatter.HasField(content, "managed-by", "anvil") {
		t.Error("HasField should return true for managed-by: anvil")
	}
	if frontmatter.HasField(content, "managed-by", "other") {
		t.Error("HasField should return false for managed-by: other")
	}
	if frontmatter.HasField(content, "nonexistent", "value") {
		t.Error("HasField should return false for nonexistent key")
	}
}

func Test_SetField(t *testing.T) {
	tests := []struct {
		name    string
		content string
		key     string
		value   string
		check   func(string) bool
		desc    string
	}{
		{
			name:    "add to existing frontmatter",
			content: "---\nmodel: high\n---\n\nBody",
			key:     "managed-by",
			value:   "anvil",
			check: func(s string) bool {
				return frontmatter.HasField(s, "managed-by", "anvil") &&
					frontmatter.Get(s, "model") == "high"
			},
			desc: "should have managed-by and preserve model",
		},
		{
			name:    "update existing field",
			content: "---\nmanaged-by: old\nmodel: high\n---\n\nBody",
			key:     "managed-by",
			value:   "anvil",
			check: func(s string) bool {
				return frontmatter.HasField(s, "managed-by", "anvil")
			},
			desc: "should update managed-by to anvil",
		},
		{
			name:    "add frontmatter to plain content",
			content: "Just plain text",
			key:     "managed-by",
			value:   "anvil",
			check: func(s string) bool {
				return frontmatter.HasField(s, "managed-by", "anvil")
			},
			desc: "should wrap content with frontmatter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := frontmatter.SetField(tt.content, tt.key, tt.value)
			if !tt.check(got) {
				t.Errorf("%s: got %q", tt.desc, got)
			}
		})
	}
}
