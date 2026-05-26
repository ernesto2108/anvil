package mcp

import (
	"bufio"
	"bytes"
	"context"
	cryptoRand "crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/ernesto2108/anvil/internal/memory"
	"github.com/ernesto2108/anvil/internal/memory/ollama"
)

// digestFromHandoff parses a handoff markdown file and writes its content as a
// digest to the runs.db. This closes the gap where handoffs created via Claude
// Code (not the CLI runner) never produce memory — the orchestrator can call
// this tool from the /task-complete flow to ensure every closed task seeds the
// memory layer.
func (s *Server) digestFromHandoff(ctx context.Context, args map[string]any) (string, error) {
	db := s.db
	if db == nil {
		return "", fmt.Errorf("database not available")
	}

	path := strArg(args, "path", "")
	if path == "" {
		return "", fmt.Errorf("path is required")
	}

	workDir, _ := os.Getwd()
	project := strArg(args, "project", filepath.Base(workDir))

	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read handoff %q: %w", path, err)
	}

	draft, taskID, err := memory.ParseHandoff(string(content), path)
	if err != nil {
		return "", fmt.Errorf("parse handoff: %w", err)
	}

	runID := "handoff_" + taskID
	now := time.Now().UTC()

	if err := upsertSyntheticHandoffRun(ctx, db, runID, taskID, draft.Summary, project, now); err != nil {
		return "", fmt.Errorf("upsert synthetic run: %w", err)
	}

	// Auto-start Ollama if needed (KEEP_ALIVE=5m so models unload after idle).
	_ = ollama.EnsureRunning(ctx, "5m")

	// Best-effort embedding via Ollama.
	var embedding []float32
	embedder := ollama.NewClient("", "")
	if embedder.Healthy(ctx) {
		if emb, embErr := embedder.Embed(ctx, draft.Summary); embErr == nil {
			embedding = emb
		}
	}

	digestID, _ := newHandoffDigestID()
	d := memory.Digest{
		ID:        digestID,
		RunID:     runID,
		Project:   project,
		Summary:   draft.Summary,
		Decisions: draft.Decisions,
		EdgeCases: draft.EdgeCases,
		Errors:    draft.Errors,
		Embedding: embedding,
		ModelUsed: "handoff-parser",
		Source:    "handoff",
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := memory.UpsertDigest(ctx, db, d); err != nil {
		return "", fmt.Errorf("upsert digest: %w", err)
	}

	type result struct {
		TaskID    string `json:"task_id"`
		RunID     string `json:"run_id"`
		Project   string `json:"project"`
		Decisions int    `json:"decisions"`
		EdgeCases int    `json:"edge_cases"`
		Errors    int    `json:"errors"`
		Embedded  bool   `json:"embedded"`
	}

	r := result{
		TaskID:    taskID,
		RunID:     runID,
		Project:   project,
		Decisions: len(draft.Decisions),
		EdgeCases: len(draft.EdgeCases),
		Errors:    len(draft.Errors),
		Embedded:  embedding != nil,
	}
	out, _ := json.MarshalIndent(r, "", "  ")
	return string(out), nil
}

// upsertSyntheticHandoffRun inserts a placeholder runs row for handoff-sourced
// digests so the FK constraint on digests.run_id is satisfied. Marked with
// task_source='handoff' so list_runs and similar can distinguish them.
func upsertSyntheticHandoffRun(ctx context.Context, db *sql.DB, runID, taskID, summary, project string, now time.Time) error {
	taskDesc := summary
	if len(taskDesc) > 500 {
		taskDesc = taskDesc[:497] + "..."
	}
	const q = `INSERT INTO runs (id, task_id, task_desc, status, started_at, ended_at, project, task_source)
		VALUES (?, ?, ?, 'success', ?, ?, ?, 'handoff')
		ON CONFLICT(id) DO UPDATE SET
			task_desc = excluded.task_desc,
			ended_at  = excluded.ended_at,
			project   = excluded.project`
	_, err := db.ExecContext(ctx, q, runID, taskID, taskDesc, now.Format(time.RFC3339Nano), now.Format(time.RFC3339Nano), project)
	return err
}

// newHandoffDigestID mirrors newDigestIDForHandoff in cli — used here to keep
// the mcp package independent of cli.
func newHandoffDigestID() (string, error) {
	var b [6]byte
	if _, err := cryptoRand.Read(b[:]); err != nil {
		return "", err
	}
	return "dh_" + hex.EncodeToString(b[:]), nil
}

// searchMemories searches past run digests using semantic (vector KNN),
// keyword (BM25 FTS5), or hybrid (RRF fusion) mode.
func (s *Server) searchMemories(ctx context.Context, args map[string]any) (string, error) {
	db := s.db
	if db == nil {
		return "", fmt.Errorf("database not available")
	}

	query := strArg(args, "query", "")
	if query == "" {
		return "", fmt.Errorf("query is required")
	}
	workDir, _ := os.Getwd()
	project := strArg(args, "project", filepath.Base(workDir))
	limit := intArg(args, "limit", 3)
	if limit > 10 {
		limit = 10
	}
	mode := strArg(args, "mode", "hybrid")

	// keyword mode never needs Ollama — serve it immediately.
	if mode == "keyword" {
		results, err := memory.SearchKeyword(ctx, db, query, project, limit)
		if err != nil {
			return "", fmt.Errorf("search memories (keyword): %w", err)
		}
		return marshalSearchResults(results)
	}

	// Auto-start Ollama if needed (KEEP_ALIVE=5m so models unload after idle).
	// EnsureRunning is a no-op when the daemon is already up, so the cost is
	// only paid on the first call after a laptop restart.
	_ = ollama.EnsureRunning(ctx, "5m")

	// Use Ollama embedder (local) — same as the runner uses.
	emb := ollama.NewClient("", "")
	ollamaOK := emb.Healthy(ctx)

	if !ollamaOK {
		// Ollama unavailable: fall back to keyword search so we still return
		// something useful, or list recent digests if keyword yields nothing.
		results, err := memory.SearchKeyword(ctx, db, query, project, limit)
		if err != nil || len(results) == 0 {
			return s.recentDigestsFallback(ctx, db, project, limit)
		}
		return marshalSearchResults(results)
	}

	var (
		results []memory.SearchResult
		err     error
	)
	if mode == "semantic" {
		results, err = memory.SearchSimilar(ctx, db, emb, query, project, limit, 0.5)
	} else {
		// default: hybrid
		results, err = memory.SearchHybrid(ctx, db, emb, query, project, limit, 0.5)
	}
	if err != nil {
		return "", fmt.Errorf("search memories (%s): %w", mode, err)
	}

	return marshalSearchResults(results)
}

// marshalSearchResults serializes a slice of SearchResult to JSON.
func marshalSearchResults(results []memory.SearchResult) (string, error) {
	type hit struct {
		RunID     string   `json:"run_id"`
		Score     float64  `json:"score"`
		Summary   string   `json:"summary"`
		Decisions []string `json:"decisions,omitempty"`
		EdgeCases []string `json:"edge_cases,omitempty"`
		Errors    []string `json:"errors,omitempty"`
		CreatedAt string   `json:"created_at"`
	}

	hits := make([]hit, 0, len(results))
	for _, r := range results {
		hits = append(hits, hit{
			RunID:     r.Digest.RunID,
			Score:     r.Score,
			Summary:   r.Digest.Summary,
			Decisions: r.Digest.Decisions,
			EdgeCases: r.Digest.EdgeCases,
			Errors:    r.Digest.Errors,
			CreatedAt: r.Digest.CreatedAt.Format(time.RFC3339),
		})
	}

	b, _ := json.MarshalIndent(hits, "", "  ")
	return string(b), nil
}

// recentDigestsFallback returns the most recent digests when Ollama is unavailable.
func (s *Server) recentDigestsFallback(ctx context.Context, db *sql.DB, project string, limit int) (string, error) {
	digests, err := memory.ListDigestsByProject(ctx, db, project)
	if err != nil {
		return "", fmt.Errorf("list digests: %w", err)
	}

	if len(digests) > limit {
		digests = digests[:limit]
	}

	type item struct {
		RunID     string   `json:"run_id"`
		Summary   string   `json:"summary"`
		Decisions []string `json:"decisions,omitempty"`
		EdgeCases []string `json:"edge_cases,omitempty"`
		CreatedAt string   `json:"created_at"`
		Note      string   `json:"note"`
	}

	items := make([]item, 0, len(digests))
	for _, d := range digests {
		items = append(items, item{
			RunID:     d.RunID,
			Summary:   d.Summary,
			Decisions: d.Decisions,
			EdgeCases: d.EdgeCases,
			CreatedAt: d.CreatedAt.Format(time.RFC3339),
			Note:      "Ollama unavailable — results not ranked by similarity",
		})
	}

	b, _ := json.MarshalIndent(items, "", "  ")
	return string(b), nil
}

// getConventions loads coding conventions for a given stack and topic.
// It reads the skill dispatcher (SKILL.md) and returns the content of the
// files matching the topic keyword.
func (s *Server) getConventions(_ context.Context, args map[string]any) (string, error) {
	stack := strArg(args, "stack", "")
	if stack == "" {
		return "", fmt.Errorf("stack is required")
	}
	topic := strArg(args, "topic", "")

	// Normalize stack name — accept "go", "golang", "typescript", "ts", etc.
	stack = normalizeStack(stack)

	skillDir := filepath.Join(s.cfg.RepoDir, "skills", stack+"-conventions")
	if _, err := os.Stat(skillDir); os.IsNotExist(err) {
		return "", fmt.Errorf("no conventions found for stack %q — available: go, react, typescript, python, rust, flutter, astro, devops", stack)
	}

	if topic == "" {
		// No topic — return the SKILL.md dispatcher so the caller can navigate.
		data, err := os.ReadFile(filepath.Join(skillDir, "SKILL.md"))
		if err != nil {
			return "", fmt.Errorf("read SKILL.md: %w", err)
		}
		return string(data), nil
	}

	// Find files matching the topic by scanning the routing table in SKILL.md.
	matches := resolveConventionFiles(skillDir, topic)

	if len(matches) == 0 {
		// No routing match — try a directory/file name match as fallback.
		matches = globConventionFiles(skillDir, topic)
	}

	if len(matches) == 0 {
		// Last resort: return the full SKILL.md.
		data, _ := os.ReadFile(filepath.Join(skillDir, "SKILL.md"))
		return fmt.Sprintf("No files matched topic %q. Returning skill dispatcher:\n\n%s", topic, data), nil
	}

	const maxTotal = 10_000 // 10KB cap across all files
	var sb strings.Builder
	total := 0
	for _, f := range matches {
		data, err := os.ReadFile(f)
		if err != nil {
			continue
		}
		rel, _ := filepath.Rel(skillDir, f)
		fmt.Fprintf(&sb, "### %s\n\n", rel)
		chunk := string(data)
		if total+len(chunk) > maxTotal {
			chunk = chunk[:maxTotal-total]
			sb.WriteString(chunk)
			sb.WriteString("\n\n[truncated]")
			break
		}
		sb.WriteString(chunk)
		sb.WriteString("\n\n")
		total += len(chunk)
	}
	return sb.String(), nil
}

// resolveConventionFiles parses the routing table in SKILL.md and returns
// the file paths that match the given topic keyword.
func resolveConventionFiles(skillDir, topic string) []string {
	data, err := os.ReadFile(filepath.Join(skillDir, "SKILL.md"))
	if err != nil {
		return nil
	}

	topicLower := strings.ToLower(topic)
	var paths []string

	scanner := bufio.NewScanner(bytes.NewReader(data))
	inTable := false
	for scanner.Scan() {
		line := scanner.Text()
		// Detect markdown table header row.
		if strings.Contains(line, "| Working on") || strings.Contains(line, "| Task") || strings.Contains(line, "| Trabajando") {
			inTable = true
			continue
		}
		if !inTable {
			continue
		}
		// Skip separator row.
		if strings.HasPrefix(strings.TrimSpace(line), "|---") {
			continue
		}
		// End of table.
		if !strings.HasPrefix(strings.TrimSpace(line), "|") {
			inTable = false
			continue
		}
		cols := strings.Split(line, "|")
		if len(cols) < 3 {
			continue
		}
		description := strings.TrimSpace(cols[1])
		filePath := strings.TrimSpace(cols[2])

		if strings.Contains(strings.ToLower(description), topicLower) {
			resolved := resolveSkillPath(skillDir, filePath)
			paths = append(paths, resolved...)
		}
	}
	return paths
}

// resolveSkillPath resolves a path entry from the routing table.
// Path entries can be a single file (rules/coding.md) or a directory (guides/concurrency/).
func resolveSkillPath(skillDir, entry string) []string {
	entry = strings.TrimSpace(entry)
	full := filepath.Join(skillDir, entry)

	fi, err := os.Stat(full)
	if err != nil {
		return nil
	}
	if fi.IsDir() {
		var files []string
		_ = filepath.Walk(full, func(path string, info os.FileInfo, err error) error {
			if err == nil && !info.IsDir() && strings.HasSuffix(path, ".md") {
				files = append(files, path)
			}
			return nil
		})
		return files
	}
	return []string{full}
}

// globConventionFiles does a fuzzy search across the skill directory for files
// whose path contains the topic keyword.
func globConventionFiles(skillDir, topic string) []string {
	topicLower := strings.ToLower(topic)
	var matches []string
	_ = filepath.Walk(skillDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".md") {
			return nil
		}
		rel, _ := filepath.Rel(skillDir, path)
		if strings.Contains(strings.ToLower(rel), topicLower) {
			matches = append(matches, path)
		}
		return nil
	})
	return matches
}

// getBacklog reads the sprint backlog from the Obsidian vault.
func (s *Server) getBacklog(_ context.Context, args map[string]any) (string, error) {
	if s.vaultDir == "" {
		return "", fmt.Errorf("vault not configured — set ANVIL_VAULT_PATH env var to the knowledge base directory")
	}

	statusFilter := strings.ToLower(strArg(args, "status", ""))

	sprintFile := filepath.Join(s.vaultDir, "02-backlog", "sprint-current.md")
	data, err := os.ReadFile(sprintFile)
	if err != nil {
		return "", fmt.Errorf("read sprint backlog: %w", err)
	}

	if statusFilter == "" {
		return string(data), nil
	}

	// Filter markdown tables by section header matching the status.
	// Sections: ## Backlog, ## TODO, ## In Progress, ## Blocked, ## In Review, ## Done
	sectionMap := map[string]string{
		"backlog":     "## Backlog",
		"todo":        "## TODO",
		"in_progress": "## In Progress",
		"blocked":     "## Blocked",
		"done":        "## Done",
	}

	header, ok := sectionMap[statusFilter]
	if !ok {
		return string(data), nil
	}

	// Extract the matching section.
	lines := strings.Split(string(data), "\n")
	var out []string
	inSection := false
	for _, line := range lines {
		if strings.TrimSpace(line) == header {
			inSection = true
		} else if inSection && strings.HasPrefix(line, "## ") {
			break
		}
		if inSection {
			out = append(out, line)
		}
	}

	if len(out) == 0 {
		return fmt.Sprintf("No section %q found in sprint backlog.", header), nil
	}
	return strings.Join(out, "\n"), nil
}

// getTask reads the documents for a specific task from the vault.
func (s *Server) getTask(_ context.Context, args map[string]any) (string, error) {
	if s.vaultDir == "" {
		return "", fmt.Errorf("vault not configured — set ANVIL_VAULT_PATH env var to the knowledge base directory")
	}

	taskID := strArg(args, "task_id", "")
	if taskID == "" {
		return "", fmt.Errorf("task_id is required")
	}

	taskDir := filepath.Join(s.vaultDir, "03-tasks", taskID)
	if _, err := os.Stat(taskDir); os.IsNotExist(err) {
		return "", fmt.Errorf("task %q not found in vault", taskID)
	}

	// Read all .md files in the task directory.
	files := []string{"task.md", "prd.md", "architecture.md", "design.md"}
	const maxPerFile = 20_000

	var sb strings.Builder
	for _, fname := range files {
		path := filepath.Join(taskDir, fname)
		data, err := os.ReadFile(path)
		if err != nil {
			continue // file doesn't exist
		}
		fmt.Fprintf(&sb, "## %s\n\n", fname)
		content := string(data)
		if len(content) > maxPerFile {
			content = content[:maxPerFile] + "\n\n[truncated]"
		}
		sb.WriteString(content)
		sb.WriteString("\n\n---\n\n")
	}

	if sb.Len() == 0 {
		return fmt.Sprintf("Task %q found but contains no readable documents.", taskID), nil
	}
	return sb.String(), nil
}

// getRecentChanges combines git log with recent pipeline runs.
func (s *Server) getRecentChanges(_ context.Context, args map[string]any) (string, error) {
	days := intArg(args, "days", 1)
	if days > 7 {
		days = 7
	}

	workDir, _ := os.Getwd()
	project := strArg(args, "project", filepath.Base(workDir))

	since := fmt.Sprintf("%d days ago", days)

	// ── Git commits ──────────────────────────────────────────────────────────
	type commitInfo struct {
		Hash    string `json:"hash"`
		Message string `json:"message"`
		Author  string `json:"author"`
		Date    string `json:"date"`
	}

	var commits []commitInfo
	gitOut, err := exec.Command("git", "-C", workDir,
		"log", "--since="+since, "--pretty=format:%h|%s|%an|%ad", "--date=short").Output()
	if err == nil {
		for _, line := range strings.Split(strings.TrimSpace(string(gitOut)), "\n") {
			if line == "" {
				continue
			}
			parts := strings.SplitN(line, "|", 4)
			if len(parts) != 4 {
				continue
			}
			commits = append(commits, commitInfo{
				Hash:    parts[0],
				Message: parts[1],
				Author:  parts[2],
				Date:    parts[3],
			})
		}
	}

	// ── Recent runs ──────────────────────────────────────────────────────────
	type runInfo struct {
		ID           string `json:"id"`
		Task         string `json:"task"`
		Status       string `json:"status"`
		StartedAt    string `json:"started_at"`
		AgentsCount  int    `json:"agents_count"`
		FilesTouched int    `json:"files_touched"`
	}

	var runs []runInfo
	if s.db != nil {
		cutoff := time.Now().AddDate(0, 0, -days).Format(time.RFC3339)
		rows, err := s.db.QueryContext(context.Background(),
			`SELECT id, task_desc, status, started_at, agents_count, files_touched
			 FROM runs WHERE project = ? AND started_at >= ?
			 ORDER BY started_at DESC LIMIT 20`, project, cutoff)
		if err == nil {
			defer func() { _ = rows.Close() }()
			for rows.Next() {
				var r runInfo
				var task sql.NullString
				var agents, files sql.NullInt64
				if err := rows.Scan(&r.ID, &task, &r.Status, &r.StartedAt, &agents, &files); err != nil {
					continue
				}
				r.Task = task.String
				r.AgentsCount = int(agents.Int64)
				r.FilesTouched = int(files.Int64)
				runs = append(runs, r)
			}
		}
	}

	if commits == nil {
		commits = []commitInfo{}
	}
	if runs == nil {
		runs = []runInfo{}
	}

	result := map[string]any{
		"period":  fmt.Sprintf("last %d day(s)", days),
		"project": project,
		"commits": commits,
		"runs":    runs,
	}
	b, _ := json.MarshalIndent(result, "", "  ")
	return string(b), nil
}

// normalizeStack maps common aliases to the canonical skill directory name.
func normalizeStack(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	switch s {
	case "golang", "go":
		return "go"
	case "ts", "typescript":
		return "typescript"
	case "js", "node", "nodejs":
		return "typescript"
	case "dart":
		return "flutter"
	case "infrastructure", "infra", "ci", "cd", "cicd":
		return "devops"
	default:
		return s
	}
}
