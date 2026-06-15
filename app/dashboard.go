package app

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hantsaniala/backlog/model"
)

type dashboardModel struct {
	backlog  *model.Backlog
	viewport viewport.Model
	prog     progress.Model
	ready    bool
	width    int
	height   int
}

func newScreenDashboard(b *model.Backlog) *dashboardModel {
	return &dashboardModel{
		backlog: b,
		prog:    newProgressBar(40),
	}
}

func (s *dashboardModel) Init() tea.Cmd { return nil }

func (s *dashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.width = msg.Width
		s.height = msg.Height
		s.ready = false
		contentWidth := s.width - 2
		barW := contentWidth - 8
		if barW < 10 {
			barW = 10
		}
		s.prog = newProgressBar(barW)
	case reloadMsg:
		s.refresh()
	}
	s.viewport, cmd = s.viewport.Update(msg)
	return s, cmd
}

func (s *dashboardModel) View() string {
	var b strings.Builder
	contentWidth := s.width - 2
	if contentWidth < 40 {
		contentWidth = 80
	}

	// Header
	b.WriteString(headerStyle.Render("Dashboard"))
	b.WriteString("\n\n")

	// Stats
	totalTasks := len(s.backlog.AllTasks)
	doneCount := 0
	for _, t := range s.backlog.AllTasks {
		if t.Status == model.StatusDone {
			doneCount++
		}
	}
	sprintCount := len(s.backlog.Current.Sprints)
	epicCount := len(s.backlog.AllEpics)

	dim := lipgloss.NewStyle().Foreground(colorTextDim)
	bright := lipgloss.NewStyle().Foreground(colorTextBright).Bold(true)
	stats := []string{
		dim.Render("Tasks") + ": " + bright.Render(fmt.Sprintf("%d", totalTasks)),
		dim.Render("Done") + ": " + bright.Render(fmt.Sprintf("%d", doneCount)),
		dim.Render("Sprints") + ": " + bright.Render(fmt.Sprintf("%d", sprintCount)),
		dim.Render("Epics") + ": " + bright.Render(fmt.Sprintf("%d", epicCount)),
	}
	b.WriteString(lipgloss.NewStyle().Padding(0, 2).Render(strings.Join(stats, "    ")))
	b.WriteString("\n\n")

	// Sprint card
	if len(s.backlog.Current.Sprints) > 0 {
		sp := s.backlog.Current.Sprints[0]
		sp.Tasks = s.backlog.TasksBySprint(sp.Name)
		total := sp.TotalPoints()
		done := sp.CompletedPoints()
		var pct float64
		if total > 0 {
			pct = float64(done) / float64(total)
		}

		sprintCard := lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(colorBorder).
			Background(colorSurface).
			Padding(0, 1).
			Width(contentWidth)

		var sprintContent strings.Builder
		sprintContent.WriteString(
			lipgloss.NewStyle().Foreground(colorPrimary).Bold(true).Render(sp.Name))
		if sp.Goal != "" {
			sprintContent.WriteString(lipgloss.NewStyle().Foreground(colorTextDim).Render(" — " + sp.Goal))
		}
		sprintContent.WriteString("\n")

		bar := s.prog.ViewAs(pct)
		sprintContent.WriteString(fmt.Sprintf("%s  %d/%d (%d%%)", bar, done, total, int(pct*100)))

		b.WriteString(sprintCard.Render(sprintContent.String()))
		b.WriteString("\n\n")
	}

	// Health
	report := s.backlog.CheckHealth()
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

	healthColor := colorSuccess
	if errCount > 0 {
		healthColor = colorError
	} else if warnCount > 0 {
		healthColor = colorWarning
	}
	b.WriteString(lipgloss.NewStyle().Foreground(healthColor).Render("● Health"))
	if errCount > 0 || warnCount > 0 {
		parts := []string{}
		if errCount > 0 {
			parts = append(parts, fmt.Sprintf("%d errors", errCount))
		}
		if warnCount > 0 {
			parts = append(parts, fmt.Sprintf("%d warnings", warnCount))
		}
		b.WriteString(lipgloss.NewStyle().Foreground(colorTextDim).Render(
			"  (" + strings.Join(parts, ", ") + ")"))
	}
	b.WriteString("\n")

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
		b.WriteString(lipgloss.NewStyle().Padding(0, 2).Foreground(sevColor).Render(msg))
		b.WriteString("\n")
	}

	b.WriteString("\n")

	// Project list
	projects := []*model.ProjectSnapshot{}
	if s.backlog.Current != nil {
		projects = append(projects, s.backlog.Current)
	}
	projects = append(projects, s.backlog.Externals...)
	for i, snap := range projects {
		counts := s.backlog.StatusCounts(snap.Config.ProjectID)
		total := 0
		for _, c := range counts {
			total += c
		}
		done := counts[model.StatusDone]
		inProg := counts[model.StatusInProgress]

		line := fmt.Sprintf("%s    %d tasks    %d done    %d in-progress",
			snap.Config.ProjectID, total, done, inProg)
		if i > 0 {
			line += " " + ExternalBadge()
		}
		b.WriteString(lipgloss.NewStyle().Foreground(colorTextDim).Render(line))
		b.WriteString("\n")
	}

	b.WriteString("\n")

	// Viewport
	content := lipgloss.NewStyle().Padding(0, 2).Render(b.String())
	w := s.width - 2
	if w < 40 {
		w = 80
	}
	viewportH := s.height - 5
	if viewportH < 10 {
		viewportH = 10
	}
	if !s.ready {
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

func (s *dashboardModel) footerHint() string {
	return " 2:tasks | 3:sprints | ::cmd | ?:help"
}

func (s *dashboardModel) footerPos() string { return "" }

func (s *dashboardModel) refresh() {
	s.ready = false
}
