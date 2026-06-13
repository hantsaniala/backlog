package app

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type sidebarItem int

const (
	sidebarDashboard sidebarItem = iota
	sidebarTasks
	sidebarSprints
)

const sidebarWidth = 22

type sidebarModel struct {
	active sidebarItem
	height int
}

func newSidebar() *sidebarModel {
	return &sidebarModel{active: sidebarDashboard, height: 24}
}

func (s *sidebarModel) SetActive(item sidebarItem) {
	s.active = item
}

func (s *sidebarModel) SetHeight(h int) {
	s.height = h
}

func (s *sidebarModel) View() string {
	items := []struct {
		item  sidebarItem
		label string
	}{
		{sidebarDashboard, "Dashboard"},
		{sidebarTasks, "Tasks"},
		{sidebarSprints, "Sprints"},
	}

	base := lipgloss.NewStyle().Width(sidebarWidth).Background(colorSurface)
	dim := lipgloss.NewStyle().Foreground(colorTextDim)
	activeDot := lipgloss.NewStyle().Foreground(colorPrimary)
	activeLabel := lipgloss.NewStyle().Bold(true).Foreground(colorTextBright)

	var lines []string
	for _, it := range items {
		var line string
		if it.item == s.active {
			dot := activeDot.Render("●")
			label := activeLabel.Render(it.label)
			line = "  " + dot + " " + label
		} else {
			dot := dim.Render("○")
			label := dim.Render(it.label)
			line = "  " + dot + " " + label
		}
		lines = append(lines, base.Render(line))
	}

	for len(lines) < s.height {
		lines = append(lines, base.Render(""))
	}

	return strings.Join(lines, "\n")
}

func screenToSidebar(s screen) sidebarItem {
	switch s {
	case screenDashboard:
		return sidebarDashboard
	case screenTaskList:
		return sidebarTasks
	case screenSprintView:
		return sidebarSprints
	default:
		return sidebarDashboard
	}
}
