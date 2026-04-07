package deploy

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ernesto2108/anvil/pkg/config"
	"github.com/ernesto2108/anvil/pkg/fileutil"
	"github.com/ernesto2108/anvil/pkg/frontmatter"
	"github.com/ernesto2108/anvil/pkg/output"
)

func Cursor(cfg *config.App) {
	projects := cfg.ExpandedCursorProjects()
	if len(projects) == 0 {
		output.Info("%s -> skipped (no cursor_projects configured)", output.Bold("Cursor"))
		return
	}

	output.Info("%s -> %d project(s)", output.Bold("Cursor"), len(projects))

	files := AgentFiles(cfg.RepoDir)

	for _, proj := range projects {
		if !fileutil.IsDir(proj) {
			output.Warn("  %s -> directory not found, skipping", proj)
			continue
		}

		rulesDir := filepath.Join(proj, ".cursor", "rules")
		os.MkdirAll(rulesDir, 0o755)

		count := 0
		for _, f := range files {
			name := filepath.Base(f)
			data, err := os.ReadFile(f)
			if err != nil {
				continue
			}

			doc := frontmatter.Parse(string(data))
			desc := doc.Fields["description"]

			adapted := fmt.Sprintf("---\ndescription: \"%s\"\nalwaysApply: false\n---\n\n%s", desc, doc.Body)

			os.WriteFile(filepath.Join(rulesDir, name), []byte(adapted), 0o644)
			count++
		}
		output.Info("  %s -> %d rules", proj, count)

		claudeMD := filepath.Join(cfg.RepoDir, config.FileClaudeMD)
		agentsMD := filepath.Join(proj, config.FileAgentsMD)
		if fileutil.Exists(claudeMD) && !fileutil.Exists(agentsMD) {
			fileutil.CopyFile(claudeMD, agentsMD)
			output.Info("  %s -> %s created", proj, config.FileAgentsMD)
		}
	}
}
