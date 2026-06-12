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
	height   int
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
		s.height = msg.Height
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

	// Active sprint ribbon
	if len(s.backlog.Current.Sprints) > 0 {
		sp := s.backlog.Current.Sprints[0]
		sp.Tasks = s.backlog.TasksBySprint(sp.Name)
		total := sp.TotalPoints()
		done := sp.CompletedPoints()
		barW := s.barWidth()
		bar := progressBar(done, total, barW)
		b.WriteString(lipgloss.NewStyle().Padding(0, 2).Render(
			fmt.Sprintf(" %s  %s", StatusBadge("in-progress"), sp.Name)))
		b.WriteString(lipgloss.NewStyle().Padding(0, 2).Foreground(colorTextDim).Render(
			fmt.Sprintf("  %s", bar)))
		b.WriteString("\n\n")
	}

	s.printProjectCard(&b, s.backlog.Current, false)

	for _, ext := range s.backlog.Externals {
		s.printProjectCard(&b, ext, true)
	}

	// Health merged
	s.printHealth(&b)

	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Padding(0, 2).Foreground(colorTextDim).Render(" Watching for changes... live reload active"))

	content := lipgloss.NewStyle().Padding(0, 2).Render(b.String())
	w := s.width - 2
	if w < 40 {
		w = 80
	}
	viewportH := s.height - 5
	if viewportH < 10 {
		viewportH = 10
	}
	if !s.ready || s.width > 0 {
		s.viewport = viewport.New(w, viewportH)
		s.viewport.SetContent(content)
		s.ready = true
	} else {
		s.viewport.SetContent(content)
	}
	vpView := s.viewport.View()
	sb := renderScrollbar(s.viewport, viewportH)
	return addScrollbar(vpView, sb)
}

func (s *dashboardModel) printHealth(b *strings.Builder) {
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

	healthColor := colorSuccess
	if errCount > 0 {
		healthColor = colorError
	} else if warnCount > 0 {
		healthColor = colorWarning
	}
	b.WriteString(lipgloss.NewStyle().Padding(0, 2).Foreground(healthColor).Bold(true).Render(" ● Health"))
	b.WriteString("\n")

	if len(report.Issues) == 0 {
		b.WriteString(lipgloss.NewStyle().Padding(0, 3).Foreground(colorSuccess).Render(" All clear"))
		b.WriteString("\n\n")
		return
	}

	drawBar := func(label string, count int, color lipgloss.Color) {
		f := count * barW / maxCount
		if f > barW {
			f = barW
		}
		bar := lipgloss.NewStyle().Foreground(color).Render(strings.Repeat("█", f) + strings.Repeat("░", barW-f))
		b.WriteString(lipgloss.NewStyle().Padding(0, 3).Render(fmt.Sprintf("%s  %-10s %s  %d", statusDot(label), label, bar, count)))
		b.WriteString("\n")
	}
	drawBar("errors", errCount, colorError)
	drawBar("warnings", warnCount, colorWarning)

	for _, issue := range report.Issues {
		var icon string
		var sevColor lipgloss.Color
		switch issue.Severity {
		case "error":
			icon = "  ✖"
			sevColor = colorError
		case "warning":
			icon = "  ⚠"
			sevColor = colorWarning
		default:
			icon = "  ℹ"
			sevColor = colorInfo
		}
		msg := fmt.Sprintf("%s %s", icon, issue.Message)
		if issue.TaskID != "" {
			msg += fmt.Sprintf(" [%s]", issue.TaskID)
		}
		b.WriteString(lipgloss.NewStyle().Padding(0, 3).Foreground(sevColor).Render(msg))
		b.WriteString("\n")
	}

	counts := s.backlog.StatusCounts("")
	totalTasks := 0
	for _, c := range counts {
		totalTasks += c
	}
	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Padding(0, 3).Foreground(colorTextDim).Render(
		fmt.Sprintf("tasks: %d | epics: %d | sprints: %d | projects: %d",
			totalTasks, len(s.backlog.AllEpics), len(s.backlog.Current.Sprints), 1+len(s.backlog.Externals))))
	b.WriteString("\n")
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

func (s *dashboardModel) printProjectCard(b *strings.Builder, snap *model.ProjectSnapshot, external bool) {
	counts := s.backlog.StatusCounts(snap.Config.ProjectID)
	title := snap.Config.ProjectID
	if external {
		title += " " + ExternalBadge()
	}
	b.WriteString(lipgloss.NewStyle().Foreground(colorTextBright).Bold(true).Padding(0, 2).Render(title))
	b.WriteString("\n")

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
		g := statusDot(string(st.status))
		bar := lipgloss.NewStyle().Foreground(st.color).Render(strings.Repeat("█", f) + strings.Repeat("░", barW-f))
		b.WriteString(lipgloss.NewStyle().Padding(0, 3).Render(fmt.Sprintf("%s %-11s %s  %d", g, st.label, bar, n)))
		b.WriteString("\n")
	}
	b.WriteString("\n")
}

func (s *dashboardModel) footerHint() string {
	return " 2:tasks | 3:sprints | ::cmd | q:quit"
}

func (s *dashboardModel) footerPos() string { return "" }

func (s *dashboardModel) refresh() {
	s.ready = false
}
