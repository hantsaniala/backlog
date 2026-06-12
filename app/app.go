package app

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/hantsaniala/backlog/config"
	"github.com/hantsaniala/backlog/model"
)

type screen int

const (
	screenDashboard screen = iota
	screenTaskList
	screenSprintView
)

// mouseState tracks the latest mouse position for hover indicators.
type mouseState struct {
	X, Y    int
	visible bool // true if last event was a mouse move
}

type Model struct {
	backlog       *model.Backlog
	currentScreen screen
	screens       map[screen]tea.Model
	width         int
	height        int
	reloading     bool

	inputMode InputMode

	// Overlay models
	palette   *paletteModel
	helpModel *helpModel
	showHelp  bool

	// Notification
	notification    string
	notificationAge int

	// Editor integration
	editorCmd string
	nvimMode  bool

	// Configuration
	conf *config.Config

	// Mouse state for hover/click
	mouse mouseState
}

func New(b *model.Backlog, cfg *config.Config) *Model {
	if cfg == nil {
		cfg = config.Default()
	}

	ApplyTheme(&cfg.Theme)

	m := &Model{
		backlog:       b,
		currentScreen: screenDashboard,
		screens:       make(map[screen]tea.Model),
		inputMode:     ModeNormal,
		helpModel:     newHelpModel(),
		conf:          cfg,
	}

	m.screens[screenDashboard] = newScreenDashboard(b)
	m.screens[screenTaskList] = newScreenTaskList(b)
	m.screens[screenSprintView] = newScreenSprintView(b)

	if os.Getenv("NVIM") != "" || os.Getenv("NVIM_LISTEN_ADDRESS") != "" {
		m.nvimMode = true
		m.editorCmd = "nvim --remote-send"
	}

	return m
}

func (m *Model) SetEditor(cmd string) {
	m.editorCmd = cmd
	if tl, ok := m.screens[screenTaskList].(*taskListModel); ok {
		tl.editorCmd = cmd
	}
}

func (m *Model) Init() tea.Cmd {
	return watchBacklogDirectories(m.backlog)
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
		// Clear mouse hover on keyboard activity
		m.mouse.visible = false

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

		switch {
		case key.Matches(msg, NormalKeys.Quit):
			return m, tea.Quit
		case key.Matches(msg, NormalKeys.Command):
			m.palette = newPaletteModel(m.width, m.height)
			m.inputMode = ModeCommandPalette
			return m, m.palette.Init()
		case key.Matches(msg, NormalKeys.One):
			m.currentScreen = screenDashboard
			return m, nil
		case key.Matches(msg, NormalKeys.Two):
			m.currentScreen = screenTaskList
			return m, nil
		case key.Matches(msg, NormalKeys.Three):
			m.currentScreen = screenSprintView
			return m, nil
		}

	case tea.MouseMsg:
		m.mouse = mouseState{X: msg.X, Y: msg.Y, visible: msg.Action == tea.MouseActionMotion}
		// Sprint view click-to-focus: left click on a panel header shifts focus.
		if m.currentScreen == screenSprintView && msg.Action == tea.MouseActionPress && msg.Button == tea.MouseButtonLeft {
			// Header = 2 lines, title block = ~4 lines, then panels start.
			panelY := msg.Y - 6
			if panelY >= 0 {
				panelW := (m.width - 4) / 2
				if panelW < 30 {
					panelW = 30
				}
				// Left panel occupies columns 3..(3+panelW), gap "  ", right panel after.
				leftEnd := 3 + panelW
				gapEnd := leftEnd + 2
				if sv, ok := m.screens[screenSprintView].(*sprintViewModel); ok {
					if msg.X >= 3 && msg.X < leftEnd {
						sv.focus = paneLeft
					} else if msg.X >= gapEnd && msg.X < gapEnd+panelW {
						sv.focus = paneRight
					}
				}
			}
		}

	case reloadMsg:
		m.reloading = false
		for _, s := range m.screens {
			if u, ok := s.(tea.Model); ok {
				u.Update(reloadMsg{})
			}
		}
		if m.conf != nil && m.conf.Git.AutoCommit && m.backlog != nil {
			root := filepath.Dir(m.backlog.Current.Root)
			go func() {
				model.GitCommit(root, m.conf.Git.CommitPrefix)
			}()
		}
		return m, tea.Batch(watchBacklogDirectories(m.backlog))

	case reloadErrorMsg:
		m.reloading = false
		return m, nil

	case notificationMsg:
		m.setNotification(msg.text)
		return m, nil

	case editorOpenMsg:
		path := taskFilePath(msg.task, msg.root)
		cmd := exec.Command("sh", "-c", fmt.Sprintf("%s %s &", msg.editor, path))
		if err := cmd.Start(); err != nil {
			m.setNotification(fmt.Sprintf("Failed to open editor: %v", err))
		} else {
			m.setNotification(fmt.Sprintf("Opened %s in editor", msg.task.ID))
		}
		return m, nil
	}

	if s, ok := m.screens[m.currentScreen]; ok {
		updated, cmd := s.Update(msg)
		m.screens[m.currentScreen] = updated
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
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
	case cmd == "git commit":
		if m.backlog != nil {
			root := filepath.Dir(m.backlog.Current.Root)
			prefix := "feat(backlog):"
			if m.conf != nil {
				prefix = m.conf.Git.CommitPrefix
			}
			if err := model.GitCommit(root, prefix); err != nil {
				m.setNotification("Git commit failed: " + err.Error())
			} else {
				m.setNotification("Changes committed")
				m.backlog.Git = model.GetGitState(root)
			}
		}
	case cmd == "git push":
		if m.backlog != nil {
			root := filepath.Dir(m.backlog.Current.Root)
			if err := model.GitPush(root); err != nil {
				m.setNotification("Git push failed: " + err.Error())
			} else {
				m.setNotification("Pushed to remote")
			}
		}
	case cmd == "git log":
		if m.backlog != nil {
			root := filepath.Dir(m.backlog.Current.Root)
			logOutput, err := model.GitLog(root)
			if err != nil {
				m.setNotification("Git log failed: " + err.Error())
			} else {
				m.setNotification("Recent commits:\n" + logOutput)
			}
		}
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

	b.WriteString(m.renderHeader())
	b.WriteString("\n")

	if m.showHelp {
		b.WriteString(m.helpModel.View(m.width, m.height))
		return b.String()
	}

	if m.inputMode == ModeCommandPalette && m.palette != nil {
		b.WriteString(m.palette.View())
		return b.String()
	}

	if s, ok := m.screens[m.currentScreen]; ok {
		b.WriteString(s.View())
	}

	b.WriteString("\n")
	b.WriteString(m.renderFooter())

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
	var leftParts, rightParts []string

	// Mode indicator
	modeLabel := ModeStyle(m.inputMode).Render(fmt.Sprintf(" %s ", m.inputMode.String()))
	leftParts = append(leftParts, modeLabel)

	// Tab bar
	tabs := []string{
		renderTab("1 Dashboard", m.currentScreen == screenDashboard),
		renderTab("2 Tasks", m.currentScreen == screenTaskList),
		renderTab("3 Sprints", m.currentScreen == screenSprintView),
	}
	leftParts = append(leftParts, strings.Join(tabs, ""))

	// Right side
	if m.nvimMode {
		rightParts = append(rightParts, lipgloss.NewStyle().
			Foreground(colorSuccess).
			Padding(0, 1).
			Render("[nvim]"))
	}
	if m.backlog != nil && m.backlog.Git != nil && m.backlog.Git.Branch != "" {
		branchLabel := m.backlog.Git.Branch
		if m.backlog.Git.Dirty {
			branchLabel += " *"
		}
		rightParts = append(rightParts, lipgloss.NewStyle().
			Foreground(colorSecondary).
			Padding(0, 1).
			Render(branchLabel))
	}
	rightParts = append(rightParts, timeStyle.Render(formatTime()))

	left := strings.Join(leftParts, " ")
	right := strings.Join(rightParts, "  ")
	avail := m.width - len(left) - len(right)
	if avail < 0 {
		avail = 0
	}

	return lipgloss.NewStyle().
		Background(colorSurface).
		Padding(0, 1).
		Render(left + strings.Repeat(" ", avail) + right)
}

func renderTab(label string, active bool) string {
	if active {
		return tabActiveStyle.Render(" " + label + " ")
	}
	return tabInactiveStyle.Render(" " + label + " ")
}

func (m *Model) renderFooter() string {
	if m.reloading {
		return lipgloss.NewStyle().
			Foreground(colorTextDim).
			Padding(0, 2).
			Render(" Loading...")
	}

	var hints string
	if s, ok := m.screens[m.currentScreen]; ok {
		type hintProvider interface{ footerHint() string }
		if hp, ok := s.(hintProvider); ok {
			hints = hp.footerHint()
		}
	}
	if hints == "" {
		hints = m.contextualHints()
	}

	posStr := ""
	if s, ok := m.screens[m.currentScreen]; ok {
		type posProvider interface{ footerPos() string }
		if pp, ok := s.(posProvider); ok {
			posStr = pp.footerPos()
		}
	}

	footerW := m.width - len(hints) - 4
	if footerW < 0 {
		footerW = 0
	}
	posFmt := lipgloss.NewStyle().Foreground(colorTextDim).Render(posStr)

	return lipgloss.NewStyle().
		Foreground(colorTextDim).
		Padding(0, 2).
		Render(hints + strings.Repeat(" ", footerW) + posFmt)
}

func (m *Model) contextualHints() string {
	switch m.inputMode {
	case ModeNormal:
		return " j/k:move | Enter:open | e:status | v:select | /:filter | ::cmd | ?:help | q:quit"
	case ModeInsert:
		return " Type filter | Esc:cancel | Enter:confirm"
	case ModeVisual:
		return " j/k:extend | Space:toggle | a:all | x:action | Esc:cancel"
	case ModeCommandPalette:
		return " Type command | Enter:execute | Esc:cancel"
	case ModeHelp:
		return " Press ? or Esc to close"
	default:
		return " ?:help | q:quit"
	}
}
