package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/ernesto2108/anvil/internal/deploy"
	"github.com/ernesto2108/anvil/internal/tui"
	"github.com/ernesto2108/anvil/pkg/config"
	"github.com/ernesto2108/anvil/pkg/fileutil"
	"github.com/ernesto2108/anvil/pkg/frontmatter"
	"github.com/ernesto2108/anvil/pkg/gitutil"
	"github.com/ernesto2108/anvil/pkg/output"
	"github.com/ernesto2108/anvil/pkg/registry"
)

func cmdRegistry(cfg *config.App, git *gitutil.Repo, args []string) {
	if len(args) == 0 {
		registryHelp()
		return
	}

	switch args[0] {
	case "list":
		registryList(cfg, args[1:])
	case "search":
		registrySearch(cfg, args[1:])
	case "browse":
		registryBrowse(cfg, git)
	default:
		output.Error("Unknown registry command: %s", args[0])
		registryHelp()
		os.Exit(1)
	}
}

func registryList(cfg *config.App, args []string) {
	regs := cfg.Manifest.Registries
	if len(regs) == 0 {
		output.Info("No registries configured. Add one to anvil.yaml")
		return
	}

	typeFilter := ""
	if len(args) > 0 {
		typeFilter = args[0]
	}

	// Check targets for installed status
	paths := deploy.ResolvePaths()
	targetPaths := enabledTargetPaths(cfg, paths)

	for _, reg := range regs {
		output.Info("%s (%s)", output.Bold(reg.Name), registrySource(reg))

		idx, err := fetchRegistry(reg, cfg.RepoDir)
		if err != nil {
			output.Error("  fetch failed: %s", err)
			continue
		}

		entries := idx.Entries
		if typeFilter != "" {
			t := strings.TrimSuffix(typeFilter, "s")
			entries = registry.FilterByType(entries, t)
		}

		if len(entries) == 0 {
			output.Info("  (no entries)")
			continue
		}

		for _, e := range entries {
			targets := perTargetStatus(e.Name, e.Type, targetPaths)
			deployed := 0
			for _, v := range targets {
				if v {
					deployed++
				}
			}

			status := ""
			if deployed == len(targetPaths) {
				status = output.Green(" (all targets)")
			} else if deployed > 0 {
				status = output.Yellow(fmt.Sprintf(" (%d/%d targets)", deployed, len(targetPaths)))
			}

			fmt.Printf("  %-12s %-20s %s%s\n",
				output.Cyan(e.Type), e.Name, e.Description, status)
		}
		fmt.Println()
	}
}

func registrySearch(cfg *config.App, args []string) {
	if len(args) == 0 {
		output.Error("Usage: registry search <query>")
		os.Exit(1)
	}
	query := strings.ToLower(strings.Join(args, " "))

	regs := cfg.Manifest.Registries
	if len(regs) == 0 {
		output.Error("No registries configured")
		os.Exit(1)
	}

	found := 0
	for _, reg := range regs {
		idx, err := fetchRegistry(reg, cfg.RepoDir)
		if err != nil {
			continue
		}

		for _, e := range idx.Entries {
			nameMatch := strings.Contains(strings.ToLower(e.Name), query)
			descMatch := strings.Contains(strings.ToLower(e.Description), query)
			if nameMatch || descMatch {
				fmt.Printf("  %-12s %-20s %s  [%s]\n",
					output.Cyan(e.Type), e.Name, e.Description, reg.Name)
				found++
			}
		}
	}

	if found == 0 {
		output.Info("No entries matching %q", query)
	}
}

func registryBrowse(cfg *config.App, git *gitutil.Repo) {
	// Resolve target paths
	paths := deploy.ResolvePaths()
	targetPaths := enabledTargetPaths(cfg, paths)

	// Ordered list of enabled targets for the TUI
	var enabledTargets []string
	for _, t := range []string{config.TargetClaude, config.TargetOpenCode, config.TargetGemini, config.TargetCodex} {
		if _, ok := targetPaths[t]; ok {
			enabledTargets = append(enabledTargets, t)
		}
	}

	// Scan everything from the local filesystem — the repo IS the source of truth
	var items []tui.Item
	items = append(items, scanAgents(cfg.RepoDir, targetPaths)...)
	items = append(items, scanSkills(cfg.RepoDir, targetPaths)...)
	items = append(items, scanCommands(cfg.RepoDir, targetPaths)...)

	if len(items) == 0 {
		output.Info("No agents, skills, or commands found in %s", cfg.RepoDir)
		return
	}

	info := tui.RepoInfo{
		Branch:   git.CurrentBranch(),
		SHA:      git.CurrentSHA(),
		Tag:      git.CurrentTag(),
		Provider: cfg.ActiveProvider(),
	}
	m := tui.NewModel(items, enabledTargets, info)
	p := tea.NewProgram(m, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		output.Error("TUI error: %s", err)
		os.Exit(1)
	}

	// Process queued actions
	result := finalModel.(tui.Model)
	queue := result.Queue()
	if len(queue) == 0 {
		return
	}

	fmt.Println()
	for _, action := range queue {
		targetPath, ok := targetPaths[action.Target]
		if !ok {
			continue
		}
		if action.Action == "install" {
			output.Info("Installing %s %q -> %s (%s)...",
				action.Item.Type, action.Item.Name, action.Target, action.Mode)
			installToTarget(cfg, action.Item, action.Target, targetPath, action.Mode)
		} else {
			output.Info("Uninstalling %s %q from %s...",
				action.Item.Type, action.Item.Name, action.Target)
			uninstallFromTarget(action.Item, action.Target, targetPath)
		}
	}

	// Sync metadata files (CLAUDE.md, GEMINI.md, AGENTS.md)
	fmt.Println()
	output.Info("Syncing metadata files...")
	syncMetadataFiles(cfg, targetPaths)

	fmt.Println()
	output.Info("Done.")
}

// scanAgents discovers agents from the repo filesystem.
func scanAgents(repoDir string, targets map[string]string) []tui.Item {
	agentDir := filepath.Join(repoDir, config.CompAgents)
	entries, err := os.ReadDir(agentDir)
	if err != nil {
		return nil
	}

	var items []tui.Item
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".md" {
			continue
		}
		name := strings.TrimSuffix(e.Name(), ".md")
		path := filepath.Join(agentDir, e.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		doc := frontmatter.Parse(string(data))
		tgts := perTargetStatus(name, "agent", targets)
		items = append(items, tui.Item{
			Name:        name,
			Type:        "agent",
			Description: doc.Fields["description"],
			Targets:     tgts,
		})
	}
	return items
}

// scanSkills discovers skills from the repo filesystem.
func scanSkills(repoDir string, targets map[string]string) []tui.Item {
	skillDir := filepath.Join(repoDir, config.CompSkills)
	entries, err := os.ReadDir(skillDir)
	if err != nil {
		return nil
	}

	var items []tui.Item
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		// Read SKILL.md for description
		skillFile := filepath.Join(skillDir, name, "SKILL.md")
		desc := ""
		if data, err := os.ReadFile(skillFile); err == nil {
			desc = frontmatter.Get(string(data), "description")
		}
		tgts := perTargetStatus(name, "skill", targets)
		items = append(items, tui.Item{
			Name:        name,
			Type:        "skill",
			Description: desc,
			Targets:     tgts,
		})
	}
	return items
}

// scanCommands discovers command files from the repo filesystem.
func scanCommands(repoDir string, targets map[string]string) []tui.Item {
	cmdDir := filepath.Join(repoDir, config.CompCommands)
	entries, err := os.ReadDir(cmdDir)
	if err != nil {
		return nil
	}

	var items []tui.Item
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() {
			// Scan .md files inside subdirectory (e.g., commands/git/*.md)
			subDir := filepath.Join(cmdDir, name)
			subEntries, err := os.ReadDir(subDir)
			if err != nil {
				continue
			}
			for _, se := range subEntries {
				if se.IsDir() || filepath.Ext(se.Name()) != ".md" || se.Name() == "SKILL.md" {
					continue
				}
				cmdName := name + "/" + strings.TrimSuffix(se.Name(), ".md")
				desc := readFrontmatterDesc(filepath.Join(subDir, se.Name()))
				tgts := perTargetStatus(cmdName, "command", targets)
				items = append(items, tui.Item{
					Name:        cmdName,
					Type:        "command",
					Description: desc,
					Targets:     tgts,
				})
			}
		} else if filepath.Ext(name) == ".md" {
			cmdName := strings.TrimSuffix(name, ".md")
			desc := readFrontmatterDesc(filepath.Join(cmdDir, name))
			tgts := perTargetStatus(cmdName, "command", targets)
			items = append(items, tui.Item{
				Name:        cmdName,
				Type:        "command",
				Description: desc,
				Targets:     tgts,
			})
		}
	}
	return items
}

// readFrontmatterDesc reads just the description from a markdown file's frontmatter.
func readFrontmatterDesc(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return frontmatter.Get(string(data), "description")
}

// enabledTargetPaths returns a map of target name -> base path for all enabled targets.
func enabledTargetPaths(cfg *config.App, paths deploy.TargetPaths) map[string]string {
	m := map[string]string{}
	if cfg.TargetEnabled(config.TargetClaude) {
		m[config.TargetClaude] = paths.Claude
	}
	if cfg.TargetEnabled(config.TargetOpenCode) {
		m[config.TargetOpenCode] = paths.OpenCode
	}
	if cfg.TargetEnabled(config.TargetGemini) {
		m[config.TargetGemini] = paths.Gemini
	}
	if cfg.TargetEnabled(config.TargetCodex) {
		m[config.TargetCodex] = paths.Codex
	}
	return m
}

// perTargetStatus checks deployment for each target individually.
func perTargetStatus(name, typ string, targets map[string]string) map[string]bool {
	result := map[string]bool{}
	for targetName, basePath := range targets {
		switch typ {
		case "agent":
			switch targetName {
			case config.TargetCursor:
				result[targetName] = fileutil.Exists(filepath.Join(basePath, ".cursor", "rules", name+".md"))
			case config.TargetGemini, config.TargetCodex:
				result[targetName] = false // no individual agent files
			default:
				result[targetName] = fileutil.Exists(filepath.Join(basePath, config.CompAgents, name+".md"))
			}
		case "skill":
			dir := filepath.Join(basePath, config.CompSkills, name)
			result[targetName] = fileutil.IsDir(dir) || fileutil.IsSymlink(dir)
		case "command":
			if targetName == config.TargetGemini {
				path := filepath.Join(basePath, config.CompCommands, name+".toml")
				result[targetName] = fileutil.Exists(path)
			} else {
				path := filepath.Join(basePath, config.CompCommands, name+".md")
				result[targetName] = fileutil.Exists(path) || fileutil.IsSymlink(path)
			}
		}
	}
	return result
}

// installToTarget deploys a single item to one target with the given mode.
// Handles per-target format adaptations (OpenCode, Gemini, Cursor, Codex).
func installToTarget(cfg *config.App, it tui.Item, targetName, targetPath string, mode tui.DeployMode) {
	switch it.Type {
	case "agent":
		installAgent(cfg, it, targetName, targetPath, mode)
	case "skill":
		installSkill(cfg, it, targetPath, mode)
	case "command":
		installCommand(cfg, it, targetName, targetPath, mode)
	}
}

func installAgent(cfg *config.App, it tui.Item, targetName, targetPath string, mode tui.DeployMode) {
	srcFile := filepath.Join(cfg.RepoDir, config.CompAgents, it.Name+".md")
	if !fileutil.Exists(srcFile) {
		output.Error("  Source not found: %s", srcFile)
		return
	}

	data, err := os.ReadFile(srcFile)
	if err != nil {
		output.Error("  Read failed: %s", err)
		return
	}

	doc := frontmatter.Parse(string(data))

	switch targetName {
	case config.TargetClaude:
		destDir := filepath.Join(targetPath, config.CompAgents)
		if err := os.MkdirAll(destDir, 0o755); err != nil {
			output.Error("  create agents dir: %s", err)
			return
		}
		destFile := filepath.Join(destDir, it.Name+".md")

		if mode == tui.ModeSymlink {
			if err := fileutil.ForceSymlink(srcFile, destFile); err != nil {
				output.Error("  symlink failed: %s", err)
				return
			}
		} else {
			content := string(data)
			content = frontmatter.SetField(content, "managed-by", "anvil")
			tier := doc.Fields["model"]
			if config.IsTier(tier) {
				if model, err := cfg.ResolveTier(tier, targetName); err == nil {
					content = frontmatter.ReplaceField(content, "model", tier, model)
				}
			}
			perm := doc.Fields["permission"]
			if config.IsPerm(perm) {
				tools := cfg.ResolvePermission(perm, targetName)
				if tools != "" {
					content = frontmatter.ReplaceField(content, "permission", perm, tools)
				}
			}
			if err := os.WriteFile(destFile, []byte(content), 0o644); err != nil {
				output.Error("  write agent: %s", err)
				return
			}
		}
		output.Info("  %s %s -> %s", output.Green("✓"), mode, destFile)

	case config.TargetOpenCode:
		// OpenCode needs mode: subagent in frontmatter
		destDir := filepath.Join(targetPath, config.CompAgents)
		if err := os.MkdirAll(destDir, 0o755); err != nil {
			output.Error("  create agents dir: %s", err)
			return
		}
		destFile := filepath.Join(destDir, it.Name+".md")

		desc := doc.Fields["description"]
		tier := doc.Fields["model"]
		perm := doc.Fields["permission"]

		resolved := tier
		if config.IsTier(tier) {
			if model, err := cfg.ResolveTier(tier, cfg.ActiveProvider()); err == nil {
				resolved = model
			}
		}
		permResolved := config.PermWrite
		if config.IsPerm(perm) {
			if p := cfg.ResolvePermission(perm, targetName); p != "" {
				permResolved = p
			}
		}

		adapted := fmt.Sprintf("---\ndescription: %s\nmanaged-by: anvil\nmode: subagent\nmodel: %s\npermission: %s\n---\n\n%s", desc, resolved, permResolved, doc.Body)
		if err := os.WriteFile(destFile, []byte(adapted), 0o644); err != nil {
			output.Error("  write agent: %s", err)
			return
		}
		output.Info("  %s copy (adapted) -> %s", output.Green("✓"), destFile)

	case config.TargetGemini:
		// Gemini doesn't use agent files directly — skip silently
		return

	case config.TargetCodex:
		// Codex uses AGENTS.md (auto-generated) — skip individual agents
		return

	case config.TargetCursor:
		// Cursor needs .cursor/rules/ format with alwaysApply
		rulesDir := filepath.Join(targetPath, ".cursor", "rules")
		if err := os.MkdirAll(rulesDir, 0o755); err != nil {
			output.Error("  create cursor rules dir: %s", err)
			return
		}
		destFile := filepath.Join(rulesDir, it.Name+".md")

		desc := doc.Fields["description"]
		adapted := fmt.Sprintf("---\ndescription: \"%s\"\nmanaged-by: anvil\nalwaysApply: false\n---\n\n%s", desc, doc.Body)
		if err := os.WriteFile(destFile, []byte(adapted), 0o644); err != nil {
			output.Error("  write cursor rule: %s", err)
			return
		}
		output.Info("  %s copy (cursor rule) -> %s", output.Green("✓"), destFile)
	}
}

func installSkill(cfg *config.App, it tui.Item, targetPath string, mode tui.DeployMode) {
	srcDir := filepath.Join(cfg.RepoDir, config.CompSkills, it.Name)
	if !fileutil.IsDir(srcDir) {
		output.Error("  Source not found: %s", srcDir)
		return
	}
	destDir := filepath.Join(targetPath, config.CompSkills)
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		output.Error("  create skills dir: %s", err)
		return
	}
	dst := filepath.Join(destDir, it.Name)

	if mode == tui.ModeSymlink {
		if err := fileutil.ForceSymlink(srcDir, dst); err != nil {
			output.Error("  symlink failed: %s", err)
			return
		}
	} else {
		fileutil.CleanPath(dst)
		if err := fileutil.CopyDir(srcDir, dst); err != nil {
			output.Error("  copy failed: %s", err)
			return
		}
	}
	output.Info("  %s %s -> %s", output.Green("✓"), mode, dst)
}

func installCommand(cfg *config.App, it tui.Item, targetName, targetPath string, mode tui.DeployMode) {
	srcFile := filepath.Join(cfg.RepoDir, config.CompCommands, it.Name+".md")
	if !fileutil.Exists(srcFile) {
		output.Error("  Source not found: %s", srcFile)
		return
	}

	switch targetName {
	case config.TargetGemini:
		// Gemini needs .toml format
		data, err := os.ReadFile(srcFile)
		if err != nil {
			output.Error("  Read failed: %s", err)
			return
		}
		doc := frontmatter.Parse(string(data))
		desc := doc.Fields["description"]
		if desc == "" {
			desc = doc.Fields["name"]
		}
		body := strings.ReplaceAll(doc.Body, "$ARGUMENTS", "{{args}}")
		body = strings.ReplaceAll(body, "$1", "{{args}}")

		toml := fmt.Sprintf("description = \"%s\"\nprompt = \"\"\"\n%s\n\"\"\"", desc, body)

		destFile := filepath.Join(targetPath, config.CompCommands, it.Name+".toml")
		if err := os.MkdirAll(filepath.Dir(destFile), 0o755); err != nil {
			output.Error("  create commands dir: %s", err)
			return
		}
		if err := os.WriteFile(destFile, []byte(toml), 0o644); err != nil {
			output.Error("  write command: %s", err)
			return
		}
		output.Info("  %s copy (toml) -> %s", output.Green("✓"), destFile)

	default:
		// Claude, OpenCode, Codex — standard .md
		destFile := filepath.Join(targetPath, config.CompCommands, it.Name+".md")
		if err := os.MkdirAll(filepath.Dir(destFile), 0o755); err != nil {
			output.Error("  create commands dir: %s", err)
			return
		}

		if mode == tui.ModeSymlink {
			if err := fileutil.ForceSymlink(srcFile, destFile); err != nil {
				output.Error("  symlink failed: %s", err)
				return
			}
		} else {
			if err := fileutil.CopyFile(srcFile, destFile); err != nil {
				output.Error("  copy failed: %s", err)
				return
			}
		}
		output.Info("  %s %s -> %s", output.Green("✓"), mode, destFile)
	}
}

// uninstallFromTarget removes a single item from one target.
func uninstallFromTarget(it tui.Item, targetName, targetPath string) {
	var path string
	switch it.Type {
	case "agent":
		switch targetName {
		case config.TargetCursor:
			path = filepath.Join(targetPath, ".cursor", "rules", it.Name+".md")
		case config.TargetGemini, config.TargetCodex:
			return // these don't have individual agent files
		default:
			path = filepath.Join(targetPath, config.CompAgents, it.Name+".md")
		}
	case "skill":
		path = filepath.Join(targetPath, config.CompSkills, it.Name)
	case "command":
		if targetName == config.TargetGemini {
			path = filepath.Join(targetPath, config.CompCommands, it.Name+".toml")
		} else {
			path = filepath.Join(targetPath, config.CompCommands, it.Name+".md")
		}
	}

	if path != "" && (fileutil.Exists(path) || fileutil.IsSymlink(path)) {
		if err := fileutil.CleanPath(path); err != nil {
			output.Error("  Remove failed: %s", err)
			return
		}
		output.Info("  %s removed %s", output.Red("✗"), path)
	}
}

// syncMetadataFiles handles CLAUDE.md, GEMINI.md, and AGENTS.md after browse actions.
func syncMetadataFiles(cfg *config.App, targetPaths map[string]string) {
	// CLAUDE.md -> symlink to claude target
	if claudePath, ok := targetPaths[config.TargetClaude]; ok {
		src := filepath.Join(cfg.RepoDir, config.FileClaudeMD)
		if fileutil.Exists(src) {
			dst := filepath.Join(claudePath, config.FileClaudeMD)
			if !fileutil.Exists(dst) && !fileutil.IsSymlink(dst) {
				fileutil.ForceSymlink(src, dst)
				output.Info("  %s %s -> %s", output.Green("✓"), config.FileClaudeMD, dst)
			}
		}
	}

	// GEMINI.md -> copy CLAUDE.md to gemini target
	if geminiPath, ok := targetPaths[config.TargetGemini]; ok {
		src := filepath.Join(cfg.RepoDir, config.FileClaudeMD)
		if fileutil.Exists(src) {
			dst := filepath.Join(geminiPath, config.FileGeminiMD)
			fileutil.CopyFile(src, dst)
			output.Info("  %s %s -> %s", output.Green("✓"), config.FileGeminiMD, dst)
		}
	}

	// AGENTS.md -> regenerate for codex target
	if codexPath, ok := targetPaths[config.TargetCodex]; ok {
		dst := filepath.Join(codexPath, config.FileAgentsMD)
		if err := deploy.GenerateAgentsMD(cfg, dst); err != nil {
			output.Error("generate AGENTS.md: %s", err)
		} else {
			output.Info("  %s %s -> %s", output.Green("✓"), config.FileAgentsMD, dst)
		}
	}
}

func registryHelp() {
	fmt.Println(`
  REGISTRY COMMANDS:
    browse                           Interactive TUI (recommended)
    list [agents|skills|commands]    List entries (non-interactive)
    search <query>                   Search across all registries

  TUI controls (browse):
    ↑/↓       navigate          tab    filter type
    enter     toggle all        c/o/g/x  toggle per target
    s         symlink/copy      /      search
    q         quit & apply

  Configure registries in anvil.yaml:
    registries:
      - name: official
        path: .                        # local (repo cloned)
      - name: community
        url: https://example.com/registry.json  # remote`)
}

func contains(list []string, item string) bool {
	for _, v := range list {
		if v == item {
			return true
		}
	}
	return false
}

// fetchRegistry resolves a registry config to an Index.
// If path is set, reads from local disk. Otherwise fetches from URL.
// repoDir is used to resolve relative paths.
func fetchRegistry(reg config.Registry, repoDir ...string) (*registry.Index, error) {
	if reg.Path != "" {
		p := reg.Path
		if strings.HasPrefix(p, "~/") {
			home, _ := os.UserHomeDir()
			p = filepath.Join(home, p[2:])
		} else if !filepath.IsAbs(p) && len(repoDir) > 0 {
			p = filepath.Join(repoDir[0], p)
		}
		// If path points to a directory, look for registry.json inside it
		info, err := os.Stat(p)
		if err != nil {
			return nil, fmt.Errorf("local path: %w", err)
		}
		if info.IsDir() {
			p = filepath.Join(p, "registry.json")
		}
		return registry.FetchLocal(p)
	}
	return registry.Fetch(reg.URL)
}

// registrySource returns a display label for the registry (path or url).
func registrySource(reg config.Registry) string {
	if reg.Path != "" {
		return reg.Path
	}
	return reg.URL
}
