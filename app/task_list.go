package app

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
	"github.com/hantsaniala/backlog/model"
)

type filterMode int

const (
	filterNone filterMode = iota
	filterSearch
	filterByStatus
	filterByPriority
	filterByType
	filterByProject
)

type taskListModel struct {
	backlog    *model.Backlog
	table      table.Model
	search     textinput.Model
	filterMode filterMode
	filterText string
	showFilter bool
	sortColumn int
	sortAsc    bool
	showAll    bool
	tasks      []*model.Task
}

func newScreenTaskList(b *model.Backlog) *taskListModel {
	ti := textinput.New()
	ti.Placeholder = "search tasks..."
	ti.CharLimit = 50

	columns := []table.Column{
		{Title: "Project", Width: 8},
		{Title: "ID", Width: 16},
		{Title: "Type", Width: 8},
		{Title: "Status", Width: 12},
		{Title: "Priority", Width: 10},
		{Title: "Assignee", Width: 12},
		{Title: "SP", Width: 4},
		{Title: "Sprint", Width: 10},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(20),
	)
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(colorBorder).
		BorderBottom(true).
		Bold(false).
		Foreground(colorTextDim)
	s.Selected = s.Selected.
		Foreground(colorTextBright).
		Background(colorPrimary).
		Bold(false)
	s.Cell = s.Cell.
		Foreground(colorText)
	t.SetStyles(s)

	tl := &taskListModel{
		backlog: b,
		table:   t,
		search:  ti,
		showAll: true,
		tasks:   b.AllTasks,
	}
	tl.refresh()
	return tl
}

func (s *taskListModel) Init() tea.Cmd { return nil }

func (s *taskListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, Keys.Filter):
			s.filterMode = filterSearch
			s.showFilter = true
			s.search.Focus()
			return s, textinput.Blink
		case key.Matches(msg, Keys.FilterMode):
			s.cycleFilterMode()
		case key.Matches(msg, Keys.Sort):
			s.cycleSortColumn()
		case key.Matches(msg, Keys.ToggleProj):
			s.showAll = !s.showAll
			s.refresh()
		case key.Matches(msg, Keys.Back):
			if s.showFilter {
				s.showFilter = false
				s.filterMode = filterNone
				s.search.Blur()
				s.search.SetValue("")
				s.refresh()
			}
		}
	case reloadMsg:
		s.tasks = s.backlog.AllTasks
		s.refresh()
	}

	if s.showFilter && s.filterMode == filterSearch {
		var cmd tea.Cmd
		s.search, cmd = s.search.Update(msg)
		s.filterText = s.search.Value()
		s.refresh()
		return s, cmd
	}

	return s, nil
}

func (s *taskListModel) View() string {
	var b strings.Builder

	b.WriteString(headerStyle.Render("  Task List"))

	if s.showFilter {
		b.WriteString(lipgloss.NewStyle().Padding(0, 2).Render("\n" + s.search.View()))
	} else {
		b.WriteString("\n")
	}
	b.WriteString("\n")

	b.WriteString(s.table.View())

	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Foreground(colorTextDim).Padding(0, 2).
		Render(fmt.Sprintf("%d tasks | / search | f filter | s sort | p toggle projects", len(s.tasks))))

	return b.String()
}

func (s *taskListModel) refresh() {
	filtered := s.filterTasks()
	sorted := s.sortTasks(filtered)
	s.tasks = sorted

	rows := make([]table.Row, 0, len(sorted))
	for _, t := range sorted {
		sp := ""
		if t.StoryPoints != nil {
			sp = fmt.Sprintf("%d", *t.StoryPoints)
		}
		projectPrefix := strings.Split(t.ID, "-")[0]
		if len(projectPrefix) > 4 {
			projectPrefix = projectPrefix[:4]
		}
		rows = append(rows, table.Row{
			projectPrefix,
			t.ID,
			string(t.Type),
			string(t.Status),
			string(t.Priority),
			t.Assignee,
			sp,
			t.Sprint,
		})
	}
	s.table.SetRows(rows)
}

func (s *taskListModel) filterTasks() []*model.Task {
	tasks := s.backlog.AllTasks

	if !s.showAll {
		var filtered []*model.Task
		for _, t := range tasks {
			if t.ProjectID == s.backlog.Current.Config.ProjectID {
				filtered = append(filtered, t)
			}
		}
		tasks = filtered
	}

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

func (s *taskListModel) sortTasks(tasks []*model.Task) []*model.Task {
	sorted := make([]*model.Task, len(tasks))
	copy(sorted, tasks)

	sort.Slice(sorted, func(i, j int) bool {
		var less bool
		switch s.sortColumn {
		case 0:
			less = sorted[i].ID < sorted[j].ID
		case 1:
			less = sorted[i].ID < sorted[j].ID
		case 2:
			less = sorted[i].Type < sorted[j].Type
		case 3:
			less = sorted[i].StatusScore() < sorted[j].StatusScore()
		case 4:
			less = sorted[i].PriorityScore() < sorted[j].PriorityScore()
		case 5:
			less = sorted[i].Assignee < sorted[j].Assignee
		case 6:
			ai, bi := 0, 0
			if sorted[i].StoryPoints != nil {
				ai = *sorted[i].StoryPoints
			}
			if sorted[j].StoryPoints != nil {
				bi = *sorted[j].StoryPoints
			}
			less = ai < bi
		default:
			less = sorted[i].ID < sorted[j].ID
		}
		if !s.sortAsc {
			return !less
		}
		return less
	})
	return sorted
}

func (s *taskListModel) cycleFilterMode() {
	switch s.filterMode {
	case filterNone:
		s.filterMode = filterByStatus
	case filterByStatus:
		s.filterMode = filterByPriority
	case filterByPriority:
		s.filterMode = filterByType
	case filterByType:
		s.filterMode = filterByProject
	case filterByProject:
		s.filterMode = filterNone
	}
	s.refresh()
}

func (s *taskListModel) cycleSortColumn() {
	s.sortColumn = (s.sortColumn + 1) % 8
	s.refresh()
}

func (s *taskListModel) SelectedTask() *model.Task {
	if len(s.tasks) == 0 {
		return nil
	}
	row := s.table.Cursor()
	if row >= 0 && row < len(s.tasks) {
		return s.tasks[row]
	}
	return nil
}
