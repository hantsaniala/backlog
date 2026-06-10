package app

import (
	"fmt"
	"strings"

	"github.com/hantsaniala/backlog/model"
)

type sidebarState struct {
	open  bool
	pinned bool
	task  *model.Task
}

func newSidebarState() *sidebarState {
	return &sidebarState{}
}

func (s *sidebarState) toggle() {
	if s.open {
		s.open = false
		s.pinned = false
	} else {
		s.open = true
		s.pinned = false
	}
}

func (s *sidebarState) pin() {
	s.pinned = !s.pinned
}

func (s *sidebarState) updateTask(t *model.Task) {
	if !s.pinned {
		s.task = t
	}
}

func (s *sidebarState) render(width int) string {
	if !s.open || s.task == nil {
		return ""
	}

	var b strings.Builder

	if s.pinned {
		b.WriteString(sidebarPinnedStyle.Render(" 📌 PINNED"))
		b.WriteString("\n")
	}

	// Header
	g := statusDot(string(s.task.Status))
	b.WriteString(sidebarHeaderStyle.Render(fmt.Sprintf(" %s %s", g, s.task.ID)))
	b.WriteString("\n")

	if s.task.Summary != "" {
		b.WriteString(sidebarValueStyle.Render(fmt.Sprintf(" %s", s.task.Summary)))
		b.WriteString("\n\n")
	}

	// Fields
	b.WriteString(sidebarFieldStyle.Render(fmt.Sprintf(" Status: %s", StatusBadge(string(s.task.Status)))))
	b.WriteString("\n")
	b.WriteString(sidebarFieldStyle.Render(fmt.Sprintf(" Type: %s", typeBadge(string(s.task.Type)))))
	b.WriteString("\n")

	if string(s.task.Priority) != "" {
		b.WriteString(sidebarFieldStyle.Render(fmt.Sprintf(" Priority: %s", string(s.task.Priority))))
		b.WriteString("\n")
	}
	if s.task.Assignee != "" {
		b.WriteString(sidebarFieldStyle.Render(fmt.Sprintf(" Assignee: %s", s.task.Assignee)))
		b.WriteString("\n")
	}
	if s.task.StoryPoints != nil {
		b.WriteString(sidebarFieldStyle.Render(fmt.Sprintf(" Points: %d", *s.task.StoryPoints)))
		b.WriteString("\n")
	}
	b.WriteString("\n")

	// Description snippet (first 5 lines)
	if s.task.Description != "" {
		b.WriteString(sidebarFieldStyle.Render("Description"))
		b.WriteString("\n")
		lines := strings.Split(s.task.Description, "\n")
		maxLines := 5
		if len(lines) > maxLines {
			lines = lines[:maxLines]
		}
		for _, line := range lines {
			b.WriteString(sidebarValueStyle.Render(fmt.Sprintf(" %s", line)))
			b.WriteString("\n")
		}
		if len(strings.Split(s.task.Description, "\n")) > maxLines {
			b.WriteString(sidebarFieldStyle.Render(" ..."))
			b.WriteString("\n")
		}
	} else if s.task.Body != "" {
		b.WriteString(sidebarFieldStyle.Render("Body"))
		b.WriteString("\n")
		lines := strings.Split(s.task.Body, "\n")
		maxLines := 5
		if len(lines) > maxLines {
			lines = lines[:maxLines]
		}
		for _, line := range lines {
			if len(line) > width-8 {
				line = line[:width-8]
			}
			b.WriteString(sidebarValueStyle.Render(fmt.Sprintf(" %s", line)))
			b.WriteString("\n")
		}
		if len(strings.Split(s.task.Body, "\n")) > maxLines {
			b.WriteString(sidebarFieldStyle.Render(" ..."))
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(sidebarFieldStyle.Render(" C-p: close · C-p again: pin"))

	renderW := width
	if renderW < 20 {
		renderW = 20
	}
	return sidebarStyle.
		Width(renderW).
		Render(b.String())
}
