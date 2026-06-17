package app

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hantsaniala/backlog/model"
)

type dashboardModel struct {
	backlog *model.Backlog
	prog    progress.Model
	width   int
	height  int
}

func newScreenDashboard(b *model.Backlog) *dashboardModel {
	return &dashboardModel{
		backlog: b,
		prog:    newProgressBar(40),
	}
}

func (s *dashboardModel) Init() tea.Cmd { return nil }

func (s *dashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.width = msg.Width
		s.height = msg.Height
		contentWidth := s.width - 2
		barW := contentWidth - 8
		if barW < 10 {
			barW = 10
		}
		s.prog = newProgressBar(barW)
	case reloadMsg:
		s.refresh()
	}
	return s, nil
}

func (s *dashboardModel) View() string {
	var b strings.Builder
	contentWidth := s.width - 2
	if contentWidth < 40 {
		contentWidth = 80
	}

	dim := lipgloss.NewStyle().Foreground(colorTextDim)
	bright := lipgloss.NewStyle().Foreground(colorTextBright).Bold(true)

	b.WriteString(headerStyle.Render("Dashboard"))
	b.WriteString("\n\n")

	// Stats line
	totalTasks := len(s.backlog.AllTasks)
	doneCount := 0
	wipCount := 0
	for _, t := range s.backlog.AllTasks {
		if t.Status == model.StatusDone {
			doneCount++
		}
		if t.Status == model.StatusInProgress {
			wipCount++
		}
	}
	sprintCount := len(s.backlog.Current.Sprints)
	epicCount := len(s.backlog.AllEpics)

	stats := []string{
		dim.Render("Tasks") + ": " + bright.Render(fmt.Sprintf("%d", totalTasks)),
		dim.Render("Done") + ": " + bright.Render(fmt.Sprintf("%d", doneCount)),
		dim.Render("WIP") + ": " + bright.Render(fmt.Sprintf("%d", wipCount)),
		dim.Render("Sprints") + ": " + bright.Render(fmt.Sprintf("%d", sprintCount)),
		dim.Render("Epics") + ": " + bright.Render(fmt.Sprintf("%d", epicCount)),
	}
	b.WriteString(lipgloss.NewStyle().Padding(0, 2).Render(strings.Join(stats, "    ")))
	b.WriteString("\n\n")

	// Status breakdown
	statusColors := map[model.Status]lipgloss.Color{
		model.StatusTodo:       colorInfo,
		model.StatusInProgress: colorWarning,
		model.StatusReview:     colorPrimary,
		model.StatusOnHold:     colorTextDim,
		model.StatusDone:       colorSuccess,
		model.StatusCancelled:  colorError,
	}
	statusOrder := []model.Status{
		model.StatusTodo, model.StatusInProgress, model.StatusReview,
		model.StatusOnHold, model.StatusDone, model.StatusCancelled,
	}
	counts := s.backlog.StatusCounts("")
	var statusParts []string
	for _, st := range statusOrder {
		c := statusColors[st]
		label := lipgloss.NewStyle().Foreground(c).Render(string(st))
		statusParts = append(statusParts, fmt.Sprintf("%s: %d", label, counts[st]))
	}
	b.WriteString(lipgloss.NewStyle().Padding(0, 2).Render(strings.Join(statusParts, "    ")))
	b.WriteString("\n\n")

	// Priority breakdown
	prioColors := map[model.Priority]lipgloss.Color{
		model.PriorityCritical: colorError,
		model.PriorityHigh:     colorWarning,
		model.PriorityMedium:   colorInfo,
		model.PriorityLow:      colorTextDim,
	}
	prioOrder := []model.Priority{model.PriorityCritical, model.PriorityHigh, model.PriorityMedium, model.PriorityLow}
	var prioParts []string
	for _, p := range prioOrder {
		count := 0
		for _, t := range s.backlog.AllTasks {
			if t.Priority == p {
				count++
			}
		}
		colored := lipgloss.NewStyle().Foreground(prioColors[p]).Render(string(p))
		prioParts = append(prioParts, fmt.Sprintf("%s: %d", colored, count))
	}
	b.WriteString(lipgloss.NewStyle().Padding(0, 2).Render(strings.Join(prioParts, "    ")))
	b.WriteString("\n\n")

	// Type breakdown
	typeOrder := []model.TaskType{model.TypeBug, model.TypeStory, model.TypeTask, model.TypeSpike, model.TypeChore}
	typeColors := map[model.TaskType]lipgloss.Color{
		model.TypeBug:   colorError,
		model.TypeStory: colorPrimary,
		model.TypeTask:  colorText,
		model.TypeSpike: colorWarning,
		model.TypeChore: colorTextDim,
	}
	typeGlyphs := map[model.TaskType]string{
		model.TypeBug:   "✖",
		model.TypeStory: "▶",
		model.TypeTask:  "●",
		model.TypeSpike: "▲",
		model.TypeChore: "○",
	}
	var typeParts []string
	for _, tp := range typeOrder {
		count := 0
		for _, t := range s.backlog.AllTasks {
			if t.Type == tp {
				count++
			}
		}
		g := typeGlyphs[tp]
		c := typeColors[tp]
		label := lipgloss.NewStyle().Foreground(c).Render(fmt.Sprintf("%s %s", g, string(tp)))
		typeParts = append(typeParts, fmt.Sprintf("%s: %d", label, count))
	}
	b.WriteString(lipgloss.NewStyle().Padding(0, 2).Render(strings.Join(typeParts, "    ")))
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
		sprintContent.WriteString(fmt.Sprintf("%s  %d/%d (%.0f%%)", bar, done, total, pct*100))

		b.WriteString(sprintCard.Render(sprintContent.String()))
		b.WriteString("\n\n")
	}

	// Health
	report := s.backlog.CheckHealth()
	errCount := 0
	warnCount := 0
	for _, issue := range report.Issues {
		switch issue.Severity {
		case "error":
			errCount++
		case "warning":
			warnCount++
		}
	}

	if errCount == 0 && warnCount == 0 {
		b.WriteString(lipgloss.NewStyle().Foreground(colorSuccess).Render("● Healthy"))
	} else {
		parts := []string{}
		if errCount > 0 {
			parts = append(parts, lipgloss.NewStyle().Foreground(colorError).Render(fmt.Sprintf("%d errors", errCount)))
		}
		if warnCount > 0 {
			parts = append(parts, lipgloss.NewStyle().Foreground(colorWarning).Render(fmt.Sprintf("%d warnings", warnCount)))
		}
		healthLine := lipgloss.NewStyle().Foreground(colorError).Render("● ") + strings.Join(parts, "  ")
		b.WriteString(healthLine)
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
				continue
			}
			msg := fmt.Sprintf("%s %s", icon, issue.Message)
			if issue.TaskID != "" {
				msg += fmt.Sprintf(" [%s]", issue.TaskID)
			}
			b.WriteString(lipgloss.NewStyle().Padding(0, 2).Foreground(sevColor).Render(msg))
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")

	content := lipgloss.NewStyle().Padding(0, 2).Render(b.String())
	return content
}

func (s *dashboardModel) footerHint() string {
	return " 2:tasks | 3:sprints | ::cmd | ?:help"
}

func (s *dashboardModel) footerPos() string { return "" }

func (s *dashboardModel) refresh() {}
