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
	height   int
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
		s.height = msg.Height
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
		content := lipgloss.NewStyle().Padding(0, 2).Render(b.String())
		w := s.width - 4
		if w < 40 {
			w = 80
		}
		viewportH := s.height - 5
		if viewportH < 10 {
			viewportH = 10
		}
		s.viewport = viewport.New(w, viewportH)
		s.viewport.SetContent(content)
		s.ready = true
		return s.viewport.View()
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
	return s.viewport.View()
}

func (s *epicTreeModel) printEpicNode(b *strings.Builder, ep *model.Epic, depth int) {
	indent := strings.Repeat(" ", depth)

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

	barW := 10
	fill := 0
	if totalSP > 0 {
		fill = doneSP * barW / totalSP
		if fill > barW {
			fill = barW
		}
	}
	bar := lipgloss.NewStyle().Foreground(colorSuccess).Render(strings.Repeat("█", fill) + strings.Repeat("░", barW-fill))

	glyph := statusDot(string(ep.Status))
	sp := ""
	if totalSP > 0 {
		sp = fmt.Sprintf(" %d/%dsp", doneSP, totalSP)
	}
	b.WriteString(lipgloss.NewStyle().Foreground(colorPrimary).Bold(true).Render(
		fmt.Sprintf("%s%s %s  %s%s", indent, glyph, ep.ID, bar, sp)))
	b.WriteString("\n")

	for _, t := range ep.Children {
		s.printTaskNode(b, t, depth+1)
	}
	b.WriteString("\n")
}

func (s *epicTreeModel) printTaskNode(b *strings.Builder, t *model.Task, depth int) {
	indent := strings.Repeat(" ", depth)
	prefix := " └ "
	if depth > 1 {
		prefix = "  \u2022 "
	}

	sp := ""
	if t.StoryPoints != nil {
		sp = fmt.Sprintf(" %dsp", *t.StoryPoints)
	}

	g := statusDot(string(t.Status))
	var typeColor lipgloss.Color
	switch t.Type {
	case model.TypeStory:
		typeColor = colorSecondary
	case model.TypeBug:
		typeColor = colorError
	default:
		typeColor = colorText
	}
	barW := 4
	fill := 0
	if t.Status == model.StatusDone {
		fill = barW
	} else if t.Status == model.StatusInProgress || t.Status == model.StatusReview {
		fill = barW / 2
	}
	bar := lipgloss.NewStyle().Foreground(colorSuccess).Render(strings.Repeat("█", fill) + strings.Repeat("░", barW-fill))

	b.WriteString(lipgloss.NewStyle().Foreground(typeColor).Render(
		fmt.Sprintf("%s%s%s %s %s%s", indent, prefix, g, t.ID, bar, sp)))
	b.WriteString("\n")
}
