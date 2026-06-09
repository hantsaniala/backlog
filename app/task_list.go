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

type flatRow struct {
	task   *model.Task
	prefix string
	level  int
}

type viewMode int

const (
	modeTree viewMode = iota
	modeDetail
	modeSearch
	modeStatusPopup
)

type taskListModel struct {
	backlog    *model.Backlog
	width      int
	height     int
	mode       viewMode

	// Tree
	rows        []flatRow
	visibleRows []flatRow
	cursor      int
	expanded    map[string]bool

	// Tree filter
	filterText string
	filterOn   bool

	// Detail
	detailView *detailView

	// Search modal
	searchModal   *searchModalModel
	searchRunning bool

	// Status popup
	statusPopup   *statusPopupModel
	statusRunning bool

	// Viewport for tree scrolling
	treeViewport viewport.Model
	treeReady    bool
}

func newScreenTaskList(b *model.Backlog) *taskListModel {
	m := &taskListModel{
		backlog:  b,
		expanded: make(map[string]bool),
	}
	m.rebuild()
	if len(m.visibleRows) > 0 {
		m.cursor = 0
	}
	return m
}

func (s *taskListModel) Init() tea.Cmd { return nil }

func (s *taskListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Sub-model routing
	if s.mode == modeSearch && s.searchRunning {
		updated, cmd := s.searchModal.Update(msg)
		s.searchModal = updated.(*searchModalModel)
		if result, ok := <-catchMsg(cmd); ok {
			if nav, ok := result.(SearchNavigateMsg); ok {
				task := nav.Task
				s.enterDetail(task)
				s.searchRunning = false
				s.mode = modeDetail
				return s, nil
			}
			// nil = cancel
			s.searchRunning = false
			s.mode = modeTree
			return s, nil
		}
		return s, cmd
	}

	if s.mode == modeStatusPopup && s.statusRunning {
		updated, cmd := s.statusPopup.Update(msg)
		s.statusPopup = updated.(*statusPopupModel)
		if result, ok := <-catchMsg(cmd); ok {
			if update, ok := result.(StatusUpdateMsg); ok {
				if s.detailView != nil {
					s.detailView.task.Status = update.NewStatus
				}
			}
			s.statusRunning = false
			s.mode = modeDetail
			return s, nil
		}
		return s, cmd
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch s.mode {
		case modeTree:
			return s.handleTreeKey(msg)
		case modeDetail:
			return s.handleDetailKey(msg)
		}

	case tea.WindowSizeMsg:
		s.width = msg.Width
		s.height = msg.Height
		s.treeReady = false
		if s.detailView != nil {
			s.detailView.width = msg.Width
			s.detailView.height = msg.Height
		}
		if s.searchModal != nil {
			s.searchModal.width = msg.Width
			s.searchModal.height = msg.Height
		}

	case reloadMsg:
		s.rebuild()
	}

	return s, nil
}

// catchMsg is a helper to synchronously get a msg from a cmd
func catchMsg(cmd tea.Cmd) <-chan tea.Msg {
	ch := make(chan tea.Msg, 1)
	if cmd == nil {
		close(ch)
		return ch
	}
	go func() {
		ch <- cmd()
		close(ch)
	}()
	return ch
}

func (s *taskListModel) handleTreeKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, Keys.Filter):
		s.mode = modeSearch
		s.searchModal = newSearchModal(s.backlog, s.width, s.height)
		s.searchRunning = true
		return s, s.searchModal.Init()

	case key.Matches(msg, Keys.Quit):
		return s, tea.Quit

	case key.Matches(msg, Keys.Up), key.Matches(msg, Keys.Down):
		if key.Matches(msg, Keys.Up) && s.cursor > 0 {
			s.cursor--
		} else if key.Matches(msg, Keys.Down) && s.cursor < len(s.visibleRows)-1 {
			s.cursor++
		}
		return s, nil

	case key.Matches(msg, Keys.Left):
		r := s.visibleRows[s.cursor]
		if r.task.Type == model.TypeEpic {
			s.expanded[r.task.ID] = false
			s.buildVisibleRows()
		}
		return s, nil

	case key.Matches(msg, Keys.Right):
		r := s.visibleRows[s.cursor]
		if r.task.Type == model.TypeEpic {
			s.expanded[r.task.ID] = true
			s.buildVisibleRows()
		}
		return s, nil

	case key.Matches(msg, Keys.Expand):
		r := s.visibleRows[s.cursor]
		if r.task.Type == model.TypeEpic {
			s.expanded[r.task.ID] = !s.expanded[r.task.ID]
			s.buildVisibleRows()
		} else {
			s.toggleDone(r.task)
		}
		return s, nil

	case key.Matches(msg, Keys.Enter):
		if len(s.visibleRows) > 0 && s.cursor >= 0 && s.cursor < len(s.visibleRows) {
			task := s.visibleRows[s.cursor].task
			s.enterDetail(task)
		}
		return s, nil
	}
	return s, nil
}

func (s *taskListModel) handleDetailKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, Keys.Back), key.Matches(msg, Keys.Quit):
		s.mode = modeTree
		return s, nil

	case key.Matches(msg, Keys.Sort), key.Matches(msg, Keys.CycleStatus):
		s.mode = modeStatusPopup
		if s.detailView != nil {
			s.statusPopup = newStatusPopup(s.detailView.task, 30)
		}
		s.statusRunning = true
		return s, nil

	case key.Matches(msg, Keys.Reload), key.Matches(msg, Keys.Related):
		if s.detailView != nil {
			s.detailView.resolveLinks()
			s.detailView.state = detailRelated
			s.detailView.relatedCursor = 0
		}
		return s, nil

	case key.Matches(msg, Keys.Enter):
		if s.detailView != nil && s.detailView.state == detailRelated && len(s.detailView.relatedItems) > 0 {
			link := s.detailView.relatedItems[s.detailView.relatedCursor]
			if link.Task != nil {
				s.detailView.previewTask = link.Task
				s.detailView.state = detailPreview
			}
		}
		return s, nil

	case key.Matches(msg, Keys.Up):
		if s.detailView != nil && s.detailView.state == detailRelated {
			if s.detailView.relatedCursor > 0 {
				s.detailView.relatedCursor--
			}
		}
		return s, nil

	case key.Matches(msg, Keys.Down):
		if s.detailView != nil && s.detailView.state == detailRelated {
			if s.detailView.relatedCursor < len(s.detailView.relatedItems)-1 {
				s.detailView.relatedCursor++
			}
		}
		return s, nil
	}
	return s, nil
}

func (s *taskListModel) enterDetail(task *model.Task) {
	dv := newDetailView(s.backlog, task, s.width, s.height)
	dv.resolveLinks()
	s.detailView = dv
	s.mode = modeDetail
}

func (s *taskListModel) toggleDone(task *model.Task) {
	if task.Status == model.StatusDone {
		task.Status = model.StatusTodo
	} else {
		task.Status = model.StatusDone
	}
	s.rebuild()
	// Sync detail if open
	if s.detailView != nil && s.detailView.task.ID == task.ID {
		s.detailView.task.Status = task.Status
	}
}

func (s *taskListModel) View() string {
	switch s.mode {
	case modeDetail:
		if s.detailView != nil {
			v := s.detailView.View()
			if s.mode == modeStatusPopup && s.statusRunning && s.statusPopup != nil {
				popup := s.statusPopup.View()
				v = lipgloss.Place(s.width, s.height, lipgloss.Center, lipgloss.Center, popup)
			}
			return v
		}
		return ""
	case modeSearch:
		if s.searchRunning && s.searchModal != nil {
			return s.searchModal.View()
		}
		return ""
	default:
		return s.renderTreeFull()
	}
}

func (s *taskListModel) renderTreeFull() string {
	var b strings.Builder

	if s.filterOn {
		b.WriteString(lipgloss.NewStyle().Padding(0, 2).Foreground(colorTextDim).Render(fmt.Sprintf(" Filter: %s", s.filterText)))
		b.WriteString("\n")
	}

	content := s.renderTree()
	treeH := s.height - 6
	if treeH < 5 {
		treeH = 5
	}

	if !s.treeReady {
		s.treeViewport = viewport.New(s.width-4, treeH)
		s.treeReady = true
	}
	s.treeViewport.Width = s.width - 4
	s.treeViewport.Height = treeH
	s.treeViewport.SetContent(content)

	b.WriteString(s.treeViewport.View())

	// Footer
	b.WriteString("\n")
	info := fmt.Sprintf("%d items | / search | j/k nav | l/h expand | space toggle done | enter detail", len(s.visibleRows))
	b.WriteString(lipgloss.NewStyle().Foreground(colorTextDim).Padding(0, 2).Render(info))

	return b.String()
}

func (s *taskListModel) renderTree() string {
	var b strings.Builder

	for i, r := range s.visibleRows {
		prefix := r.prefix
		indent := strings.Repeat(" ", r.level)

		// Expand indicator for epics
		expandSymbol := " "
		if r.task.Type == model.TypeEpic {
			if s.expanded[r.task.ID] {
				expandSymbol = "▼"
			} else {
				expandSymbol = "▶"
			}
		}

		g := statusDot(string(r.task.Status))
		statusStr := StatusBadge(string(r.task.Status))

		sp := ""
		if r.task.StoryPoints != nil {
			sp = lipgloss.NewStyle().Foreground(colorWarning).Render(fmt.Sprintf(" %dsp", *r.task.StoryPoints))
		}

		label := r.task.ID
		if r.task.Summary != "" {
			label = r.task.ID + "  " + r.task.Summary
		}

		line := fmt.Sprintf(" %s%s%s %s %s%s%s", indent, expandSymbol, prefix, g, label, sp, " "+statusStr)

		if i == s.cursor {
			line = leftBorderBar + focusedRowStyle.Render(line)
		} else {
			line = " " + line
		}

		b.WriteString(line)
		b.WriteString("\n")
	}

	return b.String()
}

func (s *taskListModel) rebuild() {
	filtered := s.filterTasks()
	s.buildTree(filtered)
	s.buildVisibleRows()
	// Clamp cursor
	if s.cursor >= len(s.visibleRows) && len(s.visibleRows) > 0 {
		s.cursor = len(s.visibleRows) - 1
	} else if len(s.visibleRows) == 0 {
		s.cursor = 0
	}
}

func (s *taskListModel) buildVisibleRows() {
	var out []flatRow
	var currentEpic string
	epicExpanded := true
	for _, r := range s.rows {
		if r.task.Type == model.TypeEpic {
			currentEpic = r.task.ID
			_, epicExpanded = s.expanded[currentEpic]
			if !epicExpanded {
				// Only show the epic row itself
				out = append(out, r)
			} else {
				out = append(out, r)
			}
		} else {
			if epicExpanded || currentEpic == "" {
				out = append(out, r)
			}
		}
	}
	s.visibleRows = out
}

func (s *taskListModel) buildTree(tasks []*model.Task) {
	epics := s.backlog.AllEpics

	var rows []flatRow
	seen := make(map[string]bool)

	// Epics first
	for _, ep := range epics {
		if ep.ProjectID != s.backlog.Current.Config.ProjectID {
			continue
		}
		epTask := epicToTask(ep)
		rows = append(rows, flatRow{task: epTask, prefix: "", level: 0})

		var children []*model.Task
		for _, t := range tasks {
			if t.Epic == ep.ID && t.ID != ep.ID {
				children = append(children, t)
			}
		}
		for _, child := range children {
			if child.Type == model.TypeStory {
				seen[child.ID] = true
				rows = append(rows, flatRow{task: child, prefix: " └", level: 1})
				for _, t := range children {
					if t.Parent == child.ID && t.ID != child.ID {
						seen[t.ID] = true
						rows = append(rows, flatRow{task: t, prefix: "   •", level: 2})
					}
				}
			}
		}
		for _, child := range children {
			if !seen[child.ID] && child.Type == model.TypeTask {
				seen[child.ID] = true
				rows = append(rows, flatRow{task: child, prefix: " └", level: 1})
			}
		}
	}

	// Stories without epic
	for _, t := range tasks {
		if !seen[t.ID] && t.Type == model.TypeStory {
			seen[t.ID] = true
			rows = append(rows, flatRow{task: t, prefix: " └", level: 1})
			for _, child := range tasks {
				if !seen[child.ID] && child.Parent == t.ID && child.ID != t.ID {
					seen[child.ID] = true
					rows = append(rows, flatRow{task: child, prefix: "   •", level: 2})
				}
			}
		}
	}

	// Remaining items
	for _, t := range tasks {
		if !seen[t.ID] {
			rows = append(rows, flatRow{task: t, prefix: "", level: 0})
		}
	}

	s.rows = rows
}

func (s *taskListModel) filterTasks() []*model.Task {
	tasks := s.backlog.AllTasks
	if s.filterText == "" {
		return tasks
	}
	lower := strings.ToLower(s.filterText)
	var filtered []*model.Task
	for _, t := range tasks {
		if strings.Contains(strings.ToLower(t.ID), lower) ||
			strings.Contains(strings.ToLower(t.Summary), lower) ||
			strings.Contains(strings.ToLower(t.Assignee), lower) {
			filtered = append(filtered, t)
		}
	}
	return filtered
}

func epicToTask(ep *model.Epic) *model.Task {
	return &model.Task{
		ID:        ep.ID,
		Type:      model.TypeEpic,
		Status:    ep.Status,
		Priority:  ep.Priority,
		Labels:    ep.Labels,
		Created:   ep.Created,
		Updated:   ep.Updated,
		Summary:   ep.Name,
		Body:      ep.Body,
		ProjectID: ep.ProjectID,
	}
}
