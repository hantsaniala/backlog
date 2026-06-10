package app

import (
	"fmt"
	"sort"

	"github.com/hantsaniala/backlog/model"
)

type visualSelection struct {
	active   bool
	selected map[string]bool // task ID -> selected
	anchor   int             // cursor position when v was pressed
	cursor   int             // current cursor
}

func newVisualSelection() *visualSelection {
	return &visualSelection{
		selected: make(map[string]bool),
	}
}

func (vs *visualSelection) activate(cursor int) {
	vs.active = true
	vs.selected = make(map[string]bool)
	vs.anchor = cursor
	vs.cursor = cursor
}

func (vs *visualSelection) cancel() {
	vs.active = false
	vs.selected = make(map[string]bool)
	vs.anchor = 0
	vs.cursor = 0
}

func (vs *visualSelection) toggle(taskID string) {
	if vs.selected[taskID] {
		delete(vs.selected, taskID)
	} else {
		vs.selected[taskID] = true
	}
	// Also set anchor if not set
	if vs.selected[taskID] {
		vs.anchor = vs.cursor
	}
}

func (vs *visualSelection) selectAll(tasks []*model.Task) {
	for _, t := range tasks {
		vs.selected[t.ID] = true
	}
}

func (vs *visualSelection) extendDown(tasks []*model.Task, newCursor int) {
	vs.cursor = newCursor
	start := min(vs.anchor, vs.cursor)
	end := max(vs.anchor, vs.cursor)
	for i := start; i <= end && i < len(tasks); i++ {
		vs.selected[tasks[i].ID] = true
	}
}

func (vs *visualSelection) extendUp(tasks []*model.Task, newCursor int) {
	vs.cursor = newCursor
	start := min(vs.anchor, vs.cursor)
	end := max(vs.anchor, vs.cursor)
	for i := start; i <= end && i < len(tasks); i++ {
		vs.selected[tasks[i].ID] = true
	}
}

func (vs *visualSelection) count() int {
	return len(vs.selected)
}

func (vs *visualSelection) selectedIDs() []string {
	var ids []string
	for id := range vs.selected {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

func (vs *visualSelection) selectedTasks(all []*model.Task) []*model.Task {
	var out []*model.Task
	seen := make(map[string]bool)
	for _, t := range all {
		if vs.selected[t.ID] && !seen[t.ID] {
			out = append(out, t)
			seen[t.ID] = true
		}
	}
	return out
}

func (vs *visualSelection) indicator() string {
	if !vs.active {
		return ""
	}
	return fmt.Sprintf(" VISUAL %d selected", vs.count())
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
