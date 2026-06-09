package app

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/help"
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

type Model struct {
	backlog       *model.Backlog
	currentScreen screen
	screens       map[screen]tea.Model
	help          help.Model
	showHelp      bool
	width         int
	height        int
	reloading     bool
	tabNames      []string
}

func New(b *model.Backlog) *Model {
	m := &Model{
		backlog:       b,
		currentScreen: screenDashboard,
		screens:       make(map[screen]tea.Model),
		help:          help.New(),
		tabNames:      []string{"1 Dashboard", "2 Tasks", "3 Sprints"},
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
		m.help.Width = msg.Width
		for _, s := range m.screens {
			if u, ok := s.(tea.Model); ok {
				u.Update(msg)
			}
		}

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, Keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, Keys.Help):
			m.showHelp = !m.showHelp
			return m, nil
		case key.Matches(msg, Keys.Reload):
			m.reloading = true
			return m, func() tea.Msg {
				err := m.backlog.Reload()
				if err != nil {
					return reloadErrorMsg{err}
				}
				return reloadMsg{}
			}
		case key.Matches(msg, Keys.One):
			m.currentScreen = screenDashboard
			return m, nil
		case key.Matches(msg, Keys.Two):
			m.currentScreen = screenTaskList
			return m, nil
		case key.Matches(msg, Keys.Three):
			m.currentScreen = screenSprintView
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

	if s, ok := m.screens[m.currentScreen]; ok {
		updated, cmd := s.Update(msg)
		m.screens[m.currentScreen] = updated
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) View() string {
	if m.reloading {
		return lipgloss.NewStyle().Foreground(colorWarning).Render("  Reloading...")
	}

	var b strings.Builder

	b.WriteString(m.renderTabs())
	b.WriteString("\n")

	if m.showHelp {
		b.WriteString(m.help.View(Keys))
		return b.String()
	}

	if s, ok := m.screens[m.currentScreen]; ok {
		b.WriteString(s.View())
	}

	b.WriteString("\n")
	footer := fmt.Sprintf("  %s | ? help | q quit | r reload",
		m.tabNames[m.currentScreen])
	b.WriteString(footerStyle.Render(footer))

	return b.String()
}

func (m *Model) renderTabs() string {
	var b strings.Builder
	for i, name := range m.tabNames {
		s := screen(i)
		if s == m.currentScreen {
			b.WriteString(tabActiveStyle.Render(name))
		} else {
			b.WriteString(tabInactiveStyle.Render(name))
		}
	}
	if m.reloading {
		b.WriteString(tabInactiveStyle.Render(" \u21BB "))
	}
	return b.String()
}
