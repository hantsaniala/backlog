package app

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
	"github.com/hantsaniala/backlog/model"
)

type sprintViewModel struct {
	backlog    *model.Backlog
	viewport   viewport.Model
	ready      bool
	sprintIdx  int
}

func newScreenSprintView(b *model.Backlog) *sprintViewModel {
	return &sprintViewModel{backlog: b}
}

func (s *sprintViewModel) Init() tea.Cmd { return nil }

func (s *sprintViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if len(s.backlog.Current.Sprints) > 0 {
			switch {
			case 			key.Matches(msg, key.NewBinding(key.WithKeys("left", "h"))):
				if s.sprintIdx > 0 {
					s.sprintIdx--
					s.ready = false
				}
			case key.Matches(msg, key.NewBinding(key.WithKeys("right", "l"))):
				if s.sprintIdx < len(s.backlog.Current.Sprints)-1 {
					s.sprintIdx++
					s.ready = false
				}
			}
		}
	case reloadMsg:
		s.ready = false
	}
	s.viewport, cmd = s.viewport.Update(msg)
	return s, cmd
}

func (s *sprintViewModel) View() string {
	var b strings.Builder
	b.WriteString(headerStyle.Render("  Sprint View"))
	b.WriteString("\n\n")

	sprints := s.backlog.Current.Sprints
	if len(sprints) == 0 {
		b.WriteString(lipgloss.NewStyle().Foreground(colorTextDim).Padding(0, 2).Render("No sprints found"))
		content := lipgloss.NewStyle().Padding(0, 2).Render(b.String())
		if !s.ready {
			s.viewport = viewport.New(80, 20)
			s.viewport.SetContent(content)
			s.ready = true
		}
		return s.viewport.View()
	}

	for i, sp := range sprints {
		if i == s.sprintIdx {
			b.WriteString(tabActiveStyle.Render(sp.Name))
		} else {
			b.WriteString(tabInactiveStyle.Render(sp.Name))
		}
	}
	b.WriteString("\n\n")

	current := sprints[s.sprintIdx]
	current.Tasks = s.backlog.TasksBySprint(current.Name)

	b.WriteString(lipgloss.NewStyle().Foreground(colorPrimary).Bold(true).Render(fmt.Sprintf("  %s", current.Name)))
	if current.Goal != "" {
		b.WriteString(lipgloss.NewStyle().Foreground(colorTextDim).Render(fmt.Sprintf(" - %s", current.Goal)))
	}
	b.WriteString("\n")
	if current.Start != "" && current.End != "" {
		b.WriteString(lipgloss.NewStyle().Foreground(colorTextDim).Render(fmt.Sprintf("  %s to %s", current.Start, current.End)))
		b.WriteString("\n")
	}

	total := current.TotalPoints()
	completed := current.CompletedPoints()
	b.WriteString(lipgloss.NewStyle().Foreground(colorTextDim).Render(fmt.Sprintf("  Progress: %d/%d points completed", completed, total)))
	b.WriteString("\n\n")

	statusGroups := []model.Status{
		model.StatusTodo, model.StatusInProgress, model.StatusReview,
		model.StatusOnHold, model.StatusDone, model.StatusCancelled,
	}
	for _, st := range statusGroups {
		var groupTasks []*model.Task
		for _, t := range current.Tasks {
			if t.Status == st {
				groupTasks = append(groupTasks, t)
			}
		}
		if len(groupTasks) == 0 {
			continue
		}
		b.WriteString(fmt.Sprintf("  %s (%d)\n", StatusBadge(string(st)), len(groupTasks)))
		for _, t := range groupTasks {
			sp := ""
			if t.StoryPoints != nil {
				sp = fmt.Sprintf(" [%dsp]", *t.StoryPoints)
			}
			b.WriteString(lipgloss.NewStyle().Foreground(colorText).Render(fmt.Sprintf("    \u2022 %s%s", t.ID, sp)))
			b.WriteString("\n")
		}
		b.WriteString("\n")
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
