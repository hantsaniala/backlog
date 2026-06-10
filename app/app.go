package app

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hantsaniala/backlog/model"
)

type screen int

const (
	screenDashboard screen = iota
	screenTaskList
	screenSprintView
)

type panelFocus int

const (
	panelMain  panelFocus = iota
	panelSidebar
)

type Model struct {
	backlog       *model.Backlog
	currentScreen screen
	screens       map[screen]tea.Model
	width         int
	height        int
	reloading     bool
	tabNames      []string

	// Navigation
	inputMode     InputMode
	navStack      *NavigationStack
	ctrlWPending  bool
	currentPanel  panelFocus
	panelMaximized bool
	savedPanels   []panelFocus

	// Overlay models
	palette    *paletteModel
	helpModel  *helpModel
	showHelp   bool
	sidebar    *sidebarState

	// Notification
	notification     string
	notificationAge  int
}

func New(b *model.Backlog) *Model {
	m := &Model{
		backlog:       b,
		currentScreen: screenDashboard,
		screens:       make(map[screen]tea.Model),
		tabNames:      []string{"1 Dashboard", "2 Tasks", "3 Sprints"},
		inputMode:     ModeNormal,
		navStack:      NewNavigationStack(),
		savedPanels:   make([]panelFocus, 0),
		sidebar:       newSidebarState(),
		helpModel:     newHelpModel(),
	}

	m.screens[screenDashboard] = newScreenDashboard(b)
	m.screens[screenTaskList] = newScreenTaskList(b)
	m.screens[screenSprintView] = newScreenSprintView(b)

	return m
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(
		watchBacklogDirectories(m.backlog),
	)
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if m.palette != nil {
			m.palette.width = msg.Width
			m.palette.height = msg.Height
		}
		for _, s := range m.screens {
			if u, ok := s.(tea.Model); ok {
				u.Update(msg)
			}
		}

	case paletteExecuteMsg:
		m.handlePaletteCommand(msg.command)
		return m, nil

	case tea.KeyMsg:
		// Handle Ctrl+w prefix state machine
		if m.ctrlWPending {
			m.ctrlWPending = false
			return m.handleCtrlW(msg)
		}

		// Handle palette overlay first
		if m.inputMode == ModeCommandPalette && m.palette != nil {
			updated, cmd := m.palette.Update(msg)
			m.palette = updated.(*paletteModel)
			if result, ok := <-catchMsg(cmd); ok {
				if exec, ok := result.(paletteExecuteMsg); ok {
					m.handlePaletteCommand(exec.command)
				}
				m.inputMode = ModeNormal
				m.palette = nil
			}
			return m, cmd
		}

		// Help overlay toggle
		if m.inputMode == ModeHelp {
			if msg.String() == "esc" || msg.String() == "?" {
				m.inputMode = ModeNormal
				m.showHelp = false
			}
			return m, nil
		}
		if m.inputMode == ModeNormal && key.Matches(msg, NormalKeys.Help) {
			m.showHelp = true
			m.inputMode = ModeHelp
			return m, nil
		}

		// Global keys
		switch {
		case key.Matches(msg, NormalKeys.Quit):
			return m, tea.Quit
		case key.Matches(msg, NormalKeys.Command):
			m.palette = newPaletteModel(m.width, m.height)
			m.inputMode = ModeCommandPalette
			return m, m.palette.Init()
		case key.Matches(msg, NormalKeys.One):
			m.navStack.Push(ViewState{Screen: m.currentScreen})
			m.currentScreen = screenDashboard
			m.inputMode = ModeNormal
			return m, nil
		case key.Matches(msg, NormalKeys.Two):
			m.navStack.Push(ViewState{Screen: m.currentScreen})
			m.currentScreen = screenTaskList
			m.inputMode = ModeNormal
			return m, nil
		case key.Matches(msg, NormalKeys.Three):
			m.navStack.Push(ViewState{Screen: m.currentScreen})
			m.currentScreen = screenSprintView
			m.inputMode = ModeNormal
			return m, nil
		case key.Matches(msg, NormalKeys.HistBack):
			if prev, ok := m.navStack.Pop(); ok {
				m.currentScreen = prev.Screen
				return m, nil
			}
			return m, nil
		case key.Matches(msg, NormalKeys.HistFwd):
			// Re-push current and try next if available
			return m, nil
		case key.Matches(msg, NormalKeys.Preview):
			m.sidebar.toggle()
			return m, nil
		case msg.String() == "ctrl+p" && m.sidebar.open:
			m.sidebar.pin()
			return m, nil
		case key.Matches(msg, NormalKeys.PanelLeft):
			// Send panel focus message to current screen
			return m, nil
		case msg.String() == "ctrl+w":
			m.ctrlWPending = true
			return m, nil
		}

	case reloadMsg:
		m.reloading = false
		for _, s := range m.screens {
			if u, ok := s.(tea.Model); ok {
				u.Update(reloadMsg{})
			}
		}
		return m, tea.Batch(watchBacklogDirectories(m.backlog))

	case reloadErrorMsg:
		m.reloading = false
		return m, nil
	}

	// Route msg to current screen
	if s, ok := m.screens[m.currentScreen]; ok {
		updated, cmd := s.Update(msg)
		m.screens[m.currentScreen] = updated
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) handleCtrlW(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "h":
		m.currentPanel = panelMain
	case "l":
		m.currentPanel = panelSidebar
	case "q":
		if m.currentPanel == panelSidebar {
			m.sidebar.open = false
		}
	case "o":
		m.panelMaximized = true
	case "r":
		m.panelMaximized = false
	}
	return m, nil
}

func (m *Model) handlePaletteCommand(cmd string) {
	m.setNotification("Executed: " + cmd)
	switch {
	case cmd == "focus dashboard":
		m.currentScreen = screenDashboard
	case cmd == "focus backlog":
		m.currentScreen = screenTaskList
	case cmd == "focus sprints":
		m.currentScreen = screenSprintView
	}
}

func (m *Model) setNotification(msg string) {
	m.notification = msg
	m.notificationAge = 0
}

func (m *Model) View() string {
	if m.reloading {
		return lipgloss.NewStyle().Foreground(colorWarning).Render("  Reloading...")
	}

	var b strings.Builder

	// Global header bar
	b.WriteString(m.renderHeader())
	b.WriteString("\n")

	// Help overlay
	if m.showHelp {
		b.WriteString(m.helpModel.View(m.width, m.height))
		return b.String()
	}

	// Command palette overlay
	if m.inputMode == ModeCommandPalette && m.palette != nil {
		b.WriteString(m.palette.View())
		return b.String()
	}

	// Main content with optional sidebar
	if m.sidebar.open {
		sidebarW := m.width * 30 / 100
		if sidebarW > 50 {
			sidebarW = 50
		}
		mainContent := ""
		if s, ok := m.screens[m.currentScreen]; ok {
			mainContent = s.View()
		}

		sideContent := m.sidebar.render(sidebarW)
		b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, mainContent, sideContent))
		b.WriteString("\n")
	} else {
		if s, ok := m.screens[m.currentScreen]; ok {
			b.WriteString(s.View())
		}
	}

	b.WriteString("\n")

	// Global footer
	b.WriteString(m.renderFooter())

	// Notification
	if m.notification != "" {
		m.notificationAge++
		if m.notificationAge < 30 {
			notifLine := notificationStyle.Render(fmt.Sprintf("  %s", m.notification))
			b.WriteString("\n")
			b.WriteString(notifLine)
		} else {
			m.notification = ""
			m.notificationAge = 0
		}
	}

	return b.String()
}

func (m *Model) renderHeader() string {
	var parts []string

	// Breadcrumb
	bc := m.navStack.Breadcrumb(m.tabNames)
	current := breadcrumbFromScreen(m.currentScreen)
	if bc != "" {
		parts = append(parts, breadcrumbStyle.Render(bc+navArrowStyle+current))
	} else {
		parts = append(parts, breadcrumbActiveStyle.Render(current))
	}

	// Mode indicator
	modeLabel := ModeStyle(m.inputMode).Render(fmt.Sprintf(" %s ", m.inputMode.String()))
	parts = append(parts, modeLabel)

	// Time
	parts = append(parts, timeStyle.Render(formatTime()))

	// Connection dot
	parts = append(parts, connectionDotStyle.Render(""))

	// Notification badge from palette
	// (handled in footer via notification field)

	return lipgloss.NewStyle().
		Background(colorSurface).
		Padding(0, 1).
		Render(strings.Join(parts, " "))
}

func (m *Model) renderFooter() string {
	hints := m.contextualHints()
	return lipgloss.NewStyle().
		Foreground(colorTextDim).
		Padding(0, 2).
		Render(hints)
}

func (m *Model) contextualHints() string {
	switch m.inputMode {
	case ModeNormal:
		return fmt.Sprintf("%s | j/k:move []:page Enter:drill /:search f:jump v:select ::cmd ?:help q:quit",
			m.tabNames[m.currentScreen])
	case ModeInsert:
		return "Type filter | Esc:cancel Tab:next field Enter:confirm"
	case ModeVisual:
		return "j/k:extend Space:toggle a:all x:action Esc:cancel"
	case ModeCommandPalette:
		return "Type command | Enter:execute Esc:cancel"
	case ModeHelp:
		return "Press ? or Esc to close"
	default:
		return m.tabNames[m.currentScreen] + " | ? help | q quit"
	}
}
