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

type sprintPane int

const (
	paneLeft sprintPane = iota
	paneRight
)

type sprintViewModel struct {
	backlog   *model.Backlog
	width     int
	height    int
	sprintIdx int

	focus        sprintPane
	leftCursor   int
	rightCursor  int

	// Left items: tasks not in any sprint
	leftItems []*model.Task
	// Right items: tasks in current sprint
	rightItems []*model.Task

	// Viewports
	leftViewport  viewport.Model
	rightViewport viewport.Model
	leftReady     bool
	rightReady    bool
}

func (s *sprintViewModel) footerHint() string {
	return " h/l:move | j/k:nav | >:to sprint | <:to backlog | e:status"
}

func (s *sprintViewModel) footerPos() string {
	if s.focus == paneLeft {
		return fmt.Sprintf("%d/%d L", s.leftCursor+1, len(s.leftItems))
	}
	return fmt.Sprintf("%d/%d R", s.rightCursor+1, len(s.rightItems))
}

func newScreenSprintView(b *model.Backlog) *sprintViewModel {
	m := &sprintViewModel{backlog: b}
	m.refresh()
	return m
}

func (s *sprintViewModel) Init() tea.Cmd { return nil }

func (s *sprintViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return s.handleKey(msg)
	case tea.WindowSizeMsg:
		s.width = msg.Width
		s.height = msg.Height
		s.leftReady = false
		s.rightReady = false
	case reloadMsg:
		s.refresh()
	}
	return s, nil
}

func (s *sprintViewModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, NormalKeys.Quit):
		return s, tea.Quit
	case msg.String() == "tab":
		if s.focus == paneLeft {
			s.focus = paneRight
		} else {
			s.focus = paneLeft
		}
		return s, nil
	case key.Matches(msg, NormalKeys.Filter):
		return s, nil // TODO: search modal
	case key.Matches(msg, NormalKeys.Up), key.Matches(msg, NormalKeys.Down):
		if s.focus == paneLeft {
			if key.Matches(msg, NormalKeys.Up) && s.leftCursor > 0 {
				s.leftCursor--
			} else if key.Matches(msg, NormalKeys.Down) && s.leftCursor < len(s.leftItems)-1 {
				s.leftCursor++
			}
		} else {
			if key.Matches(msg, NormalKeys.Up) && s.rightCursor > 0 {
				s.rightCursor--
			} else if key.Matches(msg, NormalKeys.Down) && s.rightCursor < len(s.rightItems)-1 {
				s.rightCursor++
			}
		}
		return s, nil
	case key.Matches(msg, NormalKeys.GotoBottom):
		if s.focus == paneLeft {
			s.leftCursor = len(s.leftItems) - 1
		} else {
			s.rightCursor = len(s.rightItems) - 1
		}
		return s, nil
	case msg.String() == "g":
		if s.focus == paneLeft {
			s.leftCursor = 0
		} else {
			s.rightCursor = 0
		}
		return s, nil
	case key.Matches(msg, NormalKeys.HalfDown):
		page := (s.height - 12) / 2
		if page < 1 {
			page = 1
		}
		if s.focus == paneLeft {
			s.leftCursor += page
			if s.leftCursor >= len(s.leftItems) {
				s.leftCursor = len(s.leftItems) - 1
			}
		} else {
			s.rightCursor += page
			if s.rightCursor >= len(s.rightItems) {
				s.rightCursor = len(s.rightItems) - 1
			}
		}
		return s, nil
	case key.Matches(msg, NormalKeys.HalfUp):
		page := (s.height - 12) / 2
		if page < 1 {
			page = 1
		}
		if s.focus == paneLeft {
			s.leftCursor -= page
			if s.leftCursor < 0 {
				s.leftCursor = 0
			}
		} else {
			s.rightCursor -= page
			if s.rightCursor < 0 {
				s.rightCursor = 0
			}
		}
		return s, nil
	case msg.String() == ">":
		if s.focus == paneLeft && s.leftCursor >= 0 && s.leftCursor < len(s.leftItems) {
			task := s.leftItems[s.leftCursor]
			task.Sprint = s.currentSprintName()
			s.refresh()
		}
		return s, nil
	case msg.String() == "<":
		if s.focus == paneRight && s.rightCursor >= 0 && s.rightCursor < len(s.rightItems) {
			task := s.rightItems[s.rightCursor]
			task.Sprint = ""
			s.refresh()
		}
		return s, nil
	case key.Matches(msg, NormalKeys.CycleStatus):
		if s.focus == paneRight && s.rightCursor >= 0 && s.rightCursor < len(s.rightItems) {
			task := s.rightItems[s.rightCursor]
			task.Status = nextStatus(task.Status)
			s.refresh()
		}
		return s, nil
	case key.Matches(msg, NormalKeys.Left):
		if s.sprintIdx > 0 {
			s.sprintIdx--
			s.refresh()
		}
		return s, nil
	case key.Matches(msg, NormalKeys.Right):
		if s.sprintIdx < len(s.backlog.Current.Sprints)-1 {
			s.sprintIdx++
			s.refresh()
		}
		return s, nil
	}
	return s, nil
}

func nextStatus(s model.Status) model.Status {
	switch s {
	case model.StatusTodo:
		return model.StatusInProgress
	case model.StatusInProgress:
		return model.StatusReview
	case model.StatusReview:
		return model.StatusDone
	case model.StatusDone:
		return model.StatusTodo
	default:
		return model.StatusTodo
	}
}

func (s *sprintViewModel) currentSprintName() string {
	if len(s.backlog.Current.Sprints) == 0 {
		return ""
	}
	return s.backlog.Current.Sprints[s.sprintIdx].Name
}

func (s *sprintViewModel) refresh() {
	s.leftItems = nil
	s.rightItems = nil

	sprintName := s.currentSprintName()

	for _, t := range s.backlog.AllTasks {
		if t.Sprint == sprintName {
			s.rightItems = append(s.rightItems, t)
		} else if t.Sprint == "" {
			s.leftItems = append(s.leftItems, t)
		}
	}
}

func (s *sprintViewModel) View() string {
	if len(s.backlog.Current.Sprints) == 0 {
		return lipgloss.NewStyle().Foreground(colorTextDim).Padding(2, 2).Render("No sprints found")
	}

	var b strings.Builder

	// Sprint selector
	sprints := s.backlog.Current.Sprints
	for i, sp := range sprints {
		label := sp.Name
		sp.Tasks = s.backlog.TasksBySprint(sp.Name)
		total := sp.TotalPoints()
		done := sp.CompletedPoints()
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

	// Sprint header
	current := s.backlog.Current.Sprints[s.sprintIdx]
	spTasks := s.backlog.TasksBySprint(current.Name)
	total := 0
	done := 0
	for _, t := range spTasks {
		if t.StoryPoints != nil {
			total += *t.StoryPoints
			if t.Status == model.StatusDone {
				done += *t.StoryPoints
			}
		}
	}

	barW := s.width - 20
	if barW < 10 {
		barW = 10
	}
	if barW > 40 {
		barW = 40
	}

	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(colorPrimary).Render(fmt.Sprintf(" %s", current.Name)))
	if current.Goal != "" {
		b.WriteString(lipgloss.NewStyle().Foreground(colorTextDim).Render(fmt.Sprintf(" — %s", current.Goal)))
	}
	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Padding(0, 1).Render(fmt.Sprintf(" %s", progressBar(done, total, barW))))
	pct := 0.0
	if total > 0 {
		pct = float64(done) / float64(total) * 100
	}
	b.WriteString(fmt.Sprintf("  (%.0f%%)", pct))
	b.WriteString("\n\n")

	// Two panels
	panelW := (s.width - 8) / 2
	if panelW < 30 {
		panelW = 30
	}
	panelH := s.height - 12
	if panelH < 5 {
		panelH = 5
	}

	// Left panel: unassigned items
	leftBorder := colorBorder
	if s.focus == paneLeft {
		leftBorder = colorPageSprint
	}
	leftContent := s.renderLeftPanel(panelW, panelH)
	leftPanel := lipgloss.NewStyle().
		Width(panelW).
		Border(lipgloss.NormalBorder()).
		BorderForeground(leftBorder).
		Padding(0, 1).
		Render(leftContent)

	// Right panel: sprint backlog
	rightBorder := colorBorder
	if s.focus == paneRight {
		rightBorder = colorPageSprint
	}
	rightContent := s.renderRightPanel(panelW, panelH)
	rightPanel := lipgloss.NewStyle().
		Width(panelW).
		Border(lipgloss.NormalBorder()).
		BorderForeground(rightBorder).
		Padding(0, 1).
		Render(rightContent)

	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, "  ", rightPanel))

	return b.String()
}

func (s *sprintViewModel) renderLeftPanel(w, h int) string {
	var b strings.Builder
	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(colorTextBright).Render(" Backlog (unassigned)"))
	b.WriteString(fmt.Sprintf("  %d items", len(s.leftItems)))
	b.WriteString("\n\n")

	for i, t := range s.leftItems {
		sel := " "
		if i == s.leftCursor && s.focus == paneLeft {
			sel = "▎"
		}
		g := statusDot(string(t.Status))
		sp := ""
		if t.StoryPoints != nil {
			sp = fmt.Sprintf(" %dsp", *t.StoryPoints)
		}
		line := fmt.Sprintf("%s%s %s  %s%s", sel, g, t.ID, t.Summary, sp)
		if i == s.leftCursor && s.focus == paneLeft {
			line = focusedRowStyle.Render(line)
		}
		b.WriteString(lipgloss.NewStyle().Foreground(colorText).Render(line))
		b.WriteString("\n")
	}

	return b.String()
}

func (s *sprintViewModel) renderRightPanel(w, h int) string {
	var b strings.Builder
	sprintName := s.currentSprintName()
	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(colorTextBright).Render(fmt.Sprintf(" Sprint: %s", sprintName)))
	b.WriteString(fmt.Sprintf("  %d items", len(s.rightItems)))
	b.WriteString("\n\n")

	statusGroups := []model.Status{
		model.StatusTodo, model.StatusInProgress, model.StatusReview,
		model.StatusOnHold, model.StatusDone, model.StatusCancelled,
	}

	idx := 0
	for _, st := range statusGroups {
		var group []*model.Task
		for _, t := range s.rightItems {
			if t.Status == st {
				group = append(group, t)
			}
		}
		if len(group) == 0 {
			continue
		}
		b.WriteString(fmt.Sprintf(" %s (%d)\n", StatusBadge(string(st)), len(group)))
		for _, t := range group {
			sel := " "
			if idx == s.rightCursor && s.focus == paneRight {
				sel = "▎"
			}
			g := statusDot(string(t.Status))
			sp := ""
			if t.StoryPoints != nil {
				sp = fmt.Sprintf(" %dsp", *t.StoryPoints)
			}
			assignee := ""
			if t.Assignee != "" {
				initials := strings.ToUpper(t.Assignee[:1])
				if len(t.Assignee) > 1 {
					initials = strings.ToUpper(t.Assignee[:1]) + t.Assignee[1:2]
				}
				assignee = fmt.Sprintf(" [%s]", initials)
			}
			line := fmt.Sprintf("%s  %s %s%s%s%s", sel, g, t.ID, assignee, sp, " "+t.Summary)
			if idx == s.rightCursor && s.focus == paneRight {
				line = focusedRowStyle.Render(line)
			}
			b.WriteString(lipgloss.NewStyle().Padding(0, 1).Foreground(colorText).Render(line))
			b.WriteString("\n")
			idx++
		}
		b.WriteString("\n")
	}

	return b.String()
}
