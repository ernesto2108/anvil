package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ernesto2108/anvil/internal/config"
	"github.com/ernesto2108/anvil/internal/deploy"
	"github.com/ernesto2108/anvil/internal/fileutil"
	"github.com/ernesto2108/anvil/internal/gitutil"
	"github.com/ernesto2108/anvil/internal/output"
	"github.com/ernesto2108/anvil/internal/state"
)

func main() {
	appName := detectAppName()
	output.SetAppName(appName)

	repoDir := resolveRepoDir()
	git := gitutil.New(repoDir)

	cfg, err := config.Load(repoDir, appName)
	if err != nil {
		output.Error("load config: %s", err)
		os.Exit(1)
	}

	cmd := "help"
	if len(os.Args) > 1 {
		cmd = os.Args[1]
	}
	args := os.Args[2:]

	switch cmd {
	case "deploy":
		cmdDeploy(cfg, git, args)
	case "targets":
		cmdTargets(cfg, args)
	case "provider":
		cmdProvider(cfg, args)
	case "rollback":
		cmdRollback(cfg, git)
	case "uninstall":
		cmdUninstall(cfg)
	case "status":
		cmdStatus(cfg, git)
	case "pin":
		cmdPin(cfg, git, args)
	case "unpin":
		cmdUnpin(cfg, git, args)
	case "tags":
		cmdTags(git)
	case "diff":
		cmdDiff(cfg, git)
	case "help", "-h", "--help":
		cmdHelp(appName)
	default:
		output.Error("Unknown command: %s", cmd)
		cmdHelp(appName)
		os.Exit(1)
	}
}

func title(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func detectAppName() string {
	exe := filepath.Base(os.Args[0])
	exe = strings.TrimSuffix(exe, filepath.Ext(exe))
	if exe == "forge" {
		return "forge"
	}
	return "anvil"
}

func resolveRepoDir() string {
	// 1. Check CWD first (common during development and normal usage)
	cwd, err := os.Getwd()
	if err == nil {
		appName := detectAppName()
		if fileutil.Exists(filepath.Join(cwd, appName+".yaml")) {
			return cwd
		}
	}

	// 2. Fall back to directory of the executable
	exe, err := os.Executable()
	if err != nil {
		output.Error("resolve executable path: %s", err)
		os.Exit(1)
	}
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		output.Error("resolve symlinks: %s", err)
		os.Exit(1)
	}
	return filepath.Dir(exe)
}

func loadState(cfg *config.App) *state.State {
	stateDir := filepath.Join(cfg.RepoDir, "."+cfg.Name)
	st, err := state.Load(stateDir)
	if err != nil {
		output.Error("load state: %s", err)
		os.Exit(1)
	}
	return st
}

// ──────────────────────────────────────────────
// DEPLOY
// ──────────────────────────────────────────────

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
		for _, comp := range []string{"agents", "skills", "commands"} {
			if deploy.SnapshotItem(filepath.Join(paths.Claude, comp), filepath.Join(snap, "claude", comp)) {
				found = true
			}
		}
		if deploy.SnapshotItem(filepath.Join(paths.Claude, "CLAUDE.md"), filepath.Join(snap, "claude", "CLAUDE.md")) {
			found = true
		}

		// OpenCode
		for _, comp := range []string{"agents", "commands"} {
			if deploy.SnapshotItem(filepath.Join(paths.OpenCode, comp), filepath.Join(snap, "opencode", comp)) {
				found = true
			}
		}

		// Gemini
		for _, comp := range []string{"skills", "commands"} {
			if deploy.SnapshotItem(filepath.Join(paths.Gemini, comp), filepath.Join(snap, "gemini", comp)) {
				found = true
			}
		}
		if deploy.SnapshotItem(filepath.Join(paths.Gemini, "GEMINI.md"), filepath.Join(snap, "gemini", "GEMINI.md")) {
			found = true
		}

		// Codex
		if deploy.SnapshotItem(filepath.Join(paths.Codex, "skills"), filepath.Join(snap, "codex", "skills")) {
			found = true
		}
		if deploy.SnapshotItem(filepath.Join(paths.Codex, "AGENTS.md"), filepath.Join(snap, "codex", "AGENTS.md")) {
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
	if displayVersion == "none" {
		displayVersion = sha
	}

	// Deploy to each enabled target
	if cfg.TargetEnabled("claude") {
		deploy.Claude(cfg, paths)
	}
	fmt.Println()

	if cfg.TargetEnabled("opencode") {
		deploy.OpenCode(cfg, paths)
	}
	fmt.Println()

	if cfg.TargetEnabled("gemini") {
		deploy.Gemini(cfg, paths)
	}
	fmt.Println()

	if cfg.TargetEnabled("codex") {
		deploy.Codex(cfg, paths)
	}
	fmt.Println()

	if cfg.TargetEnabled("cursor") {
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

// ─────────────────��─────────────────��──────────
// ROLLBACK
// ─────────���─────────────────���──────────────────

func cmdRollback(cfg *config.App, git *gitutil.Repo) {
	st := loadState(cfg)

	if st.PreviousVersion == "" || st.PreviousVersion == "none" {
		output.Error("No previous version to rollback to")
		os.Exit(1)
	}

	output.Warn("Rolling back to %s...", st.PreviousVersion)
	cmdDeploy(cfg, git, []string{st.PreviousVersion})
}

// ────────��───────────────────��─────────────────
// TARGETS
// ─────────────��────────────────────���───────────

func cmdTargets(cfg *config.App, args []string) {
	allTargets := cfg.AllTargets()

	if len(args) == 0 {
		fmt.Println()
		fmt.Println(output.Bold("Deploy targets:"))
		fmt.Println()
		for _, t := range allTargets {
			if cfg.TargetEnabled(t) {
				fmt.Printf("  %s %s\n", output.Green("●"), t)
			} else {
				fmt.Printf("  %s %s\n", output.Red("○"), t)
			}
		}
		fmt.Println()
		fmt.Printf("Usage:\n")
		fmt.Printf("  %s targets <tool> [tool...]    set exact targets\n", cfg.Name)
		fmt.Printf("  %s targets --add <tool>        enable one target\n", cfg.Name)
		fmt.Printf("  %s targets --rm <tool>         disable one target\n", cfg.Name)
		fmt.Printf("  %s targets all                 enable all\n", cfg.Name)
		return
	}

	mode := "set"
	var tools []string

	switch args[0] {
	case "--add":
		mode = "add"
		tools = args[1:]
	case "--rm":
		mode = "rm"
		tools = args[1:]
	case "all":
		mode = "set"
		tools = allTargets
	default:
		mode = "set"
		tools = args
	}

	if len(tools) == 0 {
		output.Error("No targets specified")
		os.Exit(1)
	}

	// Validate
	valid := make(map[string]bool)
	for _, t := range allTargets {
		valid[t] = true
	}
	for _, t := range tools {
		if !valid[t] {
			output.Error("Unknown target: %s", t)
			output.Error("Valid targets: %s", strings.Join(allTargets, ", "))
			os.Exit(1)
		}
	}

	// Apply
	requested := make(map[string]bool)
	for _, t := range tools {
		requested[t] = true
	}

	switch mode {
	case "add":
		for _, t := range tools {
			cfg.SetTargetEnabled(t, true)
		}
	case "rm":
		for _, t := range tools {
			cfg.SetTargetEnabled(t, false)
		}
	case "set":
		for _, t := range allTargets {
			cfg.SetTargetEnabled(t, requested[t])
		}
	}

	output.Info("Targets updated:")
	for _, t := range allTargets {
		if cfg.TargetEnabled(t) {
			fmt.Printf("  %s %s\n", output.Green("●"), t)
		} else {
			fmt.Printf("  %s %s\n", output.Red("○"), t)
		}
	}
	fmt.Println()
	output.Info("Run %s to apply.", output.Yellow(cfg.Name+" deploy"))
}

// ──────────────────────────────────────────────
// PROVIDER
// ──────────────────────────────────────────────

func cmdProvider(cfg *config.App, args []string) {
	if len(args) == 0 {
		current := cfg.ActiveProvider()
		fmt.Println()
		fmt.Printf("%s %s\n", output.Cyan("Current provider:"), output.Green(current))
		fmt.Println()
		fmt.Println(output.Cyan("Available providers:"))
		for _, p := range cfg.ListProviders() {
			if p == current {
				fmt.Printf("  %s (active)\n", output.Green(p))
			} else {
				fmt.Printf("  %s\n", p)
			}
		}
		fmt.Println()
		fmt.Printf("Usage: %s provider <name>\n", cfg.Name)
		return
	}

	newProvider := args[0]

	// Validate
	found := false
	for _, p := range cfg.ListProviders() {
		if p == newProvider {
			found = true
			break
		}
	}
	if !found {
		output.Error("Provider '%s' not found in %s.config.yaml", newProvider, cfg.Name)
		cmdProvider(cfg, nil)
		os.Exit(1)
	}

	if err := cfg.SetProvider(newProvider); err != nil {
		output.Error("set provider: %s", err)
		os.Exit(1)
	}

	output.Info("Provider switched to %s", output.Green(newProvider))
	fmt.Println()

	// Redeploy agents
	paths := deploy.ResolvePaths()
	output.Info("Redeploying agents...")

	if cfg.TargetEnabled("claude") {
		deploy.ClaudeAgentsOnly(cfg, paths)
	}
	fmt.Println()

	if cfg.TargetEnabled("opencode") {
		deploy.OpenCodeAgentsOnly(cfg, paths)
	}

	st := loadState(cfg)
	st.Provider = newProvider
	st.Save()

	fmt.Println()
	output.Info("Done. All agents now use %s models.", output.Green(newProvider))
}

// ────────────────────────────────────��─────────
// STATUS
// ──────────────────────────────────────────────

func cmdStatus(cfg *config.App, git *gitutil.Repo) {
	st := loadState(cfg)
	paths := deploy.ResolvePaths()

	fmt.Println()
	fmt.Println(output.Bold(title(cfg.Name) + " Status"))
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	fmt.Printf("  Repo:      %s\n", cfg.RepoDir)
	fmt.Printf("  Provider:  %s\n", output.Green(st.Provider))
	fmt.Printf("  Branch:    %s\n", git.CurrentBranch())
	fmt.Printf("  HEAD:      %s\n", git.CurrentSHA())
	fmt.Printf("  Tag:       %s\n", git.CurrentTag())
	fmt.Printf("  Deployed:  %s\n", output.Green(st.DeployedVersion))
	fmt.Printf("  Previous:  %s\n", st.PreviousVersion)
	fmt.Printf("  At:        %s\n", st.DeployedAt)
	fmt.Println()

	fmt.Println(output.Bold("Targets:"))
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Claude
	claudeIcon := output.Red("○")
	if fileutil.IsDir(filepath.Join(paths.Claude, "agents")) || fileutil.IsSymlink(filepath.Join(paths.Claude, "skills")) {
		claudeIcon = output.Green("●")
	}
	fmt.Printf("  %s %-12s %s\n", claudeIcon, "claude", paths.Claude)

	for _, comp := range []string{"agents", "skills", "commands", "CLAUDE.md"} {
		p := filepath.Join(paths.Claude, comp)
		if fileutil.IsSymlink(p) {
			fmt.Printf("    %s %s (symlink)\n", output.Green("●"), comp)
		} else if fileutil.Exists(p) {
			fmt.Printf("    %s %s (copy)\n", output.Yellow("◐"), comp)
		} else {
			fmt.Printf("    %s %s\n", output.Red("○"), comp)
		}
	}

	// OpenCode
	ocIcon := output.Red("○")
	if fileutil.IsDir(filepath.Join(paths.OpenCode, "agents")) {
		ocIcon = output.Green("●")
	}
	fmt.Printf("  %s %-12s %s\n", ocIcon, "opencode", paths.OpenCode)

	// Gemini
	gemIcon := output.Red("○")
	if fileutil.IsDir(filepath.Join(paths.Gemini, "commands")) || fileutil.IsSymlink(filepath.Join(paths.Gemini, "skills")) {
		gemIcon = output.Green("●")
	}
	fmt.Printf("  %s %-12s %s\n", gemIcon, "gemini", paths.Gemini)

	// Codex
	cdxIcon := output.Red("○")
	if fileutil.IsSymlink(filepath.Join(paths.Codex, "skills")) || fileutil.Exists(filepath.Join(paths.Codex, "AGENTS.md")) {
		cdxIcon = output.Green("●")
	}
	fmt.Printf("  %s %-12s %s\n", cdxIcon, "codex", paths.Codex)

	fmt.Println()
	fmt.Printf("  %s deployed  %s copy/pinned  %s not deployed\n", output.Green("●"), output.Yellow("◐"), output.Red("○"))

	// Recent tags
	tagsOut, err := git.Tags()
	if err == nil && tagsOut != "" {
		fmt.Println()
		fmt.Println(output.Cyan("Recent tags:"))
		for i, t := range strings.Split(tagsOut, "\n") {
			if i >= 5 {
				break
			}
			fmt.Printf("  %s\n", t)
		}
	}
	fmt.Println()
}

// ───────────��──────────────────────────────────
// PIN / UNPIN
// ──────────────────────────────────────────────

func cmdPin(cfg *config.App, git *gitutil.Repo, args []string) {
	if len(args) < 2 {
		output.Error("Usage: %s pin <component> <tag>", cfg.Name)
		os.Exit(1)
	}

	component := args[0]
	version := args[1]
	st := loadState(cfg)
	paths := deploy.ResolvePaths()

	if !git.VersionExists(version) {
		output.Error("Version '%s' not found", version)
		os.Exit(1)
	}

	// Break parent symlink if pinning a nested component
	if strings.Contains(component, "/") {
		topLevel := strings.SplitN(component, "/", 2)[0]
		parentPath := filepath.Join(paths.Claude, topLevel)
		if fileutil.IsSymlink(parentPath) {
			output.Warn("Breaking symlink %s/", topLevel)
			link, _ := os.Readlink(parentPath)
			os.Remove(parentPath)
			fileutil.CopyDir(link, parentPath)
		}
	}

	targetPath := filepath.Join(paths.Claude, component)
	fileutil.CleanPath(targetPath)
	os.MkdirAll(filepath.Dir(targetPath), 0o755)

	objType, err := git.CatFileType(version, component)
	if err != nil {
		output.Error("cat-file: %s", err)
		os.Exit(1)
	}

	if objType == "tree" {
		os.MkdirAll(targetPath, 0o755)
		if err := git.Archive(version, component, paths.Claude); err != nil {
			output.Error("archive: %s", err)
			os.Exit(1)
		}
	} else {
		content, err := git.ShowFile(version, component)
		if err != nil {
			output.Error("show file: %s", err)
			os.Exit(1)
		}
		os.WriteFile(targetPath, []byte(content), 0o644)
	}

	output.Info("Pinned %s to %s", output.Cyan(component), output.Yellow(version))

	st.SetPin(component, version)
	st.Save()
}

func cmdUnpin(cfg *config.App, git *gitutil.Repo, args []string) {
	if len(args) < 1 {
		output.Error("Usage: %s unpin <component>", cfg.Name)
		os.Exit(1)
	}

	component := args[0]
	st := loadState(cfg)
	paths := deploy.ResolvePaths()

	source := filepath.Join(cfg.RepoDir, component)
	if !fileutil.Exists(source) {
		output.Error("Component '%s' not found in repo", component)
		os.Exit(1)
	}

	targetPath := filepath.Join(paths.Claude, component)
	fileutil.CleanPath(targetPath)
	st.RemovePin(component)

	if strings.Contains(component, "/") {
		objType, err := git.CatFileType("HEAD", component)
		if err == nil {
			if objType == "tree" {
				os.MkdirAll(targetPath, 0o755)
				git.Archive("HEAD", component, paths.Claude)
			} else {
				content, err := git.ShowFile("HEAD", component)
				if err == nil {
					os.WriteFile(targetPath, []byte(content), 0o644)
				}
			}
		}

		// Restore parent symlink if no more pins in this top-level
		topLevel := strings.SplitN(component, "/", 2)[0]
		if st.PinCount(topLevel+"/") == 0 {
			parentPath := filepath.Join(paths.Claude, topLevel)
			if fileutil.IsDir(parentPath) && !fileutil.IsSymlink(parentPath) {
				os.RemoveAll(parentPath)
				os.Symlink(filepath.Join(cfg.RepoDir, topLevel), parentPath)
				output.Info("Restored %s symlink", output.Cyan(topLevel+"/"))
			}
		}
	} else {
		os.Symlink(source, targetPath)
		output.Info("Unpinned %s", output.Cyan(component))
	}

	st.Save()
}

// ─────────────────────────────────���────────────
// TAGS
// ───────────���─────────────────��────────────────

func cmdTags(git *gitutil.Repo) {
	fmt.Println(output.Cyan("Available versions:"))
	fmt.Println()

	tagsOut, err := git.Tags()
	if err != nil || tagsOut == "" {
		output.Warn("No tags yet.")
		return
	}

	for _, t := range strings.Split(tagsOut, "\n") {
		if t == "" {
			continue
		}
		d := git.TagDate(t)
		m := git.TagMessage(t)
		fmt.Printf("  %-12s %s  %s\n", t, d, m)
	}
}

// ──────────────────────────────────────────────
// DIFF
// ──────────────────────────────────────────────

func cmdDiff(cfg *config.App, git *gitutil.Repo) {
	st := loadState(cfg)

	if st.DeployedSHA == "" || st.DeployedSHA == "none" {
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

	diffOut, err := git.DiffStat(st.DeployedSHA, "agents/", "skills/", "commands/", "CLAUDE.md")
	if err != nil {
		output.Error("diff: %s", err)
		return
	}
	fmt.Println(diffOut)
}

// ──────────────────────────────────────────────
// UNINSTALL
// ──────────────────────────���───────────────────

func cmdUninstall(cfg *config.App) {
	st := loadState(cfg)
	paths := deploy.ResolvePaths()

	fmt.Println()
	output.Warn("This will remove %s from ALL targets:", cfg.Name)
	output.Warn("  Claude:   %s/{agents,skills,commands,CLAUDE.md}", paths.Claude)
	output.Warn("  OpenCode: %s/{agents,commands}", paths.OpenCode)
	output.Warn("  Gemini:   %s/{skills,commands,GEMINI.md}", paths.Gemini)
	output.Warn("  Codex:    %s/{skills,AGENTS.md}", paths.Codex)
	fmt.Println()

	fmt.Print("Continue? [y/N] ")
	reader := bufio.NewReader(os.Stdin)
	confirm, _ := reader.ReadString('\n')
	confirm = strings.TrimSpace(confirm)
	if confirm != "y" && confirm != "Y" {
		output.Info("Aborted.")
		return
	}
	fmt.Println()

	snap := st.SnapshotDir()

	// Claude
	output.Info("%s", output.Bold("Claude Code:"))
	for _, comp := range []string{"agents", "skills", "commands"} {
		deploy.RestoreItem(filepath.Join(paths.Claude, comp), filepath.Join(snap, "claude", comp))
	}
	claudeMD := filepath.Join(paths.Claude, "CLAUDE.md")
	if fileutil.IsSymlink(claudeMD) {
		os.Remove(claudeMD)
		snapMD := filepath.Join(snap, "claude", "CLAUDE.md")
		if fileutil.Exists(snapMD) {
			fileutil.CopyFile(snapMD, claudeMD)
			output.Info("  %s CLAUDE.md", output.Green("restored"))
		}
	}

	// OpenCode
	output.Info("%s", output.Bold("OpenCode:"))
	for _, comp := range []string{"agents", "commands"} {
		deploy.RestoreItem(filepath.Join(paths.OpenCode, comp), filepath.Join(snap, "opencode", comp))
	}

	// Gemini
	output.Info("%s", output.Bold("Gemini CLI:"))
	for _, comp := range []string{"skills", "commands"} {
		deploy.RestoreItem(filepath.Join(paths.Gemini, comp), filepath.Join(snap, "gemini", comp))
	}
	geminiMD := filepath.Join(paths.Gemini, "GEMINI.md")
	if fileutil.Exists(geminiMD) {
		os.Remove(geminiMD)
		snapGMD := filepath.Join(snap, "gemini", "GEMINI.md")
		if fileutil.Exists(snapGMD) {
			fileutil.CopyFile(snapGMD, geminiMD)
		}
	}

	// Codex
	output.Info("%s", output.Bold("Codex:"))
	deploy.RestoreItem(filepath.Join(paths.Codex, "skills"), filepath.Join(snap, "codex", "skills"))
	agentsMD := filepath.Join(paths.Codex, "AGENTS.md")
	if fileutil.Exists(agentsMD) {
		os.Remove(agentsMD)
		snapAMD := filepath.Join(snap, "codex", "AGENTS.md")
		if fileutil.Exists(snapAMD) {
			fileutil.CopyFile(snapAMD, agentsMD)
		}
	}
	output.Info("  removed AGENTS.md")

	// Clean repo AGENTS.md
	repoAgentsMD := filepath.Join(cfg.RepoDir, "AGENTS.md")
	if fileutil.Exists(repoAgentsMD) {
		os.Remove(repoAgentsMD)
	}

	st.Remove()
	fmt.Println()
	output.Info("%s uninstalled. Pre-existing files restored where snapshots existed.", title(cfg.Name))
	output.Info("Run %s to reinstall.", output.Yellow(cfg.Name+" deploy"))
	fmt.Println()
}

// ──────────────────��────────────────���──────────
// HELP
// ────────��─────────────────────────────────────

func cmdHelp(appName string) {
	t := title(appName)
	fmt.Printf(`
  %s - Multi-tool GitOps for AI coding configuration

  Deploys agents, skills, and commands to:
    Claude Code, OpenCode, Gemini CLI, Codex, and Cursor

  USAGE:
    %s <command> [args]

  COMMANDS:
    deploy [version]     Deploy to all enabled targets (default: HEAD)
    targets [tool...]    Show or set which tools are active
    provider [name]      Show or switch AI provider (redeploys agents)
    status               Show deployment state across all tools
    rollback             Rollback to previous version
    pin <comp> <tag>     Pin a component to specific version (Claude)
    unpin <comp>         Unpin a component (Claude)
    uninstall            Remove %s from all targets
    tags                 List available versions
    diff                 Show changes since last deploy
    help                 Show this help

  EXAMPLES:
    %s targets                             # Show active tools
    %s targets claude                      # Only use Claude Code
    %s targets claude opencode             # Claude + OpenCode
    %s targets all                         # Enable all tools
    %s deploy                              # Deploy to active tools
    %s provider gemini                     # Switch to Gemini models
    %s provider local                      # Switch to local/Ollama
    %s status                              # What's deployed where?

  TARGETS:
    claude    ~/.claude/          agents + skills + commands
    opencode  ~/.config/opencode/ agents + commands
    gemini    ~/.gemini/          skills + commands (toml)
    codex     ~/.codex/           skills + AGENTS.md (auto-generated)
    cursor    per-project         rules from agents

`, t, appName, appName, appName, appName, appName, appName, appName, appName, appName, appName)
}
