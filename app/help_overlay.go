package app

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type helpCategory struct {
	name  string
	items []helpItem
}

type helpItem struct {
	key         string
	description string
}

func newHelpModel() *helpModel {
	return &helpModel{}
}

type helpModel struct{}

func (h *helpModel) categories() []helpCategory {
	return []helpCategory{
		{
			name: "Global",
			items: []helpItem{
				{key: "q / C-c", description: "Quit application"},
				{key: "?", description: "Toggle help overlay"},
				{key: ":", description: "Open command palette"},
				{key: "/", description: "Search / filter"},
				{key: "r", description: "Related items"},
			},
		},
		{
			name: "Navigation",
			items: []helpItem{
				{key: "h/j/k/l", description: "Move left/down/up/right"},
				{key: "←/↓/↑/→", description: "Arrow keys (same)"},
				{key: "gg", description: "Jump to top"},
				{key: "G", description: "Jump to bottom"},
				{key: "C-d / C-u", description: "Half page down/up"},
				{key: "C-f / C-b", description: "Full page down/up"},
				{key: "zz", description: "Center cursor on screen"},
				{key: "f", description: "Jump hints (EasyMotion)"},
				{key: "Tab / S-Tab", description: "Next/prev panel"},
			},
		},
		{
			name: "Actions",
			items: []helpItem{
				{key: "Enter", description: "Open detail / confirm"},
				{key: "Esc / Bksp", description: "Back / cancel"},
				{key: "Space", description: "Expand/collapse / toggle"},
				{key: "s", description: "Cycle status"},
				{key: "> / <", description: "Move to/from sprint"},
				{key: "v", description: "Visual mode (multi-select)"},
			},
		},
		{
			name: "Panel Management",
			items: []helpItem{
				{key: "C-w h", description: "Focus panel left"},
				{key: "C-w j", description: "Focus panel down"},
				{key: "C-w k", description: "Focus panel up"},
				{key: "C-w l", description: "Focus panel right"},
				{key: "C-w q", description: "Close panel"},
				{key: "C-w o", description: "Maximize panel"},
				{key: "C-w r", description: "Restore panel layout"},
			},
		},
		{
			name: "Screens & History",
			items: []helpItem{
				{key: "1", description: "Dashboard"},
				{key: "2", description: "Task list / Backlog"},
				{key: "3", description: "Sprint view"},
				{key: "C-p", description: "Toggle preview sidebar"},
				{key: "M-← / M-→", description: "History back/forward"},
			},
		},
	}
}

func (h *helpModel) View(width, height int) string {
	var b strings.Builder

	b.WriteString(helpCategoryStyle.Render(" Keyboard Shortcuts"))
	b.WriteString("\n\n")

	cols := width / 30
	if cols < 2 {
		cols = 2
	}
	if cols > 3 {
		cols = 3
	}

	var allLines []string
	for _, cat := range h.categories() {
		allLines = append(allLines, "")
		allLines = append(allLines, fmt.Sprintf("  %s", cat.name))
		for _, item := range cat.items {
			allLines = append(allLines, fmt.Sprintf("    %s  %s",
				helpKeyStyle.Render(item.key),
				helpDescStyle.Render(item.description)))
		}
		allLines = append(allLines, "")
	}

	// Split into columns
	colH := (len(allLines) + cols - 1) / cols
	for i := 0; i < colH; i++ {
		for c := 0; c < cols; c++ {
			idx := c*colH + i
			if idx < len(allLines) {
				b.WriteString(fmt.Sprintf("%-30s", allLines[idx]))
			}
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(paletteDescStyle.Render(" Press ? or Esc to close"))

	modalW := width*80/100 - 4
	if modalW < 60 {
		modalW = 60
	}

	content := lipgloss.NewStyle().
		Width(modalW).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorInfo).
		Padding(1, 2).
		Render(b.String())

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, content)
}
