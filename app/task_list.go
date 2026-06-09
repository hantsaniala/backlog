package app

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
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

type focusSection int

const (
	focusList focusSection = iota
	focusClose
	focusParent
	focusDepends
	focusBlocks
	focusRelated
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

	showDetail  bool
	detailTask  *model.Task
	focus       focusSection
	depCursor   int
	parentLinks []detailLink
	depLinks    []detailLink
	blockLinks  []detailLink
	relatedLinks []detailLink
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
		// Global-ish keys: filter always works
		if key.Matches(msg, Keys.Filter) && !s.showFilter {
			s.filterMode = filterSearch
			s.showFilter = true
			s.search.Focus()
			return s, textinput.Blink
		}

		// Detail panel navigation
		if s.showDetail {
			switch {
			case key.Matches(msg, Keys.Enter):
				// Enter on a dep link opens that task in the panel
				if link := s.focusedDepLink(); link != nil && link.Task != nil {
					s.detailTask = link.Task
					s.depCursor = 0
					s.resolveLinks()
					return s, nil
				}
				// Enter on close or already on list → close
				if s.focus == focusClose || s.focus == focusList {
					s.closeDetail()
					return s, nil
				}
				// Enter on a section header with no focusable item → close
				s.closeDetail()
				return s, nil
			case key.Matches(msg, Keys.Back):
				s.closeDetail()
				return s, nil
			case key.Matches(msg, Keys.Tab):
				s.cycleFocus()
				return s, nil
			case key.Matches(msg, Keys.Up):
				s.depCursorUp()
				return s, nil
			case key.Matches(msg, Keys.Down):
				s.depCursorDown()
				return s, nil
			}
		} else {
			// No detail panel open
			switch {
			case key.Matches(msg, Keys.Enter):
				s.openDetail()
				return s, nil
			case key.Matches(msg, Keys.Back):
				return s, nil
			case key.Matches(msg, Keys.Sort):
				s.cycleSortColumn()
				return s, nil
			}
		}

	case reloadMsg:
		s.tasks = s.backlog.AllTasks
		s.refresh()
	}

	if s.showFilter && s.filterMode == filterSearch {
		var cmd2 tea.Cmd
		s.search, cmd2 = s.search.Update(msg)
		s.filterText = s.search.Value()
		s.refresh()
		return s, cmd2
	}

	// Forward to table for navigation (always)
	var tableCmd tea.Cmd
	s.table, tableCmd = s.table.Update(msg)
	return s, tableCmd
}

func (s *taskListModel) View() string {
	var b strings.Builder

	// Search bar
	if s.showFilter {
		b.WriteString(lipgloss.NewStyle().Padding(0, 2).Render(s.search.View()))
	} else {
		b.WriteString("\n")
	}

	// Main content: table + optional detail panel
	if s.showDetail && s.detailTask != nil {
		listView := s.table.View()
		detailView := renderDetailPanel(s.detailTask, s.backlog, s.depCursor, s.focusSectionName())

		// Fix table height to match
		b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, listView, detailView))
	} else {
		b.WriteString(s.table.View())
	}

	// Footer
	b.WriteString("\n")
	info := fmt.Sprintf("%d tasks | 1-5 screens | / search | s sort", len(s.tasks))
	if s.showDetail {
		info += " | Tab cycle | Enter open dep | Esc close"
	} else {
		info += " | Enter detail"
	}
	b.WriteString(lipgloss.NewStyle().Foreground(colorTextDim).Padding(0, 2).Render(info))

	return b.String()
}

// Detail panel management

func (s *taskListModel) openDetail() {
	task := s.SelectedTask()
	if task == nil {
		return
	}
	s.showDetail = true
	s.detailTask = task
	s.depCursor = 0
	s.focus = focusList
	s.resolveLinks()
}

func (s *taskListModel) closeDetail() {
	s.showDetail = false
	s.detailTask = nil
	s.depCursor = 0
}

func (s *taskListModel) resolveLinks() {
	if s.detailTask == nil {
		return
	}
	s.parentLinks = resolveDetailLinks(filterEmpty([]string{s.detailTask.Parent}), s.backlog)
	s.depLinks = resolveDetailLinks(s.detailTask.DependsOn, s.backlog)
	s.blockLinks = resolveDetailLinks(s.detailTask.Blocks, s.backlog)
	s.relatedLinks = resolveDetailLinks(s.detailTask.RelatedTo, s.backlog)
}

func filterEmpty(ids []string) []string {
	var out []string
	for _, id := range ids {
		if id != "" {
			out = append(out, id)
		}
	}
	return out
}

func (s *taskListModel) cycleFocus() {
	sections := s.activeSections()
	for i, sec := range sections {
		if sec == s.focus {
			s.focus = sections[(i+1)%len(sections)]
			s.depCursor = 0
			return
		}
	}
	s.focus = focusList
	s.depCursor = 0
}

func (s *taskListModel) activeSections() []focusSection {
	secs := []focusSection{focusList, focusClose}
	if len(s.parentLinks) > 0 {
		secs = append(secs, focusParent)
	}
	if len(s.depLinks) > 0 {
		secs = append(secs, focusDepends)
	}
	if len(s.blockLinks) > 0 {
		secs = append(secs, focusBlocks)
	}
	if len(s.relatedLinks) > 0 {
		secs = append(secs, focusRelated)
	}
	return secs
}

func (s *taskListModel) focusSectionName() string {
	switch s.focus {
	case focusList:
		return "list"
	case focusClose:
		return "close"
	case focusParent:
		return "parent"
	case focusDepends:
		return "depends"
	case focusBlocks:
		return "blocks"
	case focusRelated:
		return "related"
	default:
		return ""
	}
}

func (s *taskListModel) depCursorUp() {
	if s.depCursor > 0 {
		s.depCursor--
	}
}

func (s *taskListModel) depCursorDown() {
	count := s.focusedDepCount()
	if s.depCursor < count-1 {
		s.depCursor++
	}
}

func (s *taskListModel) focusedDepCount() int {
	switch s.focus {
	case focusParent:
		return len(s.parentLinks)
	case focusDepends:
		return len(s.depLinks)
	case focusBlocks:
		return len(s.blockLinks)
	case focusRelated:
		return len(s.relatedLinks)
	default:
		return 0
	}
}

func (s *taskListModel) focusedDepLink() *detailLink {
	var links []detailLink
	switch s.focus {
	case focusParent:
		links = s.parentLinks
	case focusDepends:
		links = s.depLinks
	case focusBlocks:
		links = s.blockLinks
	case focusRelated:
		links = s.relatedLinks
	default:
		return nil
	}
	if s.depCursor >= 0 && s.depCursor < len(links) {
		return &links[s.depCursor]
	}
	return nil
}

// Table helpers

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

func (s *taskListModel) cycleSortColumn() {
	s.sortColumn = (s.sortColumn + 1) % 8
	s.refresh()
}
