package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/ernesto2108/anvil/pkg/config"
	"github.com/ernesto2108/anvil/pkg/gitutil"
	"github.com/ernesto2108/anvil/pkg/output"
	"github.com/ernesto2108/anvil/pkg/state"
)

func cmdDiff(cfg *config.App, git *gitutil.Repo, args []string) {
	st := loadState(cfg)

	if st.DeployedSHA == "" || st.DeployedSHA == state.StateNone {
		output.Error("Nothing deployed yet")
		os.Exit(1)
	}

	current := git.CurrentSHA()
	if st.DeployedSHA == current {
		output.Info("No changes since last deploy")
		return
	}

	output.Info("Changes since deploy (%s -> %s):", st.DeployedSHA, current)
	fmt.Println()

	// Filter by component if args provided, otherwise show all
	paths := []string{config.CompAgents + "/", config.CompSkills + "/", config.CompCommands + "/", config.FileClaudeMD}
	if len(args) > 0 {
		paths = make([]string, len(args))
		for i, a := range args {
			// Ensure trailing slash for directories
			if !strings.Contains(a, ".") && !strings.HasSuffix(a, "/") {
				paths[i] = a + "/"
			} else {
				paths[i] = a
			}
		}
	}

	diffOut, err := git.DiffStat(st.DeployedSHA, paths...)
	if err != nil {
		output.Error("diff: %s", err)
		return
	}
	if diffOut == "" {
		output.Info("No changes in %s", strings.Join(args, ", "))
		return
	}
	fmt.Println(diffOut)
}
