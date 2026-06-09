package app

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/hantsaniala/backlog/model"
)

type taskDetailModel struct {
	task     *model.Task
	backlog  *model.Backlog
	viewport viewport.Model
	rendered string
	ready    bool
}

func newScreenTaskDetail(b *model.Backlog) *taskDetailModel {
	return &taskDetailModel{backlog: b}
}

func (s *taskDetailModel) Init() tea.Cmd { return nil }

func (s *taskDetailModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	s.viewport, cmd = s.viewport.Update(msg)
	return s, cmd
}

func (s *taskDetailModel) View() string {
	if s.task == nil {
		return headerStyle.Render("  Task Detail") + "\n\n  No task selected"
	}
	if !s.ready {
		s.renderTask()
	}
	return s.viewport.View()
}

func (s *taskDetailModel) renderTask() {
	t := s.task
	var b strings.Builder

	b.WriteString(headerStyle.Render(fmt.Sprintf("  %s", t.ID)))
	if t.ProjectID != s.backlog.Current.Config.ProjectID {
		b.WriteString(" ")
		b.WriteString(ExternalBadge())
	}
	b.WriteString("\n\n")

	b.WriteString(lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(colorBorder).Padding(0, 1).Render(
		lipgloss.JoinVertical(lipgloss.Top,
			lipgloss.NewStyle().Foreground(colorTextDim).Render(fmt.Sprintf("Type:     %s", TypeBadge(string(t.Type)))),
			lipgloss.NewStyle().Foreground(colorTextDim).Render(fmt.Sprintf("Status:   %s", StatusBadge(string(t.Status)))),
			lipgloss.NewStyle().Foreground(colorTextDim).Render(fmt.Sprintf("Priority: %s", PriorityBadge(string(t.Priority)))),
			fieldLine("Severity", string(t.Severity)),
			fieldLine("Assignee", t.Assignee),
			fieldLine("Reporter", t.Reporter),
			fieldLine("Story Points", fmt.Sprintf("%d", safeSP(t.StoryPoints))),
			fieldLine("Epic", t.Epic),
			fieldLine("Sprint", t.Sprint),
			fieldLine("Created", t.Created),
			fieldLine("Updated", t.Updated),
			fieldLine("Due", t.DueDate),
		),
	))
	b.WriteString("\n\n")

	if len(t.Labels) > 0 {
		b.WriteString(lipgloss.NewStyle().Foreground(colorTextDim).Render("Labels: "))
		for i, l := range t.Labels {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(l)
		}
		b.WriteString("\n\n")
	}

	if len(t.DependsOn) > 0 {
		b.WriteString(lipgloss.NewStyle().Foreground(colorWarning).Render("Depends on: "))
		b.WriteString(strings.Join(t.DependsOn, ", "))
		b.WriteString("\n")
	}
	if len(t.Blocks) > 0 {
		b.WriteString(lipgloss.NewStyle().Foreground(colorError).Render("Blocks: "))
		b.WriteString(strings.Join(t.Blocks, ", "))
		b.WriteString("\n")
	}
	if len(t.RelatedTo) > 0 {
		b.WriteString(lipgloss.NewStyle().Foreground(colorInfo).Render("Related: "))
		b.WriteString(strings.Join(t.RelatedTo, ", "))
		b.WriteString("\n")
	}

	if t.Body != "" {
		b.WriteString("\n")
		rendered, err := glamour.Render(t.Body, "dark")
		if err == nil {
			b.WriteString(rendered)
		} else {
			b.WriteString(t.Body)
		}
	}

	content := lipgloss.NewStyle().Padding(0, 2).Render(b.String())

	if !s.ready {
		s.viewport = viewport.New(80, 30)
		s.viewport.SetContent(content)
		s.ready = true
	} else {
		s.viewport.SetContent(content)
	}

	s.rendered = content
}

func fieldLine(name, value string) string {
	if value == "" || value == "0" {
		return ""
	}
	return lipgloss.NewStyle().Foreground(colorTextDim).Render(fmt.Sprintf("%-13s %s", name+":", value))
}

func safeSP(sp *int) int {
	if sp == nil {
		return 0
	}
	return *sp
}
