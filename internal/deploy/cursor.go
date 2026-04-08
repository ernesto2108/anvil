package deploy

import (
	"fmt"
	"path/filepath"

	"github.com/ernesto2108/anvil/pkg/config"
	"github.com/ernesto2108/anvil/pkg/fileutil"
)

func Cursor(cfg *config.App) {
	projects := cfg.ExpandedCursorProjects()
	if len(projects) == 0 {
		return
	}

	for _, proj := range projects {
		if !fileutil.IsDir(proj) {
			continue
		}

		ts := AddTarget(fmt.Sprintf("Cursor (%s)", filepath.Base(proj)))

		rulesDir := filepath.Join(proj, ".cursor", "rules")
		deployAgents(cfg, rulesDir, ts, adaptCursor)

		claudeMD := filepath.Join(cfg.RepoDir, config.FileClaudeMD)
		agentsMD := filepath.Join(proj, config.FileAgentsMD)
		if fileutil.Exists(claudeMD) && !fileutil.Exists(agentsMD) {
			fileutil.CopyFile(claudeMD, agentsMD)
			ts.Extras = append(ts.Extras, config.FileAgentsMD+" -> created")
		}
	}
}

// adaptCursor formats an agent file for the Cursor target.
func adaptCursor(_ *config.App, agent AgentData) string {
	desc := agent.Doc.Fields["description"]
	return fmt.Sprintf("---\ndescription: \"%s\"\nmanaged-by: anvil\nalwaysApply: false\n---\n\n%s", desc, agent.Doc.Body)
}
