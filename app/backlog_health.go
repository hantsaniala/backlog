package app

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hantsaniala/backlog/model"
)

type backlogHealthModel struct {
	backlog  *model.Backlog
	viewport viewport.Model
	ready    bool
}

func newScreenBacklogHealth(b *model.Backlog) *backlogHealthModel {
	return &backlogHealthModel{backlog: b}
}

func (s *backlogHealthModel) Init() tea.Cmd { return nil }

func (s *backlogHealthModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg.(type) {
	case reloadMsg:
		s.ready = false
	}
	s.viewport, cmd = s.viewport.Update(msg)
	return s, cmd
}

func (s *backlogHealthModel) View() string {
	var b strings.Builder
	b.WriteString(headerStyle.Render("  Backlog Health"))
	b.WriteString("\n\n")

	report := s.backlog.CheckHealth()

	if len(report.Issues) == 0 {
		b.WriteString(lipgloss.NewStyle().Foreground(colorSuccess).Bold(true).Padding(0, 2).
			Render("\u2713 No issues found. Backlog is healthy."))
	} else {
		errCount := 0
		warnCount := 0
		for _, issue := range report.Issues {
			if issue.Severity == "error" {
				errCount++
			} else {
				warnCount++
			}
		}
		b.WriteString(lipgloss.NewStyle().Foreground(colorTextDim).Render(
			fmt.Sprintf("  %d errors, %d warnings\n\n", errCount, warnCount)))

		for _, issue := range report.Issues {
			var icon string
			var sevColor lipgloss.Color
			switch issue.Severity {
			case "error":
				icon = "\u2716"
				sevColor = colorError
			case "warning":
				icon = "\u26A0"
				sevColor = colorWarning
			default:
				icon = "\u2139"
				sevColor = colorInfo
			}
			style := lipgloss.NewStyle().Foreground(sevColor)
			b.WriteString(fmt.Sprintf("  %s %s", style.Render(icon), style.Render(issue.Message)))
			if issue.TaskID != "" {
				b.WriteString(lipgloss.NewStyle().Foreground(colorTextDim).Render(fmt.Sprintf(" [%s]", issue.TaskID)))
			}
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Foreground(colorPrimary).Bold(true).Render("  Stats"))
	b.WriteString("\n\n")
	counts := s.backlog.StatusCounts("")
	total := 0
	for _, c := range counts {
		total += c
	}
	b.WriteString(lipgloss.NewStyle().Foreground(colorTextDim).Render(fmt.Sprintf("  Total tasks: %d", total)))
	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Foreground(colorTextDim).Render(fmt.Sprintf("  Projects: %d (1 current + %d external)",
		1+len(s.backlog.Externals), len(s.backlog.Externals))))
	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Foreground(colorTextDim).Render(fmt.Sprintf("  Epics: %d", len(s.backlog.AllEpics))))
	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Foreground(colorTextDim).Render(fmt.Sprintf("  Sprints: %d", len(s.backlog.Current.Sprints))))
	b.WriteString("\n")

	content := lipgloss.NewStyle().Padding(0, 2).Render(b.String())
	if !s.ready {
		s.viewport = viewport.New(80, 20)
		s.viewport.SetContent(content)
		s.ready = true
	} else {
		s.viewport.SetContent(content)
	}
	return s.viewport.View()
}
