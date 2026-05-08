package transcript

import (
	"strings"
)

// decisionKeywords is a bilingual ES+EN list of phrases that indicate a
// decision was made. Match is case-insensitive and performed on full lines
// of assistant text.
var decisionKeywords = []string{
	// Spanish
	"decidí",
	"voy a usar",
	"elegí",
	"porque",
	"en vez de",
	// English
	"i decided",
	"i'll use",
	"i chose",
	"because",
	"instead of",
}

// extractDecisions filters text lines that contain at least one decision
// keyword, deduplicates them, and returns at most 10.
// Lines must be non-empty and at least 10 characters to avoid noise.
func extractDecisions(lines []string) []string {
	seen := make(map[string]struct{})
	var out []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if len(trimmed) < 10 {
			continue
		}
		lower := strings.ToLower(trimmed)
		for _, kw := range decisionKeywords {
			if strings.Contains(lower, kw) {
				if _, exists := seen[trimmed]; !exists {
					seen[trimmed] = struct{}{}
					out = append(out, trimmed)
				}
				break
			}
		}
		if len(out) >= 10 {
			break
		}
	}
	return out
}
