package app

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/hantsaniala/backlog/config"
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

	colorJumpHintBg  = lipgloss.Color("#000000")
	colorJumpHintFg  = lipgloss.Color("#FFD700")
	colorOverlayBg   = lipgloss.Color("#1E1E2E")

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

	popupStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorPrimary).
			Padding(1, 2).
			Background(colorSurface)

	overlayStyle = lipgloss.NewStyle().
			Background(colorBg)

	focusedRowStyle = lipgloss.NewStyle().
			Background(colorSurfaceAlt)

	leftBorderBar = lipgloss.NewStyle().
			Foreground(colorPrimary).
			Render("▎")

	// ---- NEW GLOBAL NAV STYLES ----

	headerBarStyle = lipgloss.NewStyle().
			Background(colorSurface).
			Foreground(colorTextBright).
			Padding(0, 1).
			Width(160)

	modeNormalStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#2D2D44")).
				Foreground(colorSuccess).
				Bold(true).
				Padding(0, 2)

	modeInsertStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#2D2D44")).
				Foreground(colorWarning).
				Bold(true).
				Padding(0, 2)

	modeVisualStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#2D2D44")).
				Foreground(colorAccent).
				Bold(true).
				Padding(0, 2)

	modeHelpStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#2D2D44")).
			Foreground(colorInfo).
			Bold(true).
			Padding(0, 2)

	breadcrumbStyle = lipgloss.NewStyle().
			Foreground(colorTextDim).
			Padding(0, 1)

	breadcrumbActiveStyle = lipgloss.NewStyle().
				Foreground(colorTextBright).
				Bold(true).
				Padding(0, 1)

	navArrowStyle = lipgloss.NewStyle().
			Foreground(colorTextDim).
			Render(" › ")

	connectionDotStyle = lipgloss.NewStyle().
				Foreground(colorSuccess)

	timeStyle = lipgloss.NewStyle().
			Foreground(colorTextDim).
			Padding(0, 2)

	// ---- JUMP HINTS ----

	jumpHintStyle = lipgloss.NewStyle().
			Background(colorJumpHintBg).
			Foreground(colorJumpHintFg).
			Bold(true).
			Padding(0, 1)

	// ---- COMMAND PALETTE ----

	paletteStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorPrimary).
			Padding(1, 2).
			Background(colorSurface)

	paletteInputStyle = lipgloss.NewStyle().
				Foreground(colorTextBright)

	paletteResultStyle = lipgloss.NewStyle().
				Foreground(colorText).
				Padding(0, 1)

	paletteSelectedStyle = lipgloss.NewStyle().
				Foreground(colorTextBright).
				Background(colorPrimary).
				Padding(0, 1)

	paletteRecentStyle = lipgloss.NewStyle().
				Foreground(colorWarning).
				Padding(0, 1)

	paletteDescStyle = lipgloss.NewStyle().
				Foreground(colorTextDim).
				Padding(0, 1)

	// ---- HELP OVERLAY ----

	helpOverlayStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(colorInfo).
				Padding(1, 2).
				Background(colorOverlayBg)

	helpCategoryStyle = lipgloss.NewStyle().
				Foreground(colorTextBright).
				Bold(true).
				Padding(0, 1)

	helpKeyStyle = lipgloss.NewStyle().
			Foreground(colorWarning).
			Bold(true).
			Padding(0, 1)

	helpDescStyle = lipgloss.NewStyle().
			Foreground(colorText).
			Padding(0, 1)

	// ---- PREVIEW SIDEBAR ----

	sidebarStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(colorBorder).
			Padding(0, 1).
			Background(colorSurface)

	sidebarHeaderStyle = lipgloss.NewStyle().
				Foreground(colorTextBright).
				Bold(true).
				Padding(0, 1)

	sidebarFieldStyle = lipgloss.NewStyle().
				Foreground(colorTextDim).
				Padding(0, 1)

	sidebarValueStyle = lipgloss.NewStyle().
				Foreground(colorText).
				Padding(0, 1)

	sidebarPinnedStyle = lipgloss.NewStyle().
				Foreground(colorWarning).
				Padding(0, 1)

	// ---- SCROLL INDICATORS ----

	scrollUpStyle = lipgloss.NewStyle().
			Foreground(colorTextDim).
			Render(" ▲")

	scrollDownStyle = lipgloss.NewStyle().
			Foreground(colorTextDim).
			Render(" ▼")

	scrollPercentStyle = lipgloss.NewStyle().
				Foreground(colorTextDim).
				Padding(0, 1)

	// ---- PANEL FOCUS ----

	focusBorderActive = lipgloss.Color("#06B6D4")
	focusBorderInactive = lipgloss.Color("#3D3D5C")

	// ---- BULK ACTION NOTIFICATION ----

	notificationStyle = lipgloss.NewStyle().
				Foreground(colorWarning).
				Bold(true).
				Padding(0, 1)

	pageNavStyle = lipgloss.NewStyle().
			Foreground(colorTextDim).
			Padding(0, 2)

	// Per-page border colors
	colorPageBacklog = lipgloss.Color("#00ffff")
	colorPageDetail  = lipgloss.Color("#00ff88")
	colorPageSprint  = lipgloss.Color("#ff44ff")
	colorPagePopup   = lipgloss.Color("#ffff44")
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

func typeBadge(t string) string {
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
		Bold(true).
		Padding(0, 1).
		Render(t)
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

func shortStatus(s string) string {
	switch s {
	case "todo":
		return "todo"
	case "in-progress":
		return "prog"
	case "review":
		return "rvw"
	case "on-hold":
		return "hold"
	case "done":
		return "done"
	case "cancelled":
		return "canc"
	default:
		return s
	}
}

// ModeStyle returns the appropriate indicator for a given input mode.
func ModeStyle(mode InputMode) lipgloss.Style {
	switch mode {
	case ModeNormal:
		return modeNormalStyle
	case ModeInsert:
		return modeInsertStyle
	case ModeVisual:
		return modeVisualStyle
	case ModeHelp:
		return modeHelpStyle
	default:
		return modeNormalStyle
	}
}

// ApplyTheme overrides global color variables with values from config.
func ApplyTheme(t *config.ThemeConfig) {
	if t == nil {
		return
	}
	if t.Primary != "" {
		colorPrimary = lipgloss.Color(t.Primary)
	}
	if t.Secondary != "" {
		colorSecondary = lipgloss.Color(t.Secondary)
	}
	if t.Success != "" {
		colorSuccess = lipgloss.Color(t.Success)
	}
	if t.Warning != "" {
		colorWarning = lipgloss.Color(t.Warning)
	}
	if t.Error != "" {
		colorError = lipgloss.Color(t.Error)
	}
	if t.Info != "" {
		colorInfo = lipgloss.Color(t.Info)
	}
	if t.Accent != "" {
		colorAccent = lipgloss.Color(t.Accent)
	}
	if t.Background != "" {
		colorBg = lipgloss.Color(t.Background)
	}
	if t.Surface != "" {
		colorSurface = lipgloss.Color(t.Surface)
		colorSurfaceAlt = lipgloss.Color(t.Surface)
	}
	if t.Text != "" {
		colorText = lipgloss.Color(t.Text)
	}
	if t.TextDim != "" {
		colorTextDim = lipgloss.Color(t.TextDim)
	}
	if t.TextBright != "" {
		colorTextBright = lipgloss.Color(t.TextBright)
	}
	if t.Border != "" {
		colorBorder = lipgloss.Color(t.Border)
	}
}

// Conf returns the app config, or nil if not set.
func (m *Model) Conf() *config.Config {
	return m.conf
}
