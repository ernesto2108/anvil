package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// sampleItems returns a consistent set of test fixtures covering all item types.
func sampleItems() []Item {
	return []Item{
		{Name: "developer", Type: "agent", Description: "Production code", Targets: map[string]bool{"claude": true, "opencode": true}},
		{Name: "tester", Type: "agent", Description: "Test files", Targets: map[string]bool{"claude": false}},
		{Name: "go-conventions", Type: "skill", Description: "Go patterns and conventions", Targets: map[string]bool{}},
		{Name: "react-conventions", Type: "skill", Description: "React hooks and state", Targets: map[string]bool{"claude": true}},
		{Name: "git-commit", Type: "command", Description: "Git commit helper", Targets: map[string]bool{}},
	}
}

func defaultTargets() []string {
	return []string{"claude", "opencode"}
}

func pressKey(m Model, r rune) Model {
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}
	updated, _ := m.Update(msg)
	return updated.(Model)
}

func pressSpecial(m Model, kt tea.KeyType) Model {
	msg := tea.KeyMsg{Type: kt}
	updated, _ := m.Update(msg)
	return updated.(Model)
}

// ---------------------------------------------------------------------------
// 1. Model creation and initialization
// ---------------------------------------------------------------------------

func TestNewModel_Defaults(t *testing.T) {
	items := sampleItems()
	info := RepoInfo{Branch: "main", SHA: "abc1234"}
	m := NewModel(items, defaultTargets(), info)

	if m.mode != ModeSymlink {
		t.Errorf("default mode = %v, want ModeSymlink", m.mode)
	}
	if m.filter != filterAll {
		t.Errorf("default filter = %v, want filterAll", m.filter)
	}
	if m.cursor != 0 {
		t.Errorf("default cursor = %d, want 0", m.cursor)
	}
	if m.quitting {
		t.Error("default quitting should be false")
	}
	if m.searching {
		t.Error("default searching should be false")
	}
}

func TestNewModel_ItemsStoredAndFiltered(t *testing.T) {
	items := sampleItems()
	m := NewModel(items, defaultTargets(), RepoInfo{})

	if len(m.items) != len(items) {
		t.Errorf("stored items = %d, want %d", len(m.items), len(items))
	}
	// filterAll means all items appear in filtered
	if len(m.filtered) != len(items) {
		t.Errorf("filtered count = %d, want %d", len(m.filtered), len(items))
	}
}

func TestNewModel_EmptyItems(t *testing.T) {
	m := NewModel(nil, defaultTargets(), RepoInfo{})
	if len(m.filtered) != 0 {
		t.Errorf("expected 0 filtered items, got %d", len(m.filtered))
	}
	if m.cursor != 0 {
		t.Errorf("cursor should be 0 on empty list, got %d", m.cursor)
	}
}

// ---------------------------------------------------------------------------
// 2. Item.Installed()
// ---------------------------------------------------------------------------

func TestItem_Installed(t *testing.T) {
	tests := []struct {
		name    string
		targets map[string]bool
		want    bool
	}{
		{
			name:    "no targets map",
			targets: nil,
			want:    false,
		},
		{
			name:    "empty targets map",
			targets: map[string]bool{},
			want:    false,
		},
		{
			name:    "all targets false",
			targets: map[string]bool{"claude": false, "opencode": false},
			want:    false,
		},
		{
			name:    "at least one true",
			targets: map[string]bool{"claude": true, "opencode": false},
			want:    true,
		},
		{
			name:    "all targets true",
			targets: map[string]bool{"claude": true, "opencode": true},
			want:    true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			it := Item{Name: "x", Type: "agent", Targets: tc.targets}
			if got := it.Installed(); got != tc.want {
				t.Errorf("Installed() = %v, want %v", got, tc.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// 3. DeployMode
// ---------------------------------------------------------------------------

func TestDeployMode_String(t *testing.T) {
	tests := []struct {
		mode DeployMode
		want string
	}{
		{ModeSymlink, "symlink"},
		{ModeCopy, "copy"},
	}
	for _, tc := range tests {
		t.Run(tc.want, func(t *testing.T) {
			if got := tc.mode.String(); got != tc.want {
				t.Errorf("String() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestDeployMode_Label(t *testing.T) {
	tests := []struct {
		mode DeployMode
		want string
	}{
		{ModeSymlink, "symlink (linked)"},
		{ModeCopy, "copy (permanent)"},
	}
	for _, tc := range tests {
		t.Run(tc.want, func(t *testing.T) {
			if got := tc.mode.Label(); got != tc.want {
				t.Errorf("Label() = %q, want %q", got, tc.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// 4. filter
// ---------------------------------------------------------------------------

func TestFilter_String(t *testing.T) {
	tests := []struct {
		f    filter
		want string
	}{
		{filterAll, "All"},
		{filterAgents, "Agents"},
		{filterSkills, "Skills"},
		{filterCommands, "Commands"},
	}
	for _, tc := range tests {
		t.Run(tc.want, func(t *testing.T) {
			if got := tc.f.String(); got != tc.want {
				t.Errorf("String() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestFilter_Next(t *testing.T) {
	// Should cycle: All -> Agents -> Skills -> Commands -> All
	sequence := []filter{filterAll, filterAgents, filterSkills, filterCommands, filterAll}
	for i := 0; i < len(sequence)-1; i++ {
		got := sequence[i].next()
		want := sequence[i+1]
		if got != want {
			t.Errorf("filter(%d).next() = %d, want %d", sequence[i], got, want)
		}
	}
}

// ---------------------------------------------------------------------------
// 5. applyFilter (tested indirectly through the model)
// ---------------------------------------------------------------------------

func TestApplyFilter_All(t *testing.T) {
	m := NewModel(sampleItems(), defaultTargets(), RepoInfo{})
	// filterAll is the default
	if len(m.filtered) != 5 {
		t.Errorf("filterAll: got %d items, want 5", len(m.filtered))
	}
}

func TestApplyFilter_Agents(t *testing.T) {
	m := NewModel(sampleItems(), defaultTargets(), RepoInfo{})
	m.filter = filterAgents
	m.applyFilter()
	if len(m.filtered) != 2 {
		t.Errorf("filterAgents: got %d items, want 2", len(m.filtered))
	}
	for _, it := range m.filtered {
		if it.Type != "agent" {
			t.Errorf("expected agent type, got %q", it.Type)
		}
	}
}

func TestApplyFilter_Skills(t *testing.T) {
	m := NewModel(sampleItems(), defaultTargets(), RepoInfo{})
	m.filter = filterSkills
	m.applyFilter()
	if len(m.filtered) != 2 {
		t.Errorf("filterSkills: got %d items, want 2", len(m.filtered))
	}
	for _, it := range m.filtered {
		if it.Type != "skill" {
			t.Errorf("expected skill type, got %q", it.Type)
		}
	}
}

func TestApplyFilter_Commands(t *testing.T) {
	m := NewModel(sampleItems(), defaultTargets(), RepoInfo{})
	m.filter = filterCommands
	m.applyFilter()
	if len(m.filtered) != 1 {
		t.Errorf("filterCommands: got %d items, want 1", len(m.filtered))
	}
	if m.filtered[0].Type != "command" {
		t.Errorf("expected command type, got %q", m.filtered[0].Type)
	}
}

func TestApplyFilter_SearchByName(t *testing.T) {
	m := NewModel(sampleItems(), defaultTargets(), RepoInfo{})
	m.searchText = "developer"
	m.applyFilter()
	if len(m.filtered) != 1 {
		t.Errorf("search 'developer': got %d items, want 1", len(m.filtered))
	}
	if m.filtered[0].Name != "developer" {
		t.Errorf("expected 'developer', got %q", m.filtered[0].Name)
	}
}

func TestApplyFilter_SearchCaseInsensitive(t *testing.T) {
	m := NewModel(sampleItems(), defaultTargets(), RepoInfo{})
	m.searchText = "DEVELOPER"
	m.applyFilter()
	if len(m.filtered) != 1 {
		t.Errorf("case-insensitive search: got %d items, want 1", len(m.filtered))
	}
}

func TestApplyFilter_SearchByDescription(t *testing.T) {
	m := NewModel(sampleItems(), defaultTargets(), RepoInfo{})
	m.searchText = "hooks"
	m.applyFilter()
	if len(m.filtered) != 1 {
		t.Errorf("search by description 'hooks': got %d items, want 1", len(m.filtered))
	}
	if m.filtered[0].Name != "react-conventions" {
		t.Errorf("expected 'react-conventions', got %q", m.filtered[0].Name)
	}
}

func TestApplyFilter_CombinedFilterAndSearch(t *testing.T) {
	m := NewModel(sampleItems(), defaultTargets(), RepoInfo{})
	m.filter = filterSkills
	m.searchText = "go"
	m.applyFilter()
	// Only "go-conventions" matches skill type + "go" in name
	if len(m.filtered) != 1 {
		t.Errorf("combined filter+search: got %d items, want 1", len(m.filtered))
	}
	if m.filtered[0].Name != "go-conventions" {
		t.Errorf("expected 'go-conventions', got %q", m.filtered[0].Name)
	}
}

func TestApplyFilter_CursorClampedWhenFilterReduces(t *testing.T) {
	m := NewModel(sampleItems(), defaultTargets(), RepoInfo{})
	m.cursor = 4 // last item in full list
	m.filter = filterAgents
	m.applyFilter()
	// filterAgents yields 2 items, cursor must be clamped to 1
	if m.cursor >= len(m.filtered) {
		t.Errorf("cursor %d is out of bounds for filtered len %d", m.cursor, len(m.filtered))
	}
}

// ---------------------------------------------------------------------------
// 6. Update — keyboard navigation
// ---------------------------------------------------------------------------

func TestUpdate_Navigation_Down(t *testing.T) {
	m := NewModel(sampleItems(), defaultTargets(), RepoInfo{})
	m.cursor = 0
	m = pressKey(m, 'j')
	if m.cursor != 1 {
		t.Errorf("after 'j', cursor = %d, want 1", m.cursor)
	}
}

func TestUpdate_Navigation_DownArrow(t *testing.T) {
	m := NewModel(sampleItems(), defaultTargets(), RepoInfo{})
	m.cursor = 0
	m = pressSpecial(m, tea.KeyDown)
	if m.cursor != 1 {
		t.Errorf("after down arrow, cursor = %d, want 1", m.cursor)
	}
}

func TestUpdate_Navigation_Up(t *testing.T) {
	m := NewModel(sampleItems(), defaultTargets(), RepoInfo{})
	m.cursor = 2
	m = pressKey(m, 'k')
	if m.cursor != 1 {
		t.Errorf("after 'k', cursor = %d, want 1", m.cursor)
	}
}

func TestUpdate_Navigation_UpArrow(t *testing.T) {
	m := NewModel(sampleItems(), defaultTargets(), RepoInfo{})
	m.cursor = 2
	m = pressSpecial(m, tea.KeyUp)
	if m.cursor != 1 {
		t.Errorf("after up arrow, cursor = %d, want 1", m.cursor)
	}
}

func TestUpdate_Navigation_CursorNoGoBelow0(t *testing.T) {
	m := NewModel(sampleItems(), defaultTargets(), RepoInfo{})
	m.cursor = 0
	m = pressKey(m, 'k')
	if m.cursor != 0 {
		t.Errorf("cursor should stay at 0 when pressing up at top, got %d", m.cursor)
	}
}

func TestUpdate_Navigation_CursorNoExceedMax(t *testing.T) {
	m := NewModel(sampleItems(), defaultTargets(), RepoInfo{})
	m.cursor = len(m.filtered) - 1
	m = pressKey(m, 'j')
	if m.cursor != len(m.filtered)-1 {
		t.Errorf("cursor should stay at max when pressing down at bottom, got %d", m.cursor)
	}
}

func TestUpdate_Navigation_GJumpsToLast(t *testing.T) {
	m := NewModel(sampleItems(), defaultTargets(), RepoInfo{})
	m.cursor = 0
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'G'}}
	updated, _ := m.Update(msg)
	m = updated.(Model)
	want := len(m.filtered) - 1
	if m.cursor != want {
		t.Errorf("'G' should jump to last item (%d), got %d", want, m.cursor)
	}
}

func TestUpdate_Navigation_QSetsQuitting(t *testing.T) {
	m := NewModel(sampleItems(), defaultTargets(), RepoInfo{})
	m = pressKey(m, 'q')
	if !m.quitting {
		t.Error("'q' should set quitting = true")
	}
}

// ---------------------------------------------------------------------------
// 7. Update — filter tab
// ---------------------------------------------------------------------------

func TestUpdate_Tab_CyclesFilter(t *testing.T) {
	m := NewModel(sampleItems(), defaultTargets(), RepoInfo{})
	if m.filter != filterAll {
		t.Fatal("expected filterAll initially")
	}
	m = pressSpecial(m, tea.KeyTab)
	if m.filter != filterAgents {
		t.Errorf("after tab, filter = %v, want filterAgents", m.filter)
	}
	m = pressSpecial(m, tea.KeyTab)
	if m.filter != filterSkills {
		t.Errorf("after second tab, filter = %v, want filterSkills", m.filter)
	}
	m = pressSpecial(m, tea.KeyTab)
	if m.filter != filterCommands {
		t.Errorf("after third tab, filter = %v, want filterCommands", m.filter)
	}
	m = pressSpecial(m, tea.KeyTab)
	if m.filter != filterAll {
		t.Errorf("after fourth tab, filter = %v, want filterAll", m.filter)
	}
}

func TestUpdate_Tab_ReFiltersItems(t *testing.T) {
	m := NewModel(sampleItems(), defaultTargets(), RepoInfo{})
	m = pressSpecial(m, tea.KeyTab) // -> filterAgents
	if len(m.filtered) != 2 {
		t.Errorf("after tab to Agents, filtered = %d, want 2", len(m.filtered))
	}
}

// ---------------------------------------------------------------------------
// 8. Update — search mode
// ---------------------------------------------------------------------------

func TestUpdate_Search_SlashEntersSearchMode(t *testing.T) {
	m := NewModel(sampleItems(), defaultTargets(), RepoInfo{})
	m = pressKey(m, '/')
	if !m.searching {
		t.Error("'/' should set searching = true")
	}
}

func TestUpdate_Search_TypingUpdatesSearchText(t *testing.T) {
	m := NewModel(sampleItems(), defaultTargets(), RepoInfo{})
	m = pressKey(m, '/')

	// Type characters via updateSearch path
	for _, r := range "dev" {
		msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}
		updated, _ := m.Update(msg)
		m = updated.(Model)
	}

	if !strings.Contains(m.search.Value(), "dev") {
		t.Errorf("search value = %q, expected to contain 'dev'", m.search.Value())
	}
}

func TestUpdate_Search_EnterExitsSearchMode(t *testing.T) {
	m := NewModel(sampleItems(), defaultTargets(), RepoInfo{})
	m = pressKey(m, '/')
	m = pressSpecial(m, tea.KeyEnter)
	if m.searching {
		t.Error("enter should exit search mode")
	}
}

func TestUpdate_Search_EscExitsSearchMode(t *testing.T) {
	m := NewModel(sampleItems(), defaultTargets(), RepoInfo{})
	m = pressKey(m, '/')
	m = pressSpecial(m, tea.KeyEsc)
	if m.searching {
		t.Error("esc should exit search mode")
	}
}

func TestUpdate_Search_BackspaceInNormalModeClears(t *testing.T) {
	m := NewModel(sampleItems(), defaultTargets(), RepoInfo{})
	m.searchText = "dev"
	m.search.SetValue("dev")
	m.applyFilter()

	m = pressSpecial(m, tea.KeyBackspace)
	if m.searchText != "" {
		t.Errorf("backspace in normal mode should clear searchText, got %q", m.searchText)
	}
	// All items should be visible again
	if len(m.filtered) != 5 {
		t.Errorf("after backspace clear, filtered = %d, want 5", len(m.filtered))
	}
}

// ---------------------------------------------------------------------------
// 9. Update — install/uninstall toggle
// ---------------------------------------------------------------------------

func TestUpdate_Toggle_InstallUninstalled(t *testing.T) {
	items := sampleItems()
	targets := defaultTargets()
	m := NewModel(items, targets, RepoInfo{})

	// Move cursor to "go-conventions" (index 2), which has no targets deployed
	m.cursor = 2
	if m.filtered[m.cursor].Name != "go-conventions" {
		t.Fatalf("expected 'go-conventions' at index 2, got %q", m.filtered[m.cursor].Name)
	}

	m = pressSpecial(m, tea.KeyEnter)

	q := m.Queue()
	if len(q) == 0 {
		t.Fatal("expected actions in queue after install")
	}
	for _, a := range q {
		if a.Action != "install" {
			t.Errorf("expected install action, got %q", a.Action)
		}
		if a.Item.Name != "go-conventions" {
			t.Errorf("expected item 'go-conventions', got %q", a.Item.Name)
		}
	}
	// Both enabled targets should have install queued
	if len(q) != len(targets) {
		t.Errorf("expected %d install actions (one per target), got %d", len(targets), len(q))
	}
}

func TestUpdate_Toggle_UninstallInstalled(t *testing.T) {
	items := sampleItems()
	m := NewModel(items, defaultTargets(), RepoInfo{})

	// "developer" at index 0 is installed on claude and opencode
	m.cursor = 0
	if m.filtered[m.cursor].Name != "developer" {
		t.Fatalf("expected 'developer' at index 0, got %q", m.filtered[m.cursor].Name)
	}

	m = pressSpecial(m, tea.KeyEnter)

	q := m.Queue()
	if len(q) == 0 {
		t.Fatal("expected actions in queue after uninstall")
	}
	for _, a := range q {
		if a.Action != "uninstall" {
			t.Errorf("expected uninstall action, got %q", a.Action)
		}
	}
}

func TestUpdate_Toggle_SpaceKeyWorks(t *testing.T) {
	items := sampleItems()
	m := NewModel(items, defaultTargets(), RepoInfo{})
	m.cursor = 2 // go-conventions, not installed

	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}}
	updated, _ := m.Update(msg)
	m = updated.(Model)

	if len(m.Queue()) == 0 {
		t.Error("space key should enqueue install actions")
	}
}

func TestUpdate_PerTargetToggle_Claude(t *testing.T) {
	m := NewModel(sampleItems(), defaultTargets(), RepoInfo{})
	m.cursor = 2 // go-conventions, no targets set

	m = pressKey(m, 'c')

	q := m.Queue()
	if len(q) != 1 {
		t.Fatalf("expected 1 action in queue, got %d", len(q))
	}
	if q[0].Target != TargetClaude {
		t.Errorf("expected target %q, got %q", TargetClaude, q[0].Target)
	}
	if q[0].Action != "install" {
		t.Errorf("expected install, got %q", q[0].Action)
	}
}

func TestUpdate_PerTargetToggle_OpenCode(t *testing.T) {
	m := NewModel(sampleItems(), defaultTargets(), RepoInfo{})
	m.cursor = 2 // go-conventions

	m = pressKey(m, 'o')

	q := m.Queue()
	if len(q) != 1 {
		t.Fatalf("expected 1 action in queue, got %d", len(q))
	}
	if q[0].Target != TargetOpenCode {
		t.Errorf("expected target %q, got %q", TargetOpenCode, q[0].Target)
	}
}

func TestUpdate_PerTargetToggle_Gemini(t *testing.T) {
	targets := []string{"claude", "opencode", "gemini"}
	m := NewModel(sampleItems(), targets, RepoInfo{})
	m.cursor = 2 // go-conventions

	m = pressKey(m, 'g')

	q := m.Queue()
	if len(q) != 1 {
		t.Fatalf("expected 1 action in queue, got %d", len(q))
	}
	if q[0].Target != TargetGemini {
		t.Errorf("expected target %q, got %q", TargetGemini, q[0].Target)
	}
}

func TestUpdate_PerTargetToggle_Codex(t *testing.T) {
	targets := []string{"claude", "opencode", "codex"}
	m := NewModel(sampleItems(), targets, RepoInfo{})
	m.cursor = 2 // go-conventions

	m = pressKey(m, 'x')

	q := m.Queue()
	if len(q) != 1 {
		t.Fatalf("expected 1 action in queue, got %d", len(q))
	}
	if q[0].Target != TargetCodex {
		t.Errorf("expected target %q, got %q", TargetCodex, q[0].Target)
	}
}

func TestUpdate_PerTargetToggle_DisabledTargetIsNoop(t *testing.T) {
	// Only claude is enabled; pressing 'o' (opencode) should do nothing
	m := NewModel(sampleItems(), []string{"claude"}, RepoInfo{})
	m.cursor = 2 // go-conventions

	m = pressKey(m, 'o')

	q := m.Queue()
	if len(q) != 0 {
		t.Errorf("toggle on disabled target should be no-op, got %d actions", len(q))
	}
}

// ---------------------------------------------------------------------------
// 10. Deploy mode toggle
// ---------------------------------------------------------------------------

func TestUpdate_DeployModeToggle(t *testing.T) {
	m := NewModel(sampleItems(), defaultTargets(), RepoInfo{})
	if m.mode != ModeSymlink {
		t.Fatalf("initial mode should be ModeSymlink, got %v", m.mode)
	}

	m = pressKey(m, 's')
	if m.mode != ModeCopy {
		t.Errorf("after 's', mode = %v, want ModeCopy", m.mode)
	}

	m = pressKey(m, 's')
	if m.mode != ModeSymlink {
		t.Errorf("after second 's', mode = %v, want ModeSymlink", m.mode)
	}
}

// ---------------------------------------------------------------------------
// 11. Queue management
// ---------------------------------------------------------------------------

func TestEnqueue_AddsActions(t *testing.T) {
	m := NewModel(sampleItems(), defaultTargets(), RepoInfo{})
	it := m.filtered[2] // go-conventions

	m.enqueue("install", it, "claude")
	m.enqueue("install", it, "opencode")

	q := m.Queue()
	if len(q) != 2 {
		t.Errorf("expected 2 queued actions, got %d", len(q))
	}
}

func TestEnqueue_ReplacesDuplicateItemTargetCombo(t *testing.T) {
	m := NewModel(sampleItems(), defaultTargets(), RepoInfo{})
	it := m.filtered[2] // go-conventions

	m.enqueue("install", it, "claude")
	m.enqueue("uninstall", it, "claude") // same item+target — should replace

	q := m.Queue()
	if len(q) != 1 {
		t.Errorf("expected 1 action (duplicate replaced), got %d", len(q))
	}
	if q[0].Action != "uninstall" {
		t.Errorf("expected latest action 'uninstall', got %q", q[0].Action)
	}
}

func TestEnqueue_DifferentTargetsAreNotDuplicates(t *testing.T) {
	m := NewModel(sampleItems(), defaultTargets(), RepoInfo{})
	it := m.filtered[2]

	m.enqueue("install", it, "claude")
	m.enqueue("install", it, "opencode")

	q := m.Queue()
	if len(q) != 2 {
		t.Errorf("different targets should not be treated as duplicates, got %d actions", len(q))
	}
}

func TestQueue_ReturnsAccumulatedActions(t *testing.T) {
	m := NewModel(sampleItems(), defaultTargets(), RepoInfo{})
	it := m.filtered[0]

	m.enqueue("install", it, "claude")
	q := m.Queue()
	if len(q) != 1 || q[0].Item.Name != it.Name {
		t.Error("Queue() did not return expected accumulated actions")
	}
}

// ---------------------------------------------------------------------------
// 12. View rendering
// ---------------------------------------------------------------------------

func TestView_EmptyWhenQuitting(t *testing.T) {
	m := NewModel(sampleItems(), defaultTargets(), RepoInfo{})
	m.quitting = true
	if v := m.View(); v != "" {
		t.Errorf("View() when quitting = %q, want empty string", v)
	}
}

func TestRenderHeader_ContainsTitle(t *testing.T) {
	m := NewModel(sampleItems(), defaultTargets(), RepoInfo{})
	h := m.renderHeader()
	if !strings.Contains(h, "anvil registry browse") {
		t.Errorf("renderHeader() does not contain title, got: %q", h)
	}
}

func TestRenderHeader_WithRepoInfo(t *testing.T) {
	info := RepoInfo{Branch: "main", SHA: "abc1234567", Provider: "github", Tag: "v1.0.0"}
	m := NewModel(sampleItems(), defaultTargets(), info)
	h := m.renderHeader()
	if !strings.Contains(h, "main") {
		t.Errorf("renderHeader() missing branch, got: %q", h)
	}
}

func TestRenderStatus_ShowsCounts(t *testing.T) {
	m := NewModel(sampleItems(), defaultTargets(), RepoInfo{})
	s := m.renderStatus()
	// Should contain total entries count (5) and deployed count
	if !strings.Contains(s, "5") {
		t.Errorf("renderStatus() missing item count, got: %q", s)
	}
	// "deployed" should appear
	if !strings.Contains(s, "deployed") {
		t.Errorf("renderStatus() missing 'deployed', got: %q", s)
	}
}

func TestRenderStatus_ShowsMode(t *testing.T) {
	m := NewModel(sampleItems(), defaultTargets(), RepoInfo{})
	s := m.renderStatus()
	if !strings.Contains(s, "symlink") {
		t.Errorf("renderStatus() missing mode label, got: %q", s)
	}
}

func TestRenderStatus_ShowsQueueInfo(t *testing.T) {
	m := NewModel(sampleItems(), defaultTargets(), RepoInfo{})
	it := m.filtered[2]
	m.enqueue("install", it, "claude")

	s := m.renderStatus()
	if !strings.Contains(s, "Queue") {
		t.Errorf("renderStatus() missing queue info, got: %q", s)
	}
}

func TestTargetBadge_NoneDeployed(t *testing.T) {
	it := Item{Name: "x", Type: "skill", Targets: map[string]bool{}}
	badge := targetBadge(it, []string{"claude", "opencode"})
	if !strings.Contains(badge, "○") {
		t.Errorf("badge for none deployed should contain '○', got %q", badge)
	}
}

func TestTargetBadge_AllDeployed(t *testing.T) {
	it := Item{
		Name:    "x",
		Type:    "agent",
		Targets: map[string]bool{"claude": true, "opencode": true},
	}
	badge := targetBadge(it, []string{"claude", "opencode"})
	if !strings.Contains(badge, "●") {
		t.Errorf("badge for all deployed should contain '●', got %q", badge)
	}
}

func TestTargetBadge_PartialDeployed(t *testing.T) {
	it := Item{
		Name:    "x",
		Type:    "agent",
		Targets: map[string]bool{"claude": true, "opencode": false},
	}
	badge := targetBadge(it, []string{"claude", "opencode"})
	if !strings.Contains(badge, "◐") {
		t.Errorf("badge for partial deployed should contain '◐', got %q", badge)
	}
}

// ---------------------------------------------------------------------------
// 13. wordWrap
// ---------------------------------------------------------------------------

func TestWordWrap(t *testing.T) {
	tests := []struct {
		name  string
		input string
		width int
		check func(got string) bool
		desc  string
	}{
		{
			name:  "empty string",
			input: "",
			width: 40,
			check: func(got string) bool { return got == "" },
			desc:  "should return empty string",
		},
		{
			name:  "width zero",
			input: "hello world",
			width: 0,
			check: func(got string) bool { return got == "hello world" },
			desc:  "should return unchanged when width <= 0",
		},
		{
			name:  "width negative",
			input: "hello world",
			width: -5,
			check: func(got string) bool { return got == "hello world" },
			desc:  "should return unchanged when width negative",
		},
		{
			name:  "short text fits",
			input: "hello world",
			width: 40,
			check: func(got string) bool { return got == "hello world" },
			desc:  "text shorter than width should not wrap",
		},
		{
			name:  "wraps at width",
			input: "one two three four five six seven eight nine ten",
			width: 15,
			check: func(got string) bool {
				lines := strings.Split(got, "\n")
				for _, l := range lines {
					if len(l) > 15 {
						return false
					}
				}
				return len(lines) > 1
			},
			desc: "each line should be within width and there should be multiple lines",
		},
		{
			name:  "single long word",
			input: "superlongword",
			width: 5,
			check: func(got string) bool { return got == "superlongword" },
			desc:  "single word longer than width stays on one line",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := wordWrap(tc.input, tc.width)
			if !tc.check(got) {
				t.Errorf("wordWrap(%q, %d) = %q: %s", tc.input, tc.width, got, tc.desc)
			}
		})
	}
}
