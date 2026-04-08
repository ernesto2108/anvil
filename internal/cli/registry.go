package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ernesto2108/anvil/pkg/config"
	"github.com/ernesto2108/anvil/pkg/frontmatter"
	"github.com/ernesto2108/anvil/pkg/output"
	"github.com/ernesto2108/anvil/pkg/registry"
)

func cmdRegistry(cfg *config.App, args []string) {
	if len(args) == 0 {
		registryHelp()
		return
	}

	switch args[0] {
	case "list":
		registryList(cfg, args[1:])
	case "add":
		registryAdd(cfg, args[1:])
	case "search":
		registrySearch(cfg, args[1:])
	default:
		output.Error("Unknown registry command: %s", args[0])
		registryHelp()
		os.Exit(1)
	}
}

func registryList(cfg *config.App, args []string) {
	regs := cfg.Manifest.Registries
	if len(regs) == 0 {
		output.Info("No registries configured. Add one to anvil.yaml:")
		fmt.Println()
		fmt.Println("  registries:")
		fmt.Println("    - name: community")
		fmt.Println("      url: https://example.com/registry.json")
		return
	}

	typeFilter := ""
	if len(args) > 0 {
		typeFilter = args[0] // "agents" or "skills"
	}

	for _, reg := range regs {
		output.Info("%s (%s)", output.Bold(reg.Name), reg.URL)

		idx, err := registry.Fetch(reg.URL)
		if err != nil {
			output.Error("  fetch failed: %s", err)
			continue
		}

		entries := idx.Entries
		if typeFilter != "" {
			// Normalize: "agents" -> "agent"
			t := strings.TrimSuffix(typeFilter, "s")
			entries = registry.FilterByType(entries, t)
		}

		if len(entries) == 0 {
			output.Info("  (no entries)")
			continue
		}

		// Check which ones are already installed locally
		localAgents := listLocalNames(cfg.RepoDir, config.CompAgents)
		localSkills := listLocalNames(cfg.RepoDir, config.CompSkills)

		for _, e := range entries {
			status := ""
			if e.Type == "agent" && contains(localAgents, e.Name) {
				status = output.Green(" (installed)")
			} else if e.Type == "skill" && contains(localSkills, e.Name) {
				status = output.Green(" (installed)")
			}

			author := ""
			if e.Author != "" {
				author = fmt.Sprintf(" by %s", e.Author)
			}

			fmt.Printf("  %-12s %-20s %s%s%s\n",
				output.Cyan(e.Type), e.Name, e.Description, author, status)
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
		idx, err := registry.Fetch(reg.URL)
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

func registryAdd(cfg *config.App, args []string) {
	if len(args) == 0 {
		output.Error("Usage: registry add <name> [--from <registry>]")
		os.Exit(1)
	}

	name := args[0]
	fromRegistry := ""
	for i, a := range args {
		if a == "--from" && i+1 < len(args) {
			fromRegistry = args[i+1]
		}
	}

	regs := cfg.Manifest.Registries
	if len(regs) == 0 {
		output.Error("No registries configured")
		os.Exit(1)
	}

	// Search all registries for the entry
	var entry *registry.Entry
	var source string
	for _, reg := range regs {
		if fromRegistry != "" && reg.Name != fromRegistry {
			continue
		}

		idx, err := registry.Fetch(reg.URL)
		if err != nil {
			output.Warn("  %s: fetch failed, skipping", reg.Name)
			continue
		}

		if found := registry.FindByName(idx.Entries, name); found != nil {
			entry = found
			source = reg.Name
			break
		}
	}

	if entry == nil {
		output.Error("Entry %q not found in any registry", name)
		os.Exit(1)
	}

	output.Info("Found %s %q in %s", entry.Type, entry.Name, output.Bold(source))

	// Download and install
	data, err := registry.DownloadFile(entry.URL)
	if err != nil {
		output.Error("Download failed: %s", err)
		os.Exit(1)
	}

	// Determine destination based on type
	var destDir string
	switch entry.Type {
	case "agent":
		destDir = filepath.Join(cfg.RepoDir, config.CompAgents)
	case "skill":
		destDir = filepath.Join(cfg.RepoDir, config.CompSkills, entry.Name)
	default:
		output.Error("Unknown entry type: %s", entry.Type)
		os.Exit(1)
	}

	os.MkdirAll(destDir, 0o755)

	var destFile string
	if entry.Type == "agent" {
		destFile = filepath.Join(destDir, entry.Name+".md")

		// Check for collision with existing agent
		if _, err := os.Stat(destFile); err == nil {
			existing, _ := os.ReadFile(destFile)
			content := string(existing)
			managedBy := frontmatter.Get(content, "managed-by")
			if managedBy != "anvil" && managedBy != "" {
				output.Error("Agent %q already exists and is not managed by anvil. Use --force to overwrite.", entry.Name)
				os.Exit(1)
			}
		}

		// Stamp with registry source
		content := string(data)
		content = frontmatter.SetField(content, "managed-by", "registry:"+source)
		os.WriteFile(destFile, []byte(content), 0o644)
	} else {
		// For skills, write as SKILL.md
		destFile = filepath.Join(destDir, "SKILL.md")
		os.WriteFile(destFile, data, 0o644)
	}

	output.Info("Installed %s %q -> %s", entry.Type, entry.Name, destFile)
	output.Info("Run %s to update all targets", output.Cyan("anvil deploy"))
}

func registryHelp() {
	fmt.Println(`
  REGISTRY COMMANDS:
    registry list [agents|skills]    List available agents/skills from registries
    registry search <query>          Search across all registries
    registry add <name> [--from r]   Download and install from registry

  Configure registries in anvil.yaml:
    registries:
      - name: community
        url: https://example.com/registry.json

  Registry JSON format:
    {
      "name": "My Registry",
      "entries": [
        {
          "name": "my-agent",
          "type": "agent",
          "description": "Does something useful",
          "version": "1.0.0",
          "url": "https://raw.githubusercontent.com/.../my-agent.md",
          "author": "someone"
        }
      ]
    }`)
}

func listLocalNames(repoDir, component string) []string {
	dir := filepath.Join(repoDir, component)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var names []string
	for _, e := range entries {
		name := e.Name()
		name = strings.TrimSuffix(name, ".md")
		names = append(names, name)
	}
	return names
}

func contains(list []string, item string) bool {
	for _, v := range list {
		if v == item {
			return true
		}
	}
	return false
}
