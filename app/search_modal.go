package app

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hantsaniala/backlog/model"
)

type SearchNavigateMsg struct {
	Task *model.Task
}

type searchModalModel struct {
	backlog  *model.Backlog
	input    textinput.Model
	results  []*model.Task
	cursor   int
	width    int
	height   int
	allTasks []*model.Task
}

func newSearchModal(b *model.Backlog, width, height int) *searchModalModel {
	ti := textinput.New()
	ti.Placeholder = "search by title, ID, or description..."
	ti.CharLimit = 100
	ti.Width = width - 10
	ti.Focus()

	return &searchModalModel{
		backlog:  b,
		input:    ti,
		results:  b.AllTasks,
		width:    width,
		height:   height,
		allTasks: b.AllTasks,
	}
}

func (m *searchModalModel) Init() tea.Cmd { return textinput.Blink }

func (m *searchModalModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "ctrl+c":
			return m, func() tea.Msg { return nil }
		case "enter":
			if len(m.results) > 0 && m.cursor >= 0 && m.cursor < len(m.results) {
				task := m.results[m.cursor]
				return m, func() tea.Msg {
					return SearchNavigateMsg{Task: task}
				}
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

func (m *searchModalModel) filter() {
	query := strings.ToLower(m.input.Value())
	if query == "" {
		m.results = m.allTasks
		m.cursor = 0
		return
	}
	var filtered []*model.Task
	for _, t := range m.allTasks {
		if strings.Contains(strings.ToLower(t.ID), query) ||
			strings.Contains(strings.ToLower(t.Summary), query) ||
			strings.Contains(strings.ToLower(t.Description), query) {
			filtered = append(filtered, t)
		}
	}
	m.results = filtered
	if m.cursor >= len(m.results) {
		m.cursor = len(m.results) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
}

func (m *searchModalModel) View() string {
	modalW := m.width * 80 / 100
	if modalW < 50 {
		modalW = 50
	}
	modalH := m.height * 60 / 100
	if modalH < 15 {
		modalH = 15
	}

	var b strings.Builder

	// Input line
	prompt := fmt.Sprintf(" 🔍 Filter: ")
	status := fmt.Sprintf("%d results", len(m.results))
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Center,
		lipgloss.NewStyle().Foreground(colorTextBright).Render(prompt),
		m.input.View(),
		lipgloss.NewStyle().Foreground(colorTextDim).PaddingLeft(1).Render(status),
	))
	b.WriteString("\n\n")

	// Separator
	sep := lipgloss.NewStyle().Foreground(colorBorder).Render(strings.Repeat("─", modalW-6))
	b.WriteString("  " + sep + "\n\n")

	// Results
	if len(m.results) == 0 {
		b.WriteString(lipgloss.NewStyle().Foreground(colorTextDim).Padding(1, 2).Render("No matching items"))
	} else {
		maxResults := 10
		if maxResults > len(m.results) {
			maxResults = len(m.results)
		}
		for i := 0; i < maxResults; i++ {
			t := m.results[i]
			g := statusDot(string(t.Status))
			label := fmt.Sprintf("%s %s  %s", typeDot(string(t.Type)), t.ID, t.Summary)
			if t.Summary == "" {
				label = fmt.Sprintf("%s %s", typeDot(string(t.Type)), t.ID)
			}
			line := fmt.Sprintf("  %s  %s  %s", g, label, StatusBadge(string(t.Status)))
			if i == m.cursor {
				line = lipgloss.NewStyle().
					Foreground(colorTextBright).
					Background(colorPrimary).
					Padding(0, 1).
					Render(line)
			} else {
				line = lipgloss.NewStyle().Foreground(colorText).Padding(0, 1).Render(line)
			}
			b.WriteString(line)
			b.WriteString("\n")
		}
		if len(m.results) > maxResults {
			b.WriteString(lipgloss.NewStyle().Foreground(colorTextDim).Padding(1, 2).Render(
				fmt.Sprintf("... and %d more results", len(m.results)-maxResults)))
		}
	}

	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Foreground(colorTextDim).Render(" Enter: navigate  |  Esc: cancel"))

	content := lipgloss.NewStyle().
		Width(modalW).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorPrimary).
		Padding(1, 2).
		Background(colorSurface).
		Render(b.String())

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}
