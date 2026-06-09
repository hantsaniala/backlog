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
	width    int
}

func newScreenBacklogHealth(b *model.Backlog) *backlogHealthModel {
	return &backlogHealthModel{backlog: b}
}

func (s *backlogHealthModel) Init() tea.Cmd { return nil }

func (s *backlogHealthModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.width = msg.Width
		s.ready = false
	case reloadMsg:
		s.ready = false
	}
	s.viewport, cmd = s.viewport.Update(msg)
	return s, cmd
}

func (s *backlogHealthModel) View() string {
	var b strings.Builder
	b.WriteString(headerStyle.Render("  5 Health"))
	b.WriteString("\n\n")

	report := s.backlog.CheckHealth()

	barW := s.barWidth()

	errCount := 0
	warnCount := 0
	infoCount := 0
	for _, issue := range report.Issues {
		switch issue.Severity {
		case "error":
			errCount++
		case "warning":
			warnCount++
		default:
			infoCount++
		}
	}

	// Severity bars
	maxCount := errCount
	if warnCount > maxCount {
		maxCount = warnCount
	}
	if infoCount > maxCount {
		maxCount = infoCount
	}
	if maxCount == 0 {
		maxCount = 1
	}

	drawBar := func(label string, count int, color lipgloss.Color) {
		f := count * barW / maxCount
		if f > barW {
			f = barW
		}
		bar := lipgloss.NewStyle().Foreground(color).Render(strings.Repeat("█", f) + strings.Repeat("░", barW-f))
		b.WriteString(lipgloss.NewStyle().Padding(0, 2).Render(fmt.Sprintf("  %-10s %s  %d", label, bar, count)))
		b.WriteString("\n")
	}

	if len(report.Issues) == 0 {
		b.WriteString(lipgloss.NewStyle().Foreground(colorSuccess).Bold(true).Padding(0, 2).Render(" ✓ No issues found. Backlog is healthy."))
		b.WriteString("\n\n")
	} else {
		drawBar("errors", errCount, colorError)
		drawBar("warnings", warnCount, colorWarning)
		drawBar("info", infoCount, colorInfo)
		b.WriteString("\n")

		for _, issue := range report.Issues {
			var icon string
			var sevColor lipgloss.Color
			switch issue.Severity {
			case "error":
				icon = " ✖"
				sevColor = colorError
			case "warning":
				icon = " ⚠"
				sevColor = colorWarning
			default:
				icon = " ℹ"
				sevColor = colorInfo
			}
			msg := fmt.Sprintf("%s %s", icon, issue.Message)
			if issue.TaskID != "" {
				msg += fmt.Sprintf(" [%s]", issue.TaskID)
			}
			b.WriteString(lipgloss.NewStyle().Padding(0, 2).Foreground(sevColor).Render(msg))
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	// Stats grid
	b.WriteString(lipgloss.NewStyle().Foreground(colorTextBright).Bold(true).Padding(0, 2).Render("Stats"))
	b.WriteString("\n\n")

	counts := s.backlog.StatusCounts("")
	totalTasks := 0
	for _, c := range counts {
		totalTasks += c
	}

	stats := []struct {
		label string
		value string
	}{
		{"Tasks", fmt.Sprintf("%d", totalTasks)},
		{"Epics", fmt.Sprintf("%d", len(s.backlog.AllEpics))},
		{"Sprints", fmt.Sprintf("%d", len(s.backlog.Current.Sprints))},
		{"Projects", fmt.Sprintf("%d", 1+len(s.backlog.Externals))},
	}

	// Two-column grid
	for i, stat := range stats {
		if i%2 == 0 {
			b.WriteString(lipgloss.NewStyle().Padding(0, 2).Render(
				fmt.Sprintf("  %-12s %s", stat.label+":", stat.value)))
		} else {
			b.WriteString(lipgloss.NewStyle().Render(
				fmt.Sprintf("    %-12s %s", stat.label+":", stat.value)))
		}
		if i%2 == 1 || i == len(stats)-1 {
			b.WriteString("\n")
		}
	}

	content := lipgloss.NewStyle().Padding(0, 2).Render(b.String())
	w := s.width - 4
	if w < 40 {
		w = 80
	}
	if !s.ready || s.width > 0 {
		s.viewport = viewport.New(w, 20)
		s.viewport.SetContent(content)
		s.ready = true
	} else {
		s.viewport.SetContent(content)
	}
	return s.viewport.View()
}

func (s *backlogHealthModel) barWidth() int {
	w := s.width - 22
	if w < 8 {
		return 8
	}
	if w > 30 {
		return 30
	}
	return w
}
