package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ernesto2108/anvil/internal/memory"
	"github.com/ernesto2108/anvil/pkg/config"
	"github.com/ernesto2108/anvil/pkg/output"
)

// exportFlags holds parsed flags for the export command.
type exportFlags struct {
	project string
	format  string
	outDir  string
}

func parseExportFlags(args []string) (exportFlags, []string) {
	f := exportFlags{format: "md"}
	var rest []string
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--project", "-p":
			if i+1 < len(args) {
				i++
				f.project = args[i]
			}
		case "--format", "-f":
			if i+1 < len(args) {
				i++
				f.format = args[i]
			}
		case "--output", "-o":
			if i+1 < len(args) {
				i++
				f.outDir = args[i]
			}
		default:
			rest = append(rest, args[i])
		}
	}
	return f, rest
}

func cmdExport(cfg *config.App, args []string) {
	flags, rest := parseExportFlags(args)

	if flags.format != "md" && flags.format != "json" {
		output.Error("--format must be 'md' or 'json', got: %s", flags.format)
		os.Exit(1)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		output.Error("resolve home dir: %s", err)
		os.Exit(1)
	}
	dbPath := filepath.Join(home, ".anvil", "runs.db")

	db, err := openDashboardDB(dbPath, false)
	if err != nil {
		output.Error("open store: %s", err)
		os.Exit(1)
	}
	defer func() { _ = db.Close() }()

	ctx := context.Background()

	// Single run-id export.
	if len(rest) > 0 {
		runID := rest[0]
		d, err := memory.GetDigestByRunID(ctx, db, runID)
		if err != nil {
			output.Error("get digest: %s", err)
			os.Exit(1)
		}
		if d == nil {
			output.Warn("No digest found for run %s", runID)
			return
		}
		fmt.Print(renderDigest(*d, flags.format))
		return
	}

	// Project export.
	if flags.project == "" {
		workDir, _ := os.Getwd()
		flags.project = filepath.Base(workDir)
	}

	digests, err := memory.ListDigestsByProject(ctx, db, flags.project)
	if err != nil {
		output.Error("list digests: %s", err)
		os.Exit(1)
	}

	if len(digests) == 0 {
		output.Info("No digests found for project %s", flags.project)
		return
	}

	if flags.outDir != "" {
		if err := os.MkdirAll(flags.outDir, 0o755); err != nil {
			output.Error("create output dir: %s", err)
			os.Exit(1)
		}
		for _, d := range digests {
			ext := flags.format
			filename := filepath.Join(flags.outDir, d.RunID+"."+ext)
			content := renderDigest(d, flags.format)
			if err := os.WriteFile(filename, []byte(content), 0o644); err != nil {
				output.Error("write %s: %s", filename, err)
				os.Exit(1)
			}
			output.Info("Wrote %s", filename)
		}
		return
	}

	// Print all to stdout separated by ---
	for i, d := range digests {
		if i > 0 {
			fmt.Println("---")
		}
		fmt.Print(renderDigest(d, flags.format))
	}
}

// renderDigest formats a digest as Markdown or JSON.
func renderDigest(d memory.Digest, format string) string {
	switch format {
	case "json":
		return renderDigestJSON(d)
	default:
		return renderDigestMD(d)
	}
}

func renderDigestMD(d memory.Digest) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# Digest: %s\n\n", d.RunID)
	fmt.Fprintf(&b, "**Project:** %s\n", d.Project)
	fmt.Fprintf(&b, "**Model:** %s\n", d.ModelUsed)
	fmt.Fprintf(&b, "**Created:** %s\n", d.CreatedAt.Format("2006-01-02 15:04:05 UTC"))
	fmt.Fprintf(&b, "\n## Summary\n\n%s\n", d.Summary)

	if len(d.Decisions) > 0 {
		b.WriteString("\n## Decisions\n\n")
		for _, dec := range d.Decisions {
			fmt.Fprintf(&b, "- %s\n", dec)
		}
	}

	if len(d.EdgeCases) > 0 {
		b.WriteString("\n## Edge Cases\n\n")
		for _, ec := range d.EdgeCases {
			fmt.Fprintf(&b, "- %s\n", ec)
		}
	}

	if len(d.Errors) > 0 {
		b.WriteString("\n## Errors\n\n")
		for _, e := range d.Errors {
			fmt.Fprintf(&b, "- %s\n", e)
		}
	}

	b.WriteString("\n")
	return b.String()
}

// digestJSON is a JSON-friendly view of Digest that omits the embedding.
type digestJSON struct {
	ID        string   `json:"id"`
	RunID     string   `json:"run_id"`
	Project   string   `json:"project"`
	Summary   string   `json:"summary"`
	Decisions []string `json:"decisions"`
	EdgeCases []string `json:"edge_cases"`
	Errors    []string `json:"errors"`
	ModelUsed string   `json:"model_used"`
	CreatedAt string   `json:"created_at"`
	UpdatedAt string   `json:"updated_at"`
}

func renderDigestJSON(d memory.Digest) string {
	v := digestJSON{
		ID:        d.ID,
		RunID:     d.RunID,
		Project:   d.Project,
		Summary:   d.Summary,
		Decisions: d.Decisions,
		EdgeCases: d.EdgeCases,
		Errors:    d.Errors,
		ModelUsed: d.ModelUsed,
		CreatedAt: d.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: d.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		output.Warn("marshal digest: %s", err)
		return "{}"
	}
	return string(b) + "\n"
}
