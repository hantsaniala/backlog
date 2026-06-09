package app

import (
	"fmt"
	"sort"
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

	b.WriteString(headerStyle.Render("  Dashboard"))
	b.WriteString("\n\n")

	s.printProjectCard(&b, s.backlog.Current, false)

	for _, ext := range s.backlog.Externals {
		s.printProjectCard(&b, ext, true)
	}

	health := s.backlog.CheckHealth()
	if len(health.Issues) > 0 {
		b.WriteString(lipgloss.NewStyle().
			Foreground(colorWarning).
			Bold(true).
			Render("\n  \u26A0 Backlog Health Issues"))
		b.WriteString("\n")
		for _, issue := range health.Issues {
			icon := "  \u2022"
			style := lipgloss.NewStyle().Foreground(colorText)
			if issue.Severity == "error" {
				icon = "  \u2716"
				style = lipgloss.NewStyle().Foreground(colorError)
			} else if issue.Severity == "warning" {
				icon = "  \u26A0"
				style = lipgloss.NewStyle().Foreground(colorWarning)
			}
			taskRef := ""
			if issue.TaskID != "" {
				taskRef = " [" + issue.TaskID + "]"
			}
			b.WriteString(style.Render(fmt.Sprintf("%s %s%s", icon, issue.Message, taskRef)))
			b.WriteString("\n")
		}
	}

	b.WriteString("\n  Watching for file changes... live reload active")

	content := lipgloss.NewStyle().Padding(0, 2).Render(b.String())

	if !s.ready || s.width > 0 {
		w := s.width - 4
		if w < 40 {
			w = 80
		}
		s.viewport = viewport.New(w, 20)
		s.viewport.SetContent(content)
		s.ready = true
	} else {
		s.viewport.SetContent(content)
	}

	return s.viewport.View()
}

func (s *dashboardModel) printProjectCard(b *strings.Builder, snap *model.ProjectSnapshot, external bool) {
	counts := s.backlog.StatusCounts(snap.Config.ProjectID)

	title := fmt.Sprintf("  %s (%s)", snap.Config.ProjectID, snap.Config.Name)
	if external {
		title += " " + ExternalBadge()
	}
	b.WriteString(lipgloss.NewStyle().
		Foreground(colorPrimary).
		Bold(true).
		Render(title))
	b.WriteString("\n")

	total := 0
	for _, c := range counts {
		total += c
	}
	statusLine := fmt.Sprintf("  Total: %d  |  todo: %d  prog: %d  review: %d  hold: %d  done: %d  cancelled: %d",
		total, counts["todo"], counts["in-progress"], counts["review"], counts["on-hold"], counts["done"], counts["cancelled"])
	b.WriteString(lipgloss.NewStyle().Foreground(colorTextDim).Render(statusLine))
	b.WriteString("\n\n")
}

func (s *dashboardModel) refresh() {
	s.ready = false
}

type summaryItem struct {
	label string
	value int
}

func sortSummary(items []summaryItem) {
	sort.Slice(items, func(i, j int) bool {
		return items[i].value > items[j].value
	})
}
