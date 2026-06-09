package app

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	colorPrimary    = lipgloss.Color("#7C3AED")
	colorSecondary  = lipgloss.Color("#06B6D4")
	colorSuccess    = lipgloss.Color("#10B981")
	colorWarning    = lipgloss.Color("#F59E0B")
	colorError      = lipgloss.Color("#EF4444")
	colorInfo       = lipgloss.Color("#3B82F6")
	colorAccent     = lipgloss.Color("#A78BFA")

	colorBg         = lipgloss.Color("#1E1E2E")
	colorSurface    = lipgloss.Color("#2D2D44")
	colorSurfaceAlt = lipgloss.Color("#25253D")
	colorText       = lipgloss.Color("#E2E8F0")
	colorTextDim    = lipgloss.Color("#64748B")
	colorTextBright = lipgloss.Color("#F8FAFC")
	colorBorder     = lipgloss.Color("#3D3D5C")

	footerStyle = lipgloss.NewStyle().
			Height(1).
			Foreground(colorTextDim).
			PaddingLeft(1)

	headerStyle = lipgloss.NewStyle().
			Foreground(colorTextBright).
			Bold(true).
			Padding(0, 1)

	tabActiveStyle = lipgloss.NewStyle().
			Foreground(colorTextBright).
			Background(colorPrimary).
			Bold(true).
			Padding(0, 2)

	tabInactiveStyle = lipgloss.NewStyle().
				Foreground(colorTextDim).
				Background(colorSurface).
				Padding(0, 2)

	statusColors = map[string]lipgloss.Color{
		"todo":        colorInfo,
		"in-progress": colorWarning,
		"review":      colorSecondary,
		"on-hold":     colorTextDim,
		"done":        colorSuccess,
		"cancelled":   colorError,
	}

	statusGlyph = map[string]string{
		"todo":        "◌",
		"in-progress": "◎",
		"review":      "◐",
		"on-hold":     "◷",
		"done":        "●",
		"cancelled":   "⊗",
	}
)

func StatusBadge(status string) string {
	c, ok := statusColors[status]
	if !ok {
		c = colorTextDim
	}
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#1E1E2E")).
		Background(c).
		Bold(true).
		Padding(0, 1).
		Render(status)
}

func statusDot(status string) string {
	c, ok := statusColors[status]
	if !ok {
		c = colorTextDim
	}
	g, ok := statusGlyph[status]
	if !ok {
		g = "?"
	}
	return lipgloss.NewStyle().Foreground(c).Render(g)
}

func priorityDot(priority string) string {
	var c lipgloss.Color
	switch priority {
	case "critical":
		c = colorError
	case "high":
		c = colorWarning
	case "medium":
		c = colorInfo
	default:
		c = colorTextDim
	}
	return lipgloss.NewStyle().Foreground(c).Render("◆")
}

func typeDot(t string) string {
	var c lipgloss.Color
	switch t {
	case "bug":
		c = colorError
	case "story":
		c = colorSecondary
	case "spike":
		c = colorWarning
	case "chore":
		c = colorTextDim
	default:
		c = colorPrimary
	}
	return lipgloss.NewStyle().Foreground(c).Render("●")
}

func sectionHeader(title string, color lipgloss.Color, width int) string {
	line := lipgloss.NewStyle().Foreground(color).Render(strings.Repeat("─", width-2))
	return fmt.Sprintf(" %s %s %s", lipgloss.NewStyle().Foreground(color).Render("┃"), title, line)
}

func progressBar(filled, total int, width int) string {
	if total <= 0 {
		total = 1
	}
	if filled > total {
		filled = total
	}
	ratio := float64(filled) / float64(total)
	fillChars := int(ratio * float64(width))
	if fillChars > width {
		fillChars = width
	}
	var c lipgloss.Color
	switch {
	case ratio >= 0.9:
		c = colorSuccess
	case ratio >= 0.5:
		c = colorWarning
	default:
		c = colorInfo
	}
	bar := strings.Repeat("█", fillChars) + strings.Repeat("░", width-fillChars)
	colored := lipgloss.NewStyle().Foreground(c).Render(bar)
	return fmt.Sprintf("%s %d/%d", colored, filled, total)
}

func ExternalBadge() string {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#1E1E2E")).
		Background(colorTextDim).
		Padding(0, 1).
		Render("ext")
}
