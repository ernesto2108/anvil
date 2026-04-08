package cli

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/ernesto2108/anvil/pkg/config"
	"github.com/ernesto2108/anvil/pkg/gitutil"
	"github.com/ernesto2108/anvil/pkg/output"
)

func cmdSelfUpdate(cfg *config.App, git *gitutil.Repo, args []string) {
	// Step 1: Pull latest changes
	output.Info("Pulling latest changes...")
	if err := git.Pull(); err != nil {
		output.Error("git pull: %s", err)
		output.Warn("Try running 'git pull' manually in %s", cfg.RepoDir)
		os.Exit(1)
	}

	newSHA := git.CurrentSHA()
	output.Info("At %s (%s)", output.Cyan(git.CurrentBranch()), newSHA)

	// Step 2: Rebuild binary
	output.Info("Building %s...", cfg.Name)
	build := exec.Command("go", "build", "-o", cfg.Name, "./cmd/"+cfg.Name)
	build.Dir = cfg.RepoDir
	build.Stdout = os.Stdout
	build.Stderr = os.Stderr
	if err := build.Run(); err != nil {
		output.Error("go build: %s", err)
		os.Exit(1)
	}

	fmt.Println()
	output.Info("Updated. Symlinked targets are already in sync.")
	output.Info("Run %s to review and update targets.", output.Cyan(cfg.Name+" browse"))
}
