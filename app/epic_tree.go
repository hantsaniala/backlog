package app

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hantsaniala/backlog/model"
)

type epicTreeModel struct {
	backlog  *model.Backlog
	viewport viewport.Model
	ready    bool
	width    int
}

func newScreenEpicTree(b *model.Backlog) *epicTreeModel {
	return &epicTreeModel{backlog: b}
}

func (s *epicTreeModel) Init() tea.Cmd { return nil }

func (s *epicTreeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (s *epicTreeModel) View() string {
	var b strings.Builder
	b.WriteString(headerStyle.Render("  3 Epics"))
	b.WriteString("\n\n")

	if len(s.backlog.AllEpics) == 0 {
		b.WriteString(lipgloss.NewStyle().Foreground(colorTextDim).Padding(0, 2).Render("No epics found"))
	} else {
		for _, ep := range s.backlog.AllEpics {
			s.printEpicNode(&b, ep, 0)
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

func (s *epicTreeModel) printEpicNode(b *strings.Builder, ep *model.Epic, depth int) {
	indent := strings.Repeat("  ", depth)

	totalSP := 0
	doneSP := 0
	for _, t := range ep.Children {
		if t.StoryPoints != nil {
			totalSP += *t.StoryPoints
			if t.Status == model.StatusDone {
				doneSP += *t.StoryPoints
			}
		}
	}

	barW := 12
	fill := doneSP
	if totalSP > 0 {
		fill = doneSP * barW / totalSP
		if fill > barW {
			fill = barW
		}
	}
	bar := lipgloss.NewStyle().Foreground(colorSuccess).Render(strings.Repeat("█", fill) + strings.Repeat("░", barW-fill))

	epicLine := fmt.Sprintf("%s%s %s  %s", indent, StatusBadge(string(ep.Status)), ep.ID, bar)
	if totalSP > 0 {
		epicLine += fmt.Sprintf("  %d/%dsp", doneSP, totalSP)
	}
	b.WriteString(lipgloss.NewStyle().Foreground(colorPrimary).Bold(true).Render(epicLine))
	b.WriteString("\n")

	for _, t := range ep.Children {
		s.printTaskNode(b, t, depth+1)
	}
	b.WriteString("\n")
}

func (s *epicTreeModel) printTaskNode(b *strings.Builder, t *model.Task, depth int) {
	indent := strings.Repeat("  ", depth)
	prefix := "   └ "
	if depth > 1 {
		prefix = "    • "
	}

	sp := ""
	if t.StoryPoints != nil {
		sp = fmt.Sprintf(" [%dsp]", *t.StoryPoints)
	}

	barStatus := 0
	switch t.Status {
	case model.StatusDone:
		barStatus = 2
	case model.StatusInProgress, model.StatusReview:
		barStatus = 1
	}
	barW := 6
	fill := barStatus * barW / 2
	bar := lipgloss.NewStyle().Foreground(colorSuccess).Render(strings.Repeat("█", fill) + strings.Repeat("░", barW-fill))

	line := fmt.Sprintf("%s%s %s%s  %s%s", indent, prefix, StatusBadge(string(t.Status)), t.ID, bar, sp)

	var typeColor lipgloss.Color
	switch t.Type {
	case model.TypeStory:
		typeColor = colorSecondary
	case model.TypeBug:
		typeColor = colorError
	default:
		typeColor = colorText
	}
	b.WriteString(lipgloss.NewStyle().Foreground(typeColor).Render(line))
	b.WriteString("\n")
}
