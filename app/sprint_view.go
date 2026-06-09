package app

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hantsaniala/backlog/model"
)

type sprintViewModel struct {
	backlog    *model.Backlog
	viewport   viewport.Model
	ready      bool
	sprintIdx  int
	width      int
	height     int
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
			case key.Matches(msg, Keys.Left):
				if s.sprintIdx > 0 {
					s.sprintIdx--
					s.ready = false
				}
			case key.Matches(msg, Keys.Right):
				if s.sprintIdx < len(s.backlog.Current.Sprints)-1 {
					s.sprintIdx++
					s.ready = false
				}
			}
		}
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

func (s *sprintViewModel) View() string {
	var b strings.Builder
	b.WriteString(headerStyle.Render("  4 Sprints"))
	b.WriteString("\n\n")

	sprints := s.backlog.Current.Sprints
	if len(sprints) == 0 {
		b.WriteString(lipgloss.NewStyle().Foreground(colorTextDim).Padding(0, 2).Render("No sprints found"))
		content := lipgloss.NewStyle().Padding(0, 2).Render(b.String())
		w := s.width - 4
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
		}
		return s.viewport.View()
	}

	for i, sp := range sprints {
		sp.Tasks = s.backlog.TasksBySprint(sp.Name)
		total := sp.TotalPoints()
		done := sp.CompletedPoints()
		label := sp.Name
		if total > 0 {
			label = fmt.Sprintf("%s (%d/%d)", sp.Name, done, total)
		}
		if i == s.sprintIdx {
			b.WriteString(tabActiveStyle.Render(fmt.Sprintf(" %s ", label)))
		} else {
			b.WriteString(tabInactiveStyle.Render(fmt.Sprintf(" %s ", label)))
		}
	}
	b.WriteString("\n\n")

	current := sprints[s.sprintIdx]
	current.Tasks = s.backlog.TasksBySprint(current.Name)
	total := current.TotalPoints()
	completed := current.CompletedPoints()

	header := fmt.Sprintf(" %s", current.Name)
	if current.Goal != "" {
		header += fmt.Sprintf(" — %s", current.Goal)
	}
	b.WriteString(lipgloss.NewStyle().Foreground(colorPrimary).Bold(true).Render(header))
	b.WriteString("\n")

	barW := s.barWidth()
	b.WriteString(lipgloss.NewStyle().Padding(0, 1).Render(fmt.Sprintf(" %s", progressBar(completed, total, barW))))
	pct := 0.0
	if total > 0 {
		pct = float64(completed) / float64(total) * 100
	}
	b.WriteString(fmt.Sprintf(" (%.0f%%)", pct))
	b.WriteString("\n\n")

	statusGroups := []model.Status{
		model.StatusInProgress, model.StatusReview,
		model.StatusTodo, model.StatusOnHold,
		model.StatusDone, model.StatusCancelled,
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
		g := statusDot(string(st))
		b.WriteString(lipgloss.NewStyle().Padding(0, 1).Render(fmt.Sprintf("%s %s (%d)", StatusBadge(string(st)), g, len(groupTasks))))
		b.WriteString("\n")
		for _, t := range groupTasks {
			spStr := ""
			if t.StoryPoints != nil {
				spStr = fmt.Sprintf(" %dsp", *t.StoryPoints)
			}
			glyph := statusDot(string(t.Status))
			title := t.ID
			if t.Summary != "" {
				title = t.Summary
			}
			b.WriteString(lipgloss.NewStyle().Padding(0, 2).Foreground(colorText).Render(
				fmt.Sprintf(" %s  %s%s", glyph, title, spStr)))
			b.WriteString("\n")
		}
		b.WriteString("\n")
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

func (s *sprintViewModel) barWidth() int {
	w := s.width - 16
	if w < 10 {
		return 10
	}
	if w > 40 {
		return 40
	}
	return w
}
