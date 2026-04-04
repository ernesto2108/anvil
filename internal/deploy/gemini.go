package deploy

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ernesto2108/anvil/internal/config"
	"github.com/ernesto2108/anvil/internal/fileutil"
	"github.com/ernesto2108/anvil/internal/frontmatter"
	"github.com/ernesto2108/anvil/internal/output"
)

func Gemini(cfg *config.App, paths TargetPaths) {
	target := paths.Gemini
	output.Info("%s -> %s", output.Bold("Gemini CLI"), target)

	output.Info("  agents -> skipped (not supported)")
	DeploySkillsSymlink(cfg.RepoDir, target)
	deployGeminiCommands(cfg, target)
	deployGeminiMD(cfg, target)
}

func adaptCommandGemini(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("read %s: %w", src, err)
	}

	doc := frontmatter.Parse(string(data))
	desc := doc.Fields["description"]
	if desc == "" {
		desc = doc.Fields["name"]
	}

	body := doc.Body
	body = strings.ReplaceAll(body, "$ARGUMENTS", "{{args}}")
	body = strings.ReplaceAll(body, "$1", "{{args}}")

	content := fmt.Sprintf("description = \"%s\"\nprompt = \"\"\"\n%s\n\"\"\"", desc, body)
	return os.WriteFile(dst, []byte(content), 0o644)
}

func deployGeminiCommands(cfg *config.App, target string) {
	cmdSrc := filepath.Join(cfg.RepoDir, "commands")
	if !fileutil.IsDir(cmdSrc) {
		return
	}

	cmdDst := filepath.Join(target, "commands")
	fileutil.CleanPath(cmdDst)
	os.MkdirAll(cmdDst, 0o755)

	count := 0

	// Flat commands
	entries, _ := os.ReadDir(cmdSrc)
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if filepath.Ext(e.Name()) != ".md" {
			continue
		}
		name := strings.TrimSuffix(e.Name(), ".md")
		src := filepath.Join(cmdSrc, e.Name())
		dst := filepath.Join(cmdDst, name+".toml")
		if err := adaptCommandGemini(src, dst); err != nil {
			output.Error("adapt command %s: %s", e.Name(), err)
			continue
		}
		count++
	}

	// Subdirectory commands
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		subSrc := filepath.Join(cmdSrc, e.Name())
		subDst := filepath.Join(cmdDst, e.Name())
		os.MkdirAll(subDst, 0o755)

		subEntries, _ := os.ReadDir(subSrc)
		for _, se := range subEntries {
			if filepath.Ext(se.Name()) != ".md" {
				continue
			}
			name := strings.TrimSuffix(se.Name(), ".md")
			src := filepath.Join(subSrc, se.Name())
			dst := filepath.Join(subDst, name+".toml")
			if err := adaptCommandGemini(src, dst); err != nil {
				output.Error("adapt command %s/%s: %s", e.Name(), se.Name(), err)
				continue
			}
			count++
		}
	}

	output.Info("  %d commands -> toml", count)
}

func deployGeminiMD(cfg *config.App, target string) {
	src := filepath.Join(cfg.RepoDir, "CLAUDE.md")
	if !fileutil.Exists(src) {
		return
	}
	dst := filepath.Join(target, "GEMINI.md")
	if err := fileutil.CopyFile(src, dst); err != nil {
		output.Error("copy GEMINI.md: %s", err)
		return
	}
	output.Info("  GEMINI.md -> copied from CLAUDE.md")
}
