package app

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/hantsaniala/backlog/model"
)

type detailLink struct {
	Label string
	Task  *model.Task
}

func resolveDetailLinks(ids []string, backlog *model.Backlog) []detailLink {
	var links []detailLink
	for _, id := range ids {
		t := backlog.TaskByID(id)
		label := id
		if t != nil && t.Summary != "" {
			label = id + " - " + t.Summary
		}
		links = append(links, detailLink{Label: label, Task: t})
	}
	return links
}

func renderDetailPanel(task *model.Task, backlog *model.Backlog, focusIdx int, depFocus string, width int) string {
	if task == nil {
		return ""
	}
	if width < 30 {
		width = 30
	}
	if width > 60 {
		width = 60
	}

	var b strings.Builder

	// ID line with glyphs
	glyph := statusDot(string(task.Status))
	pDot := priorityDot(string(task.Priority))
	tDot := typeDot(string(task.Type))
	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(colorTextBright).Render(
		fmt.Sprintf(" %s %s  %s %s", glyph, tDot, task.ID, pDot)))
	if task.ProjectID != backlog.Current.Config.ProjectID {
		b.WriteString(" " + ExternalBadge())
	}
	b.WriteString("\n")

	// Fields
	fields := []string{
		fmt.Sprintf("  %s  %s", typeDot(string(task.Type)), string(task.Type)),
		fmt.Sprintf("  %s  %s", statusDot(string(task.Status)), string(task.Status)),
		fmt.Sprintf("  %s  %s", priorityDot(string(task.Priority)), string(task.Priority)),
	}
	if str := string(task.Severity); str != "" {
		fields = append(fields, fmt.Sprintf("  severity: %s", str))
	}
	if task.Assignee != "" {
		fields = append(fields, fmt.Sprintf("  assignee: %s", task.Assignee))
	}
	if task.StoryPoints != nil {
		fields = append(fields, fmt.Sprintf("  sp: %d", *task.StoryPoints))
	}
	if task.Epic != "" {
		fields = append(fields, fmt.Sprintf("  epic: %s", task.Epic))
	}
	if task.Sprint != "" {
		fields = append(fields, fmt.Sprintf("  sprint: %s", task.Sprint))
	}
	b.WriteString(strings.Join(fields, "\n"))
	b.WriteString("\n\n")

	// Summary
	if task.Summary != "" {
		b.WriteString(lipgloss.NewStyle().Foreground(colorTextDim).Italic(true).Render(" " + task.Summary))
		b.WriteString("\n")
	}

	// Body
	if task.Body != "" {
		b.WriteString("\n")
		rendered, err := glamour.Render(task.Body, "dark")
		if err == nil {
			b.WriteString(rendered)
		}
	}

	b.WriteString("\n")

	printLinks(&b, "parent", []string{task.Parent}, backlog, focusIdx, depFocus, "parent")
	printLinks(&b, "epic", []string{task.Epic}, backlog, focusIdx, depFocus, "epic")
	printLinks(&b, "depends", task.DependsOn, backlog, focusIdx, depFocus, "depends")
	printLinks(&b, "blocks", task.Blocks, backlog, focusIdx, depFocus, "blocks")
	printLinks(&b, "related", task.RelatedTo, backlog, focusIdx, depFocus, "related")

	content := b.String()
	detailStyle := lipgloss.NewStyle().
		Width(width).
		Border(lipgloss.NormalBorder()).
		BorderForeground(colorPrimary).
		Padding(0, 1)

	return detailStyle.Render(content)
}

func printLinks(b *strings.Builder, title string, ids []string, backlog *model.Backlog, focusIdx int, depFocus, section string) {
	var filtered []string
	for _, id := range ids {
		if id != "" {
			filtered = append(filtered, id)
		}
	}
	if len(filtered) == 0 {
		return
	}

	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(colorTextDim).Render(" " + title))
	b.WriteString("\n")
	for i, id := range filtered {
		t := backlog.TaskByID(id)
		label := id
		if t != nil && t.Summary != "" {
			label = id + " - " + t.Summary
		}
		prefix := "   ▸ "
		if depFocus == section && focusIdx == i {
			hl := lipgloss.NewStyle().
				Foreground(colorTextBright).
				Background(colorPrimary).
				Padding(0, 1)
			b.WriteString(hl.Render(prefix + label))
		} else {
			b.WriteString(lipgloss.NewStyle().Foreground(colorText).Render(prefix + label))
		}
		b.WriteString("\n")
	}
}
