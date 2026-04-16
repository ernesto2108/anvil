package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ernesto2108/anvil/internal/memory"
	"github.com/ernesto2108/anvil/pkg/config"
	"github.com/ernesto2108/anvil/pkg/output"
)

// digestsFlags holds parsed flags for the digests command.
type digestsFlags struct {
	project string
	last    int
}

func parseDigestsFlags(args []string) digestsFlags {
	f := digestsFlags{last: 10}
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--project", "-p":
			if i+1 < len(args) {
				i++
				f.project = args[i]
			}
		case "--last", "-n":
			if i+1 < len(args) {
				i++
				fmt.Sscanf(args[i], "%d", &f.last) //nolint:errcheck
			}
		}
	}
	return f
}

func cmdDigests(cfg *config.App, args []string) {
	// Subcommand dispatch: `anvil digests show <run_id>` prints the full
	// digest body (summary + decisions + edge cases + errors) so the user
	// can see exactly what would be injected as memory in a future run.
	if len(args) > 0 && args[0] == "show" {
		cmdDigestsShow(args[1:])
		return
	}

	flags := parseDigestsFlags(args)

	if flags.project == "" {
		workDir, _ := os.Getwd()
		flags.project = filepath.Base(workDir)
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

	digests, err := memory.ListDigestsByProject(ctx, db, flags.project)
	if err != nil {
		output.Error("list digests: %s", err)
		os.Exit(1)
	}

	if len(digests) == 0 {
		output.Info("No digests found for project %s", flags.project)
		return
	}

	// Apply --last limit.
	if flags.last > 0 && len(digests) > flags.last {
		digests = digests[:flags.last]
	}

	output.Info("Digests for project %s (%d shown)", flags.project, len(digests))
	fmt.Println()
	fmt.Printf("  %-30s  %-19s  %s\n", "RUN_ID", "CREATED", "SUMMARY")
	fmt.Printf("  %-30s  %-19s  %s\n", "------------------------------", "-------------------", "------------------------------------------------------------")

	for _, d := range digests {
		runID := d.RunID
		if len(runID) > 30 {
			runID = runID[:27] + "..."
		}
		created := d.CreatedAt.Format("2006-01-02 15:04:05")
		summary := d.Summary
		if len(summary) > 60 {
			summary = summary[:57] + "..."
		}
		fmt.Printf("  %-30s  %-19s  %s\n", runID, created, summary)
	}
	fmt.Println()
}

// cmdDigestsShow prints the full body of a single digest, identified by run_id.
// Useful for inspecting what gets injected as memory into the next run.
func cmdDigestsShow(args []string) {
	if len(args) == 0 {
		output.Error("usage: anvil digests show <run_id>")
		os.Exit(1)
	}
	runID := args[0]

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

	d, err := memory.GetDigestByRunID(context.Background(), db, runID)
	if err != nil {
		output.Error("get digest: %s", err)
		os.Exit(1)
	}
	if d == nil {
		output.Error("No digest found for run %s", runID)
		os.Exit(1)
	}

	embStatus := "absent"
	if len(d.Embedding) > 0 {
		embStatus = fmt.Sprintf("%d dims", len(d.Embedding))
	}

	fmt.Println()
	fmt.Printf("  Digest:    %s\n", d.ID)
	fmt.Printf("  Run:       %s\n", d.RunID)
	fmt.Printf("  Project:   %s\n", d.Project)
	fmt.Printf("  Model:     %s\n", d.ModelUsed)
	fmt.Printf("  Created:   %s\n", d.CreatedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("  Embedding: %s\n", embStatus)
	fmt.Println()
	fmt.Println("  Summary:")
	fmt.Printf("    %s\n", d.Summary)
	if len(d.Decisions) > 0 {
		fmt.Println()
		fmt.Println("  Decisions:")
		for _, x := range d.Decisions {
			fmt.Printf("    - %s\n", x)
		}
	}
	if len(d.EdgeCases) > 0 {
		fmt.Println()
		fmt.Println("  Edge cases:")
		for _, x := range d.EdgeCases {
			fmt.Printf("    - %s\n", x)
		}
	}
	if len(d.Errors) > 0 {
		fmt.Println()
		fmt.Println("  Errors:")
		for _, x := range d.Errors {
			fmt.Printf("    - %s\n", x)
		}
	}
	fmt.Println()
}
