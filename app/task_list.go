package app

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hantsaniala/backlog/model"
)

type flatRow struct {
	task   *model.Task
	prefix string
	level  int
	epicID string // which epic owns this row; "" for root-level or orphan items
}

type viewMode int

const (
	modeTree viewMode = iota
	modeDetail
	modeSearch
	modeStatusPopup
	modeInlineFilter
	modeAssign
)

type taskListModel struct {
	backlog     *model.Backlog
	backlogRoot string
	editorCmd   string
	width       int
	height      int
	mode        viewMode

	rows        []flatRow
	visibleRows []flatRow
	cursor      int
	expanded    map[string]bool

	filterText string
	filterOn   bool

	inlineInput  textinput.Model
	filterHistory []string

	detailView *detailView

	searchModal   *searchModalModel
	searchRunning bool

	statusPopup   *statusPopupModel
	statusRunning bool

	treeViewport viewport.Model
	treeReady    bool

	jumpHints *jumpHintState
	visualSel *visualSelection
	bulkAssignMode bool

	pendingG     bool
	pendingZ     bool
	pendingY     bool
	pendingM     bool
	pendingQuote bool

	marks          map[string]string
	lastFilterText string
}

func newScreenTaskList(b *model.Backlog) *taskListModel {
	ti := textinput.New()
	ti.Placeholder = "filter by ID, title, or assignee..."
	ti.CharLimit = 100
	ti.Width = 50

		m := &taskListModel{
		backlog:       b,
		backlogRoot:   b.Current.Root,
		expanded:      make(map[string]bool),
		inlineInput:   ti,
		filterHistory: make([]string, 0),
		jumpHints:     newJumpHintState(),
		visualSel:     newVisualSelection(),
		marks:         make(map[string]string),
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
		return s.handleSearchUpdate(msg)
	}
	if s.mode == modeStatusPopup && s.statusRunning {
		return s.handleStatusPopupUpdate(msg)
	}
	if s.mode == modeInlineFilter {
		return s.handleInlineFilterUpdate(msg)
	}
	if s.mode == modeAssign {
		return s.handleAssignUpdate(msg)
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
		s.inlineInput.Width = s.width - 20
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

// --- Sub-model update helpers ---

func (s *taskListModel) handleSearchUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		s.searchRunning = false
		s.mode = modeTree
		return s, nil
	}
	return s, cmd
}

func (s *taskListModel) handleStatusPopupUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (s *taskListModel) handleAssignUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			s.bulkAssignMode = false
			s.mode = modeTree
			return s, nil
		case "enter":
			assignee := s.inlineInput.Value()
			if assignee != "" {
				if s.bulkAssignMode {
					// Bulk assign all selected items
					count := 0
					for _, t := range s.backlog.AllTasks {
						if s.visualSel.selected[t.ID] {
							t.Assignee = assignee
							count++
						}
					}
					s.visualSel.cancel()
					s.bulkAssignMode = false
					s.mode = modeTree
					return s, notifyCmd(fmt.Sprintf("Assigned %d items to %s", count, assignee))
				} else if s.cursor >= 0 && s.cursor < len(s.visibleRows) {
					task := s.visibleRows[s.cursor].task
					task.Assignee = assignee
				}
			}
			s.mode = modeTree
			return s, nil
		}
	}
	var cmd tea.Cmd
	s.inlineInput, cmd = s.inlineInput.Update(msg)
	return s, cmd
}

func (s *taskListModel) handleInlineFilterUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			s.mode = modeTree
			return s, nil
	case "enter":
		s.filterText = s.inlineInput.Value()
		s.filterOn = s.filterText != ""
		s.lastFilterText = s.filterText
		s.addFilterHistory(s.filterText)
		s.mode = modeTree
		s.rebuild()
		return s, nil
		}
	}
	var cmd tea.Cmd
	s.inlineInput, cmd = s.inlineInput.Update(msg)
	// Live filter
	s.filterText = s.inlineInput.Value()
	s.filterOn = s.filterText != ""
	s.lastFilterText = s.filterText
	s.rebuild()
	return s, cmd
}

func (s *taskListModel) addFilterHistory(f string) {
	if f == "" {
		return
	}
	s.filterHistory = append([]string{f}, s.filterHistory...)
	if len(s.filterHistory) > 5 {
		s.filterHistory = s.filterHistory[:5]
	}
}

// --- Main key handlers ---

func (s *taskListModel) handleTreeKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Jump hints active: route all keypresses to jump buffer
	if s.jumpHints.active {
		return s.handleJumpKey(msg)
	}

	// Visual mode active: route to visual handler
	if s.visualSel.active {
		return s.handleVisualKey(msg)
	}

	// Pending multi-key sequences
	if s.pendingG {
		s.pendingG = false
		if msg.String() == "g" || msg.String() == "G" {
			// gg or gG — both go to top
			s.cursor = 0
			s.clampCursor()
			return s, nil
		}
	}
	if s.pendingZ {
		s.pendingZ = false
		ch := msg.String()
		if ch == "z" {
			s.centerCursor()
			return s, nil
		}
		if ch == "t" && len(s.visibleRows) > 0 {
			s.cursor = int(s.treeViewport.YOffset)
			s.clampCursor()
			return s, nil
		}
		if ch == "b" {
			bottom := int(s.treeViewport.YOffset) + s.treeViewport.Height - 1
			if bottom >= len(s.visibleRows) {
				bottom = len(s.visibleRows) - 1
			}
			s.cursor = bottom
			s.clampCursor()
			return s, nil
		}
	}
	if s.pendingY {
		s.pendingY = false
		if len(s.visibleRows) > 0 && s.cursor >= 0 && s.cursor < len(s.visibleRows) {
			task := s.visibleRows[s.cursor].task
			if msg.String() == "k" {
				copyToClipboard(task.ID)
				return s, notifyCmd(fmt.Sprintf("Copied %s", task.ID))
			}
		}
	}
	if s.pendingM {
		s.pendingM = false
		ch := msg.String()
		if len(ch) == 1 && ch >= "a" && ch <= "z" && len(s.visibleRows) > 0 && s.cursor >= 0 && s.cursor < len(s.visibleRows) {
			s.marks[ch] = s.visibleRows[s.cursor].task.ID
			return s, notifyCmd(fmt.Sprintf("Mark %s set on %s", ch, s.visibleRows[s.cursor].task.ID))
		}
	}
	if s.pendingQuote {
		s.pendingQuote = false
		ch := msg.String()
		if len(ch) == 1 && ch >= "a" && ch <= "z" {
			if taskID, ok := s.marks[ch]; ok {
				for i, r := range s.visibleRows {
					if r.task.ID == taskID {
						s.cursor = i
						s.clampCursor()
						return s, notifyCmd(fmt.Sprintf("Jumped to mark %s (%s)", ch, taskID))
					}
				}
				return s, notifyCmd(fmt.Sprintf("Mark %s item %s not visible", ch, taskID))
			}
			return s, notifyCmd(fmt.Sprintf("No mark %s", ch))
		}
	}

	switch {
	// Quit
	case key.Matches(msg, NormalKeys.Quit):
		return s, tea.Quit

	// Navigation
	case key.Matches(msg, NormalKeys.Up):
		if s.cursor > 0 {
			s.cursor--
		}
		s.clampCursor()
		return s, nil

	case key.Matches(msg, NormalKeys.Down):
		if s.cursor < len(s.visibleRows)-1 {
			s.cursor++
		}
		s.clampCursor()
		return s, nil

	case key.Matches(msg, NormalKeys.GotoTop):
		// Single G — go to bottom
		s.cursor = len(s.visibleRows) - 1
		s.clampCursor()
		return s, nil

	case msg.String() == "g":
		s.pendingG = true
		return s, nil

	case msg.String() == "z":
		s.pendingZ = true
		return s, nil

	case key.Matches(msg, NormalKeys.HalfDown):
		page := (s.height - 8) / 2
		if page < 1 {
			page = 1
		}
		s.cursor += page
		s.clampCursor()
		return s, nil

	case key.Matches(msg, NormalKeys.HalfUp):
		page := (s.height - 8) / 2
		if page < 1 {
			page = 1
		}
		s.cursor -= page
		s.clampCursor()
		return s, nil

	// Page navigation
	case msg.String() == "]":
		page := s.height - 8
		if page < 1 {
			page = 1
		}
		s.cursor += page
		s.clampCursor()
		// Also scroll viewport
		if s.treeReady {
			s.treeViewport.YOffset += page
		}
		return s, nil

	case msg.String() == "[":
		page := s.height - 8
		if page < 1 {
			page = 1
		}
		s.cursor -= page
		s.clampCursor()
		if s.treeReady {
			s.treeViewport.YOffset -= page
			if s.treeViewport.YOffset < 0 {
				s.treeViewport.YOffset = 0
			}
		}
		return s, nil

	// Back / cancel
	case key.Matches(msg, NormalKeys.Back):
		if s.filterOn {
			s.filterText = ""
			s.filterOn = false
			s.rebuild()
		}
		return s, nil

	// Update lastFilterText when clearing filter in tree mode
	case msg.String() == "C-l":
		s.filterText = ""
		s.filterOn = false
		s.lastFilterText = ""
		s.rebuild()
		return s, nil

	// Expand / collapse
	case key.Matches(msg, NormalKeys.Left):
		if len(s.visibleRows) > 0 && s.cursor >= 0 && s.cursor < len(s.visibleRows) {
			r := s.visibleRows[s.cursor]
			if r.task.Type == model.TypeEpic {
				s.expanded[r.task.ID] = false
				s.buildVisibleRows()
				s.clampCursor()
			}
		}
		return s, nil

	case key.Matches(msg, NormalKeys.Right):
		if len(s.visibleRows) > 0 && s.cursor >= 0 && s.cursor < len(s.visibleRows) {
			r := s.visibleRows[s.cursor]
			if r.task.Type == model.TypeEpic {
				s.expanded[r.task.ID] = true
				s.buildVisibleRows()
				s.clampCursor()
			}
		}
		return s, nil

	case key.Matches(msg, NormalKeys.Expand):
		if len(s.visibleRows) > 0 && s.cursor >= 0 && s.cursor < len(s.visibleRows) {
			r := s.visibleRows[s.cursor]
			if r.task.Type == model.TypeEpic {
				oldVal := s.expanded[r.task.ID]
				s.expanded[r.task.ID] = !oldVal
				s.buildVisibleRows()
				s.clampCursor()
			} else {
				s.toggleDone(r.task)
			}
		}
		return s, nil

	// Detail
	case key.Matches(msg, NormalKeys.Enter):
		if len(s.visibleRows) > 0 && s.cursor >= 0 && s.cursor < len(s.visibleRows) {
			task := s.visibleRows[s.cursor].task
			s.enterDetail(task)
		}
		return s, nil

	// Filter / search
	case key.Matches(msg, NormalKeys.Filter):
		// If already filtered, pressing / again opens inline filter with history
		s.inlineInput.SetValue(s.filterText)
		s.inlineInput.Focus()
		s.mode = modeInlineFilter
		return s, s.inlineInput.Focus()

	// Jump hints
	case key.Matches(msg, NormalKeys.Jump):
		s.jumpHints.activate(len(s.visibleRows))
		return s, nil

	// Copy key (pending y)
	case msg.String() == "y":
		s.pendingY = true
		return s, nil

	// Mark set key (pending letter)
	case msg.String() == "m":
		s.pendingM = true
		return s, nil

	// Mark jump key (pending letter)
	case msg.String() == "'":
		s.pendingQuote = true
		return s, nil

	// Search next/prev
	case key.Matches(msg, NormalKeys.SearchNext):
		return s.handleSearchRepeat(1)

	case key.Matches(msg, NormalKeys.SearchPrev):
		return s.handleSearchRepeat(-1)

	// Block motion — jump between epics
	case key.Matches(msg, NormalKeys.BlockUp):
		epicIdx := -1
		for i := s.cursor - 1; i >= 0; i-- {
			if s.visibleRows[i].task.Type == model.TypeEpic {
				epicIdx = i
				break
			}
		}
		if epicIdx >= 0 {
			s.cursor = epicIdx
			s.clampCursor()
		}
		return s, nil

	case key.Matches(msg, NormalKeys.BlockDown):
		epicIdx := -1
		for i := s.cursor + 1; i < len(s.visibleRows); i++ {
			if s.visibleRows[i].task.Type == model.TypeEpic {
				epicIdx = i
				break
			}
		}
		if epicIdx >= 0 {
			s.cursor = epicIdx
			s.clampCursor()
		}
		return s, nil

	// Word search — search for word under cursor
	case key.Matches(msg, NormalKeys.WordSearch):
		if len(s.visibleRows) > 0 && s.cursor >= 0 && s.cursor < len(s.visibleRows) {
			word := s.visibleRows[s.cursor].task.ID
			if word != "" {
				s.filterText = word
				s.filterOn = true
				s.lastFilterText = word
				s.rebuild()
				return s, notifyCmd(fmt.Sprintf("Searching for %q", word))
			}
		}
		return s, nil

	case key.Matches(msg, NormalKeys.WordSearchRev):
		if len(s.visibleRows) > 0 && s.cursor >= 0 && s.cursor < len(s.visibleRows) {
			word := s.visibleRows[s.cursor].task.ID
			if word != "" {
				s.filterText = word
				s.filterOn = true
				s.lastFilterText = word
				s.rebuild()
				s.cursor = len(s.visibleRows) - 1
				return s, notifyCmd(fmt.Sprintf("Searching for %q (reverse)", word))
			}
		}
		return s, nil

	// Visual mode
	case key.Matches(msg, NormalKeys.Visual):
		s.visualSel.activate(s.cursor)
		return s, nil

	// Quick assign
	case msg.String() == "a":
		if len(s.visibleRows) > 0 && s.cursor >= 0 && s.cursor < len(s.visibleRows) {
			s.inlineInput.SetValue("")
			s.inlineInput.Placeholder = "assignee name..."
			s.inlineInput.Focus()
			s.mode = modeAssign
		}
		return s, nil

	// Cycle status
	case key.Matches(msg, NormalKeys.CycleStatus):
		if len(s.visibleRows) > 0 && s.cursor >= 0 && s.cursor < len(s.visibleRows) {
			task := s.visibleRows[s.cursor].task
			task.Status = nextStatus(task.Status)
			s.rebuild()
		}
		return s, nil

	// Related items from list
	case key.Matches(msg, NormalKeys.Related):
		if len(s.visibleRows) > 0 && s.cursor >= 0 && s.cursor < len(s.visibleRows) {
			task := s.visibleRows[s.cursor].task
			dv := newDetailView(s.backlog, task, s.width, s.height)
			dv.state = detailRelated
			dv.relatedCursor = 0
			s.detailView = dv
			s.mode = modeDetail
		}
		return s, nil
	}

	return s, nil
}

func (s *taskListModel) handleJumpKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "esc" {
		s.jumpHints.cancel()
		return s, nil
	}
	// Single character
	if len(msg.String()) == 1 {
		r := rune(msg.String()[0])
		target, ok := s.jumpHints.pushRune(r)
		if ok && target >= 0 && target < len(s.visibleRows) {
			s.cursor = target
			s.clampCursor()
			task := s.visibleRows[target].task
			return s, jumpNotificationCmd(task.ID)
		}
	}
	return s, nil
}

func (s *taskListModel) handleVisualKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, VisualKeys.Cancel):
		s.visualSel.cancel()
		return s, nil
	case key.Matches(msg, VisualKeys.Toggle):
		if s.cursor >= 0 && s.cursor < len(s.visibleRows) {
			s.visualSel.toggle(s.visibleRows[s.cursor].task.ID)
		}
		return s, nil
	case key.Matches(msg, VisualKeys.SelectAll):
		s.visualSel.selectAll(s.backlog.AllTasks)
		return s, nil
	case key.Matches(msg, VisualKeys.Action):
		// Bulk status change: cycle all selected items
		if len(s.visualSel.selected) > 0 {
			for _, t := range s.backlog.AllTasks {
				if s.visualSel.selected[t.ID] {
					t.Status = nextStatus(t.Status)
				}
			}
			selCount := len(s.visualSel.selected)
			s.visualSel.cancel()
			s.rebuild()
			return s, notifyCmd(fmt.Sprintf("Updated %d items", selCount))
		}
		return s, nil
	case key.Matches(msg, VisualKeys.Assign):
		// Bulk assign: show assign input for all selected items
		if len(s.visualSel.selected) > 0 {
			s.inlineInput.SetValue("")
			s.inlineInput.Placeholder = "assignee name..."
			s.inlineInput.Focus()
			s.bulkAssignMode = true
			s.mode = modeAssign
		}
		return s, nil
	case key.Matches(msg, VisualKeys.Down):
		if s.cursor < len(s.visibleRows)-1 {
			s.cursor++
			s.visualSel.extendDown(s.backlog.AllTasks, s.cursor)
		}
		return s, nil
	case key.Matches(msg, VisualKeys.Up):
		if s.cursor > 0 {
			s.cursor--
			s.visualSel.extendUp(s.backlog.AllTasks, s.cursor)
		}
		return s, nil
	}
	return s, nil
}

// --- Detail key handler ---

func (s *taskListModel) handleDetailKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, NormalKeys.Back):
		if s.detailView != nil {
			if s.detailView.state == detailPreview {
				s.detailView.state = detailRelated
				s.detailView.previewTask = nil
				return s, nil
			}
			if s.detailView.state == detailRelated {
				s.detailView.state = detailNormal
				s.detailView.relatedItems = nil
				return s, nil
			}
		}
		s.mode = modeTree
		s.detailView = nil
		return s, nil

	case msg.String() == "tab":
		if s.detailView != nil && len(s.detailView.subtaskTasks) > 0 {
			s.detailView.subtaskFocus = !s.detailView.subtaskFocus
		}
		return s, nil

	case key.Matches(msg, NormalKeys.Expand):
		if s.detailView != nil && s.detailView.subtaskFocus && len(s.detailView.subtaskTasks) > 0 {
			st := s.detailView.subtaskTasks[s.detailView.subtaskCursor]
			if st.Status == model.StatusDone {
				st.Status = model.StatusTodo
			} else {
				st.Status = model.StatusDone
			}
		}
		return s, nil

	case key.Matches(msg, NormalKeys.CycleStatus):
		s.mode = modeStatusPopup
		if s.detailView != nil {
			s.statusPopup = newStatusPopup(s.detailView.task, 30)
		}
		s.statusRunning = true
		return s, nil

	case key.Matches(msg, NormalKeys.Related):
		if s.detailView != nil {
			s.detailView.resolveLinks()
			s.detailView.state = detailRelated
			s.detailView.relatedCursor = 0
		}
		return s, nil

	case key.Matches(msg, NormalKeys.Enter):
		if s.detailView != nil && s.detailView.state == detailRelated && len(s.detailView.relatedItems) > 0 {
			link := s.detailView.relatedItems[s.detailView.relatedCursor]
			if link.Task != nil {
				s.detailView.previewTask = link.Task
				s.detailView.state = detailPreview
			}
		}
		return s, nil

	case key.Matches(msg, NormalKeys.OpenInEditor):
		if s.detailView != nil && s.detailView.task != nil {
			return s, editorOpenCmd(s.detailView.task, s.editorCmd, s.backlogRoot)
		}
		return s, nil

	case key.Matches(msg, NormalKeys.SearchNext):
		return s.handleSearchRepeat(1)

	case key.Matches(msg, NormalKeys.SearchPrev):
		return s.handleSearchRepeat(-1)

	case key.Matches(msg, NormalKeys.BlockUp):
		return s, nil

	case key.Matches(msg, NormalKeys.BlockDown):
		return s, nil

	case key.Matches(msg, NormalKeys.Up):
		if s.detailView != nil && s.detailView.subtaskFocus && len(s.detailView.subtaskTasks) > 0 {
			if s.detailView.subtaskCursor > 0 {
				s.detailView.subtaskCursor--
			}
		} else if s.detailView != nil && s.detailView.state == detailRelated {
			if s.detailView.relatedCursor > 0 {
				s.detailView.relatedCursor--
			}
		}
		return s, nil

	case key.Matches(msg, NormalKeys.Down):
		if s.detailView != nil && s.detailView.subtaskFocus && len(s.detailView.subtaskTasks) > 0 {
			if s.detailView.subtaskCursor < len(s.detailView.subtaskTasks)-1 {
				s.detailView.subtaskCursor++
			}
		} else if s.detailView != nil && s.detailView.state == detailRelated {
			if s.detailView.relatedCursor < len(s.detailView.relatedItems)-1 {
				s.detailView.relatedCursor++
			}
		}
		return s, nil

	case key.Matches(msg, NormalKeys.HalfDown):
		if s.detailView != nil {
			s.detailView.scrollOffset += (s.height - 8) / 2
		}
		return s, nil

	case key.Matches(msg, NormalKeys.HalfUp):
		if s.detailView != nil {
			s.detailView.scrollOffset -= (s.height - 8) / 2
			if s.detailView.scrollOffset < 0 {
				s.detailView.scrollOffset = 0
			}
		}
		return s, nil
	}
	return s, nil
}

// --- Actions ---

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
	if s.detailView != nil && s.detailView.task.ID == task.ID {
		s.detailView.task.Status = task.Status
	}
}

func (s *taskListModel) centerCursor() {
	// Centers the viewport on the cursor by adjusting scroll position
	half := (s.height - 8) / 2
	if half < 1 {
		half = 1
	}
	target := s.cursor - half
	if target < 0 {
		target = 0
	}
	if s.treeReady {
		s.treeViewport.YOffset = target
	}
}

// --- View ---

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
	case modeInlineFilter:
		return s.renderTreeFull()
	default:
		return s.renderTreeFull()
	}
}

func (s *taskListModel) renderTreeFull() string {
	var b strings.Builder

	// Assign input bar
	if s.mode == modeAssign {
		b.WriteString(lipgloss.NewStyle().Padding(0, 2).Foreground(colorTextDim).Render(" Assign to: "))
		b.WriteString(s.inlineInput.View())
		b.WriteString("\n")
	} else if s.mode == modeInlineFilter {
		b.WriteString(lipgloss.NewStyle().Padding(0, 2).Foreground(colorTextDim).Render(" / "))
		b.WriteString(s.inlineInput.View())
		b.WriteString("\n")
	} else if s.filterOn {
		b.WriteString(lipgloss.NewStyle().Padding(0, 2).Foreground(colorTextDim).Render(
			fmt.Sprintf(" Filter: %s  (%d items)", s.filterText, len(s.visibleRows))))
		b.WriteString("\n")
	}

	// Visual mode indicator
	if s.visualSel.active {
		b.WriteString(lipgloss.NewStyle().Padding(0, 2).Foreground(colorAccent).Bold(true).Render(
			s.visualSel.indicator()))
		b.WriteString("\n")
	}

	// Main tree content
	content := s.renderTree()
	treeH := s.height - 8
	if s.mode == modeInlineFilter || s.filterOn || s.visualSel.active {
		treeH = s.height - 9
	}
	if treeH < 5 {
		treeH = 5
	}

	if !s.treeReady {
		s.treeViewport = viewport.New(s.width-2, treeH)
		s.treeReady = true
	}
	s.treeViewport.Width = s.width - 2
	s.treeViewport.Height = treeH
	s.treeViewport.SetContent(content)

	b.WriteString(s.treeViewport.View())

	// Scroll indicators
	below := len(s.visibleRows) - int(s.treeViewport.YOffset) - treeH
	if int(s.treeViewport.YOffset) > 0 {
		b.WriteString(scrollUpStyle)
	}
	if below > 0 {
		b.WriteString(scrollDownStyle)
	}

	// Page indicator
	totalItems := len(s.visibleRows)
	pageNum := int(s.treeViewport.YOffset)/treeH + 1
	totalPages := (totalItems + treeH - 1) / treeH
	if totalPages < 1 {
		totalPages = 1
	}
	pageStr := fmt.Sprintf("Page %d/%d  [ ] prev/next  ", pageNum, totalPages)
	b.WriteString("\n")
	b.WriteString(pageNavStyle.Render(pageStr))

	// Contextual info
	b.WriteString("\n")
	info := fmt.Sprintf("%d items | / search | j/k nav | h/l expand | space toggle | enter detail | Esc back | q quit",
		totalItems)
	if s.filterOn {
		info = fmt.Sprintf("🔍 %q (%d) ", s.filterText, totalItems) +
			" | " + info
	}
	if s.jumpHints.active {
		info += s.jumpHints.bufferDisplay()
	}
	b.WriteString(lipgloss.NewStyle().Foreground(colorTextDim).Padding(0, 2).Render(info))

	return b.String()
}

func (s *taskListModel) renderTree() string {
	var b strings.Builder

	if len(s.visibleRows) == 0 {
		if s.filterOn {
			b.WriteString(lipgloss.NewStyle().Foreground(colorTextDim).Padding(1, 2).Render(
				fmt.Sprintf("No items match %q", s.filterText)))
			b.WriteString("\n")
			b.WriteString(lipgloss.NewStyle().Foreground(colorTextDim).Padding(1, 2).Render("  Esc to clear filter"))
		} else {
			b.WriteString(lipgloss.NewStyle().Foreground(colorTextDim).Padding(1, 2).Render("~ No items ~"))
		}
		return b.String()
	}

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

		var labelStyle lipgloss.Style
		if r.task.Type == model.TypeEpic {
			labelStyle = lipgloss.NewStyle().Bold(true).Foreground(colorAccent)
		} else if r.task.Type == model.TypeStory {
			labelStyle = lipgloss.NewStyle().Foreground(colorSecondary)
		} else {
			labelStyle = lipgloss.NewStyle().Foreground(colorText)
		}
		label = labelStyle.Render(label)

		line := fmt.Sprintf(" %s%s%s %s %s%s%s", indent, expandSymbol, prefix, g, label, sp, " "+statusStr)

		// Visual mode highlight
		if s.visualSel.active && s.visualSel.selected[r.task.ID] {
			line = lipgloss.NewStyle().Background(colorSurface).Render(line)
		}

		if i == s.cursor {
			line = leftBorderBar + focusedRowStyle.Render(line)
		} else {
			line = " " + line
		}

		// Jump hint overlay
		line = 	s.jumpHints.hintOverlay(i, line)

		// Mark indicator
		for markLetter, markTaskID := range s.marks {
			if markTaskID == r.task.ID {
				markStyle := lipgloss.NewStyle().Foreground(colorWarning).Bold(true)
				line += " " + markStyle.Render(fmt.Sprintf("[%s]", markLetter))
				break
			}
		}

		b.WriteString(line)
		b.WriteString("\n")
	}

	return b.String()
}

func (s *taskListModel) footerPos() string {
	if len(s.visibleRows) == 0 {
		return ""
	}
	return fmt.Sprintf("%d,%d  %d%%", s.cursor+1, len(s.visibleRows), (s.cursor+1)*100/len(s.visibleRows))
}

func (s *taskListModel) footerHint() string {
	if s.mode == modeDetail && s.detailView != nil {
		if s.detailView.state == detailRelated {
			return " j/k:navigate | Enter:preview | Esc:back"
		}
		if s.detailView.state == detailPreview {
			return " Esc:back"
		}
		if s.statusRunning {
			return " j/k:select | Enter:confirm | Esc:cancel"
		}
		return " j/k:scroll | e:status | o:open | r:links | C-d/u:scroll | Esc:back"
	}
	if s.mode == modeAssign {
		return " Type assignee | Enter:confirm | Esc:cancel"
	}
	return " j/k:move | Enter:view | Space:toggle | e:status | /:filter | ::cmd | ?:help"
}

// --- Tree building ---

func (s *taskListModel) handleSearchRepeat(dir int) (tea.Model, tea.Cmd) {
	if s.lastFilterText == "" {
		return s, nil
	}
	s.filterText = s.lastFilterText
	s.filterOn = true
	s.rebuild()
	if len(s.visibleRows) == 0 {
		return s, notifyCmd(fmt.Sprintf("No matches for %q", s.lastFilterText))
	}
	if dir > 0 {
		// Find the first match after current cursor
		for i := s.cursor + 1; i < len(s.visibleRows); i++ {
			if matchFilter(s.visibleRows[i].task, s.lastFilterText) {
				s.cursor = i
				s.clampCursor()
				return s, nil
			}
		}
		// Wrap to first match
		for i := 0; i <= s.cursor; i++ {
			if matchFilter(s.visibleRows[i].task, s.lastFilterText) {
				s.cursor = i
				s.clampCursor()
				return s, notifyCmd("Search wrapped to top")
			}
		}
	} else {
		// Find the first match before current cursor
		for i := s.cursor - 1; i >= 0; i-- {
			if matchFilter(s.visibleRows[i].task, s.lastFilterText) {
				s.cursor = i
				s.clampCursor()
				return s, nil
			}
		}
		// Wrap to last match
		for i := len(s.visibleRows) - 1; i >= s.cursor; i-- {
			if matchFilter(s.visibleRows[i].task, s.lastFilterText) {
				s.cursor = i
				s.clampCursor()
				return s, notifyCmd("Search wrapped to bottom")
			}
		}
	}
	return s, notifyCmd(fmt.Sprintf("No matches for %q", s.lastFilterText))
}

func matchFilter(task *model.Task, filter string) bool {
	lower := strings.ToLower(filter)
	return strings.Contains(strings.ToLower(task.ID), lower) ||
		strings.Contains(strings.ToLower(task.Summary), lower) ||
		strings.Contains(strings.ToLower(task.Assignee), lower)
}

func (s *taskListModel) clampCursor() {
	if len(s.visibleRows) == 0 {
		s.cursor = 0
		return
	}
	if s.cursor >= len(s.visibleRows) {
		s.cursor = len(s.visibleRows) - 1
	}
}

func (s *taskListModel) rebuild() {
	filtered := s.filterTasks()
	s.buildTree(filtered)
	s.buildVisibleRows()
	if s.cursor >= len(s.visibleRows) && len(s.visibleRows) > 0 {
		s.cursor = len(s.visibleRows) - 1
	} else if len(s.visibleRows) == 0 {
		s.cursor = 0
	}
}

func (s *taskListModel) buildVisibleRows() {
	var out []flatRow
	for _, r := range s.rows {
		if r.task.Type == model.TypeEpic {
			out = append(out, r)
		} else if r.epicID == "" {
			// Root-level item (orphan story/remaining task) — always show
			out = append(out, r)
		} else {
			// Child item — show only if epic is expanded
			expanded := s.expanded[r.epicID]
			if expanded {
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
		epTask := model.EpicToTask(ep)
		rows = append(rows, flatRow{task: epTask, prefix: "", level: 0, epicID: ""})
		seen[epTask.ID] = true

		var children []*model.Task
		for _, t := range tasks {
			if t.Epic == ep.ID && t.ID != ep.ID {
				children = append(children, t)
			}
		}
		sort.Slice(children, func(i, j int) bool {
			di := children[i].Status == model.StatusDone
			dj := children[j].Status == model.StatusDone
			if di != dj {
				return !di
			}
			return children[i].ID < children[j].ID
		})
		for _, child := range children {
			if child.Type == model.TypeStory {
				seen[child.ID] = true
				rows = append(rows, flatRow{task: child, prefix: " └", level: 1, epicID: ep.ID})
				for _, t := range children {
					if t.Parent == child.ID && t.ID != child.ID {
						seen[t.ID] = true
						rows = append(rows, flatRow{task: t, prefix: "   •", level: 2, epicID: ep.ID})
					}
				}
			}
		}
		for _, child := range children {
			if !seen[child.ID] && child.Type == model.TypeTask {
				seen[child.ID] = true
				rows = append(rows, flatRow{task: child, prefix: " └", level: 1, epicID: ep.ID})
			}
		}
	}

	// Stories without epic
	for _, t := range tasks {
		if !seen[t.ID] && t.Type == model.TypeStory {
			seen[t.ID] = true
			rows = append(rows, flatRow{task: t, prefix: " └", level: 1, epicID: ""})
			var storyChildren []*model.Task
			for _, child := range tasks {
				if !seen[child.ID] && child.Parent == t.ID && child.ID != t.ID {
					storyChildren = append(storyChildren, child)
				}
			}
			sort.Slice(storyChildren, func(i, j int) bool {
				di := storyChildren[i].Status == model.StatusDone
				dj := storyChildren[j].Status == model.StatusDone
				if di != dj {
					return !di
				}
				return storyChildren[i].ID < storyChildren[j].ID
			})
			for _, child := range storyChildren {
				seen[child.ID] = true
				rows = append(rows, flatRow{task: child, prefix: "   •", level: 2, epicID: ""})
			}
		}
	}

	// Remaining items
	var remaining []*model.Task
	for _, t := range tasks {
		if !seen[t.ID] {
			remaining = append(remaining, t)
		}
	}
	sort.Slice(remaining, func(i, j int) bool {
		di := remaining[i].Status == model.StatusDone
		dj := remaining[j].Status == model.StatusDone
		if di != dj {
			return !di
		}
		return remaining[i].ID < remaining[j].ID
	})
	for _, t := range remaining {
		rows = append(rows, flatRow{task: t, prefix: "", level: 0, epicID: ""})
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
