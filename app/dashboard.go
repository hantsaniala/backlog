package app

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hantsaniala/backlog/model"
)

type dashboardModel struct {
	backlog  *model.Backlog
	viewport viewport.Model
	ready    bool
	width    int
}

func newScreenDashboard(b *model.Backlog) *dashboardModel {
	return &dashboardModel{backlog: b}
}

func (s *dashboardModel) Init() tea.Cmd { return nil }

func (s *dashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.width = msg.Width
		s.ready = false
	case reloadMsg:
		s.refresh()
	}
	s.viewport, cmd = s.viewport.Update(msg)
	return s, cmd
}

func (s *dashboardModel) View() string {
	var b strings.Builder

	b.WriteString(headerStyle.Render("  1 Dashboard"))
	b.WriteString("\n\n")

	s.printProjectCard(&b, s.backlog.Current, false, s.width)

	for _, ext := range s.backlog.Externals {
		s.printProjectCard(&b, ext, true, s.width)
	}

	// Health summary
	health := s.backlog.CheckHealth()
	barW := s.barWidth()
	b.WriteString(lipgloss.NewStyle().Foreground(colorTextBright).Bold(true).Padding(0, 2).Render("Backlog Health"))
	b.WriteString("\n\n")
	if len(health.Issues) == 0 {
		b.WriteString(lipgloss.NewStyle().Padding(0, 2).Foreground(colorSuccess).Render(" ✓ No issues found"))
	} else {
		errCount := 0
		warnCount := 0
		for _, issue := range health.Issues {
			if issue.Severity == "error" {
				errCount++
			} else {
				warnCount++
			}
		}
		if errCount > 0 {
			f := errCount
			if f > barW {
				f = barW
			}
			errBar := lipgloss.NewStyle().Foreground(colorError).Render(strings.Repeat("█", f) + strings.Repeat("░", barW-f))
			b.WriteString(lipgloss.NewStyle().Padding(0, 2).Render(fmt.Sprintf("  errors    %s  %d", errBar, errCount)))
			b.WriteString("\n")
		}
		if warnCount > 0 {
			f := warnCount
			if f > barW {
				f = barW
			}
			warnBar := lipgloss.NewStyle().Foreground(colorWarning).Render(strings.Repeat("█", f) + strings.Repeat("░", barW-f))
			b.WriteString(lipgloss.NewStyle().Padding(0, 2).Render(fmt.Sprintf("  warnings  %s  %d", warnBar, warnCount)))
			b.WriteString("\n")
		}
		// First 3 issues
		limit := 3
		if len(health.Issues) < limit {
			limit = len(health.Issues)
		}
		b.WriteString("\n")
		for i := 0; i < limit; i++ {
			issue := health.Issues[i]
			var sevColor lipgloss.Color
			icon := " •"
			if issue.Severity == "error" {
				sevColor = colorError
				icon = " ✖"
			} else {
				sevColor = colorWarning
				icon = " ⚠"
			}
			msg := fmt.Sprintf("%s %s", icon, issue.Message)
			if issue.TaskID != "" {
				msg += fmt.Sprintf(" [%s]", issue.TaskID)
			}
			b.WriteString(lipgloss.NewStyle().Padding(0, 2).Foreground(sevColor).Render(msg))
			b.WriteString("\n")
		}
		if len(health.Issues) > limit {
			b.WriteString(lipgloss.NewStyle().Padding(0, 3).Foreground(colorTextDim).Render(fmt.Sprintf("+%d more issues (screen 5)", len(health.Issues)-limit)))
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Padding(0, 2).Foreground(colorTextDim).Render(" Watching for file changes... live reload active"))

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

func (s *dashboardModel) barWidth() int {
	w := s.width - 24
	if w < 8 {
		return 8
	}
	if w > 30 {
		return 30
	}
	return w
}

func (s *dashboardModel) printProjectCard(b *strings.Builder, snap *model.ProjectSnapshot, external bool, termWidth int) {
	counts := s.backlog.StatusCounts(snap.Config.ProjectID)
	total := 0
	for _, c := range counts {
		total += c
	}

	title := snap.Config.ProjectID + " (" + snap.Config.Name + ")"
	if external {
		title += " ext"
	}
	b.WriteString(lipgloss.NewStyle().Foreground(colorTextBright).Bold(true).Padding(0, 2).Render(title))
	b.WriteString("\n\n")

	barW := s.barWidth()

	statusOrder := []struct {
		label  string
		status model.Status
		color  lipgloss.Color
	}{
		{"todo", model.StatusTodo, colorInfo},
		{"in-progress", model.StatusInProgress, colorWarning},
		{"review", model.StatusReview, colorSecondary},
		{"on-hold", model.StatusOnHold, colorTextDim},
		{"done", model.StatusDone, colorSuccess},
		{"cancelled", model.StatusCancelled, colorError},
	}

	for _, st := range statusOrder {
		n := counts[st.status]
		if n == 0 {
			continue
		}
		f := n
		if f > barW {
			f = barW
		}
		bar := lipgloss.NewStyle().Foreground(st.color).Render(strings.Repeat("█", f) + strings.Repeat("░", barW-f))
		line := fmt.Sprintf("  %-12s %s  %d", st.label, bar, n)
		b.WriteString(lipgloss.NewStyle().Padding(0, 2).Render(line))
		b.WriteString("\n")
	}
	b.WriteString("\n")
}

func (s *dashboardModel) refresh() {
	s.ready = false
}
