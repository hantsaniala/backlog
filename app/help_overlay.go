package app

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type helpItem struct {
	key         string
	description string
}

func newHelpModel() *helpModel {
	return &helpModel{}
}

type helpModel struct{}

func (h *helpModel) categories() [][]helpItem {
	return [][]helpItem{
		{
			{key: "j/k", description: "Move up/down"},
			{key: "h/l", description: "Collapse/expand"},
			{key: "Enter", description: "Open / confirm"},
			{key: "Esc", description: "Back / cancel"},
			{key: "Space", description: "Toggle status"},
			{key: "q", description: "Quit"},
		},
		{
			{key: "1/2/3", description: "Switch screens"},
			{key: "/", description: "Filter tasks"},
			{key: ":", description: "Command palette"},
			{key: "e", description: "Cycle status"},
			{key: "o", description: "Open in editor"},
			{key: "v", description: "Visual select"},
			{key: "f", description: "Jump hints"},
		},
		{
			{key: "gg / G", description: "Top / bottom"},
			{key: "C-d / C-u", description: "Half page"},
			{key: "C-f / C-b", description: "Full page"},
			{key: "zz/zt/zb", description: "Center/top/bottom"},
			{key: "n / N", description: "Next/prev match"},
			{key: "{ / }", description: "Prev/next epic"},
			{key: "m / '", description: "Set / jump mark"},
			{key: "* / #", description: "Search word"},
		},
	}
}

func (h *helpModel) View(width, height int) string {
	var b strings.Builder

	b.WriteString(helpCategoryStyle.Render(" Keyboard Shortcuts"))
	b.WriteString("\n\n")

	cols := 3

	var allLines []string
	for _, cat := range h.categories() {
		allLines = append(allLines, "")
		for _, item := range cat {
			allLines = append(allLines, fmt.Sprintf("  %s  %s",
				helpKeyStyle.Render(item.key),
				helpDescStyle.Render(item.description)))
		}
	}

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
