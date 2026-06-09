package app

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	// Colors
	colorPrimary   = lipgloss.Color("#7C3AED") // purple
	colorSecondary = lipgloss.Color("#06B6D4") // cyan
	colorSuccess   = lipgloss.Color("#10B981") // green
	colorWarning   = lipgloss.Color("#F59E0B") // amber
	colorError     = lipgloss.Color("#EF4444") // red
	colorInfo      = lipgloss.Color("#3B82F6") // blue

	colorBg        = lipgloss.Color("#1E1E2E")
	colorSurface   = lipgloss.Color("#2D2D44")
	colorText      = lipgloss.Color("#E2E8F0")
	colorTextDim   = lipgloss.Color("#64748B")
	colorTextBright = lipgloss.Color("#F8FAFC")
	colorBorder    = lipgloss.Color("#3D3D5C")

	// Layout
	appStyle = lipgloss.NewStyle().
		Margin(0, 0).
		Padding(1, 2)

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

	// Status badge colors
	statusColors = map[string]lipgloss.Color{
		"todo":        colorInfo,
		"in-progress": colorWarning,
		"review":      colorSecondary,
		"on-hold":     colorTextDim,
		"done":        colorSuccess,
		"cancelled":   colorError,
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

func PriorityBadge(priority string) string {
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
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#1E1E2E")).
		Background(c).
		Padding(0, 1).
		Render(priority)
}

func TypeBadge(t string) string {
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
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#1E1E2E")).
		Background(c).
		Padding(0, 1).
		Render(t)
}

func ExternalBadge() string {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#1E1E2E")).
		Background(colorTextDim).
		Padding(0, 1).
		Render("external")
}
