package transcript

import "strings"

// genericNames is the set of directory names too generic to use as a project
// name. ResolveProject uses IsGeneric to reject these before falling back to
// the next resolution strategy.
var genericNames = map[string]struct{}{
	"projects":  {},
	"src":       {},
	"home":      {},
	"users":     {},
	"workspace": {},
	"code":      {},
	"dev":       {},
}

// IsGeneric reports whether name should be rejected as a project identifier
// because it is too common to be meaningful. Comparison is case-insensitive.
func IsGeneric(name string) bool {
	_, ok := genericNames[strings.ToLower(name)]
	return ok
}
