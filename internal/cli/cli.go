package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ernesto2108/anvil/pkg/config"
	"github.com/ernesto2108/anvil/pkg/fileutil"
	"github.com/ernesto2108/anvil/pkg/gitutil"
	"github.com/ernesto2108/anvil/pkg/output"
	"github.com/ernesto2108/anvil/pkg/state"
)

func Run(args []string) {
	appName := detectAppName(args[0])
	output.SetAppName(appName)

	cmd := "help"
	var cmdArgs []string
	if len(args) > 1 {
		cmd = args[1]
	}
	if len(args) > 2 {
		cmdArgs = args[2:]
	}

	// Commands that don't need repo config — launch early to avoid config errors
	// when run outside a project directory (no anvil.yaml in cwd).
	switch cmd {
	case "emit":
		cmdEmit(nil)
		return
	case "verify-direct":
		cmdVerifyDirect(nil)
		return
	case "clean-empty-runs":
		cmdCleanEmptyRuns(nil, cmdArgs)
		return
	case "capture":
		cmdCapture(nil, cmdArgs)
		return
	case "mcp-server":
		repoDir := resolveRepoDir(appName)
		cfg, err := config.Load(repoDir, appName)
		if err != nil {
			// Run with a minimal config — inventory tools still work from repoDir.
			cfg = &config.App{Name: appName, RepoDir: repoDir}
		}
		cmdMCPServer(cfg)
		return
	}

	repoDir := resolveRepoDir(appName)
	git := gitutil.New(repoDir)

	cfg, err := config.Load(repoDir, appName)
	if err != nil {
		output.Error("load config: %s", err)
		os.Exit(1)
	}

	switch cmd {
	case "init":
		cmdInit(cfg, git)
	case "browse":
		registryBrowse(cfg, git)
	case "targets":
		cmdTargets(cfg, cmdArgs)
	case "provider":
		cmdProvider(cfg, cmdArgs)
	case "uninstall":
		cmdUninstall(cfg)
	case "status":
		cmdStatus(cfg, git)
	case "pin":
		cmdPin(cfg, git, cmdArgs)
	case "unpin":
		cmdUnpin(cfg, git, cmdArgs)
	case "run":
		cmdRun(cfg, cmdArgs)
	case "quick", "bug", "feat", "design", "epic", "db", "infra":
		cmdPreset(cfg, cmd, cmdArgs)
	case "diff":
		cmdDiff(cfg, git, cmdArgs)
	case "registry":
		cmdRegistry(cfg, git, cmdArgs)
	case "update":
		cmdSelfUpdate(cfg, git, cmdArgs)
	case "doctor":
		cmdDoctor(cfg, git)
	case "stats":
		cmdStats(cfg)
	case "dream":
		cmdDream(cfg, cmdArgs)
	case "export":
		cmdExport(cfg, cmdArgs)
	case "digests":
		cmdDigests(cfg, cmdArgs)
	case "migrate":
		flags, _ := parseRunFlags(cmdArgs)
		cmdMigrate(flags.forceMigrate)
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

func detectAppName(exe string) string {
	name := filepath.Base(exe)
	name = strings.TrimSuffix(name, filepath.Ext(name))
	if name == "forge" {
		return "forge"
	}
	return "anvil"
}

func resolveRepoDir(appName string) string {
	cwd, err := os.Getwd()
	if err == nil {
		if fileutil.Exists(filepath.Join(cwd, appName+".yaml")) {
			return cwd
		}
	}

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

func printCheck(ok bool, msg string, args ...any) {
	icon := output.Green("●")
	if !ok {
		icon = output.Red("✗")
	}
	fmt.Printf("  %s %s\n", icon, fmt.Sprintf(msg, args...))
}
