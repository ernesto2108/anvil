package imports

import (
	"regexp"
)

// TaskOrigin holds the detected source and URL of a task.
type TaskOrigin struct {
	Source string // "linear", "jira", "github", "obsidian", "unknown", ""
	URL    string // extracted or reconstructed URL, "" if none
	ID     string // extracted task ID (e.g. "ENG-123", "#78", "DASH-FEAT-019"), "" if none
}

var (
	reLinearURL   = regexp.MustCompile(`https://linear\.app/[^\s]+`)
	reJiraURL     = regexp.MustCompile(`https://[^/\s]+\.atlassian\.net/browse/[^\s]+`)
	reGitHubURL   = regexp.MustCompile(`https://github\.com/([^/\s]+)/([^/\s]+)/issues/(\d+)`)
	reObsidian    = regexp.MustCompile(`\[\[([^\]]+)\]\]`)
	reObsidianPath = regexp.MustCompile(`(?:^|/)([^/\s]+)/task\.md(?:\s|$)`)
	reGitHubShort = regexp.MustCompile(`(?:^|[^/\w])([A-Za-z0-9_-]+/[A-Za-z0-9_-]+)#(\d+)`)
	reHashIssue   = regexp.MustCompile(`(?:^|[^\w/])#(\d+)(?:\b|$)`)
	reTrackerID   = regexp.MustCompile(`\b([A-Z]{2,6}-\d+)\b`)
)

// DetectTaskOrigin analyzes taskID, taskDesc, and firstPrompt to determine
// the task origin. trackerHint is the value of anvil.yaml's task_tracker field
// ("linear", "jira", or "" if not configured).
// Returns zero-value TaskOrigin (all empty strings) if nothing is detected.
func DetectTaskOrigin(taskID, taskDesc, firstPrompt, trackerHint string) TaskOrigin {
	inputs := []string{taskID, taskDesc, firstPrompt}

	for _, input := range inputs {
		if o, ok := detectInText(input, trackerHint); ok {
			return o
		}
	}
	return TaskOrigin{}
}

// detectInText attempts all detection rules on a single input string.
// Returns the match and true if found, or zero value and false otherwise.
func detectInText(text, trackerHint string) (TaskOrigin, bool) {
	// 1. Explicit URLs — check before anything else.
	if m := reLinearURL.FindString(text); m != "" {
		return TaskOrigin{Source: "linear", URL: m}, true
	}
	if m := reJiraURL.FindString(text); m != "" {
		return TaskOrigin{Source: "jira", URL: m}, true
	}
	if m := reGitHubURL.FindStringSubmatch(text); m != nil {
		return TaskOrigin{
			Source: "github",
			URL:    m[0],
			ID:     "#" + m[3],
		}, true
	}

	// 2. Obsidian patterns.
	if m := reObsidian.FindStringSubmatch(text); m != nil {
		return TaskOrigin{Source: "obsidian", ID: m[1]}, true
	}
	if m := reObsidianPath.FindStringSubmatch(text); m != nil {
		return TaskOrigin{Source: "obsidian", ID: m[1]}, true
	}

	// 3. GitHub shorthand — org/repo#num before bare #num.
	if m := reGitHubShort.FindStringSubmatch(text); m != nil {
		repo := m[1]
		num := m[2]
		return TaskOrigin{
			Source: "github",
			URL:    "https://github.com/" + repo + "/issues/" + num,
			ID:     "#" + num,
		}, true
	}
	if m := reHashIssue.FindStringSubmatch(text); m != nil {
		return TaskOrigin{Source: "github", ID: "#" + m[1]}, true
	}

	// 4. Ambiguous tracker ID (ABC-123 pattern).
	if m := reTrackerID.FindString(text); m != "" {
		source := "unknown"
		switch trackerHint {
		case "linear":
			source = "linear"
		case "jira":
			source = "jira"
		}
		return TaskOrigin{Source: source, ID: m}, true
	}

	return TaskOrigin{}, false
}
