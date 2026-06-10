package app

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type paletteCommand struct {
	label       string
	description string
	aliases     []string
}

var paletteCommands = []paletteCommand{
	{label: "status todo", description: "Set item status to Todo"},
	{label: "status in-progress", description: "Set item status to In Progress"},
	{label: "status review", description: "Set item status to Review"},
	{label: "status done", description: "Set item status to Done"},
	{label: "status on-hold", description: "Set item status to On Hold"},
	{label: "status cancelled", description: "Set item status to Cancelled"},
	{label: "assign", description: "Assign current item to a user"},
	{label: "sprint", description: "Move item to a sprint"},
	{label: "export csv", description: "Export backlog as CSV"},
	{label: "export json", description: "Export backlog as JSON"},
	{label: "filter status=todo", description: "Filter by status: todo", aliases: []string{"f status todo"}},
	{label: "filter status=in-progress", description: "Filter by status: in progress"},
	{label: "filter status=done", description: "Filter by status: done"},
	{label: "filter priority=high", description: "Filter by priority: high"},
	{label: "filter priority=critical", description: "Filter by priority: critical"},
	{label: "focus dashboard", description: "Switch to dashboard view"},
	{label: "focus backlog", description: "Switch to backlog view"},
	{label: "focus sprints", description: "Switch to sprint view"},
}

type paletteExecuteMsg struct {
	command string
}

type historyNavigateMsg struct {
	index int
}

type paletteModel struct {
	input    textinput.Model
	cursor   int
	results  []paletteCommand
	recent   []string // last 5 executed commands
	width    int
	height   int
	allCmds  []paletteCommand

	// History mode (g h)
	historyMode   bool
	historyItems  []ViewState
}

func newPaletteModel(width, height int) *paletteModel {
	ti := textinput.New()
	ti.Placeholder = "Type a command..."
	ti.CharLimit = 100
	ti.Width = width - 20
	ti.Focus()

	m := &paletteModel{
		input:   ti,
		width:   width,
		height:  height,
		allCmds: paletteCommands,
		recent:  make([]string, 0),
	}
	m.filter()
	return m
}

func (m *paletteModel) Init() tea.Cmd { return textinput.Blink }

func (m *paletteModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "ctrl+c":
			return m, func() tea.Msg { return nil }
		case "enter":
			if m.historyMode {
				if m.cursor >= 0 && m.cursor < len(m.historyItems) {
					return m, func() tea.Msg { return historyNavigateMsg{index: m.cursor} }
				}
				return m, nil
			}
			if len(m.results) > 0 && m.cursor >= 0 && m.cursor < len(m.results) {
				cmd := m.results[m.cursor].label
				m.addRecent(cmd)
				return m, func() tea.Msg { return paletteExecuteMsg{command: cmd} }
			}
			return m, nil
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
			return m, nil
		case "down", "j":
			if m.cursor < len(m.results)-1 {
				m.cursor++
			}
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	m.filter()
	return m, cmd
}

func (m *paletteModel) filter() {
	query := strings.ToLower(m.input.Value())

	if m.historyMode {
		// In history mode, filter by screen name or task ID
		type histItem struct {
			label string
			idx   int
		}
		var filtered []histItem
		for i, vs := range m.historyItems {
			label := breadcrumbFromScreen(vs.Screen)
			if vs.TaskID != "" {
				label = vs.TaskID
			}
			if query == "" || strings.Contains(strings.ToLower(label), query) {
				filtered = append(filtered, histItem{label: label, idx: i})
			}
		}
		// Fake paletteCommand results for View
		var cmds []paletteCommand
		for _, fi := range filtered {
			label := fmt.Sprintf("(history entry %d)", fi.idx+1)
			if m.historyItems[fi.idx].Screen == screenDashboard {
				label = "Dashboard"
			} else if m.historyItems[fi.idx].Screen == screenTaskList {
				label = "Backlog"
			} else if m.historyItems[fi.idx].Screen == screenSprintView {
				label = "Sprints"
			}
			if m.historyItems[fi.idx].TaskID != "" {
				label += " - " + m.historyItems[fi.idx].TaskID
			}
			if m.historyItems[fi.idx].FilterText != "" {
				label += " [" + m.historyItems[fi.idx].FilterText + "]"
			}
			cmds = append(cmds, paletteCommand{label: label, description: "jump to this view"})
		}
		// Assign cursor from filtered items back to original index
		m.results = cmds
		if m.cursor >= len(m.results) {
			m.cursor = len(m.results) - 1
		}
		if m.cursor < 0 {
			m.cursor = 0
		}
		return
	}

	if query == "" {
		m.results = m.allCmds
		m.cursor = 0
		return
	}

	// Fuzzy-ish: prefix match + contains
	var matched []paletteCommand
	for _, c := range m.allCmds {
		normalized := strings.ToLower(c.label)
		if strings.HasPrefix(normalized, query) || strings.Contains(normalized, query) {
			matched = append(matched, c)
			continue
		}
		for _, a := range c.aliases {
			na := strings.ToLower(a)
			if strings.HasPrefix(na, query) || strings.Contains(na, query) {
				matched = append(matched, c)
				break
			}
		}
	}
	// Sort by prefix match first, then contains
	sort.Slice(matched, func(i, j int) bool {
		ni := strings.ToLower(matched[i].label)
		nj := strings.ToLower(matched[j].label)
		pi := strings.HasPrefix(ni, query)
		pj := strings.HasPrefix(nj, query)
		if pi != pj {
			return pi
		}
		return ni < nj
	})
	m.results = matched
	if m.cursor >= len(m.results) {
		m.cursor = len(m.results) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
}

func (m *paletteModel) addRecent(cmd string) {
	m.recent = append([]string{cmd}, m.recent...)
	if len(m.recent) > 5 {
		m.recent = m.recent[:5]
	}
}

func (m *paletteModel) View() string {
	modalW := m.width * 70 / 100
	if modalW < 50 {
		modalW = 50
	}
	modalH := m.height * 50 / 100
	if modalH < 15 {
		modalH = 15
	}

	var b strings.Builder

	if m.historyMode {
		b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(colorTextBright).Render(" Navigation History"))
		b.WriteString(fmt.Sprintf("  %d entries", len(m.historyItems)))
	} else {
		b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(colorTextBright).Render(" Command Palette"))
		b.WriteString(fmt.Sprintf("  %d commands", len(m.results)))
	}
	b.WriteString("\n\n")

	// Input
	b.WriteString(lipgloss.NewStyle().Foreground(colorTextDim).Render(" : "))
	b.WriteString(m.input.View())
	b.WriteString("\n\n")

	// Recent commands (only in normal mode)
	if !m.historyMode && len(m.recent) > 0 && m.input.Value() == "" {
		b.WriteString(paletteRecentStyle.Render(" Recent"))
		b.WriteString("\n")
		for _, r := range m.recent {
			b.WriteString(fmt.Sprintf("  %s\n", r))
		}
		b.WriteString("\n")
	}

	// Results
	maxResults := modalH - 8
	if maxResults > len(m.results) {
		maxResults = len(m.results)
	}
	for i := 0; i < maxResults; i++ {
		c := m.results[i]
		line := fmt.Sprintf("  %s  %s", c.label, paletteDescStyle.Render(c.description))
		if i == m.cursor {
			line = paletteSelectedStyle.Render(line)
		} else {
			line = paletteResultStyle.Render(line)
		}
		b.WriteString(line)
		b.WriteString("\n")
	}
	if len(m.results) > maxResults {
		b.WriteString(paletteDescStyle.Render(fmt.Sprintf("  ... %d more", len(m.results)-maxResults)))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	if m.historyMode {
		b.WriteString(paletteDescStyle.Render(" ↓↑ navigate | Enter jump | Esc cancel"))
	} else {
		b.WriteString(paletteDescStyle.Render(" ↓↑ navigate | Enter execute | Esc cancel"))
	}

	content := lipgloss.NewStyle().
		Width(modalW).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorPrimary).
		Padding(1, 2).
		Background(colorSurface).
		Render(b.String())

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}
