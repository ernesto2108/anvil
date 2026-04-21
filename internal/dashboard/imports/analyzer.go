package imports

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

// Edge represents a detected dependency between two files in the same run.
type Edge struct {
	SourcePath string // relative path to project root (e.g. "internal/handler/metrics.go")
	TargetPath string // relative path to project root
	Relation   string // always "imports" for now
}

// Analyzer extracts import edges from source files.
// It caches the Go module path per CWD to avoid repeated go.mod reads.
// mu guards modCache for concurrent PostToolUse calls.
type Analyzer struct {
	mu       sync.RWMutex
	cwd      string
	modCache map[string]string // cwd -> module path cache
}

// New creates an Analyzer rooted at the given working directory.
func New(cwd string) *Analyzer {
	return &Analyzer{
		cwd:      cwd,
		modCache: make(map[string]string),
	}
}

// Analyze reads the file at filePath, extracts its imports, and returns
// edges only for targets that exist in touchedPaths (files touched in the same run).
// filePath must be relative to cwd.
// Returns nil, nil if the file extension is not supported or the file doesn't exist.
func (a *Analyzer) Analyze(filePath string, touchedPaths map[string]bool) ([]Edge, error) {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".go":
		return a.analyzeGo(filePath, touchedPaths)
	case ".ts", ".tsx", ".js", ".jsx":
		return a.analyzeTS(filePath, touchedPaths)
	case ".rs":
		return a.analyzeRust(filePath, touchedPaths)
	case ".py":
		return a.analyzePython(filePath, touchedPaths)
	default:
		return nil, nil
	}
}

// readHeadFile reads up to maxLines lines from the file at absPath.
// Returns nil, nil on file-not-found or permission errors (graceful degradation).
func readHeadFile(absPath string, maxLines int) ([]string, error) {
	f, err := os.Open(absPath)
	if err != nil {
		if os.IsNotExist(err) || os.IsPermission(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close() //nolint:errcheck

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() && len(lines) < maxLines {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

// ---- Go ----

var (
	reGoSingleImport = regexp.MustCompile(`^\s*import\s+"([^"]+)"`)
	reGoImportPath   = regexp.MustCompile(`^\s*"([^"]+)"`)
)

// goModulePath returns the module path declared in go.mod inside cwd.
// Result is cached. Returns "" if go.mod cannot be read.
// Uses double-checked locking to be safe under concurrent access.
func (a *Analyzer) goModulePath() string {
	// Fast path: read lock to check cache.
	a.mu.RLock()
	mod, ok := a.modCache[a.cwd]
	a.mu.RUnlock()
	if ok {
		return mod
	}

	// Slow path: compute and cache under write lock.
	a.mu.Lock()
	defer a.mu.Unlock()
	// Re-check after acquiring write lock to avoid duplicate work.
	if mod, ok = a.modCache[a.cwd]; ok {
		return mod
	}

	data, err := os.ReadFile(filepath.Join(a.cwd, "go.mod"))
	if err != nil {
		a.modCache[a.cwd] = ""
		return ""
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "module ") {
			mod = strings.TrimSpace(strings.TrimPrefix(line, "module "))
			a.modCache[a.cwd] = mod
			return mod
		}
	}
	a.modCache[a.cwd] = ""
	return ""
}

func (a *Analyzer) analyzeGo(filePath string, touchedPaths map[string]bool) ([]Edge, error) {
	modPath := a.goModulePath()
	if modPath == "" {
		return nil, nil
	}

	absPath := filepath.Join(a.cwd, filePath)
	lines, err := readHeadFile(absPath, 100)
	if err != nil {
		return nil, err
	}
	if lines == nil {
		return nil, nil
	}

	var imports []string
	inBlock := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if !inBlock {
			if m := reGoSingleImport.FindStringSubmatch(line); m != nil {
				imports = append(imports, m[1])
				continue
			}
			if trimmed == "import (" {
				inBlock = true
				continue
			}
		} else {
			if trimmed == ")" {
				inBlock = false
				continue
			}
			if m := reGoImportPath.FindStringSubmatch(line); m != nil {
				imports = append(imports, m[1])
			}
		}
	}

	var edges []Edge
	prefix := modPath + "/"
	for _, imp := range imports {
		if !strings.HasPrefix(imp, prefix) && imp != modPath {
			continue
		}
		// Strip module prefix to get package directory relative to cwd
		pkgDir := strings.TrimPrefix(imp, prefix)
		// Match against touchedPaths: any file whose directory equals pkgDir
		for touched := range touchedPaths {
			if filepath.Dir(touched) == pkgDir {
				edges = append(edges, Edge{
					SourcePath: filePath,
					TargetPath: touched,
					Relation:   "imports",
				})
				break // one edge per import path is enough
			}
		}
	}
	return edges, nil
}

// ---- TypeScript / JavaScript ----

var (
	reTSFrom    = regexp.MustCompile(`from\s+['"]([^'"]+)['"]`)
	reTSRequire = regexp.MustCompile(`require\(['"]([^'"]+)['"]\)`)
)

func (a *Analyzer) analyzeTS(filePath string, touchedPaths map[string]bool) ([]Edge, error) {
	absPath := filepath.Join(a.cwd, filePath)
	lines, err := readHeadFile(absPath, 100)
	if err != nil {
		return nil, err
	}
	if lines == nil {
		return nil, nil
	}

	fileDir := filepath.Dir(filePath)
	seen := map[string]bool{}
	var edges []Edge

	addIfRelative := func(imp string) {
		if !strings.HasPrefix(imp, "./") && !strings.HasPrefix(imp, "../") {
			return
		}
		// Resolve relative to file's directory (paths are relative to cwd)
		resolved := filepath.Clean(filepath.Join(fileDir, imp))
		for touched := range touchedPaths {
			if seen[touched] {
				continue
			}
			// Strip extension from touched for comparison
			touchedNoExt := strings.TrimSuffix(touched, filepath.Ext(touched))
			if touched == resolved || touchedNoExt == resolved {
				seen[touched] = true
				edges = append(edges, Edge{
					SourcePath: filePath,
					TargetPath: touched,
					Relation:   "imports",
				})
			}
		}
	}

	for _, line := range lines {
		if m := reTSFrom.FindStringSubmatch(line); m != nil {
			addIfRelative(m[1])
		}
		if m := reTSRequire.FindStringSubmatch(line); m != nil {
			addIfRelative(m[1])
		}
	}
	return edges, nil
}

// ---- Rust ----

var (
	reRustMod   = regexp.MustCompile(`^\s*mod\s+(\w+)\s*;`)
	reRustCrate = regexp.MustCompile(`^\s*use\s+crate::([\w:]+)`)
	reRustSuper = regexp.MustCompile(`^\s*use\s+super::([\w:]+)`)
)

func (a *Analyzer) analyzeRust(filePath string, touchedPaths map[string]bool) ([]Edge, error) {
	absPath := filepath.Join(a.cwd, filePath)
	lines, err := readHeadFile(absPath, 100)
	if err != nil {
		return nil, err
	}
	if lines == nil {
		return nil, nil
	}

	fileDir := filepath.Dir(filePath)
	seen := map[string]bool{}
	var edges []Edge

	addEdge := func(candidate string) {
		if seen[candidate] || !touchedPaths[candidate] {
			return
		}
		seen[candidate] = true
		edges = append(edges, Edge{
			SourcePath: filePath,
			TargetPath: candidate,
			Relation:   "imports",
		})
	}

	for _, line := range lines {
		if m := reRustMod.FindStringSubmatch(line); m != nil {
			name := m[1]
			// mod name; → name.rs or name/mod.rs relative to file's directory
			addEdge(filepath.Join(fileDir, name+".rs"))
			addEdge(filepath.Join(fileDir, name, "mod.rs"))
			continue
		}
		if m := reRustCrate.FindStringSubmatch(line); m != nil {
			// use crate::x::y → src/x/y.rs
			rel := strings.ReplaceAll(m[1], "::", "/")
			addEdge(filepath.Join("src", rel+".rs"))
			continue
		}
		if m := reRustSuper.FindStringSubmatch(line); m != nil {
			// use super::x → resolve relative to parent directory
			rel := strings.ReplaceAll(m[1], "::", "/")
			parent := filepath.Dir(fileDir)
			addEdge(filepath.Join(parent, rel+".rs"))
		}
	}
	return edges, nil
}

// ---- Python ----

var (
	rePyFromImport = regexp.MustCompile(`^\s*from\s+([\.\w]+)\s+import`)
	rePyImport     = regexp.MustCompile(`^\s*import\s+([\w\.]+)`)
)

func (a *Analyzer) analyzePython(filePath string, touchedPaths map[string]bool) ([]Edge, error) {
	absPath := filepath.Join(a.cwd, filePath)
	lines, err := readHeadFile(absPath, 100)
	if err != nil {
		return nil, err
	}
	if lines == nil {
		return nil, nil
	}

	fileDir := filepath.Dir(filePath)
	seen := map[string]bool{}
	var edges []Edge

	addPyEdge := func(modStr string) {
		if !strings.HasPrefix(modStr, ".") {
			// Absolute import: match if it corresponds to a touched file's path
			candidate := strings.ReplaceAll(modStr, ".", "/") + ".py"
			if !seen[candidate] && touchedPaths[candidate] {
				seen[candidate] = true
				edges = append(edges, Edge{
					SourcePath: filePath,
					TargetPath: candidate,
					Relation:   "imports",
				})
			}
			return
		}
		// Relative import: count leading dots
		dots := 0
		for _, ch := range modStr {
			if ch == '.' {
				dots++
			} else {
				break
			}
		}
		rest := modStr[dots:]
		// Walk up 'dots' directory levels from fileDir
		base := fileDir
		for i := 1; i < dots; i++ {
			base = filepath.Dir(base)
		}
		var candidate string
		if rest == "" {
			candidate = filepath.Join(base, "__init__.py")
		} else {
			rel := strings.ReplaceAll(rest, ".", "/")
			candidate = filepath.Join(base, rel+".py")
		}
		if !seen[candidate] && touchedPaths[candidate] {
			seen[candidate] = true
			edges = append(edges, Edge{
				SourcePath: filePath,
				TargetPath: candidate,
				Relation:   "imports",
			})
		}
	}

	for _, line := range lines {
		if m := rePyFromImport.FindStringSubmatch(line); m != nil {
			addPyEdge(m[1])
			continue
		}
		if m := rePyImport.FindStringSubmatch(line); m != nil {
			addPyEdge(m[1])
		}
	}
	return edges, nil
}
