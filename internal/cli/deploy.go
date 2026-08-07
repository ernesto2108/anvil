package cli

import (
	"os"

	"github.com/ernesto2108/anvil/internal/deploy"
	"github.com/ernesto2108/anvil/pkg/config"
	"github.com/ernesto2108/anvil/pkg/gitutil"
	"github.com/ernesto2108/anvil/pkg/output"
	"github.com/ernesto2108/anvil/pkg/state"
)

// cmdDeploy deploys components to all enabled targets and records the
// deployed SHA in state.json so 'diff' and 'doctor' can detect drift.
func cmdDeploy(cfg *config.App, git *gitutil.Repo, args []string) {
	component := "all"
	if len(args) > 0 {
		component = args[0]
	}
	if component != "all" && component != "agents" && component != "skills" {
		output.Error("component must be one of: all, agents, skills")
		os.Exit(1)
	}

	paths := deploy.ResolvePaths()
	deploy.StartSummary()

	switch component {
	case "agents":
		if cfg.TargetEnabled(config.TargetClaude) {
			deploy.ClaudeAgentsOnly(cfg, paths)
		}
		if cfg.TargetEnabled(config.TargetOpenCode) {
			deploy.OpenCodeAgentsOnly(cfg, paths)
		}
	case "skills":
		if cfg.TargetEnabled(config.TargetClaude) {
			ts := deploy.AddTarget(config.TargetClaude + " (skills)")
			deploy.DeploySkillsSymlink(cfg, paths.Claude, ts)
		}
		if cfg.TargetEnabled(config.TargetGemini) {
			ts := deploy.AddTarget(config.TargetGemini + " (skills)")
			deploy.DeploySkillsSymlink(cfg, paths.Gemini, ts)
		}
		if cfg.TargetEnabled(config.TargetCodex) {
			ts := deploy.AddTarget(config.TargetCodex + " (skills)")
			deploy.DeploySkillsSymlink(cfg, paths.Codex, ts)
		}
	default:
		if cfg.TargetEnabled(config.TargetClaude) {
			deploy.Claude(cfg, paths)
		}
		if cfg.TargetEnabled(config.TargetOpenCode) {
			deploy.OpenCode(cfg, paths)
		}
		if cfg.TargetEnabled(config.TargetGemini) {
			deploy.Gemini(cfg, paths)
		}
		if cfg.TargetEnabled(config.TargetCodex) {
			deploy.Codex(cfg, paths)
		}
	}

	deploy.PrintSummary()
	RecordDeployState(cfg, git)
}

// RecordDeployState persists the current git SHA/branch/tag as the deployed
// version so 'diff' and 'doctor' have a baseline to compare against.
func RecordDeployState(cfg *config.App, git *gitutil.Repo) {
	st := loadState(cfg)
	version := git.CurrentTag()
	if version == "" || version == state.StateNone {
		version = git.CurrentSHA()
	}
	st.RecordDeploy(version, git.CurrentSHA(), git.CurrentBranch(), cfg.ActiveProvider(), deploy.DeployedTargets(cfg))
	if err := st.Save(); err != nil {
		output.Error("save deploy state: %s", err)
	}
}
