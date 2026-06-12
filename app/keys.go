package app

import "github.com/charmbracelet/bubbles/key"

type normalKeyMap struct {
	Up          key.Binding
	Down        key.Binding
	Left        key.Binding
	Right       key.Binding
	PrevTab     key.Binding
	NextTab     key.Binding
	Enter       key.Binding
	Back        key.Binding
	Quit        key.Binding
	Help        key.Binding
	Filter      key.Binding
	Command     key.Binding
	Jump        key.Binding
	Visual      key.Binding
	Reload      key.Binding
	Expand      key.Binding
	Related     key.Binding
	CycleStatus key.Binding
	MoveRight   key.Binding
	MoveLeft    key.Binding
	One         key.Binding
	Two         key.Binding
	Three       key.Binding
	Sort        key.Binding
	Confirm     key.Binding

	// Vim motions
	GotoTop    key.Binding
	GotoBottom key.Binding
	PageDown   key.Binding
	PageUp     key.Binding
	HalfDown   key.Binding
	HalfUp     key.Binding
	Center     key.Binding
	HistBack   key.Binding
	HistFwd    key.Binding
	Preview    key.Binding

	// Legacy alias for sprint view
	Tab        key.Binding

	// Panel management
	PanelLeft  key.Binding
	PanelDown  key.Binding
	PanelUp    key.Binding
	PanelRight key.Binding
	PanelClose key.Binding
	PanelMax   key.Binding
	PanelRest  key.Binding

	// Editor
	OpenInEditor key.Binding

	// Vim motions
	MarkSet    key.Binding
	MarkJump   key.Binding
	SearchNext key.Binding
	SearchPrev key.Binding
	BlockUp    key.Binding
	BlockDown  key.Binding
	ScrollTop  key.Binding
	ScrollBot  key.Binding
	WordSearch key.Binding
	WordSearchRev key.Binding
}

type insertKeyMap struct {
	Cancel   key.Binding
	Tab      key.Binding
	LineHome key.Binding
	LineEnd  key.Binding
	Enter    key.Binding
}

type visualKeyMap struct {
	Cancel    key.Binding
	Down      key.Binding
	Up        key.Binding
	Toggle   key.Binding
	SelectAll key.Binding
	Action   key.Binding
	Assign   key.Binding
}

var NormalKeys = normalKeyMap{
	Up:          key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
	Down:        key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
	Left:        key.NewBinding(key.WithKeys("left", "h"), key.WithHelp("←/h", "prev / collapse")),
	Right:       key.NewBinding(key.WithKeys("right", "l"), key.WithHelp("→/l", "next / expand")),
	PrevTab:     key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "switch panel")),
	NextTab:     key.NewBinding(key.WithKeys("shift+tab"), key.WithHelp("S-tab", "prev panel")),
	Enter:       key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "open / confirm")),
	Back:        key.NewBinding(key.WithKeys("esc", "backspace"), key.WithHelp("Esc", "back / cancel")),
	Quit:        key.NewBinding(key.WithKeys("ctrl+c", "q"), key.WithHelp("q", "quit")),
	Help:        key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
	Filter:      key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "search / filter")),
	Command:     key.NewBinding(key.WithKeys(":"), key.WithHelp(":", "command palette")),
	Jump:        key.NewBinding(key.WithKeys("f"), key.WithHelp("f", "jump to item")),
	Visual:      key.NewBinding(key.WithKeys("v"), key.WithHelp("v", "visual select")),
	Reload:      key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "related items")),
	Expand:      key.NewBinding(key.WithKeys(" "), key.WithHelp("space", "expand / toggle")),
	Related:     key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "related items")),
	CycleStatus: key.NewBinding(key.WithKeys("s", "e"), key.WithHelp("s/e", "cycle status")),
	MoveRight:   key.NewBinding(key.WithKeys(">", "."), key.WithHelp(">", "move to right")),
	MoveLeft:    key.NewBinding(key.WithKeys("<", ","), key.WithHelp("<", "move to left")),
	One:         key.NewBinding(key.WithKeys("1"), key.WithHelp("1", "dashboard")),
	Two:         key.NewBinding(key.WithKeys("2"), key.WithHelp("2", "tasks")),
	Three:       key.NewBinding(key.WithKeys("3"), key.WithHelp("3", "sprints")),
	Sort:        key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "cycle status")),
	Confirm:     key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "confirm")),
	GotoTop:     key.NewBinding(key.WithKeys("g", "G"), key.WithHelp("gg", "top")),
	GotoBottom:  key.NewBinding(key.WithKeys("G", "G"), key.WithHelp("G", "bottom")),
	PageDown:    key.NewBinding(key.WithKeys("ctrl+f"), key.WithHelp("C-f", "page down")),
	PageUp:      key.NewBinding(key.WithKeys("ctrl+b"), key.WithHelp("C-b", "page up")),
	HalfDown:    key.NewBinding(key.WithKeys("ctrl+d"), key.WithHelp("C-d", "half down")),
	HalfUp:      key.NewBinding(key.WithKeys("ctrl+u"), key.WithHelp("C-u", "half up")),
	Center:      key.NewBinding(key.WithKeys("z", "z"), key.WithHelp("zz", "center cursor")),
	HistBack:    key.NewBinding(key.WithKeys("alt+left"), key.WithHelp("M-←", "history back")),
	HistFwd:     key.NewBinding(key.WithKeys("alt+right"), key.WithHelp("M-→", "history forward")),
	Preview:     key.NewBinding(key.WithKeys("ctrl+p"), key.WithHelp("C-p", "toggle preview")),
	Tab:         key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "switch panel")),
	PanelLeft:   key.NewBinding(key.WithKeys("ctrl+w", "h"), key.WithHelp("C-w h", "panel left")),
	PanelDown:   key.NewBinding(key.WithKeys("ctrl+w", "j"), key.WithHelp("C-w j", "panel down")),
	PanelUp:     key.NewBinding(key.WithKeys("ctrl+w", "k"), key.WithHelp("C-w k", "panel up")),
	PanelRight:  key.NewBinding(key.WithKeys("ctrl+w", "l"), key.WithHelp("C-w l", "panel right")),
	PanelClose:  key.NewBinding(key.WithKeys("ctrl+w", "q"), key.WithHelp("C-w q", "close panel")),
	PanelMax:    key.NewBinding(key.WithKeys("ctrl+w", "o"), key.WithHelp("C-w o", "max panel")),
	PanelRest:   key.NewBinding(key.WithKeys("ctrl+w", "r"), key.WithHelp("C-w r", "restore layout")),

	OpenInEditor: key.NewBinding(key.WithKeys("o"), key.WithHelp("o", "open in editor")),

	MarkSet:     key.NewBinding(key.WithKeys("m"), key.WithHelp("m[a-z]", "set mark")),
	MarkJump:    key.NewBinding(key.WithKeys("'"), key.WithHelp("'[a-z]", "jump to mark")),
	SearchNext:  key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "next match")),
	SearchPrev:  key.NewBinding(key.WithKeys("N"), key.WithHelp("N", "prev match")),
	BlockUp:     key.NewBinding(key.WithKeys("{"), key.WithHelp("{", "prev epic")),
	BlockDown:   key.NewBinding(key.WithKeys("}"), key.WithHelp("}", "next epic")),
	ScrollTop:   key.NewBinding(key.WithKeys("z", "t"), key.WithHelp("zt", "scroll cursor top")),
	ScrollBot:   key.NewBinding(key.WithKeys("z", "b"), key.WithHelp("zb", "scroll cursor bottom")),
	WordSearch:  key.NewBinding(key.WithKeys("*"), key.WithHelp("*", "search word forward")),
	WordSearchRev: key.NewBinding(key.WithKeys("#"), key.WithHelp("#", "search word backward")),
}

var InsertKeys = insertKeyMap{
	Cancel:   key.NewBinding(key.WithKeys("esc"), key.WithHelp("Esc", "cancel")),
	Tab:      key.NewBinding(key.WithKeys("tab"), key.WithHelp("Tab", "next field")),
	LineHome: key.NewBinding(key.WithKeys("ctrl+a"), key.WithHelp("C-a", "line start")),
	LineEnd:  key.NewBinding(key.WithKeys("ctrl+e"), key.WithHelp("C-e", "line end")),
	Enter:    key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "confirm")),
}

var VisualKeys = visualKeyMap{
	Cancel:    key.NewBinding(key.WithKeys("esc"), key.WithHelp("Esc", "cancel")),
	Down:      key.NewBinding(key.WithKeys("j", "down"), key.WithHelp("j", "extend down")),
	Up:        key.NewBinding(key.WithKeys("k", "up"), key.WithHelp("k", "extend up")),
	Toggle:    key.NewBinding(key.WithKeys(" "), key.WithHelp("space", "toggle item")),
	SelectAll: key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "select all")),
	Action:    key.NewBinding(key.WithKeys("x"), key.WithHelp("x", "bulk status")),
	Assign:    key.NewBinding(key.WithKeys("c"), key.WithHelp("c", "bulk assign")),
}

func ModeKeyMap(mode InputMode) interface{} {
	switch mode {
	case ModeNormal:
		return NormalKeys
	case ModeInsert:
		return InsertKeys
	case ModeVisual:
		return VisualKeys
	default:
		return NormalKeys
	}
}

func NormalHelp() []key.Binding {
	return []key.Binding{
		NormalKeys.Up, NormalKeys.Down, NormalKeys.Left, NormalKeys.Right,
		NormalKeys.Enter, NormalKeys.Back, NormalKeys.Quit,
	}
}

func NormalFullHelp() [][]key.Binding {
	return [][]key.Binding{
		{NormalKeys.Up, NormalKeys.Down, NormalKeys.Left, NormalKeys.Right, NormalKeys.Expand},
		{NormalKeys.GotoTop, NormalKeys.GotoBottom, NormalKeys.HalfDown, NormalKeys.HalfUp, NormalKeys.Center},
		{NormalKeys.Jump, NormalKeys.Visual, NormalKeys.Filter, NormalKeys.Command, NormalKeys.Help},
		{NormalKeys.One, NormalKeys.Two, NormalKeys.Three},
		{NormalKeys.HistBack, NormalKeys.HistFwd, NormalKeys.Preview},
		{NormalKeys.PanelLeft, NormalKeys.PanelRight, NormalKeys.PanelMax, NormalKeys.PanelRest},
		{NormalKeys.Enter, NormalKeys.Back, NormalKeys.Quit},
	}
}

func InsertHelp() []key.Binding {
	return []key.Binding{
		InsertKeys.Cancel, InsertKeys.Tab, InsertKeys.Enter,
		InsertKeys.LineHome, InsertKeys.LineEnd,
	}
}

func VisualHelp() []key.Binding {
	return []key.Binding{
		VisualKeys.Cancel, VisualKeys.Down, VisualKeys.Up,
		VisualKeys.Toggle, VisualKeys.SelectAll, VisualKeys.Action,
	}
}
