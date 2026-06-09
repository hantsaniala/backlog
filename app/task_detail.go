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
		if t != nil {
			if t.Summary != "" {
				label = id + " - " + t.Summary
			}
		}
		links = append(links, detailLink{Label: label, Task: t})
	}
	return links
}

func renderDetailPanel(task *model.Task, backlog *model.Backlog, focusIdx int, depFocus string) string {
	if task == nil {
		return ""
	}

	var b strings.Builder

	width := 40

	closeBtn := "[X]"
	b.WriteString(lipgloss.NewStyle().Width(width).Align(lipgloss.Right).Render(closeBtn))
	b.WriteString("\n")

	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(colorTextBright).Render(" " + task.ID))
	if task.ProjectID != backlog.Current.Config.ProjectID {
		b.WriteString(" " + ExternalBadge())
	}
	b.WriteString("\n")

	b.WriteString(fieldLine("Type", string(task.Type)))
	b.WriteString("\n")
	b.WriteString(fieldLine("Status", string(task.Status)))
	b.WriteString("\n")
	b.WriteString(fieldLine("Priority", string(task.Priority)))
	b.WriteString("\n")
	if t := string(task.Severity); t != "" {
		b.WriteString(fieldLine("Severity", t))
		b.WriteString("\n")
	}
	if task.Assignee != "" {
		b.WriteString(fieldLine("Assignee", task.Assignee))
		b.WriteString("\n")
	}
	if task.StoryPoints != nil {
		b.WriteString(fieldLine("SP", fmt.Sprintf("%d", *task.StoryPoints)))
		b.WriteString("\n")
	}
	if task.Epic != "" {
		b.WriteString(fieldLine("Epic", task.Epic))
		b.WriteString("\n")
	}
	if task.Sprint != "" {
		b.WriteString(fieldLine("Sprint", task.Sprint))
		b.WriteString("\n")
	}

	b.WriteString("\n")

	if task.Summary != "" {
		b.WriteString(lipgloss.NewStyle().Foreground(colorTextDim).Italic(true).Render(" " + task.Summary))
		b.WriteString("\n\n")
	}

	if task.Body != "" {
		rendered, err := glamour.Render(task.Body, "dark")
		if err == nil {
			b.WriteString(rendered)
		}
	}

	b.WriteString("\n")

	printLinks(&b, "Parent", []string{task.Parent}, backlog, focusIdx, depFocus, "parent")
	printLinks(&b, "Depends on", task.DependsOn, backlog, focusIdx, depFocus, "depends")
	printLinks(&b, "Blocks", task.Blocks, backlog, focusIdx, depFocus, "blocks")
	printLinks(&b, "Related", task.RelatedTo, backlog, focusIdx, depFocus, "related")

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
		prefix := "  ▸ "
		if depFocus == section && focusIdx == i {
			b.WriteString(lipgloss.NewStyle().Foreground(colorPrimary).Render(prefix + label))
		} else {
			b.WriteString(lipgloss.NewStyle().Foreground(colorText).Render(prefix + label))
		}
		b.WriteString("\n")
	}
}

func fieldLine(name, value string) string {
	if value == "" || value == "0" {
		return ""
	}
	return lipgloss.NewStyle().Foreground(colorTextDim).Render(fmt.Sprintf("%-13s %s", name+":", value))
}
