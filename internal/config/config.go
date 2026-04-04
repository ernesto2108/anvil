package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Target struct {
	Enabled bool   `yaml:"enabled"`
	Path    string `yaml:"path"`
}

type Component struct {
	Tag string `yaml:"tag"`
}

type FileEntry struct {
	Tag string `yaml:"tag"`
}

type Manifest struct {
	Targets        map[string]Target    `yaml:"targets"`
	CursorProjects []string             `yaml:"cursor_projects"`
	Components     map[string]Component `yaml:"components"`
	Files          map[string]FileEntry `yaml:"files"`
	Ignore         []string             `yaml:"ignore"`
}

type TierMap struct {
	High   string `yaml:"high"`
	Medium string `yaml:"medium"`
	Low    string `yaml:"low"`
}

type PermMap struct {
	Read    string `yaml:"read"`
	Write   string `yaml:"write"`
	Execute string `yaml:"execute"`
}

type ProviderConfig struct {
	Provider    string             `yaml:"provider"`
	Providers   map[string]TierMap `yaml:"providers"`
	Permissions map[string]PermMap `yaml:"permissions"`
}

type App struct {
	Name       string
	RepoDir    string
	Manifest   Manifest
	Provider   ProviderConfig
	manifestRaw []byte
	configPath  string
}

func Load(repoDir, appName string) (*App, error) {
	manifestPath := filepath.Join(repoDir, appName+".yaml")
	configPath := filepath.Join(repoDir, appName+".config.yaml")

	manifestData, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("read manifest %s: %w", manifestPath, err)
	}

	var manifest Manifest
	if err := yaml.Unmarshal(manifestData, &manifest); err != nil {
		return nil, fmt.Errorf("parse manifest: %w", err)
	}

	configData, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("read config %s: %w", configPath, err)
	}

	var provider ProviderConfig
	if err := yaml.Unmarshal(configData, &provider); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	return &App{
		Name:        appName,
		RepoDir:     repoDir,
		Manifest:    manifest,
		Provider:    provider,
		manifestRaw: manifestData,
		configPath:  configPath,
	}, nil
}

func (a *App) TargetEnabled(name string) bool {
	t, ok := a.Manifest.Targets[name]
	return ok && t.Enabled
}

func (a *App) AllTargets() []string {
	return []string{"claude", "opencode", "gemini", "codex", "cursor"}
}

func (a *App) ActiveProvider() string {
	return a.Provider.Provider
}

func (a *App) ListProviders() []string {
	var names []string
	for k := range a.Provider.Providers {
		names = append(names, k)
	}
	return names
}

func (a *App) ResolveTier(tier, provider string) (string, error) {
	if provider == "" {
		provider = a.Provider.Provider
	}
	tm, ok := a.Provider.Providers[provider]
	if !ok {
		return "", fmt.Errorf("provider %q not found in config", provider)
	}

	switch tier {
	case "high":
		return tm.High, nil
	case "medium":
		return tm.Medium, nil
	case "low":
		return tm.Low, nil
	default:
		return tier, nil
	}
}

func (a *App) ResolvePermission(perm, tool string) string {
	pm, ok := a.Provider.Permissions[tool]
	if !ok {
		return ""
	}
	switch perm {
	case "read":
		return pm.Read
	case "write":
		return pm.Write
	case "execute":
		return pm.Execute
	default:
		return ""
	}
}

func (a *App) SetTargetEnabled(name string, enabled bool) error {
	content := string(a.manifestRaw)

	oldVal := "true"
	newVal := "false"
	if enabled {
		oldVal = "false"
		newVal = "true"
	}

	inBlock := false
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == name+":" && strings.HasPrefix(line, "  ") && !strings.HasPrefix(line, "    ") {
			inBlock = true
			continue
		}
		if inBlock {
			if strings.Contains(line, "enabled: "+oldVal) {
				lines[i] = strings.Replace(line, "enabled: "+oldVal, "enabled: "+newVal, 1)
				break
			}
			if !strings.HasPrefix(line, "    ") && trimmed != "" {
				break
			}
		}
	}

	result := strings.Join(lines, "\n")
	manifestPath := filepath.Join(a.RepoDir, a.Name+".yaml")
	if err := os.WriteFile(manifestPath, []byte(result), 0o644); err != nil {
		return fmt.Errorf("write manifest: %w", err)
	}

	a.manifestRaw = []byte(result)
	t := a.Manifest.Targets[name]
	t.Enabled = enabled
	a.Manifest.Targets[name] = t
	return nil
}

func (a *App) SetProvider(name string) error {
	data, err := os.ReadFile(a.configPath)
	if err != nil {
		return fmt.Errorf("read config: %w", err)
	}

	content := string(data)
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if strings.HasPrefix(line, "provider:") {
			lines[i] = "provider: " + name
			break
		}
	}

	result := strings.Join(lines, "\n")
	if err := os.WriteFile(a.configPath, []byte(result), 0o644); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	a.Provider.Provider = name
	return nil
}

func (a *App) ExpandedCursorProjects() []string {
	home, _ := os.UserHomeDir()
	var result []string
	for _, p := range a.Manifest.CursorProjects {
		if strings.HasPrefix(p, "~/") {
			p = filepath.Join(home, p[2:])
		}
		result = append(result, p)
	}
	return result
}
