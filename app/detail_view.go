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
	relatedItems  []detailLink
	relatedCursor int

	// Preview popup (within related)
	previewTask *model.Task

	// Viewport for scrollable content
	view      viewport.Model
	viewReady bool

	// Focus indicator
	focused bool

	// Sub-tasks
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
	contentW := d.width - 6

	// Header: dots + ID + summary
	glyph := statusDot(string(d.task.Status))
	tDot := typeDot(string(d.task.Type))
	pDot := priorityDot(string(d.task.Priority))
	summary := d.task.Summary
	if summary != "" {
		summary = "  " + summary
	}
	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(colorTextBright).Render(
		fmt.Sprintf(" %s%s%s %s%s", glyph, tDot, pDot, d.task.ID, summary)))
	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Foreground(colorTextDim).Render(
		fmt.Sprintf(" %s  %s", StatusBadge(string(d.task.Status)), typeBadge(string(d.task.Type)))))
	b.WriteString("\n\n")

	// Body: full-width rendered markdown with width constraint
	if d.task.Body != "" {
		rendered, err := glamour.Render(d.task.Body, "dark")
		if err == nil {
			b.WriteString(lipgloss.NewStyle().Width(contentW).Padding(0, 1).Render(rendered))
		} else {
			b.WriteString(lipgloss.NewStyle().Width(contentW).Padding(0, 1).Foreground(colorText).Render(d.task.Body))
		}
	} else {
		b.WriteString(lipgloss.NewStyle().Foreground(colorTextDim).Render(" No description"))
	}
	b.WriteString("\n\n")

	// Separator
	sep := lipgloss.NewStyle().Foreground(colorTextDim).Render(strings.Repeat("─", contentW))
	b.WriteString(fmt.Sprintf("  %s\n", sep))

	// Inspector: compact inline fields
	var fields []string
	addField := func(label, val string) {
		if val == "" {
			return
		}
		fields = append(fields, fmt.Sprintf("%s: %s", label, val))
	}
	addField("Assignee", d.task.Assignee)
	addField("Priority", string(d.task.Priority))
	if d.task.StoryPoints != nil {
		addField("Points", fmt.Sprintf("%d", *d.task.StoryPoints))
	}
	if d.task.Sprint != "" {
		addField("Sprint", d.task.Sprint)
	}

	b.WriteString(lipgloss.NewStyle().Padding(0, 1).Foreground(colorText).Render(
		strings.Join(fields, "   ")))
	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Padding(0, 1).Foreground(colorTextDim).Render(
		fmt.Sprintf("Created: %s", d.task.Created)))
	if d.task.Updated != "" {
		b.WriteString(lipgloss.NewStyle().Foreground(colorTextDim).Render(
			fmt.Sprintf("   Updated: %s", d.task.Updated)))
	}
	b.WriteString("\n")

	// Labels
	if len(d.task.Labels) > 0 {
		b.WriteString("\n")
		b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(colorTextDim).Padding(0, 1).Render("Labels"))
		b.WriteString("\n")
		for _, l := range d.task.Labels {
			b.WriteString(fmt.Sprintf("  • %s\n", l))
		}
	}

	// Sub-tasks
	if len(d.subtaskTasks) > 0 {
		b.WriteString("\n")
		subSep := lipgloss.NewStyle().Foreground(colorTextDim).Render(strings.Repeat("─", contentW))
		b.WriteString(fmt.Sprintf("  %s\n", subSep))
		b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(colorTextDim).Padding(0, 1).Render("Sub-tasks"))
		b.WriteString("\n")
		for i, st := range d.subtaskTasks {
			cb := "☐"
			if st.Status == model.StatusDone || st.Status == model.StatusCancelled {
				cb = "☑"
			}
			assignee := ""
			if st.Assignee != "" {
				assignee = fmt.Sprintf(" (%s)", st.Assignee)
			}
			line := fmt.Sprintf("  %s %s %s%s", cb, st.ID, st.Summary, assignee)
			if i == d.subtaskCursor {
				line = lipgloss.NewStyle().Foreground(colorTextBright).Background(colorSurface).Render(" " + line)
			}
			b.WriteString(line)
			b.WriteString("\n")
		}
	}

	// Links
	if len(d.relatedItems) > 0 {
		b.WriteString("\n")
		linkSep := lipgloss.NewStyle().Foreground(colorTextDim).Render(strings.Repeat("─", contentW))
		b.WriteString(fmt.Sprintf("  %s\n", linkSep))
		b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(colorTextDim).Padding(0, 1).Render("Links"))
		b.WriteString("\n")
		for _, link := range d.relatedItems {
			g := "◌"
			if link.Task != nil {
				g = statusDot(string(link.Task.Status))
			}
			b.WriteString(fmt.Sprintf("  ▸ %s %s\n", g, link.Label))
		}
	}

	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Foreground(colorTextDim).Render(
		" r:links  o:open  C-d/u:scroll  Esc:focus tree"))

	content := lipgloss.NewStyle().Padding(0, 1).Render(b.String())
	borderColor := colorBorder
	if d.focused {
		borderColor = colorPrimary
	}
	wrapped := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(0, 1).
		Render(content)

	vpW := d.width - 4
	vpH := d.height - 2
	if vpH < 5 {
		vpH = 5
	}
	if !d.viewReady || d.view.Width != vpW {
		d.view = viewport.New(vpW, vpH)
		d.viewReady = true
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

	popupW := d.width
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

	return lipgloss.Place(d.width+40, d.height, lipgloss.Center, lipgloss.Center, content)
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

	popupW := d.width
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

	return lipgloss.Place(d.width+40, d.height, lipgloss.Center, lipgloss.Center, content)
}

func fieldLine(name, value string) string {
	if value == "" || value == "0" || value == "\x00" {
		return ""
	}
	return fmt.Sprintf("  %s %s: %s\n", lipgloss.NewStyle().Foreground(colorTextDim).Render("┃"), name, value)
}
