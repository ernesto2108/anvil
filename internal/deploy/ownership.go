package deploy

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ernesto2108/anvil/pkg/frontmatter"
	"github.com/ernesto2108/anvil/pkg/output"
)

const (
	ManagedByKey   = "managed-by"
	ManagedByValue = "anvil"
)

// CollisionAction represents what to do with a conflicting file.
type CollisionAction int

const (
	ActionKeep    CollisionAction = iota // keep user's version
	ActionReplace                        // replace with anvil's
	ActionRename                         // deploy anvil's with a different name
)

// CollisionResolution holds the decision for one file.
type CollisionResolution struct {
	Name      string
	Action    CollisionAction
	NewName   string // only used when Action == ActionRename
}

// CollisionResult holds the result of a collision check.
type CollisionResult struct {
	Resolutions []CollisionResolution
	Safe        []string
}

// IsAnvilSymlink returns true if path is a symlink (symlinks in agent dirs
// are always created by anvil, even if the target lacks managed-by frontmatter).
func IsAnvilSymlink(path string) bool {
	fi, err := os.Lstat(path)
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeSymlink != 0
}

// IsManagedByAnvil returns true if the file was deployed by anvil.
// This includes files with managed-by: anvil frontmatter AND symlinks
// (which are always created by anvil in deploy target directories).
func IsManagedByAnvil(path string) bool {
	if IsAnvilSymlink(path) {
		return true
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	return frontmatter.HasField(string(data), ManagedByKey, ManagedByValue)
}

// CleanManagedFiles removes only files with managed-by: anvil from the directory.
func CleanManagedFiles(dir string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	var removed []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		path := filepath.Join(dir, e.Name())
		if IsManagedByAnvil(path) {
			os.Remove(path)
			removed = append(removed, e.Name())
		}
	}
	return removed
}

// ResolveCollisions detects collisions and prompts the user interactively.
// Returns the resolution for each collision and the list of safe files.
func ResolveCollisions(anvilFiles []string, targetDir string, reader *bufio.Reader) CollisionResult {
	var result CollisionResult

	// Gather collisions first
	var collisions []string
	for _, f := range anvilFiles {
		name := filepath.Base(f)
		targetPath := filepath.Join(targetDir, name)

		if _, err := os.Stat(targetPath); os.IsNotExist(err) {
			result.Safe = append(result.Safe, name)
			continue
		}

		if IsManagedByAnvil(targetPath) {
			result.Safe = append(result.Safe, name)
		} else {
			collisions = append(collisions, name)
		}
	}

	if len(collisions) == 0 {
		return result
	}

	// Show header
	fmt.Println()
	output.Warn("Found %d agent(s) with the same name as yours:", len(collisions))
	for _, c := range collisions {
		fmt.Printf("    %s\n", c)
	}
	fmt.Println()

	// Ask bulk or individual
	fmt.Printf("  Resolve: [a]ll at once  [o]ne by one  > ")
	mode := readChoice(reader)

	if mode == "a" {
		fmt.Printf("  Action for all %d: [k]eep yours  [r]eplace with anvil  [R]ename anvil copies  > ", len(collisions))
		action := readChoice(reader)
		for _, name := range collisions {
			result.Resolutions = append(result.Resolutions, resolveOne(name, action, reader))
		}
	} else {
		for _, name := range collisions {
			result.Resolutions = append(result.Resolutions, promptOne(name, targetDir, anvilFiles, reader))
		}
	}

	fmt.Println()
	return result
}

func promptOne(name, targetDir string, anvilFiles []string, reader *bufio.Reader) CollisionResolution {
	fmt.Printf("  %s %s\n", output.Yellow("▸"), name)
	fmt.Printf("    [k]eep yours  [r]eplace with anvil  [R]ename anvil copy  [d]iff  > ")

	for {
		choice := readChoice(reader)
		switch choice {
		case "d":
			showDiff(name, targetDir, anvilFiles)
			fmt.Printf("    [k]eep  [r]eplace  [R]ename  > ")
			continue
		default:
			return resolveOne(name, choice, reader)
		}
	}
}

func resolveOne(name, choice string, reader *bufio.Reader) CollisionResolution {
	switch choice {
	case "r":
		return CollisionResolution{Name: name, Action: ActionReplace}
	case "R":
		base := strings.TrimSuffix(name, ".md")
		suggested := base + "-anvil.md"
		fmt.Printf("    Rename to [%s]: ", suggested)
		custom := readChoice(reader)
		if custom == "" {
			custom = suggested
		}
		if !strings.HasSuffix(custom, ".md") {
			custom += ".md"
		}
		return CollisionResolution{Name: name, Action: ActionRename, NewName: custom}
	default: // "k" or anything else = keep
		return CollisionResolution{Name: name, Action: ActionKeep}
	}
}

func showDiff(name, targetDir string, anvilFiles []string) {
	existing := filepath.Join(targetDir, name)

	// Find the source file in anvilFiles
	var source string
	for _, f := range anvilFiles {
		if filepath.Base(f) == name {
			source = f
			break
		}
	}
	if source == "" {
		return
	}

	fmt.Println()
	cmd := exec.Command("diff", "--color=always", "-u",
		"--label", "yours: "+name,
		"--label", "anvil: "+name,
		existing, source)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run() // exit code 1 means files differ — that's OK
	fmt.Println()
}

func readChoice(reader *bufio.Reader) string {
	line, _ := reader.ReadString('\n')
	return strings.TrimSpace(line)
}

// ApplyResolutions returns which files to skip, which to deploy, and which to rename.
func ApplyResolutions(resolutions []CollisionResolution) (skip []string, replace []string, renames map[string]string) {
	renames = make(map[string]string)
	for _, r := range resolutions {
		switch r.Action {
		case ActionKeep:
			skip = append(skip, r.Name)
		case ActionReplace:
			replace = append(replace, r.Name)
		case ActionRename:
			renames[r.Name] = r.NewName
		}
	}
	return
}

// ShouldDeployFile returns true if the file should be deployed (not skipped).
func ShouldDeployFile(name string, skip []string) bool {
	for _, s := range skip {
		if s == name {
			return false
		}
	}
	return true
}

// GetDeployName returns the actual filename to use (handles renames).
func GetDeployName(name string, renames map[string]string) string {
	if newName, ok := renames[name]; ok {
		return newName
	}
	return name
}

// StampManagedBy adds managed-by: anvil to a content string.
func StampManagedBy(content string) string {
	return frontmatter.SetField(content, ManagedByKey, ManagedByValue)
}
