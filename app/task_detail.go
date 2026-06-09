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

	// Close button row
	closeBtn := lipgloss.NewStyle().Width(width - 2).Align(lipgloss.Right).Foreground(colorTextDim).Render("[X]")
	b.WriteString(closeBtn)
	b.WriteString("\n")

	// ID header
	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(colorTextBright).Render(" " + task.ID))
	if task.ProjectID != backlog.Current.Config.ProjectID {
		b.WriteString(" " + ExternalBadge())
	}
	b.WriteString("\n")

	// Separator
	sep := lipgloss.NewStyle().Foreground(colorBorder).Render(strings.Repeat("─", width-2))
	b.WriteString(lipgloss.NewStyle().Padding(0, 1).Render(sep))
	b.WriteString("\n")

	// Fields
	if str := string(task.Type); str != "" {
		b.WriteString(labelValue("Type", str, width))
	}
	if str := string(task.Status); str != "" {
		b.WriteString(labelValue("Status", str, width))
	}
	if str := string(task.Priority); str != "" {
		b.WriteString(labelValue("Priority", str, width))
	}
	if str := string(task.Severity); str != "" {
		b.WriteString(labelValue("Severity", str, width))
	}
	if task.Assignee != "" {
		b.WriteString(labelValue("Assignee", task.Assignee, width))
	}
	if task.StoryPoints != nil {
		b.WriteString(labelValue("SP", fmt.Sprintf("%d", *task.StoryPoints), width))
	}
	if task.Epic != "" {
		b.WriteString(labelValue("Epic", task.Epic, width))
	}
	if task.Sprint != "" {
		b.WriteString(labelValue("Sprint", task.Sprint, width))
	}
	b.WriteString("\n")

	// Summary
	if task.Summary != "" {
		b.WriteString(lipgloss.NewStyle().Padding(0, 1).Foreground(colorTextDim).Italic(true).Render(task.Summary))
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

	// Separator before links
	b.WriteString("\n")

	// Links
	printLinks(&b, "Parent", []string{task.Parent}, backlog, focusIdx, depFocus, "parent", width)
	printLinks(&b, "Depends on", task.DependsOn, backlog, focusIdx, depFocus, "depends", width)
	printLinks(&b, "Blocks", task.Blocks, backlog, focusIdx, depFocus, "blocks", width)
	printLinks(&b, "Related", task.RelatedTo, backlog, focusIdx, depFocus, "related", width)

	content := b.String()
	detailStyle := lipgloss.NewStyle().
		Width(width).
		Border(lipgloss.NormalBorder()).
		BorderForeground(colorPrimary).
		Padding(0, 1)

	return detailStyle.Render(content)
}

func printLinks(b *strings.Builder, title string, ids []string, backlog *model.Backlog, focusIdx int, depFocus, section string, width int) {
	var filtered []string
	for _, id := range ids {
		if id != "" {
			filtered = append(filtered, id)
		}
	}
	if len(filtered) == 0 {
		return
	}

	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(colorTextDim).Padding(0, 1).Render(title))
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
			b.WriteString(lipgloss.NewStyle().Padding(0, 1).Foreground(colorText).Render(prefix + label))
		}
		b.WriteString("\n")
	}
}
