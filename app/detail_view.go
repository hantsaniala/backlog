package app

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/hantsaniala/backlog/model"
)

type detailViewState int

const (
	detailNormal detailViewState = iota
	detailRelated
	detailPreview
)

type detailView struct {
	task    *model.Task
	backlog *model.Backlog
	width   int
	height  int

	state detailViewState

	// Related items popup
	relatedSection string
	relatedCursor  int
	relatedItems   []detailLink

	// Preview popup (within related)
	previewTask *model.Task

	// Scroll tracking
	scrollOffset int

	// Viewport for scrollable content
	view      viewport.Model
	viewReady bool

	// Sub-tasks checklist
	subtaskFocus  bool
	subtaskCursor int
	subtaskTasks  []*model.Task
}

func newDetailView(b *model.Backlog, t *model.Task, w, h int) *detailView {
	d := &detailView{
		task:    t,
		backlog: b,
		width:   w,
		height:  h,
	}
	d.resolveChildren()
	return d
}

func (d *detailView) resolveChildren() {
	d.subtaskTasks = nil
	if d.task == nil {
		return
	}
	for _, id := range d.task.Children {
		if id == "" {
			continue
		}
		t := d.backlog.TaskByID(id)
		if t != nil {
			d.subtaskTasks = append(d.subtaskTasks, t)
		}
	}
}

func (d *detailView) resolveLinks() {
	d.relatedItems = nil
	if d.task == nil {
		return
	}
	addLinks := func(ids []string) {
		for _, id := range ids {
			if id == "" {
				continue
			}
			t := d.backlog.TaskByID(id)
			label := id
			if t != nil && t.Summary != "" {
				label = id + " - " + t.Summary
			}
			d.relatedItems = append(d.relatedItems, detailLink{Label: label, Task: t})
		}
	}
	addLinks([]string{d.task.Parent})
	addLinks([]string{d.task.Epic})
	addLinks(d.task.DependsOn)
	addLinks(d.task.Blocks)
	addLinks(d.task.RelatedTo)
}

func (d *detailView) View() string {
	if d.task == nil {
		return ""
	}
	if d.state == detailRelated && len(d.relatedItems) > 0 {
		return d.renderRelatedPopup()
	}
	if d.state == detailPreview && d.previewTask != nil {
		return d.renderPreviewPopup()
	}
	return d.renderMain()
}

func (d *detailView) renderMain() string {
	var b strings.Builder

	// Header
	glyph := statusDot(string(d.task.Status))
	tDot := typeDot(string(d.task.Type))
	pDot := priorityDot(string(d.task.Priority))
	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(colorTextBright).Render(
		fmt.Sprintf(" %s %s  %s %s", glyph, tDot, d.task.ID, pDot)))
	if d.task.Summary != "" {
		b.WriteString(lipgloss.NewStyle().Foreground(colorTextDim).Render("  " + d.task.Summary))
	}
	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Foreground(colorTextDim).Render(
		fmt.Sprintf(" %s  %s", StatusBadge(string(d.task.Status)), typeBadge(string(d.task.Type)))))
	b.WriteString("\n\n")

	// Two-column layout
	colW := (d.width - 6) / 2
	if colW < 30 {
		colW = 30
	}

	// Left column: description + body
	var left strings.Builder
	if d.task.Body != "" {
		rendered, err := glamour.Render(d.task.Body, "dark")
		if err == nil {
			left.WriteString(lipgloss.NewStyle().Padding(0, 1).Render(rendered))
		} else {
			left.WriteString(lipgloss.NewStyle().Padding(0, 1).Foreground(colorText).Render(d.task.Body))
		}
	} else {
		left.WriteString(lipgloss.NewStyle().Foreground(colorTextDim).Render(" No description"))
	}

	// Right column: metadata
	var right strings.Builder
	right.WriteString(lipgloss.NewStyle().Bold(true).Foreground(colorTextDim).Render(" Details"))
	right.WriteString("\n")
	right.WriteString(fieldLine("Assignee", d.task.Assignee))
	right.WriteString(fieldLine("Reporter", d.task.Reporter))
	right.WriteString(fieldLine("Priority", string(d.task.Priority)))
	if d.task.StoryPoints != nil {
		right.WriteString(fieldLine("Points", fmt.Sprintf("%d", *d.task.StoryPoints)))
	}
	if d.task.Sprint != "" {
		right.WriteString(fieldLine("Sprint", d.task.Sprint))
	}
	right.WriteString(fmt.Sprintf("  %s Created: %s\n", lipgloss.NewStyle().Foreground(colorTextDim).Render("┃"), d.task.Created))
	right.WriteString(fmt.Sprintf("  %s Updated: %s\n", lipgloss.NewStyle().Foreground(colorTextDim).Render("┃"), d.task.Updated))

	if len(d.task.Labels) > 0 {
		right.WriteString("\n")
		right.WriteString(lipgloss.NewStyle().Bold(true).Foreground(colorTextDim).Render(" Labels"))
		right.WriteString("\n")
		for _, l := range d.task.Labels {
			right.WriteString(fmt.Sprintf("  • %s\n", l))
		}
	}

	// Sub-tasks
	if len(d.subtaskTasks) > 0 {
		right.WriteString("\n")
		title := " Sub-tasks"
		if d.subtaskFocus {
			title += " [focused]"
		}
		right.WriteString(lipgloss.NewStyle().Bold(true).Foreground(colorTextDim).Render(title))
		right.WriteString("\n")
		for i, st := range d.subtaskTasks {
			cb := "[ ]"
			if st.Status == model.StatusDone || st.Status == model.StatusCancelled {
				cb = "[x]"
			}
			assignee := ""
			if st.Assignee != "" {
				assignee = fmt.Sprintf(" (%s)", st.Assignee)
			}
			line := fmt.Sprintf("  %s %s %s%s", cb, st.ID, st.Summary, assignee)
			if i == d.subtaskCursor {
				if d.subtaskFocus {
					line = lipgloss.NewStyle().Foreground(colorTextBright).Background(colorSurface).Render(" " + line)
				} else {
					line = lipgloss.NewStyle().Foreground(colorTextDim).Render(line)
				}
			}
			right.WriteString(line)
			right.WriteString("\n")
		}
	}

	// Links
	if len(d.relatedItems) > 0 {
		right.WriteString("\n")
		right.WriteString(lipgloss.NewStyle().Bold(true).Foreground(colorTextDim).Render(" Links"))
		right.WriteString("\n")
		for _, link := range d.relatedItems {
			g := "◌"
			if link.Task != nil {
				g = statusDot(string(link.Task.Status))
			}
			right.WriteString(fmt.Sprintf("  ▸ %s %s\n", g, link.Label))
		}
	}

	leftCol := lipgloss.NewStyle().
		Width(colW).
		Border(lipgloss.NormalBorder()).
		BorderForeground(colorBorder).
		Padding(0, 1).
		Render(left.String())

	rightCol := lipgloss.NewStyle().
		Width(colW).
		Border(lipgloss.NormalBorder()).
		BorderForeground(colorBorder).
		Padding(0, 1).
		Render(right.String())

	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, leftCol, rightCol))

	// Footer help
	b.WriteString("\n")
	scrollPct := ""
	if d.scrollOffset > 0 {
		// Estimate scroll percentage from offset
		pct := d.scrollOffset * 100 / 200
		if pct > 99 {
			pct = 99
		}
		scrollPct = fmt.Sprintf(" %d%%", pct)
	}
	b.WriteString(lipgloss.NewStyle().Foreground(colorTextDim).Render(
		fmt.Sprintf(" r:related  C-d/u:scroll%s  Esc/q:back", scrollPct)))

	content := lipgloss.NewStyle().Padding(0, 2).Render(b.String())
	detailStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorPrimary).
		Padding(0, 1)

	wrapped := detailStyle.Render(content)

	vpW := d.width - 4
	vpH := d.height - 6
	if vpH < 10 {
		vpH = 10
	}
	if !d.viewReady || d.view.Width != vpW {
		d.view = viewport.New(vpW, vpH)
		d.viewReady = true
	}
	if d.scrollOffset > 0 {
		d.view.YOffset = d.scrollOffset
	}
	d.view.SetContent(wrapped)
	vpView := d.view.View()
	sb := renderScrollbar(d.view, vpH)
	return addScrollbar(vpView, sb)
}

func (d *detailView) renderRelatedPopup() string {
	var b strings.Builder
	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(colorTextBright).Render(" Related Items"))
	b.WriteString(fmt.Sprintf(" (%d)", len(d.relatedItems)))
	b.WriteString("\n\n")

	for i, link := range d.relatedItems {
		g := "◌"
		if link.Task != nil {
			g = statusDot(string(link.Task.Status))
		}
		line := fmt.Sprintf(" %s %s", g, link.Label)
		if i == d.relatedCursor {
			line = lipgloss.NewStyle().
				Foreground(colorTextBright).
				Background(colorPrimary).
				Padding(0, 1).
				Render(line)
		} else {
			line = lipgloss.NewStyle().Foreground(colorText).Padding(0, 1).Render(line)
		}
		b.WriteString(line)
		b.WriteString("\n")
	}
	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Foreground(colorTextDim).Render(" Enter: preview  |  Esc: back"))

	popupW := d.width * 60 / 100
	if popupW < 40 {
		popupW = 40
	}
	content := lipgloss.NewStyle().
		Width(popupW).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorPrimary).
		Padding(1, 2).
		Background(colorSurface).
		Render(b.String())

	return lipgloss.Place(d.width, d.height, lipgloss.Center, lipgloss.Center, content)
}

func (d *detailView) renderPreviewPopup() string {
	t := d.previewTask
	if t == nil {
		return ""
	}
	var b strings.Builder

	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(colorTextBright).Render(
		fmt.Sprintf(" Preview: %s", t.ID)))
	b.WriteString("\n")

	g := statusDot(string(t.Status))
	b.WriteString(fmt.Sprintf(" %s %s", g, StatusBadge(string(t.Status))))
	if t.Summary != "" {
		b.WriteString(fmt.Sprintf("  %s", t.Summary))
	}
	b.WriteString("\n\n")

	// Fields
	b.WriteString(fieldLine("Type", string(t.Type)))
	b.WriteString(fieldLine("Priority", string(t.Priority)))
	if t.Assignee != "" {
		b.WriteString(fieldLine("Assignee", t.Assignee))
	}
	if t.StoryPoints != nil {
		b.WriteString(fieldLine("Points", fmt.Sprintf("%d", *t.StoryPoints)))
	}
	b.WriteString("\n")

	if t.Body != "" {
		rendered, err := glamour.Render(t.Body, "dark")
		if err == nil {
			b.WriteString(rendered)
		}
	}

	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Foreground(colorTextDim).Render(" Esc: back"))

	popupW := d.width * 60 / 100
	if popupW < 40 {
		popupW = 40
	}
	content := lipgloss.NewStyle().
		Width(popupW).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorAccent).
		Padding(1, 2).
		Background(colorSurface).
		Render(b.String())

	return lipgloss.Place(d.width, d.height, lipgloss.Center, lipgloss.Center, content)
}

func fieldLine(name, value string) string {
	if value == "" || value == "0" || value == "\x00" {
		return ""
	}
	return fmt.Sprintf("  %s %s: %s\n", lipgloss.NewStyle().Foreground(colorTextDim).Render("┃"), name, value)
}
