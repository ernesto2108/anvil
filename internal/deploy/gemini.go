package deploy

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ernesto2108/anvil/pkg/config"
	"github.com/ernesto2108/anvil/pkg/fileutil"
	"github.com/ernesto2108/anvil/pkg/frontmatter"
	"github.com/ernesto2108/anvil/pkg/output"
)

func Gemini(cfg *config.App, paths TargetPaths) {
	target := paths.Gemini
	ts := AddTarget("Gemini CLI")

	DeploySkillsSymlink(cfg, target, ts)
	deployGeminiCommands(cfg, target, ts)
	deployGeminiMD(cfg, target, ts)
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

func deployGeminiCommands(cfg *config.App, target string, ts *TargetStats) {
	cmdSrc := filepath.Join(cfg.RepoDir, config.CompCommands)
	if !fileutil.IsDir(cmdSrc) {
		return
	}

	cmdDst := filepath.Join(target, config.CompCommands)
	fileutil.CleanPath(cmdDst)
	if err := os.MkdirAll(cmdDst, 0o755); err != nil {
		output.Error("create gemini commands dir: %s", err)
		return
	}

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
		if err := os.MkdirAll(subDst, 0o755); err != nil {
			output.Error("create gemini subcommands dir %s: %s", e.Name(), err)
			continue
		}

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

	ts.Commands.Deployed = count
}

func deployGeminiMD(cfg *config.App, target string, ts *TargetStats) {
	src := filepath.Join(cfg.RepoDir, config.FileClaudeMD)
	if !fileutil.Exists(src) {
		return
	}
	dst := filepath.Join(target, config.FileGeminiMD)
	if err := fileutil.CopyFile(src, dst); err != nil {
		output.Error("copy %s: %s", config.FileGeminiMD, err)
		return
	}
	ts.Extras = append(ts.Extras, config.FileGeminiMD+" -> copied")
}
