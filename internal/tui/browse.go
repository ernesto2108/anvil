package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Target keys used for per-target toggling.
const (
	TargetClaude   = "claude"
	TargetOpenCode = "opencode"
	TargetGemini   = "gemini"
	TargetCodex    = "codex"
	TargetCursor   = "cursor"
)

// RepoInfo holds git/provider context displayed in the TUI header.
type RepoInfo struct {
	Branch   string
	SHA      string
	Tag      string
	Provider string
}

// DeployMode controls how items are deployed to targets.
type DeployMode int

const (
	ModeCopy    DeployMode = iota // permanent copy
	ModeSymlink                   // symlink (temporary, always in sync)
)

func (d DeployMode) String() string {
	if d == ModeSymlink {
		return "symlink"
	}
	return "copy"
}

func (d DeployMode) Label() string {
	if d == ModeSymlink {
		return "symlink (linked)"
	}
	return "copy (permanent)"
}

// Item represents an agent, skill, or command entry in the browse list.
type Item struct {
	Name        string
	Type        string // "agent", "skill", or "command"
	Description string
	Version     string
	Author      string
	URL         string
	Targets     map[string]bool // target name -> deployed
}

// Installed returns true if deployed to at least one target.
func (it Item) Installed() bool {
	for _, v := range it.Targets {
		if v {
			return true
		}
	}
	return false
}

// ActionResult is the outcome after quitting the TUI.
type ActionResult struct {
	Action string // "install" or "uninstall"
	Target string // specific target, or "" for all enabled
	Mode   DeployMode
	Item   Item
}

type filter int

const (
	filterAll filter = iota
	filterAgents
	filterSkills
	filterCommands
)

func (f filter) String() string {
	switch f {
	case filterAgents:
		return "Agents"
	case filterSkills:
		return "Skills"
	case filterCommands:
		return "Commands"
	default:
		return "All"
	}
}

func (f filter) next() filter { return (f + 1) % 4 }

// Model is the bubbletea model for the registry browser.
type Model struct {
	items      []Item
	filtered   []Item
	cursor     int
	filter     filter
	searching  bool
	search     textinput.Model
	searchText string
	width      int
	height     int
	queue      []ActionResult
	quitting   bool
	mode       DeployMode
	targets    []string // ordered list of enabled targets
	info       RepoInfo
}

// NewModel creates a browse model from the given items, enabled targets, and repo info.
func NewModel(items []Item, enabledTargets []string, info RepoInfo) Model {
	ti := textinput.New()
	ti.Placeholder = "search..."
	ti.CharLimit = 60

	m := Model{
		items:   items,
		search:  ti,
		targets: enabledTargets,
		mode:    ModeSymlink, // default to symlink
		info:    info,
	}
	m.applyFilter()
	return m
}

// Queue returns the accumulated install/uninstall actions.
func (m Model) Queue() []ActionResult {
	return m.queue
}

func (m *Model) applyFilter() {
	m.filtered = nil
	query := strings.ToLower(m.searchText)

	for _, it := range m.items {
		switch m.filter {
		case filterAgents:
			if it.Type != "agent" {
				continue
			}
		case filterSkills:
			if it.Type != "skill" {
				continue
			}
		case filterCommands:
			if it.Type != "command" {
				continue
			}
		}

		if query != "" {
			name := strings.ToLower(it.Name)
			desc := strings.ToLower(it.Description)
			if !strings.Contains(name, query) && !strings.Contains(desc, query) {
				continue
			}
		}

		m.filtered = append(m.filtered, it)
	}

	if m.cursor >= len(m.filtered) {
		m.cursor = max(0, len(m.filtered)-1)
	}
}

// -- styles --

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205")).
			Padding(0, 1)

	tabActiveStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205")).
			Underline(true)

	tabInactiveStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("245"))

	selectedStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("62"))

	normalStyle = lipgloss.NewStyle()

	typeAgentStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("81")).
			Width(8)

	typeSkillStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")).
			Width(8)

	typeCmdStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("183")).
			Width(8)

	detailTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("15"))

	detailLabelStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("245"))

	detailValueStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("252"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	statusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")).
			Padding(0, 1)

	searchStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205"))

	targetOnStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true)
	targetOffStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	modeStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("141")).Bold(true)
	queueInstall   = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true)
	queueUninstall = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
)

func typeStyle(typ string) lipgloss.Style {
	switch typ {
	case "skill":
		return typeSkillStyle
	case "command":
		return typeCmdStyle
	default:
		return typeAgentStyle
	}
}

// -- tea.Model interface --

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		if m.searching {
			return m.updateSearch(msg)
		}
		return m.updateNormal(msg)
	}
	return m, nil
}

func (m Model) updateSearch(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, key.NewBinding(key.WithKeys("enter", "esc"))):
		m.searching = false
		m.searchText = m.search.Value()
		m.search.Blur()
		m.applyFilter()
		return m, nil
	default:
		var cmd tea.Cmd
		m.search, cmd = m.search.Update(msg)
		m.searchText = m.search.Value()
		m.applyFilter()
		return m, cmd
	}
}

func (m Model) updateNormal(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c", "esc":
		m.quitting = true
		return m, tea.Quit

	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.filtered)-1 {
			m.cursor++
		}

	case "tab":
		m.filter = m.filter.next()
		m.applyFilter()

	case "/":
		m.searching = true
		m.search.Focus()
		return m, m.search.Cursor.BlinkCmd()

	case "backspace":
		if m.searchText != "" {
			m.searchText = ""
			m.search.SetValue("")
			m.applyFilter()
		}

	case "enter", " ":
		// Toggle all enabled targets
		if len(m.filtered) > 0 {
			it := m.filtered[m.cursor]
			if it.Installed() {
				// Uninstall from all targets where it's deployed
				for t, deployed := range it.Targets {
					if deployed {
						m.enqueue("uninstall", it, t)
					}
				}
				m.setAllTargets(it.Name, false)
			} else {
				// Install to all enabled targets
				for _, t := range m.targets {
					m.enqueue("install", it, t)
				}
				m.setAllTargets(it.Name, true)
			}
		}

	// Per-target toggles
	case "c":
		m.toggleTarget(TargetClaude)
	case "o":
		m.toggleTarget(TargetOpenCode)
	case "g":
		if m.cursor == 0 {
			// If at top, don't conflict with vim g
			m.toggleTarget(TargetGemini)
		} else {
			m.toggleTarget(TargetGemini)
		}
	case "x":
		m.toggleTarget(TargetCodex)
	case "r":
		m.toggleTarget(TargetCursor)

	// Batch deploy: install ALL filtered items to a single target
	case "C":
		m.batchToggleTarget(TargetClaude)
	case "O":
		m.batchToggleTarget(TargetOpenCode)
	case "X":
		m.batchToggleTarget(TargetCodex)
	case "R":
		m.batchToggleTarget(TargetCursor)

	// Deploy mode toggle
	case "s":
		if m.mode == ModeCopy {
			m.mode = ModeSymlink
		} else {
			m.mode = ModeCopy
		}

	case "G":
		m.cursor = max(0, len(m.filtered)-1)
	}

	return m, nil
}

func (m *Model) toggleTarget(target string) {
	if len(m.filtered) == 0 {
		return
	}
	// Check target is enabled
	enabled := false
	for _, t := range m.targets {
		if t == target {
			enabled = true
			break
		}
	}
	if !enabled {
		return
	}

	it := m.filtered[m.cursor]
	currentlyOn := it.Targets[target]

	if currentlyOn {
		m.enqueue("uninstall", it, target)
		m.setTarget(it.Name, target, false)
	} else {
		m.enqueue("install", it, target)
		m.setTarget(it.Name, target, true)
	}
}

// batchToggleTarget installs or uninstalls ALL filtered items for a single target.
// If any filtered item is NOT deployed to the target, it installs all of them.
// If all are already deployed, it uninstalls all of them.
func (m *Model) batchToggleTarget(target string) {
	if len(m.filtered) == 0 {
		return
	}
	// Check target is enabled
	enabled := false
	for _, t := range m.targets {
		if t == target {
			enabled = true
			break
		}
	}
	if !enabled {
		return
	}

	// Determine direction: if any item is not deployed, we install all; otherwise uninstall all
	allDeployed := true
	for _, it := range m.filtered {
		if !it.Targets[target] {
			allDeployed = false
			break
		}
	}

	if allDeployed {
		for _, it := range m.filtered {
			m.enqueue("uninstall", it, target)
			m.setTarget(it.Name, target, false)
		}
	} else {
		for _, it := range m.filtered {
			if !it.Targets[target] {
				m.enqueue("install", it, target)
				m.setTarget(it.Name, target, true)
			}
		}
	}
}

func (m *Model) enqueue(action string, it Item, target string) {
	// Remove same item+target if already queued
	var kept []ActionResult
	for _, a := range m.queue {
		if a.Item.Name == it.Name && a.Target == target {
			continue
		}
		kept = append(kept, a)
	}
	kept = append(kept, ActionResult{Action: action, Target: target, Mode: m.mode, Item: it})
	m.queue = kept
}

func (m *Model) setTarget(name, target string, val bool) {
	for i := range m.items {
		if m.items[i].Name == name {
			if m.items[i].Targets == nil {
				m.items[i].Targets = map[string]bool{}
			}
			m.items[i].Targets[target] = val
		}
	}
	for i := range m.filtered {
		if m.filtered[i].Name == name {
			if m.filtered[i].Targets == nil {
				m.filtered[i].Targets = map[string]bool{}
			}
			m.filtered[i].Targets[target] = val
		}
	}
}

func (m *Model) setAllTargets(name string, val bool) {
	for i := range m.items {
		if m.items[i].Name == name {
			if m.items[i].Targets == nil {
				m.items[i].Targets = map[string]bool{}
			}
			for _, t := range m.targets {
				m.items[i].Targets[t] = val
			}
		}
	}
	for i := range m.filtered {
		if m.filtered[i].Name == name {
			if m.filtered[i].Targets == nil {
				m.filtered[i].Targets = map[string]bool{}
			}
			for _, t := range m.targets {
				m.filtered[i].Targets[t] = val
			}
		}
	}
}

func (m Model) View() string {
	if m.quitting {
		return ""
	}

	listWidth := 42
	if m.width > 100 {
		listWidth = 50
	}
	detailWidth := m.width - listWidth - 4
	if detailWidth < 20 {
		detailWidth = 20
	}

	header := m.renderHeader()
	tabs := m.renderTabs()

	listHeight := m.height - 6
	if listHeight < 5 {
		listHeight = 5
	}

	list := m.renderList(listWidth, listHeight)
	detail := m.renderDetail(detailWidth, listHeight)

	body := lipgloss.JoinHorizontal(lipgloss.Top,
		list,
		"  ",
		detail,
	)

	status := m.renderStatus()
	help := m.renderHelp()

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		tabs,
		"",
		body,
		status,
		help,
	)
}

func (m Model) renderHeader() string {
	title := titleStyle.Render("anvil registry browse")

	// Build context line: branch (sha) | provider | tag
	var parts []string

	branch := m.info.Branch
	sha := m.info.SHA
	if len(sha) > 7 {
		sha = sha[:7]
	}
	if branch != "" {
		parts = append(parts, detailValueStyle.Render(branch)+detailLabelStyle.Render("@"+sha))
	}

	if m.info.Provider != "" {
		parts = append(parts, modeStyle.Render(m.info.Provider))
	}

	if m.info.Tag != "" && m.info.Tag != "none" {
		parts = append(parts, targetOnStyle.Render(m.info.Tag))
	}

	if len(parts) == 0 {
		return title
	}

	ctx := detailLabelStyle.Render("  ") + strings.Join(parts, detailLabelStyle.Render(" | "))
	return title + ctx
}

func (m Model) renderTabs() string {
	tabs := []filter{filterAll, filterAgents, filterSkills, filterCommands}
	parts := make([]string, len(tabs))
	for i, t := range tabs {
		label := fmt.Sprintf(" %s ", t.String())
		if t == m.filter {
			parts[i] = tabActiveStyle.Render(label)
		} else {
			parts[i] = tabInactiveStyle.Render(label)
		}
	}

	line := strings.Join(parts, "  ")
	if m.searching {
		line += "  " + searchStyle.Render("/") + m.search.View()
	} else if m.searchText != "" {
		line += "  " + searchStyle.Render("/ "+m.searchText)
	}
	return line
}

func (m Model) renderList(width, height int) string {
	var sb strings.Builder

	start := 0
	if m.cursor >= height {
		start = m.cursor - height + 1
	}
	end := start + height
	if end > len(m.filtered) {
		end = len(m.filtered)
	}

	for i := start; i < end; i++ {
		it := m.filtered[i]

		// Badge: show target-count based indicators
		badge := targetBadge(it, m.targets)

		nameStr := it.Name
		if len(nameStr) > width-14 {
			nameStr = nameStr[:width-17] + "..."
		}

		line := fmt.Sprintf(" %s %s %s",
			badge,
			typeStyle(it.Type).Render(it.Type),
			nameStr,
		)

		if i == m.cursor {
			pad := width - lipgloss.Width(line)
			if pad > 0 {
				line += strings.Repeat(" ", pad)
			}
			line = selectedStyle.Render(line)
		} else {
			line = normalStyle.Render(line)
		}

		sb.WriteString(line)
		if i < end-1 {
			sb.WriteByte('\n')
		}
	}

	rendered := strings.Count(sb.String(), "\n") + 1
	for rendered < height {
		sb.WriteByte('\n')
		rendered++
	}

	return sb.String()
}

// targetBadge returns a colored badge showing deployment status.
func targetBadge(it Item, enabledTargets []string) string {
	deployed := 0
	for _, t := range enabledTargets {
		if it.Targets[t] {
			deployed++
		}
	}
	total := len(enabledTargets)

	if deployed == 0 {
		return targetOffStyle.Render("○")
	}
	if deployed == total {
		return targetOnStyle.Render("●")
	}
	// Partially deployed
	return lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true).Render("◐")
}

func (m Model) renderDetail(width, height int) string {
	if len(m.filtered) == 0 {
		return detailLabelStyle.Render("No entries found")
	}

	it := m.filtered[m.cursor]
	var sb strings.Builder

	sb.WriteString(detailTitleStyle.Render(it.Name))
	sb.WriteString("\n\n")

	sb.WriteString(detailLabelStyle.Render("Type:      "))
	sb.WriteString(detailValueStyle.Render(it.Type))
	sb.WriteString("\n")

	if it.Version != "" {
		sb.WriteString(detailLabelStyle.Render("Version:   "))
		sb.WriteString(detailValueStyle.Render(it.Version))
		sb.WriteString("\n")
	}

	if it.Author != "" {
		sb.WriteString(detailLabelStyle.Render("Author:    "))
		sb.WriteString(detailValueStyle.Render(it.Author))
		sb.WriteString("\n")
	}

	sb.WriteString(detailLabelStyle.Render("Mode:      "))
	sb.WriteString(modeStyle.Render(m.mode.Label()))
	sb.WriteString("\n\n")

	// Per-target status
	sb.WriteString(detailLabelStyle.Render("Targets:\n"))
	targetKeys := []struct {
		name string
		key  string
	}{
		{TargetClaude, "c"},
		{TargetOpenCode, "o"},
		{TargetGemini, "g"},
		{TargetCodex, "x"},
		{TargetCursor, "r"},
	}

	for _, tk := range targetKeys {
		enabled := false
		for _, t := range m.targets {
			if t == tk.name {
				enabled = true
				break
			}
		}
		if !enabled {
			continue
		}

		deployed := it.Targets[tk.name]
		var icon, label string
		if deployed {
			icon = targetOnStyle.Render("  ✓")
			label = targetOnStyle.Render(fmt.Sprintf(" %-10s", tk.name))
		} else {
			icon = targetOffStyle.Render("  ○")
			label = targetOffStyle.Render(fmt.Sprintf(" %-10s", tk.name))
		}
		keyHint := detailLabelStyle.Render(fmt.Sprintf(" [%s]", tk.key))
		sb.WriteString(icon + label + keyHint + "\n")
	}

	sb.WriteString("\n")
	sb.WriteString(detailLabelStyle.Render("Description:\n"))
	desc := wordWrap(it.Description, width)
	sb.WriteString(detailValueStyle.Render(desc))

	// Show queued actions for this item
	var queued []string
	for _, a := range m.queue {
		if a.Item.Name == it.Name {
			if a.Action == "install" {
				queued = append(queued, queueInstall.Render(fmt.Sprintf("  + %s (%s)", a.Target, a.Mode)))
			} else {
				queued = append(queued, queueUninstall.Render(fmt.Sprintf("  - %s", a.Target)))
			}
		}
	}
	if len(queued) > 0 {
		sb.WriteString("\n\n")
		sb.WriteString(detailLabelStyle.Render("Queued changes:\n"))
		sb.WriteString(strings.Join(queued, "\n"))
	}

	return sb.String()
}

func (m Model) renderStatus() string {
	total := len(m.filtered)
	installed := 0
	for _, it := range m.filtered {
		if it.Installed() {
			installed++
		}
	}

	modeLabel := modeStyle.Render("[" + m.mode.String() + "]")

	queueInfo := ""
	if len(m.queue) > 0 {
		installs := 0
		uninstalls := 0
		for _, a := range m.queue {
			if a.Action == "install" {
				installs++
			} else {
				uninstalls++
			}
		}
		var parts []string
		if installs > 0 {
			parts = append(parts, fmt.Sprintf("+%d", installs))
		}
		if uninstalls > 0 {
			parts = append(parts, fmt.Sprintf("-%d", uninstalls))
		}
		queueInfo = fmt.Sprintf("  |  Queue: %s", strings.Join(parts, ", "))
	}

	return statusBarStyle.Render(fmt.Sprintf(
		"%d entries  |  %d deployed  %s%s",
		total, installed, modeLabel, queueInfo,
	))
}

func (m Model) renderHelp() string {
	keys := []string{
		"↑/↓ nav",
		"tab filter",
		"enter all targets",
		"c/o/g/x/r per target",
		"C/O/X/R all→target",
		"s mode",
		"/ search",
		"esc/q quit",
	}
	return helpStyle.Render("  " + strings.Join(keys, "  •  "))
}

// wordWrap wraps text to the given width.
func wordWrap(s string, width int) string {
	if width <= 0 {
		return s
	}
	words := strings.Fields(s)
	if len(words) == 0 {
		return ""
	}

	var lines []string
	line := words[0]
	for _, w := range words[1:] {
		if len(line)+1+len(w) > width {
			lines = append(lines, line)
			line = w
		} else {
			line += " " + w
		}
	}
	lines = append(lines, line)
	return strings.Join(lines, "\n")
}
