package app

import (
	"math"
	"strings"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
	"github.com/hantsaniala/backlog/config"
)

var (
	colorPrimary    = lipgloss.Color("#14b8a6")
	colorSuccess    = lipgloss.Color("#22c55e")
	colorWarning    = lipgloss.Color("#eab308")
	colorError      = lipgloss.Color("#ef4444")
	colorInfo       = lipgloss.Color("#06b6d4")
	colorAccent     = lipgloss.Color("#14b8a6")

	colorBg         = lipgloss.Color("#18181b")
	colorSurface    = lipgloss.Color("#27272a")
	colorSurfaceAlt = lipgloss.Color("#1f1f23")
	colorText       = lipgloss.Color("#e4e4e7")
	colorTextDim    = lipgloss.Color("#71717a")
	colorTextBright = lipgloss.Color("#fafafa")
	colorBorder     = lipgloss.Color("#3f3f46")

	headerStyle = lipgloss.NewStyle().
			Foreground(colorTextBright).
			Bold(true).
			Padding(0, 1)

	tabActiveStyle = lipgloss.NewStyle().
			Foreground(colorTextBright).
			Background(colorPrimary).
			Bold(true).
			Padding(0, 1)

	tabInactiveStyle = lipgloss.NewStyle().
				Foreground(colorTextDim).
				Background(colorSurface).
				Padding(0, 1)

	statusColors = map[string]lipgloss.Color{
		"todo":        colorInfo,
		"in-progress": colorWarning,
		"review":      colorPrimary,
		"on-hold":     colorTextDim,
		"done":        colorSuccess,
		"cancelled":   colorError,
	}

	focusedRowStyle = lipgloss.NewStyle().
			Background(colorSurfaceAlt)

	leftBorderBar = lipgloss.NewStyle().
			Foreground(colorPrimary).
			Render(">")

	jumpHintStyle = lipgloss.NewStyle().
			Background(colorSurface).
			Foreground(colorAccent).
			Bold(true).
			Padding(0, 1)

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

	notificationStyle = lipgloss.NewStyle().
				Foreground(colorWarning).
				Bold(true).
				Padding(0, 1)
)

func StatusBadge(status string) string {
	c, ok := statusColors[status]
	if !ok {
		c = colorTextDim
	}
	return lipgloss.NewStyle().
		Foreground(colorBg).
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
	var g string
	switch status {
	case "todo":
		g = "□"
	case "in-progress":
		g = "■"
	case "review":
		g = "▣"
	case "on-hold":
		g = "–"
	case "done":
		g = "✓"
	case "cancelled":
		g = "✗"
	default:
		g = "●"
	}
	return lipgloss.NewStyle().Foreground(c).Render(g)
}

func priorityLabel(priority string) string {
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
	return lipgloss.NewStyle().Foreground(c).Render(priority)
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
	var g string
	switch t {
	case "bug":
		c = colorError
		g = "✖"
	case "story":
		c = colorInfo
		g = "▶"
	case "spike":
		c = colorWarning
		g = "▲"
	case "chore":
		c = colorTextDim
		g = "○"
	case "epic":
		c = colorAccent
		g = "◆"
	default:
		c = colorText
		g = "●"
	}
	return lipgloss.NewStyle().Foreground(c).Render(g)
}

func typeBadge(t string) string {
	var c lipgloss.Color
	switch t {
	case "bug":
		c = colorError
	case "story":
		c = colorPrimary
	case "spike":
		c = colorWarning
	case "chore":
		c = colorTextDim
	default:
		c = colorPrimary
	}
	return lipgloss.NewStyle().
		Foreground(colorBg).
		Background(c).
		Bold(true).
		Padding(0, 1).
		Render(t)
}

func labelBadge(label string) string {
	return lipgloss.NewStyle().
		Foreground(colorText).
		Background(colorSurface).
		Padding(0, 1).
		Render(label)
}

func ExternalBadge() string {
	return lipgloss.NewStyle().
		Foreground(colorBg).
		Background(colorTextDim).
		Padding(0, 1).
		Render("ext")
}

// ApplyTheme overrides global color variables with values from config.
func ApplyTheme(t *config.ThemeConfig) {
	if t == nil {
		return
	}
	if t.Primary != "" {
		colorPrimary = lipgloss.Color(t.Primary)
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

// newProgressBar returns a progress.Model with solid fill and no percentage.
func newProgressBar(width int) progress.Model {
	return progress.New(
		progress.WithSolidFill("#22c55e"),
		progress.WithFillCharacters('█', '░'),
		progress.WithoutPercentage(),
		progress.WithWidth(width),
	)
}

func renderTaskProgress(done, total int) string {
	if total == 0 {
		return ""
	}
	pct := float64(done) / float64(total)
	w := 8
	filled := int(pct * float64(w))
	var bar strings.Builder
	for i := 0; i < w; i++ {
		if i < filled {
			bar.WriteString("▓")
		} else {
			bar.WriteString("░")
		}
	}
	return lipgloss.NewStyle().Foreground(colorSuccess).Render(bar.String())
}

// renderScrollbar returns a column of scrollbar characters for a viewport.
// Returns one character per line of height. "█" for the thumb, "│" for track.
func renderScrollbar(vp viewport.Model, height int) string {
	if height <= 0 {
		return ""
	}
	total := vp.TotalLineCount()
	if height >= total {
		return strings.Repeat("│\n", height-1) + "│"
	}
	thumbH := int(math.Max(1, float64(height*height)/float64(total)))
	scrollRange := total - height
	thumbPos := vp.YOffset * (height - thumbH) / max(1, scrollRange)
	var sb strings.Builder
	for i := 0; i < height; i++ {
		if i >= thumbPos && i < thumbPos+thumbH {
			sb.WriteString("█")
		} else {
			sb.WriteString("│")
		}
		if i < height-1 {
			sb.WriteByte('\n')
		}
	}
	return sb.String()
}

// addScrollbar appends a rendered scrollbar column to a viewport's output.
func addScrollbar(vpView string, scrollbarStr string) string {
	lines := strings.Split(vpView, "\n")
	sbLines := strings.Split(scrollbarStr, "\n")
	maxLen := len(lines)
	if len(sbLines) > maxLen {
		maxLen = len(sbLines)
	}
	// Pad both slices to equal length with empty strings
	for len(lines) < maxLen {
		lines = append(lines, "")
	}
	for len(sbLines) < maxLen {
		sbLines = append(sbLines, "│")
	}
	for i := range lines {
		if i < len(sbLines) {
			lines[i] = lines[i] + " " + sbLines[i]
		}
	}
	return strings.Join(lines, "\n")
}
