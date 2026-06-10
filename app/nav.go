package app

import (
	"strings"
	"time"
)

type InputMode int

const (
	ModeNormal InputMode = iota
	ModeInsert
	ModeVisual
	ModeHelp
	ModeCommandPalette
)

func (m InputMode) String() string {
	switch m {
	case ModeNormal:
		return "NORMAL"
	case ModeInsert:
		return "INSERT"
	case ModeVisual:
		return "VISUAL"
	case ModeHelp:
		return "HELP"
	case ModeCommandPalette:
		return "CMD"
	default:
		return "NORMAL"
	}
}

type ViewState struct {
	Screen     screen
	TaskID     string
	ScrollPos  int
	FilterText string
}

type NavigationStack struct {
	stack []ViewState
}

func NewNavigationStack() *NavigationStack {
	return &NavigationStack{stack: make([]ViewState, 0)}
}

func (ns *NavigationStack) Push(v ViewState) {
	if len(ns.stack) >= 50 {
		ns.stack = ns.stack[1:]
	}
	ns.stack = append(ns.stack, v)
}

func (ns *NavigationStack) Pop() (ViewState, bool) {
	if len(ns.stack) == 0 {
		return ViewState{}, false
	}
	v := ns.stack[len(ns.stack)-1]
	ns.stack = ns.stack[:len(ns.stack)-1]
	return v, true
}

func (ns *NavigationStack) Peek() (ViewState, bool) {
	if len(ns.stack) == 0 {
		return ViewState{}, false
	}
	return ns.stack[len(ns.stack)-1], true
}

func (ns *NavigationStack) Size() int {
	return len(ns.stack)
}

func (ns *NavigationStack) Breadcrumb(screenNames []string) string {
	if len(ns.stack) == 0 {
		return ""
	}
	var parts []string
	for _, v := range ns.stack {
		label := screenNames[v.Screen]
		if v.TaskID != "" {
			label = v.TaskID
		}
		parts = append(parts, label)
	}
	return strings.Join(parts, " / ")
}

func breadcrumbFromScreen(s screen) string {
	switch s {
	case screenDashboard:
		return "Dashboard"
	case screenTaskList:
		return "Backlog"
	case screenSprintView:
		return "Sprints"
	default:
		return "Unknown"
	}
}

func formatTime() string {
	return time.Now().Format("15:04")
}
