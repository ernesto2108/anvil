package memory

import (
	"bufio"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

// ParseHandoff extracts a DigestDraft from a handoff markdown file. The handoff
// format is defined in skills/handoff/template.md — sections like "## Decisiones
// tomadas", "### Edge cases descubiertos", and "## Notas" map directly to the
// Digest schema, avoiding an LLM round-trip for the structured fields.
//
// Returns the draft and the extracted TASK-ID (or filename slug if no TASK-ID
// is present). The caller is responsible for combining this with an embedding
// and persisting via UpsertDigest.
//
// The parser is deliberately lenient: missing sections produce empty slices,
// not errors. Only a completely empty body returns an error.
func ParseHandoff(content, sourcePath string) (DigestDraft, string, error) {
	if strings.TrimSpace(content) == "" {
		return DigestDraft{}, "", fmt.Errorf("handoff: empty content")
	}

	taskID := extractTaskID(content, sourcePath)
	title := extractH1Title(content)

	decisions := extractListSection(content, "## Decisiones tomadas")
	edgeCases := extractListSection(content, "### Edge cases descubiertos")
	if len(edgeCases) == 0 {
		// Some handoffs put edge cases under "Edge cases descubiertos" without ###
		edgeCases = extractListSection(content, "## Edge cases descubiertos")
	}
	errors := extractErrorsFromNotes(content)

	summary := buildSummary(title, taskID, decisions, sourcePath)

	return DigestDraft{
		Summary:   summary,
		Decisions: decisions,
		EdgeCases: edgeCases,
		Errors:    errors,
	}, taskID, nil
}

// taskIDPattern matches IDs like MCP-001, DASH-FEAT-007, MCP-001-F7, BLT-123.
var taskIDPattern = regexp.MustCompile(`\b[A-Z]{2,6}(?:-[A-Z0-9]{1,8}){1,3}\b`)

// extractTaskID looks for a TASK-ID pattern in the H1 title or filename.
// Falls back to the filename slug if no pattern is found.
func extractTaskID(content, sourcePath string) string {
	if h1 := extractH1Title(content); h1 != "" {
		if m := taskIDPattern.FindString(h1); m != "" {
			return m
		}
	}
	base := strings.TrimSuffix(filepath.Base(sourcePath), filepath.Ext(sourcePath))
	if m := taskIDPattern.FindString(base); m != "" {
		return m
	}
	if base == "" {
		return "unknown"
	}
	return base
}

// extractH1Title returns the text of the first H1 (`# ...`), trimmed and with
// any leading "Handoff:" / "Handoff —" prefix removed.
func extractH1Title(content string) string {
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "# ") && !strings.HasPrefix(line, "## ") {
			title := strings.TrimSpace(strings.TrimPrefix(line, "# "))
			title = strings.TrimPrefix(title, "Handoff:")
			title = strings.TrimPrefix(title, "Handoff —")
			title = strings.TrimPrefix(title, "Handoff -")
			return strings.TrimSpace(title)
		}
	}
	return ""
}

// extractListSection returns the bullet/numbered items inside a section header.
// The section ends at the next header at the same level or shallower (`## ` or
// `# `). Empty bullets are skipped. Sub-bullets (indented) are concatenated to
// their parent on a single line for compact embeddings.
func extractListSection(content, header string) []string {
	headerLevel := 0
	for _, c := range header {
		if c == '#' {
			headerLevel++
		} else {
			break
		}
	}
	if headerLevel == 0 {
		return nil
	}

	scanner := bufio.NewScanner(strings.NewReader(content))
	var (
		inSection bool
		items     []string
		current   strings.Builder
	)

	flush := func() {
		s := strings.TrimSpace(current.String())
		if s != "" {
			items = append(items, s)
		}
		current.Reset()
	}

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if !inSection {
			if trimmed == header {
				inSection = true
			}
			continue
		}

		// End of section: another header at same or shallower level.
		if isHeaderAtOrAbove(trimmed, headerLevel) {
			flush()
			break
		}

		// Top-level bullet/numbered item: starts at column 0-1 with `-`, `*`, or `N.`
		if isTopLevelItem(line) {
			flush()
			current.WriteString(stripBulletPrefix(trimmed))
			continue
		}

		// Continuation: sub-bullet or indented prose. Append to current item.
		if strings.HasPrefix(line, "  ") || strings.HasPrefix(line, "\t") {
			if trimmed != "" && current.Len() > 0 {
				current.WriteString(" ")
				current.WriteString(stripBulletPrefix(trimmed))
			}
		}
	}
	flush()
	return items
}

// isHeaderAtOrAbove returns true if the line is a markdown header at the given
// level or shallower (e.g. for level 2, both `## ...` and `# ...` end the section).
func isHeaderAtOrAbove(line string, level int) bool {
	if !strings.HasPrefix(line, "#") {
		return false
	}
	for i := 1; i <= level; i++ {
		prefix := strings.Repeat("#", i) + " "
		if strings.HasPrefix(line, prefix) {
			return true
		}
	}
	return false
}

// numberedItemPattern matches list items like "1.", "12.", "1)" at the start of a line.
var numberedItemPattern = regexp.MustCompile(`^\s*\d+[.)]\s`)

// isTopLevelItem returns true if the line is a top-level bullet (`-`, `*`) or
// numbered item (`1.`, `2.`) with no leading spaces.
func isTopLevelItem(line string) bool {
	if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") {
		return true
	}
	return numberedItemPattern.MatchString(line) && !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t")
}

// stripBulletPrefix removes the leading `-`, `*`, or `N.` from a list item.
func stripBulletPrefix(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "- ") {
		return s[2:]
	}
	if strings.HasPrefix(s, "* ") {
		return s[2:]
	}
	if loc := numberedItemPattern.FindStringIndex(s); loc != nil && loc[0] == 0 {
		return strings.TrimSpace(s[loc[1]:])
	}
	return s
}

// errorKeywords identifies notes that describe a real error/blocker rather
// than informational notes.
var errorKeywords = []string{"bug", "blocker", "broke", "broken", "fail", "regression", "crash", "panic", "error:"}

// extractErrorsFromNotes returns bullets from "## Notas" that mention error-like
// keywords. The full note section is too noisy to ingest as Errors — most
// "notes" are workarounds or context, not errors. Filtering by keyword keeps
// only the genuinely problematic items.
func extractErrorsFromNotes(content string) []string {
	all := extractListSection(content, "## Notas")
	if len(all) == 0 {
		return nil
	}
	var out []string
	for _, item := range all {
		lower := strings.ToLower(item)
		for _, kw := range errorKeywords {
			if strings.Contains(lower, kw) {
				out = append(out, item)
				break
			}
		}
	}
	return out
}

// buildSummary composes the embedded summary text. Format:
//
//	<Title>. Decisiones clave: <first 2-3 decisions joined with "; ">.
//
// The title carries the TASK-ID and topic; the decisions provide semantic
// signal for similarity search beyond keyword matches on the title alone.
func buildSummary(title, taskID string, decisions []string, sourcePath string) string {
	if title == "" {
		title = taskID
	}
	if title == "" || title == "unknown" {
		// Last resort: use the filename.
		base := strings.TrimSuffix(filepath.Base(sourcePath), filepath.Ext(sourcePath))
		if base != "" {
			title = base
		} else {
			title = "handoff"
		}
	}

	if len(decisions) == 0 {
		return title + "."
	}

	keep := 3
	if len(decisions) < keep {
		keep = len(decisions)
	}
	parts := make([]string, 0, keep)
	for _, d := range decisions[:keep] {
		// Trim long decisions to keep the summary readable for embeddings.
		d = strings.TrimSpace(d)
		if len(d) > 200 {
			d = d[:197] + "..."
		}
		parts = append(parts, d)
	}
	return title + ". Decisiones clave: " + strings.Join(parts, "; ") + "."
}
