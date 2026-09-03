package deploy_test

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ernesto2108/anvil/internal/deploy"
	"github.com/ernesto2108/anvil/pkg/config"
	"github.com/ernesto2108/anvil/pkg/fileutil"
	"github.com/ernesto2108/anvil/pkg/state"
)

// mockStdin sets a fake stdin reader for interactive collision prompts in tests.
func mockStdin(input string) {
	deploy.SetStdinReader(bufio.NewReader(strings.NewReader(input)))
}

// resetStdin restores the real stdin reader.
func resetStdin() {
	deploy.SetStdinReader(bufio.NewReader(os.Stdin))
}

// setupFakeRepo creates a minimal repo structure with agents, skills, commands, and CLAUDE.md.
func setupFakeRepo(t *testing.T) string {
	t.Helper()
	repo := t.TempDir()

	// agents
	agentDir := filepath.Join(repo, config.CompAgents)
	os.MkdirAll(agentDir, 0o755)
	os.WriteFile(filepath.Join(agentDir, "developer.md"), []byte("---\npermission: execute\ndescription: writes code\n---\n\nYou are a developer."), 0o644)
	os.WriteFile(filepath.Join(agentDir, "tester.md"), []byte("---\npermission: execute\ndescription: writes tests\n---\n\nYou are a tester."), 0o644)

	// skills
	skillDir := filepath.Join(repo, config.CompSkills, "go-conventions")
	os.MkdirAll(skillDir, 0o755)
	os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("# Go conventions"), 0o644)

	// commands
	cmdDir := filepath.Join(repo, config.CompCommands)
	os.MkdirAll(cmdDir, 0o755)
	os.WriteFile(filepath.Join(cmdDir, "commit.md"), []byte("---\nname: commit\ndescription: commit changes\n---\n\nCreate a commit."), 0o644)

	// CLAUDE.md
	os.WriteFile(filepath.Join(repo, config.FileClaudeMD), []byte("# Project instructions"), 0o644)

	return repo
}

// setupFakeTargets creates empty target directories simulating ~/.claude, ~/.gemini, etc.
func setupFakeTargets(t *testing.T) deploy.TargetPaths {
	t.Helper()
	base := t.TempDir()
	return deploy.TargetPaths{
		Claude:   filepath.Join(base, ".claude"),
		OpenCode: filepath.Join(base, ".config", "opencode"),
		Gemini:   filepath.Join(base, ".gemini"),
		Codex:    filepath.Join(base, ".codex"),
	}
}

func loadTestConfig(t *testing.T, repoDir string) *config.App {
	t.Helper()

	manifest := `targets:
  claude:
    enabled: true
  opencode:
    enabled: true
  gemini:
    enabled: true
  codex:
    enabled: true
  cursor:
    enabled: false
`
	cfg := `provider: claude
providers:
  - claude-opus-4-6
  - claude-sonnet-4-6
  - claude-haiku-4-5-20251001
permissions:
  claude:
    read: "Read,Glob,Grep"
    write: "Read,Glob,Grep,Write,Edit"
    execute: "*"
  opencode:
    read: read
    write: write
    execute: execute
`
	os.WriteFile(filepath.Join(repoDir, "anvil.yaml"), []byte(manifest), 0o644)
	os.WriteFile(filepath.Join(repoDir, "anvil.config.yaml"), []byte(cfg), 0o644)

	app, err := config.Load(repoDir, "anvil")
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	return app
}

func Test_DeployClaude_CreatesAgentsSkillsCommands(t *testing.T) {
	repo := setupFakeRepo(t)
	paths := setupFakeTargets(t)
	cfg := loadTestConfig(t, repo)

	os.MkdirAll(paths.Claude, 0o755)
	deploy.StartSummary()
	deploy.Claude(cfg, paths)

	// agents should be a directory with resolved files
	agentDir := filepath.Join(paths.Claude, config.CompAgents)
	if !fileutil.IsDir(agentDir) {
		t.Fatal("agents directory not created")
	}
	entries, _ := os.ReadDir(agentDir)
	if len(entries) != 2 {
		t.Errorf("expected 2 agent files, got %d", len(entries))
	}

	// Verify no model field is emitted in the deployed agent content
	data, _ := os.ReadFile(filepath.Join(agentDir, "developer.md"))
	content := string(data)
	if containsStr(content, "model:") {
		t.Error("deployed agent should not contain a 'model:' field")
	}

	// skills should be a directory with individual symlinks
	skillsPath := filepath.Join(paths.Claude, config.CompSkills)
	if !fileutil.IsDir(skillsPath) {
		t.Error("skills should be a directory")
	}
	// go-conventions should be an individual symlink
	goConv := filepath.Join(skillsPath, "go-conventions")
	if !fileutil.IsSymlink(goConv) {
		t.Error("skills/go-conventions should be an individual symlink")
	}

	// commands should be a directory with individual symlinks
	cmdsPath := filepath.Join(paths.Claude, config.CompCommands)
	if !fileutil.IsDir(cmdsPath) {
		t.Error("commands should be a directory")
	}
	commitCmd := filepath.Join(cmdsPath, "commit.md")
	if !fileutil.IsSymlink(commitCmd) {
		t.Error("commands/commit.md should be an individual symlink")
	}

	// CLAUDE.md should be a symlink
	claudeMD := filepath.Join(paths.Claude, config.FileClaudeMD)
	if !fileutil.IsSymlink(claudeMD) {
		t.Error("CLAUDE.md should be a symlink")
	}
}

func Test_DeployCodex_GeneratesAgentsMD(t *testing.T) {
	repo := setupFakeRepo(t)
	paths := setupFakeTargets(t)
	cfg := loadTestConfig(t, repo)

	deploy.StartSummary()
	deploy.Codex(cfg, paths)

	agentsMD := filepath.Join(paths.Codex, config.FileAgentsMD)
	if !fileutil.Exists(agentsMD) {
		t.Fatal("AGENTS.md not generated")
	}

	data, _ := os.ReadFile(agentsMD)
	content := string(data)
	if !containsStr(content, "developer") {
		t.Error("AGENTS.md should reference developer agent")
	}
	if !containsStr(content, "tester") {
		t.Error("AGENTS.md should reference tester agent")
	}
}

func Test_DeployGemini_CopiesCommandsAsToml(t *testing.T) {
	repo := setupFakeRepo(t)
	paths := setupFakeTargets(t)
	cfg := loadTestConfig(t, repo)

	os.MkdirAll(paths.Gemini, 0o755)
	deploy.StartSummary()
	deploy.Gemini(cfg, paths)

	// commands should be a directory with .toml files
	cmdsDir := filepath.Join(paths.Gemini, config.CompCommands)
	if !fileutil.IsDir(cmdsDir) {
		t.Fatal("gemini commands directory not created")
	}

	tomlFile := filepath.Join(cmdsDir, "commit.toml")
	if !fileutil.Exists(tomlFile) {
		t.Fatal("commit.toml not created")
	}

	data, _ := os.ReadFile(tomlFile)
	if !containsStr(string(data), "description") {
		t.Error("toml file should contain description")
	}

	// GEMINI.md should exist as copy
	geminiMD := filepath.Join(paths.Gemini, config.FileGeminiMD)
	if !fileutil.Exists(geminiMD) {
		t.Fatal("GEMINI.md not created")
	}
}

func Test_SnapshotAndRestore_FullCycle(t *testing.T) {
	repo := setupFakeRepo(t)
	paths := setupFakeTargets(t)
	cfg := loadTestConfig(t, repo)

	// Simulate pre-existing user-managed agent in claude target
	os.MkdirAll(paths.Claude, 0o755)
	originalAgents := filepath.Join(paths.Claude, config.CompAgents)
	os.MkdirAll(originalAgents, 0o755)
	os.WriteFile(filepath.Join(originalAgents, "custom.md"), []byte("---\nname: custom\n---\n\nmy custom agent"), 0o644)

	// Step 1: Snapshot the originals
	stateDir := filepath.Join(t.TempDir(), ".anvil")
	st, err := state.Load(stateDir)
	if err != nil {
		t.Fatalf("load state: %v", err)
	}
	snap := st.SnapshotDir()

	deploy.SnapshotItem(
		filepath.Join(paths.Claude, config.CompAgents),
		filepath.Join(snap, config.TargetClaude, config.CompAgents),
	)
	st.MarkSnapshotComplete()

	// Step 2: Deploy — user chooses to keep their custom.md
	mockStdin("a\nk\n")
	defer resetStdin()
	deploy.StartSummary()
	deploy.Claude(cfg, paths)

	entries, _ := os.ReadDir(filepath.Join(paths.Claude, config.CompAgents))
	hasCustom := false
	hasDeveloper := false
	for _, e := range entries {
		if e.Name() == "custom.md" {
			hasCustom = true
		}
		if e.Name() == "developer.md" {
			hasDeveloper = true
		}
	}
	if !hasCustom {
		t.Error("deploy should have preserved user-managed custom.md")
	}
	if !hasDeveloper {
		t.Error("deploy should have added anvil-managed developer.md")
	}

	// Verify custom.md was NOT stamped with managed-by
	customData, _ := os.ReadFile(filepath.Join(originalAgents, "custom.md"))
	if containsStr(string(customData), "managed-by: anvil") {
		t.Error("user-managed file should not have managed-by: anvil")
	}

	// Verify developer.md WAS stamped with managed-by
	devData, _ := os.ReadFile(filepath.Join(originalAgents, "developer.md"))
	if !containsStr(string(devData), "managed-by: anvil") {
		t.Error("anvil-deployed file should have managed-by: anvil")
	}

	// Step 3: Restore (uninstall)
	deploy.RestoreItem(
		filepath.Join(paths.Claude, config.CompAgents),
		filepath.Join(snap, config.TargetClaude, config.CompAgents),
	)

	// Verify original custom agent is restored
	restored, err := os.ReadFile(filepath.Join(paths.Claude, config.CompAgents, "custom.md"))
	if err != nil {
		t.Fatalf("read restored file: %v", err)
	}
	if !containsStr(string(restored), "my custom agent") {
		t.Errorf("restored content = %q, want to contain 'my custom agent'", string(restored))
	}
}

func Test_DeployClaude_KeepsOnCollision(t *testing.T) {
	repo := setupFakeRepo(t)
	paths := setupFakeTargets(t)
	cfg := loadTestConfig(t, repo)

	os.MkdirAll(paths.Claude, 0o755)

	agentDir := filepath.Join(paths.Claude, config.CompAgents)
	os.MkdirAll(agentDir, 0o755)
	os.WriteFile(filepath.Join(agentDir, "developer.md"), []byte("---\nname: developer\n---\n\nmy custom developer"), 0o644)

	// Simulate: "a" (all at once), "k" (keep)
	mockStdin("a\nk\n")
	defer resetStdin()

	deploy.StartSummary()
	deploy.Claude(cfg, paths)

	// Verify user's developer.md was preserved
	data, _ := os.ReadFile(filepath.Join(agentDir, "developer.md"))
	content := string(data)
	if containsStr(content, "managed-by: anvil") {
		t.Error("user's developer.md should not have been overwritten")
	}
	if !containsStr(content, "my custom developer") {
		t.Error("user's original content should be preserved")
	}

	// Verify tester.md was still deployed
	testerData, err := os.ReadFile(filepath.Join(agentDir, "tester.md"))
	if err != nil {
		t.Fatal("tester.md should have been deployed")
	}
	if !containsStr(string(testerData), "managed-by: anvil") {
		t.Error("tester.md should be stamped with managed-by: anvil")
	}
}

func Test_DeployClaude_ReplacesOnCollision(t *testing.T) {
	repo := setupFakeRepo(t)
	paths := setupFakeTargets(t)
	cfg := loadTestConfig(t, repo)

	os.MkdirAll(paths.Claude, 0o755)

	agentDir := filepath.Join(paths.Claude, config.CompAgents)
	os.MkdirAll(agentDir, 0o755)
	os.WriteFile(filepath.Join(agentDir, "developer.md"), []byte("---\nname: developer\n---\n\nmy custom developer"), 0o644)

	// Simulate: "a" (all at once), "r" (replace)
	mockStdin("a\nr\n")
	defer resetStdin()

	deploy.StartSummary()
	deploy.Claude(cfg, paths)

	// Verify developer.md was replaced with anvil's version
	data, _ := os.ReadFile(filepath.Join(agentDir, "developer.md"))
	content := string(data)
	if !containsStr(content, "managed-by: anvil") {
		t.Error("developer.md should now have managed-by: anvil")
	}
	if containsStr(content, "my custom developer") {
		t.Error("user's old content should have been replaced")
	}
}

func Test_DeployClaude_UpdatesManagedAgents(t *testing.T) {
	repo := setupFakeRepo(t)
	paths := setupFakeTargets(t)
	cfg := loadTestConfig(t, repo)

	os.MkdirAll(paths.Claude, 0o755)

	agentDir := filepath.Join(paths.Claude, config.CompAgents)
	os.MkdirAll(agentDir, 0o755)
	os.WriteFile(filepath.Join(agentDir, "developer.md"), []byte("---\nmanaged-by: anvil\nmodel: old-model\n---\n\nold version"), 0o644)

	deploy.StartSummary()
	deploy.Claude(cfg, paths)

	// Verify developer.md was updated (old content replaced)
	data, _ := os.ReadFile(filepath.Join(agentDir, "developer.md"))
	content := string(data)
	if containsStr(content, "old version") {
		t.Error("old managed content should have been replaced")
	}
	if !containsStr(content, "managed-by: anvil") {
		t.Error("updated file should still have managed-by: anvil")
	}
}

func Test_DeployOpenCode_AdaptsAgentFormat(t *testing.T) {
	repo := setupFakeRepo(t)
	paths := setupFakeTargets(t)
	cfg := loadTestConfig(t, repo)

	os.MkdirAll(paths.OpenCode, 0o755)
	deploy.StartSummary()
	deploy.OpenCode(cfg, paths)

	agentDir := filepath.Join(paths.OpenCode, config.CompAgents)
	if !fileutil.IsDir(agentDir) {
		t.Fatal("opencode agents directory not created")
	}

	// Verify adapted format (should have mode: subagent)
	data, _ := os.ReadFile(filepath.Join(agentDir, "developer.md"))
	content := string(data)
	if !containsStr(content, "mode: subagent") {
		t.Error("opencode agent should contain 'mode: subagent'")
	}

	// commands should be copied (not symlinked)
	cmdsDir := filepath.Join(paths.OpenCode, config.CompCommands)
	if !fileutil.IsDir(cmdsDir) {
		t.Fatal("opencode commands directory not created")
	}
	if fileutil.IsSymlink(cmdsDir) {
		t.Error("opencode commands should be copied, not symlinked")
	}
}

func Test_StateRecordDeploy_TracksPrevious(t *testing.T) {
	stateDir := filepath.Join(t.TempDir(), ".anvil")
	st, err := state.Load(stateDir)
	if err != nil {
		t.Fatalf("load state: %v", err)
	}

	// Initial state
	if st.DeployedVersion != state.StateNone {
		t.Errorf("initial version = %q, want %q", st.DeployedVersion, state.StateNone)
	}

	// First deploy
	st.RecordDeploy("v1.0.0", "abc123", "main", "claude", []string{config.TargetClaude})
	if err := st.Save(); err != nil {
		t.Fatalf("save: %v", err)
	}

	if st.DeployedVersion != "v1.0.0" {
		t.Errorf("deployed = %q, want %q", st.DeployedVersion, "v1.0.0")
	}
	if st.PreviousVersion != state.StateNone {
		t.Errorf("previous = %q, want %q", st.PreviousVersion, state.StateNone)
	}

	// Second deploy (simulates update)
	st.RecordDeploy("v1.1.0", "def456", "main", "claude", []string{config.TargetClaude})
	if err := st.Save(); err != nil {
		t.Fatalf("save: %v", err)
	}

	if st.DeployedVersion != "v1.1.0" {
		t.Errorf("deployed = %q, want %q", st.DeployedVersion, "v1.1.0")
	}
	if st.PreviousVersion != "v1.0.0" {
		t.Errorf("previous = %q, want %q", st.PreviousVersion, "v1.0.0")
	}

	// Reload from disk and verify persistence
	st2, err := state.Load(stateDir)
	if err != nil {
		t.Fatalf("reload state: %v", err)
	}
	if st2.DeployedVersion != "v1.1.0" {
		t.Errorf("reloaded deployed = %q, want %q", st2.DeployedVersion, "v1.1.0")
	}
	if st2.PreviousVersion != "v1.0.0" {
		t.Errorf("reloaded previous = %q, want %q", st2.PreviousVersion, "v1.0.0")
	}
}

func Test_DeployClaude_IncludeFilter(t *testing.T) {
	repo := setupFakeRepo(t)
	paths := setupFakeTargets(t)

	// Config with include filter: only deploy developer
	manifest := `targets:
  claude:
    enabled: true
  opencode:
    enabled: false
  gemini:
    enabled: false
  codex:
    enabled: false
  cursor:
    enabled: false
components:
  agents:
    tag: "HEAD"
    include:
      - developer
  skills:
    tag: "HEAD"
  commands:
    tag: "HEAD"
`
	cfg := `provider: claude
providers:
  - claude-opus-4-6
  - claude-sonnet-4-6
  - claude-haiku-4-5-20251001
permissions:
  claude:
    read: "Read,Glob,Grep"
    write: "Read,Glob,Grep,Write,Edit"
    execute: "*"
`
	os.WriteFile(filepath.Join(repo, "anvil.yaml"), []byte(manifest), 0o644)
	os.WriteFile(filepath.Join(repo, "anvil.config.yaml"), []byte(cfg), 0o644)

	app, err := config.Load(repo, "anvil")
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	os.MkdirAll(paths.Claude, 0o755)
	deploy.StartSummary()
	deploy.Claude(app, paths)

	agentDir := filepath.Join(paths.Claude, config.CompAgents)
	entries, _ := os.ReadDir(agentDir)
	if len(entries) != 1 {
		t.Errorf("expected 1 agent (include filter), got %d", len(entries))
	}
	if len(entries) > 0 && entries[0].Name() != "developer.md" {
		t.Errorf("expected developer.md, got %s", entries[0].Name())
	}
}

func Test_DeployClaude_ExcludeFilter(t *testing.T) {
	repo := setupFakeRepo(t)
	paths := setupFakeTargets(t)

	manifest := `targets:
  claude:
    enabled: true
  opencode:
    enabled: false
  gemini:
    enabled: false
  codex:
    enabled: false
  cursor:
    enabled: false
components:
  agents:
    tag: "HEAD"
    exclude:
      - tester
  skills:
    tag: "HEAD"
  commands:
    tag: "HEAD"
`
	cfg := `provider: claude
providers:
  - claude-opus-4-6
  - claude-sonnet-4-6
  - claude-haiku-4-5-20251001
permissions:
  claude:
    read: "Read,Glob,Grep"
    write: "Read,Glob,Grep,Write,Edit"
    execute: "*"
`
	os.WriteFile(filepath.Join(repo, "anvil.yaml"), []byte(manifest), 0o644)
	os.WriteFile(filepath.Join(repo, "anvil.config.yaml"), []byte(cfg), 0o644)

	app, err := config.Load(repo, "anvil")
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	os.MkdirAll(paths.Claude, 0o755)
	deploy.StartSummary()
	deploy.Claude(app, paths)

	agentDir := filepath.Join(paths.Claude, config.CompAgents)
	entries, _ := os.ReadDir(agentDir)
	if len(entries) != 1 {
		t.Errorf("expected 1 agent (exclude filter), got %d", len(entries))
	}
	if len(entries) > 0 && entries[0].Name() != "developer.md" {
		t.Errorf("expected developer.md, got %s", entries[0].Name())
	}
}

func Test_DeployClaude_UserSkillsPreserved(t *testing.T) {
	repo := setupFakeRepo(t)
	paths := setupFakeTargets(t)
	cfg := loadTestConfig(t, repo)

	os.MkdirAll(paths.Claude, 0o755)

	// Pre-create a user-managed skill
	userSkill := filepath.Join(paths.Claude, config.CompSkills, "my-custom-skill")
	os.MkdirAll(userSkill, 0o755)
	os.WriteFile(filepath.Join(userSkill, "SKILL.md"), []byte("# My skill"), 0o644)

	deploy.StartSummary()
	deploy.Claude(cfg, paths)

	// Verify user skill was preserved
	if !fileutil.IsDir(userSkill) {
		t.Error("user-managed skill should be preserved")
	}
	data, _ := os.ReadFile(filepath.Join(userSkill, "SKILL.md"))
	if string(data) != "# My skill" {
		t.Error("user skill content should be unchanged")
	}

	// Verify anvil skill was linked
	goConv := filepath.Join(paths.Claude, config.CompSkills, "go-conventions")
	if !fileutil.IsSymlink(goConv) {
		t.Error("anvil skill should be symlinked")
	}
}

func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstr(s, substr))
}

func containsSubstr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
