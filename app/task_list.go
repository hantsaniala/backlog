package app

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
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
	focusContent
	focusParent
	focusEpic
	focusDepends
	focusBlocks
	focusRelated
)

type flatRow struct {
	task      *model.Task
	prefix    string
	level     int
}

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
	rows       []flatRow

	detailTask  *model.Task
	focus       focusSection
	depCursor   int
	parentLinks  []detailLink
	epicLinks    []detailLink
	depLinks     []detailLink
	blockLinks   []detailLink
	relatedLinks []detailLink

	width          int
	height         int
	detailViewport viewport.Model
	detailReady    bool
}

func newScreenTaskList(b *model.Backlog) *taskListModel {
	ti := textinput.New()
	ti.Placeholder = "search tasks..."
	ti.CharLimit = 50

	columns := []table.Column{
		{Title: "Proj", Width: 6},
		{Title: "ID", Width: 14},
		{Title: "St", Width: 4},
		{Title: "Title", Width: 40},
		{Title: "SP", Width: 4},
		{Title: "Sprint", Width: 8},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(10),
	)
	st := table.DefaultStyles()
	st.Header = st.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(colorBorder).
		BorderBottom(true).
		Bold(false).
		Foreground(colorTextDim)
	st.Selected = st.Selected.
		Foreground(colorTextBright).
		Background(colorPrimary).
		Bold(false)
	st.Cell = st.Cell.
		Foreground(colorText)
	t.SetStyles(st)

	tl := &taskListModel{
		backlog: b,
		table:   t,
		search:  ti,
		showAll: true,
	}
	tl.rebuild()
	tl.openDetail()
	return tl
}

func (s *taskListModel) Init() tea.Cmd { return nil }

func (s *taskListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if key.Matches(msg, Keys.Filter) && !s.showFilter {
			s.filterMode = filterSearch
			s.showFilter = true
			s.search.Focus()
			return s, textinput.Blink
		}

		switch {
		case key.Matches(msg, Keys.Enter):
			if link := s.focusedDepLink(); link != nil && link.Task != nil {
				s.detailTask = link.Task
				s.depCursor = 0
				s.resolveLinks()
				s.updateDetailContent()
				return s, nil
			}
			if s.focus == focusList {
				s.focus = focusContent
				return s, nil
			}
		case key.Matches(msg, Keys.Tab):
			s.cycleFocus()
			return s, nil
		case key.Matches(msg, Keys.Sort):
			s.sortColumn = (s.sortColumn + 1) % 7
			s.rebuild()
			return s, nil
		case key.Matches(msg, Keys.Up):
			if s.focus == focusContent {
				s.detailViewport, _ = s.detailViewport.Update(msg)
			} else if s.focus == focusList {
				s.table, _ = s.table.Update(msg)
				s.syncDetailFromList()
			} else {
				s.depCursorUp()
			}
			return s, nil
		case key.Matches(msg, Keys.Down):
			if s.focus == focusContent {
				s.detailViewport, _ = s.detailViewport.Update(msg)
			} else if s.focus == focusList {
				s.table, _ = s.table.Update(msg)
				s.syncDetailFromList()
			} else {
				s.depCursorDown()
			}
			return s, nil
		}

	case tea.WindowSizeMsg:
		s.width = msg.Width
		s.height = msg.Height
		s.detailReady = false
		s.rebuild()

	case reloadMsg:
		s.rebuild()
	}

	if s.showFilter && s.filterMode == filterSearch {
		if km, ok := msg.(tea.KeyMsg); ok && key.Matches(km, Keys.Back) {
			s.showFilter = false
			s.filterMode = filterNone
			s.search.Blur()
			s.search.SetValue("")
			s.filterText = ""
			s.rebuild()
			return s, nil
		}
		oldText := s.filterText
		var cmd2 tea.Cmd
		s.search, cmd2 = s.search.Update(msg)
		s.filterText = s.search.Value()
		if s.filterText != oldText {
			s.rebuild()
		}
		return s, cmd2
	}

	var tableCmd tea.Cmd
	s.table, tableCmd = s.table.Update(msg)
	return s, tableCmd
}

func (s *taskListModel) View() string {
	var b strings.Builder

	if s.showFilter {
		b.WriteString(lipgloss.NewStyle().Padding(0, 2).Render(s.search.View()))
	} else {
		b.WriteString("\n")
	}

	listView := s.table.View()
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, listView, s.detailViewport.View()))

	b.WriteString("\n")
	info := fmt.Sprintf("%d items | 1-5 screens | / search | s sort | Tab cycle | Enter open dep", len(s.rows))
	b.WriteString(lipgloss.NewStyle().Foreground(colorTextDim).Padding(0, 2).Render(info))

	return b.String()
}

func (s *taskListModel) rebuild() {
	filtered := s.filterTasks()
	s.buildTree(filtered)

	showType := false
	for _, r := range s.rows {
		if r.task.Type != model.TypeTask {
			showType = true
			break
		}
	}

	avail := 80
	if s.width > 0 {
		avail = s.width - 4
		detailW := s.width * 2 / 5
		if detailW < 35 {
			detailW = 35
		}
		if detailW > 55 {
			detailW = 55
		}
		avail = s.width - 4 - detailW
	}
	if avail < 40 {
		avail = 40
	}

	fixedWidth := 6 + 14 + 4 + 4 + 8 // Proj+ID+St+SP+Sprint
	if showType {
		fixedWidth += 4
	}
	titleW := avail - fixedWidth
	if titleW < 10 {
		titleW = 10
	}

	cols := []table.Column{
		{Title: "Proj", Width: 6},
		{Title: "ID", Width: 14},
	}
	if showType {
		cols = append(cols, table.Column{Title: "T", Width: 4})
	}
	cols = append(cols,
		table.Column{Title: "St", Width: 4},
		table.Column{Title: "Title", Width: titleW},
		table.Column{Title: "SP", Width: 4},
		table.Column{Title: "Sprint", Width: 8},
	)
	s.table.SetColumns(cols)

	rows := make([]table.Row, 0, len(s.rows))
	for _, r := range s.rows {
		t := r.task
		sp := ""
		if t.StoryPoints != nil {
			sp = fmt.Sprintf("%d", *t.StoryPoints)
		}
		pid := strings.Split(t.ID, "-")[0]
		if len(pid) > 4 {
			pid = pid[:4]
		}

		title := truncate(r.prefix+t.Summary, titleW)
		if title == "" {
			title = truncate(r.prefix+t.ID, titleW)
		}

		glyph := statusDot(string(t.Status))

		row := table.Row{pid, t.ID}
		if showType {
			row = append(row, typeDot(string(t.Type)))
		}
		row = append(row, glyph, title, sp, t.Sprint)
		rows = append(rows, row)
	}
	s.table.SetRows(rows)
	tableH := s.height - 6
	if tableH < 5 {
		tableH = 5
	}
	s.table.SetHeight(tableH)
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	if n <= 3 {
		return s[:n]
	}
	return s[:n-1] + "\u2026"
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

// Build a nested tree from epics → stories → tasks, then flat items
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

		// Children of this epic (stories + tasks with epic set)
		var children []*model.Task
		for _, t := range tasks {
			if t.Epic == ep.ID && t.ID != ep.ID {
				children = append(children, t)
			}
		}
		for _, child := range children {
			if child.Type == model.TypeStory {
				seen[child.ID] = true
				rows = append(rows, flatRow{task: child, prefix: "  \u251C ", level: 1})
				// Tasks under this story
				for _, t := range children {
					if t.Parent == child.ID && t.ID != child.ID {
						seen[t.ID] = true
						rows = append(rows, flatRow{task: t, prefix: "  \u2502   \u2514 ", level: 2})
					}
				}
			}
		}
		// Tasks directly under epic (not under a story)
		for _, child := range children {
			if !seen[child.ID] && child.Type == model.TypeTask {
				seen[child.ID] = true
				rows = append(rows, flatRow{task: child, prefix: "  \u2514 ", level: 1})
			}
		}
	}

	// Remaining items (no epic, or external)
	for _, t := range tasks {
		if !seen[t.ID] {
			rows = append(rows, flatRow{task: t, prefix: "", level: 0})
		}
	}

	s.rows = rows
}

// Detail panel management

func (s *taskListModel) openDetail() {
	task := s.SelectedTask()
	if task == nil {
		return
	}
	s.detailTask = task
	s.depCursor = 0
	s.focus = focusList
	s.resolveLinks()
	s.updateDetailContent()
	s.rebuild()
}

func (s *taskListModel) syncDetailFromList() {
	task := s.SelectedTask()
	if task != nil && task != s.detailTask {
		s.detailTask = task
		s.depCursor = 0
		s.resolveLinks()
		s.updateDetailContent()
	}
}

func (s *taskListModel) updateDetailContent() {
	detailW := s.width * 2 / 5
	if detailW < 35 {
		detailW = 35
	}
	if detailW > 55 {
		detailW = 55
	}
	content := renderDetailPanel(s.detailTask, s.backlog, s.depCursor, s.focusSectionName(), detailW)
	viewportH := s.height - 6
	if viewportH < 10 {
		viewportH = 10
	}
	if !s.detailReady {
		s.detailViewport = viewport.New(detailW, viewportH)
		s.detailViewport.SetContent(content)
		s.detailReady = true
	} else {
		s.detailViewport.SetContent(content)
		s.detailViewport.GotoTop()
	}
}

func (s *taskListModel) resolveLinks() {
	if s.detailTask == nil {
		return
	}
	s.parentLinks = resolveDetailLinks(filterEmpty([]string{s.detailTask.Parent}), s.backlog)
	s.epicLinks = resolveDetailLinks(filterEmpty([]string{s.detailTask.Epic}), s.backlog)
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
	secs := []focusSection{focusList, focusContent}
	if len(s.parentLinks) > 0 {
		secs = append(secs, focusParent)
	}
	if len(s.epicLinks) > 0 {
		secs = append(secs, focusEpic)
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
	case focusContent:
		return "content"
	case focusParent:
		return "parent"
	case focusEpic:
		return "epic"
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
	if s.depCursor < s.focusedDepCount()-1 {
		s.depCursor++
	}
}

func (s *taskListModel) focusedDepCount() int {
	switch s.focus {
	case focusParent:
		return len(s.parentLinks)
	case focusEpic:
		return len(s.epicLinks)
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
	case focusEpic:
		links = s.epicLinks
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

func (s *taskListModel) SelectedTask() *model.Task {
	if len(s.rows) == 0 {
		return nil
	}
	row := s.table.Cursor()
	if row >= 0 && row < len(s.rows) {
		return s.rows[row].task
	}
	return nil
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
