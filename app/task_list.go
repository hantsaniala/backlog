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
	modeSearch
	modeInlineFilter
)

type taskListModel struct {
	backlog     *model.Backlog
	backlogRoot string
	editorCmd   string
	width       int
	height      int
	mode        viewMode

	detailOpen     bool

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

	treeViewport viewport.Model
	treeReady    bool

	jumpHints *jumpHintState

	pendingG     bool
	pendingZ     bool
	pendingY     bool
	pendingM     bool
	pendingQuote bool

	marks          map[string]string
	lastFilterText string

	// Mouse hover
	hoverLine int // -1 = no hover
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
		marks:         make(map[string]string),
		hoverLine:     -1,
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
	if s.mode == modeInlineFilter {
		return s.handleInlineFilterUpdate(msg)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		s.hoverLine = -1
	if s.mode == modeTree {
		return s.handleTreeKey(msg)
	}

	case tea.WindowSizeMsg:
		s.width = msg.Width
		s.height = msg.Height
		s.treeReady = false
		s.inlineInput.Width = s.width - 20
		if s.detailView != nil {
			detailW := msg.Width - (msg.Width*40)/100
			s.detailView.width = detailW
			s.detailView.height = msg.Height
		}
		if s.searchModal != nil {
			s.searchModal.width = msg.Width
			s.searchModal.height = msg.Height
		}

	case reloadMsg:
		s.rebuild()

	case tea.MouseMsg:
		// Detail mouse wheel scroll
		if s.detailOpen && s.detailView != nil {
			if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonWheelUp {
				s.detailView.view.LineUp(3)
			}
			if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonWheelDown {
				s.detailView.view.LineDown(3)
			}
			// Still fall through to tree hover logic
		}

		// Hover: convert terminal coords to tree-relative.
		treeY := msg.Y - 2
		if treeY >= 0 && treeY < s.treeViewport.Height {
			bufLine := treeY + int(s.treeViewport.YOffset)
			if bufLine >= 0 && bufLine < len(s.visibleRows) {
				s.hoverLine = bufLine
			} else {
				s.hoverLine = -1
			}
		} else {
			s.hoverLine = -1
		}
		if msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft {
			if s.hoverLine >= 0 && s.hoverLine < len(s.visibleRows) {
				s.cursor = s.hoverLine
			}
		}
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
			return s, nil
		}
		s.searchRunning = false
		s.mode = modeTree
		return s, nil
	}
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

	// Detail state shortcut routing
	if s.detailOpen && s.detailView != nil {
		switch {
		case key.Matches(msg, NormalKeys.Related):
			if s.detailView.state == detailRelated {
				s.detailView.state = detailNormal
				s.detailView.relatedItems = nil
			} else {
				s.detailView.resolveLinks()
				s.detailView.state = detailRelated
				s.detailView.relatedCursor = 0
			}
			return s, nil
		case key.Matches(msg, NormalKeys.Enter):
			if s.detailView.state == detailRelated && len(s.detailView.relatedItems) > 0 {
				link := s.detailView.relatedItems[s.detailView.relatedCursor]
				if link.Task != nil {
					s.detailView.previewTask = link.Task
					s.detailView.state = detailPreview
				}
				return s, nil
			}
		case key.Matches(msg, NormalKeys.HalfDown):
			s.detailView.view.HalfViewDown()
			return s, nil
		case key.Matches(msg, NormalKeys.HalfUp):
			s.detailView.view.HalfViewUp()
			return s, nil
		}
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
		if s.detailOpen && s.detailView != nil {
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
			s.detailOpen = false
			s.detailView = nil
			return s, nil
		}
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

	case key.Matches(msg, NormalKeys.OpenInEditor):
		if s.detailOpen && s.detailView != nil {
			return s, editorOpenCmd(s.detailView.task, s.editorCmd, s.backlogRoot)
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

	// Related items from list
	case key.Matches(msg, NormalKeys.Related):
		if s.detailOpen && s.detailView != nil {
			s.detailView.resolveLinks()
			s.detailView.state = detailRelated
			s.detailView.relatedCursor = 0
		} else if len(s.visibleRows) > 0 && s.cursor >= 0 && s.cursor < len(s.visibleRows) {
			task := s.visibleRows[s.cursor].task
			dv := newDetailView(s.backlog, task, s.width, s.height)
			dv.state = detailRelated
			dv.relatedCursor = 0
			s.detailView = dv
			s.detailOpen = true
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

// --- Actions ---

func (s *taskListModel) DetailOpen() bool {
	return s.detailOpen
}

func (s *taskListModel) enterDetail(task *model.Task) {
	dv := newDetailView(s.backlog, task, s.width, s.height)
	dv.resolveLinks()
	s.detailView = dv
	s.detailOpen = true
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
	case modeSearch:
		if s.searchRunning && s.searchModal != nil {
			return s.searchModal.View()
		}
		return ""
	default:
		return s.renderSplitView()
	}
}

func (s *taskListModel) renderSplitView() string {
	treeW := (s.width * 40) / 100
	if treeW < 30 {
		treeW = 30
	}
	detailW := s.width - treeW

	treeContent := s.renderTreeFullWithWidth(treeW)

	if s.detailOpen && s.detailView != nil {
		s.detailView.width = detailW
		s.detailView.height = s.height
		return lipgloss.JoinHorizontal(lipgloss.Top, treeContent, s.detailView.View())
	}

	sep := lipgloss.NewStyle().
		Border(lipgloss.Border{Left: "│"}, false, false, false, true).
		BorderForeground(colorBorder).
		Render(s.renderDetailPlaceholder(detailW, s.height))
	return lipgloss.JoinHorizontal(lipgloss.Top, treeContent, sep)
}

func (s *taskListModel) renderDetailPlaceholder(w, h int) string {
	ph := lipgloss.NewStyle().
		Width(w - 2).
		Height(h - 2).
		Foreground(colorTextDim).
		Align(lipgloss.Center, lipgloss.Center).
		Render("Select a task to view details")
	return ph
}

func (s *taskListModel) renderTreeFullWithWidth(w int) string {
	var b strings.Builder

	if s.mode == modeInlineFilter {
		b.WriteString(lipgloss.NewStyle().Padding(0, 2).Foreground(colorTextDim).Render(" / "))
		b.WriteString(s.inlineInput.View())
		b.WriteString("\n")
	} else if s.filterOn {
		b.WriteString(lipgloss.NewStyle().Padding(0, 2).Foreground(colorTextDim).Render(
			fmt.Sprintf(" Filter: %s  (%d items)", s.filterText, len(s.visibleRows))))
		b.WriteString("\n")
	}

	// Main tree content
	content := s.renderTree()
	treeH := s.height
	if s.filterOn || s.mode == modeInlineFilter {
		treeH = s.height - 1
	}
	if treeH < 5 {
		treeH = 5
	}

	vpW := w - 3
	if !s.treeReady {
		s.treeViewport = viewport.New(vpW, treeH)
		s.treeReady = true
	}
	s.treeViewport.Width = vpW
	s.treeViewport.Height = treeH
	s.treeViewport.SetContent(content)

	vpView := s.treeViewport.View()
	scrollbarStr := renderScrollbar(s.treeViewport, treeH)
	b.WriteString(addScrollbar(vpView, scrollbarStr))

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
		indent := strings.Repeat(" ", r.level)

		// Expand indicator and branch prefix
		expandSymbol := " "
		if r.task.Type == model.TypeEpic {
			if s.expanded[r.task.ID] {
				expandSymbol = "▼"
			} else {
				expandSymbol = "▶"
			}
		}

		g := statusDot(string(r.task.Status))

		sp := ""
		if r.task.StoryPoints != nil {
			sp = lipgloss.NewStyle().Foreground(colorWarning).Render(fmt.Sprintf(" %dsp", *r.task.StoryPoints))
		}

		// Branch prefix for non-root items
		branchPrefix := r.prefix
		if r.level > 0 && branchPrefix == "" {
			branchPrefix = "├─"
		}

		label := r.task.ID
		if r.task.Summary != "" {
			label = r.task.ID + "  " + r.task.Summary
		}

		var labelStyle lipgloss.Style
		if r.task.Type == model.TypeEpic {
			labelStyle = lipgloss.NewStyle().Bold(true).Foreground(colorAccent)
		} else if r.task.Type == model.TypeStory {
			labelStyle = lipgloss.NewStyle().Foreground(colorPrimary)
		} else {
			labelStyle = lipgloss.NewStyle().Foreground(colorText)
		}
		label = labelStyle.Render(label)

		t := typeDot(string(r.task.Type))
		line := fmt.Sprintf(" %s%s%s%s %s %s%s", indent, expandSymbol, branchPrefix, t, g, label, sp)

		if i == s.cursor {
			line = leftBorderBar + focusedRowStyle.Render(line)
		} else if i == s.hoverLine && s.hoverLine >= 0 {
			// Subtle hover background; only when mouse recently moved
			line = " " + lipgloss.NewStyle().Background(colorSurfaceAlt).Render(line)
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
	if s.detailOpen {
		if s.detailView != nil && s.detailView.state == detailRelated {
			return " j/k:navigate | Enter:preview | Esc:back"
		}
		if s.detailView != nil && s.detailView.state == detailPreview {
			return " Esc:back"
		}
		return " o:open | r:links | C-d/u:scroll | Esc:close | Tab:nav"
	}
	pageInfo := ""
	if len(s.visibleRows) > 0 && s.treeReady {
		treeH := s.height - 8
		if s.filterOn {
			treeH = s.height - 9
		}
		if treeH < 1 {
			treeH = 1
		}
		pageNum := int(s.treeViewport.YOffset)/treeH + 1
		totalPages := (len(s.visibleRows) + treeH - 1) / treeH
		pageInfo = fmt.Sprintf(" %d items | Page %d/%d |", len(s.visibleRows), pageNum, totalPages)
	}
	return fmt.Sprintf("%s j/k:move | Enter:open | /:filter | ::cmd | ?:help | Tab:nav", pageInfo)
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
	s.scrollCursorIntoView()
}

func (s *taskListModel) scrollCursorIntoView() {
	if !s.treeReady || len(s.visibleRows) == 0 {
		return
	}
	cursor := s.cursor
	offset := int(s.treeViewport.YOffset)
	height := s.treeViewport.Height
	if cursor < offset {
		s.treeViewport.YOffset = cursor
	} else if cursor >= offset+height {
		s.treeViewport.YOffset = cursor - height + 1
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

		// Collect level-1 children with their grandchildren
		type childGroup struct {
			task     *model.Task
			grandchildren []*model.Task
		}
		var groups []childGroup
		for _, child := range children {
			if child.Type == model.TypeStory {
				var gc []*model.Task
				for _, t := range children {
					if t.Parent == child.ID && t.ID != child.ID {
						gc = append(gc, t)
					}
				}
				sort.Slice(gc, func(i, j int) bool {
					di := gc[i].Status == model.StatusDone
					dj := gc[j].Status == model.StatusDone
					if di != dj {
						return !di
					}
					return gc[i].ID < gc[j].ID
				})
				groups = append(groups, childGroup{task: child, grandchildren: gc})
			}
		}
		for _, child := range children {
			if !seen[child.ID] && child.Type != model.TypeStory {
				groups = append(groups, childGroup{task: child})
			}
		}

		for gi, g := range groups {
			isLast := gi == len(groups)-1
			p := "├─"
			if isLast {
				p = "└─"
			}
			seen[g.task.ID] = true
			rows = append(rows, flatRow{task: g.task, prefix: p, level: 1, epicID: ep.ID})

			for ci, gc := range g.grandchildren {
				gcPrefix := "├─"
				if ci == len(g.grandchildren)-1 {
					gcPrefix = "└─"
				}
				// Include ancestor pipe if parent is not the last sibling
				if !isLast {
					gcPrefix = "│ " + gcPrefix
				} else {
					gcPrefix = "  " + gcPrefix
				}
				seen[gc.ID] = true
				rows = append(rows, flatRow{task: gc, prefix: gcPrefix, level: 2, epicID: ep.ID})
			}
		}
	}

	// Stories without epic
	for _, t := range tasks {
		if !seen[t.ID] && t.Type == model.TypeStory {
			seen[t.ID] = true
			rows = append(rows, flatRow{task: t, prefix: "├─", level: 1, epicID: ""})
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
			for ci, child := range storyChildren {
				gcPrefix := "├─"
				if ci == len(storyChildren)-1 {
					gcPrefix = "└─"
				}
				seen[child.ID] = true
				rows = append(rows, flatRow{task: child, prefix: gcPrefix, level: 2, epicID: ""})
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
