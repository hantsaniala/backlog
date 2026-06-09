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
}

func newScreenEpicTree(b *model.Backlog) *epicTreeModel {
	return &epicTreeModel{backlog: b}
}

func (s *epicTreeModel) Init() tea.Cmd { return nil }

func (s *epicTreeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg.(type) {
	case reloadMsg:
		s.ready = false
	}
	s.viewport, cmd = s.viewport.Update(msg)
	return s, cmd
}

func (s *epicTreeModel) View() string {
	var b strings.Builder
	b.WriteString(headerStyle.Render("  Epic Tree"))
	b.WriteString("\n\n")

	if len(s.backlog.AllEpics) == 0 {
		b.WriteString(lipgloss.NewStyle().Foreground(colorTextDim).Padding(0, 2).Render("No epics found"))
		content := lipgloss.NewStyle().Padding(0, 2).Render(b.String())
		if !s.ready {
			s.viewport = viewport.New(80, 20)
			s.viewport.SetContent(content)
			s.ready = true
		}
		return s.viewport.View()
	}

	for _, ep := range s.backlog.AllEpics {
		s.printEpicNode(&b, ep, 0)
	}

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

func (s *epicTreeModel) printEpicNode(b *strings.Builder, ep *model.Epic, depth int) {
	indent := strings.Repeat("  ", depth)

	statusStr := StatusBadge(string(ep.Status))
	storyCount := ep.StoryCount()
	epicLine := fmt.Sprintf("%s%s %s (%d stories)", indent, statusStr, ep.ID, storyCount)
	b.WriteString(lipgloss.NewStyle().Foreground(colorPrimary).Bold(true).Render(epicLine))
	if ep.ProjectID != s.backlog.Current.Config.ProjectID {
		b.WriteString(" ")
		b.WriteString(ExternalBadge())
	}
	b.WriteString("\n")

	for _, t := range ep.Children {
		s.printTaskNode(b, t, depth+1)
	}

	b.WriteString("\n")
}

func (s *epicTreeModel) printTaskNode(b *strings.Builder, t *model.Task, depth int) {
	indent := strings.Repeat("  ", depth)
	prefix := "\u2514 "
	if depth > 1 {
		prefix = "\u2022 "
	}

	sp := ""
	if t.StoryPoints != nil {
		sp = fmt.Sprintf(" [%dsp]", *t.StoryPoints)
	}

	line := fmt.Sprintf("%s%s%s %s%s", indent, prefix, StatusBadge(string(t.Status)), t.ID, sp)

	var typeColor lipgloss.Color
	switch t.Type {
	case model.TypeStory:
		typeColor = colorSecondary
	case model.TypeBug:
		typeColor = colorError
	default:
		typeColor = colorTextDim
	}
	b.WriteString(lipgloss.NewStyle().Foreground(typeColor).Render(line))
	b.WriteString("\n")
}
