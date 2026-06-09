package app

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hantsaniala/backlog/model"
)

type StatusUpdateMsg struct {
	TaskID    string
	OldStatus model.Status
	NewStatus model.Status
}

type statusPopupModel struct {
	task     *model.Task
	cursor   int
	width    int
	statuses []model.Status
}

func newStatusPopup(task *model.Task, width int) *statusPopupModel {
	statuses := []model.Status{
		model.StatusTodo,
		model.StatusInProgress,
		model.StatusReview,
		model.StatusDone,
	}
	cursor := 0
	for i, s := range statuses {
		if s == task.Status {
			cursor = i
			break
		}
	}
	return &statusPopupModel{task: task, cursor: cursor, width: width, statuses: statuses}
}

func (m *statusPopupModel) Init() tea.Cmd { return nil }

func (m *statusPopupModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
			return m, nil
		case "down", "j":
			if m.cursor < len(m.statuses)-1 {
				m.cursor++
			}
			return m, nil
		case "enter":
			oldStatus := m.task.Status
			newStatus := m.statuses[m.cursor]
			return m, func() tea.Msg {
				return StatusUpdateMsg{
					TaskID:    m.task.ID,
					OldStatus: oldStatus,
					NewStatus: newStatus,
				}
			}
		case "esc":
			return m, func() tea.Msg { return nil }
		}
	}
	return m, nil
}

func (m *statusPopupModel) View() string {
	var b strings.Builder
	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(colorTextBright).Render(" Update Status"))
	b.WriteString("\n\n")

	for i, s := range m.statuses {
		g := statusDot(string(s))
		selected := i == m.cursor
		prefix := "  "
		if selected {
			prefix = " ✓"
		}
		line := fmt.Sprintf("%s %s %s", prefix, g, strings.ToUpper(string(s)))
		if selected {
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

	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Foreground(colorTextDim).Render(" Enter confirm | Esc cancel"))

	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorPrimary).
		Padding(1, 2).
		Background(colorSurface).
		Width(m.width - 4)

	return style.Render(b.String())
}
