package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/ernesto2108/anvil/pkg/output"
)

// taskContext holds the enriched metadata gathered during preflight.
// It is passed to the runner so agents receive structured prompts.
type taskContext struct {
	Stack      string
	Objective  string
	Files      string // comma-separated or "por descubrir"
	Complexity string // "Small (N pts)", "Medium (N pts)", "Large (N pts)"
}

// preflight asks the user for any missing task metadata before running a pipeline.
// If autoApprove is true, it skips interactive questions and uses defaults.
func preflight(task string, autoApprove bool) taskContext {
	tc := taskContext{
		Objective: task,
	}

	if autoApprove {
		tc.Stack = "auto"
		tc.Files = "por descubrir"
		tc.Complexity = "Medium (5 pts)"
		return tc
	}

	reader := bufio.NewReader(os.Stdin)

	fmt.Println()
	output.Info("Preflight — necesito algunos datos para armar los prompts de los agentes.")
	fmt.Println()

	// Stack
	tc.Stack = ask(reader, "¿Stack?", "Go, React, Flutter, Python, Rust, Astro, otro")
	if tc.Stack == "" {
		tc.Stack = "auto"
	}

	// Files
	tc.Files = ask(reader, "¿Archivos afectados?", "rutas separadas por coma, o 'no sé'")
	if tc.Files == "" || strings.EqualFold(tc.Files, "no sé") || strings.EqualFold(tc.Files, "no se") {
		tc.Files = "por descubrir"
	}

	// Complexity
	tc.Complexity = ask(reader, "¿Complejidad?", "small, medium, large — default: medium")
	tc.Complexity = normalizeComplexity(tc.Complexity)

	// Show summary and confirm
	fmt.Println()
	output.Info("Prompt que recibirán los agentes:")
	fmt.Println()
	fmt.Printf("  Objetivo:    %s\n", tc.Objective)
	fmt.Printf("  Stack:       %s\n", tc.Stack)
	fmt.Printf("  Archivos:    %s\n", tc.Files)
	fmt.Printf("  Complejidad: %s\n", tc.Complexity)
	fmt.Println()

	confirm := ask(reader, "¿Dale?", "sí/no")
	confirm = strings.ToLower(strings.TrimSpace(confirm))
	if confirm != "" && confirm != "sí" && confirm != "si" && confirm != "s" && confirm != "dale" && confirm != "y" && confirm != "yes" {
		output.Error("Cancelado por el usuario.")
		os.Exit(0)
	}

	fmt.Println()
	return tc
}

func ask(reader *bufio.Reader, question, hint string) string {
	fmt.Printf("  %s (%s): ", question, hint)
	line, _ := reader.ReadString('\n')
	return strings.TrimSpace(line)
}

func normalizeComplexity(input string) string {
	switch strings.ToLower(strings.TrimSpace(input)) {
	case "small", "s", "chica", "1", "2", "3":
		return "Small (3 pts)"
	case "large", "l", "grande", "8", "13":
		return "Large (8 pts)"
	case "medium", "m", "media", "5", "":
		return "Medium (5 pts)"
	default:
		return "Medium (5 pts)"
	}
}
