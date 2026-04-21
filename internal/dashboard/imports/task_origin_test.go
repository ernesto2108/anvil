package imports

import (
	"testing"
)

func TestDetectTaskOrigin(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		taskID      string
		taskDesc    string
		firstPrompt string
		trackerHint string
		want        TaskOrigin
	}{
		{
			name:   "linear URL in taskID",
			taskID: "https://linear.app/myteam/issue/ENG-123/do-something",
			want: TaskOrigin{
				Source: "linear",
				URL:    "https://linear.app/myteam/issue/ENG-123/do-something",
			},
		},
		{
			name:     "jira URL in taskDesc",
			taskDesc: "See ticket at https://myorg.atlassian.net/browse/PROJ-456 for details",
			want: TaskOrigin{
				Source: "jira",
				URL:    "https://myorg.atlassian.net/browse/PROJ-456",
			},
		},
		{
			name:        "github issue URL in firstPrompt",
			firstPrompt: "Fix the bug described in https://github.com/myorg/myrepo/issues/78",
			want: TaskOrigin{
				Source: "github",
				URL:    "https://github.com/myorg/myrepo/issues/78",
				ID:     "#78",
			},
		},
		{
			name:        "#78 standalone in prompt",
			firstPrompt: "close #78 please",
			want: TaskOrigin{
				Source: "github",
				ID:     "#78",
			},
		},
		{
			name:        "org/repo#42 shorthand",
			firstPrompt: "relates to myorg/myrepo#42",
			want: TaskOrigin{
				Source: "github",
				URL:    "https://github.com/myorg/myrepo/issues/42",
				ID:     "#42",
			},
		},
		{
			name:        "obsidian wikilink",
			firstPrompt: "task from [[DASH-FEAT-019]]",
			want: TaskOrigin{
				Source: "obsidian",
				ID:     "DASH-FEAT-019",
			},
		},
		{
			name:        "obsidian task.md path",
			firstPrompt: "see 03-tasks/DASH-FEAT-019/task.md for context",
			want: TaskOrigin{
				Source: "obsidian",
				ID:     "DASH-FEAT-019",
			},
		},
		{
			name:        "ENG-123 with trackerHint linear",
			taskID:      "ENG-123",
			trackerHint: "linear",
			want: TaskOrigin{
				Source: "linear",
				ID:     "ENG-123",
			},
		},
		{
			name:        "PROJ-456 with trackerHint jira",
			taskID:      "PROJ-456",
			trackerHint: "jira",
			want: TaskOrigin{
				Source: "jira",
				ID:     "PROJ-456",
			},
		},
		{
			name:   "ENG-123 with no trackerHint",
			taskID: "ENG-123",
			want: TaskOrigin{
				Source: "unknown",
				ID:     "ENG-123",
			},
		},
		{
			name:        "no match anywhere returns zero value",
			taskID:      "my-task",
			taskDesc:    "something vague",
			firstPrompt: "do the thing",
			want:        TaskOrigin{},
		},
		{
			name:        "taskID checked before firstPrompt — priority order",
			taskID:      "https://linear.app/team/issue/ENG-001/task-in-id",
			firstPrompt: "https://github.com/org/repo/issues/9",
			want: TaskOrigin{
				Source: "linear",
				URL:    "https://linear.app/team/issue/ENG-001/task-in-id",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := DetectTaskOrigin(tc.taskID, tc.taskDesc, tc.firstPrompt, tc.trackerHint)
			if got != tc.want {
				t.Errorf("DetectTaskOrigin(%q, %q, %q, %q)\n  got  %+v\n  want %+v",
					tc.taskID, tc.taskDesc, tc.firstPrompt, tc.trackerHint, got, tc.want)
			}
		})
	}
}
