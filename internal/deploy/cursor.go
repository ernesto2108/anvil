package deploy

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ernesto2108/anvil/pkg/config"
	"github.com/ernesto2108/anvil/pkg/fileutil"
	"github.com/ernesto2108/anvil/pkg/frontmatter"
)

func Cursor(cfg *config.App) {
	projects := cfg.ExpandedCursorProjects()
	if len(projects) == 0 {
		return
	}

	files := FilteredAgentFiles(cfg)

	for _, proj := range projects {
		if !fileutil.IsDir(proj) {
			continue
		}

		ts := AddTarget(fmt.Sprintf("Cursor (%s)", filepath.Base(proj)))

		rulesDir := filepath.Join(proj, ".cursor", "rules")
		os.MkdirAll(rulesDir, 0o755)

		result := ResolveCollisions(files, rulesDir, stdinReader)
		CleanManagedFiles(rulesDir)
		skip, _, renames := ApplyResolutions(result.Resolutions)

		for _, f := range files {
			name := filepath.Base(f)
			if !ShouldDeployFile(name, skip) {
				ts.Agents.Preserved++
				continue
			}

			data, err := os.ReadFile(f)
			if err != nil {
				continue
			}

			doc := frontmatter.Parse(string(data))
			desc := doc.Fields["description"]

			deployName := GetDeployName(name, renames)
			adapted := fmt.Sprintf("---\ndescription: \"%s\"\nmanaged-by: anvil\nalwaysApply: false\n---\n\n%s", desc, doc.Body)

			os.WriteFile(filepath.Join(rulesDir, deployName), []byte(adapted), 0o644)
			ts.Agents.Deployed++
		}

		claudeMD := filepath.Join(cfg.RepoDir, config.FileClaudeMD)
		agentsMD := filepath.Join(proj, config.FileAgentsMD)
		if fileutil.Exists(claudeMD) && !fileutil.Exists(agentsMD) {
			fileutil.CopyFile(claudeMD, agentsMD)
			ts.Extras = append(ts.Extras, config.FileAgentsMD+" -> created")
		}
	}
}
