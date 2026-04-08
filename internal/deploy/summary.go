package deploy

import (
	"fmt"
	"strings"

	"github.com/ernesto2108/anvil/pkg/output"
)

// ComponentStats tracks deploy results for one component type.
type ComponentStats struct {
	Deployed  int
	Skipped   int
	Preserved int // user-managed files that were kept
	Names     []string
}

// TargetStats tracks deploy results for one target.
type TargetStats struct {
	Name     string
	Agents   ComponentStats
	Skills   ComponentStats
	Commands ComponentStats
	Extras   []string // additional items like CLAUDE.md, AGENTS.md
}

// DeploySummary collects stats across all targets for a final report.
type DeploySummary struct {
	Targets []TargetStats
}

var activeSummary *DeploySummary

// StartSummary initializes a new deploy summary.
func StartSummary() {
	activeSummary = &DeploySummary{}
}

// AddTarget adds a target stats entry and returns a pointer for mutation.
func AddTarget(name string) *TargetStats {
	if activeSummary == nil {
		StartSummary()
	}
	ts := TargetStats{Name: name}
	activeSummary.Targets = append(activeSummary.Targets, ts)
	return &activeSummary.Targets[len(activeSummary.Targets)-1]
}

// PrintSummary prints the final deploy report.
func PrintSummary() {
	if activeSummary == nil || len(activeSummary.Targets) == 0 {
		return
	}

	fmt.Println()
	output.Info("%s", output.Bold("Deploy Summary"))
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	for _, t := range activeSummary.Targets {
		fmt.Printf("  %s\n", output.Bold(t.Name))

		// Agents
		if t.Agents.Deployed > 0 || t.Agents.Preserved > 0 {
			parts := []string{}
			if t.Agents.Deployed > 0 {
				parts = append(parts, fmt.Sprintf("%s deployed", output.Green(fmt.Sprintf("%d", t.Agents.Deployed))))
			}
			if t.Agents.Preserved > 0 {
				parts = append(parts, fmt.Sprintf("%d user-managed", t.Agents.Preserved))
			}
			if t.Agents.Skipped > 0 {
				parts = append(parts, fmt.Sprintf("%d filtered", t.Agents.Skipped))
			}
			fmt.Printf("    Agents:   %s\n", strings.Join(parts, ", "))
		}

		// Skills
		if t.Skills.Deployed > 0 || t.Skills.Preserved > 0 {
			parts := []string{}
			if t.Skills.Deployed > 0 {
				parts = append(parts, fmt.Sprintf("%s linked", output.Green(fmt.Sprintf("%d", t.Skills.Deployed))))
			}
			if t.Skills.Preserved > 0 {
				parts = append(parts, fmt.Sprintf("%d user-managed", t.Skills.Preserved))
			}
			if t.Skills.Skipped > 0 {
				parts = append(parts, fmt.Sprintf("%d filtered", t.Skills.Skipped))
			}
			fmt.Printf("    Skills:   %s\n", strings.Join(parts, ", "))
		}

		// Commands
		if t.Commands.Deployed > 0 || t.Commands.Preserved > 0 {
			parts := []string{}
			if t.Commands.Deployed > 0 {
				parts = append(parts, fmt.Sprintf("%s linked", output.Green(fmt.Sprintf("%d", t.Commands.Deployed))))
			}
			if t.Commands.Preserved > 0 {
				parts = append(parts, fmt.Sprintf("%d user-managed", t.Commands.Preserved))
			}
			fmt.Printf("    Commands: %s\n", strings.Join(parts, ", "))
		}

		// Extras
		for _, e := range t.Extras {
			fmt.Printf("    %s\n", e)
		}

		fmt.Println()
	}
}
