package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ernesto2108/anvil/pkg/config"
	"github.com/ernesto2108/anvil/internal/deploy"
	"github.com/ernesto2108/anvil/pkg/gitutil"
	"github.com/ernesto2108/anvil/pkg/output"
	"github.com/ernesto2108/anvil/pkg/state"
)

func cmdDeploy(cfg *config.App, git *gitutil.Repo, args []string) {
	version := "HEAD"
	if len(args) > 0 {
		version = args[0]
	}

	st := loadState(cfg)
	paths := deploy.ResolvePaths()

	output.Info("Deploying %s...", cfg.Name)
	fmt.Println()

	// Pre-install snapshot (first deploy only)
	if !st.SnapshotComplete() {
		output.Info("Saving pre-install snapshot of existing files...")
		found := false
		snap := st.SnapshotDir()

		// Claude
		for _, comp := range []string{config.CompAgents, config.CompSkills, config.CompCommands} {
			if deploy.SnapshotItem(filepath.Join(paths.Claude, comp), filepath.Join(snap, config.TargetClaude, comp)) {
				found = true
			}
		}
		if deploy.SnapshotItem(filepath.Join(paths.Claude, config.FileClaudeMD), filepath.Join(snap, config.TargetClaude, config.FileClaudeMD)) {
			found = true
		}

		// OpenCode
		for _, comp := range []string{config.CompAgents, config.CompCommands} {
			if deploy.SnapshotItem(filepath.Join(paths.OpenCode, comp), filepath.Join(snap, config.TargetOpenCode, comp)) {
				found = true
			}
		}

		// Gemini
		for _, comp := range []string{config.CompSkills, config.CompCommands} {
			if deploy.SnapshotItem(filepath.Join(paths.Gemini, comp), filepath.Join(snap, config.TargetGemini, comp)) {
				found = true
			}
		}
		if deploy.SnapshotItem(filepath.Join(paths.Gemini, config.FileGeminiMD), filepath.Join(snap, config.TargetGemini, config.FileGeminiMD)) {
			found = true
		}

		// Codex
		if deploy.SnapshotItem(filepath.Join(paths.Codex, config.CompSkills), filepath.Join(snap, config.TargetCodex, config.CompSkills)) {
			found = true
		}
		if deploy.SnapshotItem(filepath.Join(paths.Codex, config.FileAgentsMD), filepath.Join(snap, config.TargetCodex, config.FileAgentsMD)) {
			found = true
		}

		if !found {
			output.Info("  No existing files found — clean install")
		}
		st.MarkSnapshotComplete()
		fmt.Println()
	}

	// Checkout specific version if not HEAD
	if version != "HEAD" {
		if !git.VersionExists(version) {
			output.Error("Version '%s' not found", version)
			os.Exit(1)
		}
		output.Warn("Checking out %s...", version)
		if err := git.Checkout(version); err != nil {
			output.Error("checkout: %s", err)
			os.Exit(1)
		}
	}

	sha := git.CurrentSHA()
	tag := git.CurrentTag()
	displayVersion := tag
	if displayVersion == state.StateNone {
		displayVersion = sha
	}

	// Deploy to each enabled target
	if cfg.TargetEnabled(config.TargetClaude) {
		deploy.Claude(cfg, paths)
	}
	fmt.Println()

	if cfg.TargetEnabled(config.TargetOpenCode) {
		deploy.OpenCode(cfg, paths)
	}
	fmt.Println()

	if cfg.TargetEnabled(config.TargetGemini) {
		deploy.Gemini(cfg, paths)
	}
	fmt.Println()

	if cfg.TargetEnabled(config.TargetCodex) {
		deploy.Codex(cfg, paths)
	}
	fmt.Println()

	if cfg.TargetEnabled(config.TargetCursor) {
		deploy.Cursor(cfg)
	}

	// Update state
	st.RecordDeploy(displayVersion, sha, git.CurrentBranch(), cfg.ActiveProvider(), deploy.DeployedTargets(cfg))
	if err := st.Save(); err != nil {
		output.Error("save state: %s", err)
	}

	fmt.Println()
	output.Info("Deployed %s (%s) to all enabled targets", output.Cyan(displayVersion), sha)
}

func cmdRollback(cfg *config.App, git *gitutil.Repo) {
	st := loadState(cfg)

	if st.PreviousVersion == "" || st.PreviousVersion == state.StateNone {
		output.Error("No previous version to rollback to")
		os.Exit(1)
	}

	output.Warn("Rolling back to %s...", st.PreviousVersion)
	cmdDeploy(cfg, git, []string{st.PreviousVersion})
}
