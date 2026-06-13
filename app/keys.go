package app

import "github.com/charmbracelet/bubbles/key"

type normalKeyMap struct {
	Up          key.Binding
	Down        key.Binding
	Left        key.Binding
	Right       key.Binding
	Enter       key.Binding
	Back        key.Binding
	Quit        key.Binding
	Help        key.Binding
	Filter      key.Binding
	Command     key.Binding
	Jump        key.Binding
	Expand      key.Binding
	Related     key.Binding
	One         key.Binding
	Two         key.Binding
	Three       key.Binding
	OpenInEditor key.Binding

	GotoTop      key.Binding
	GotoBottom   key.Binding
	PageDown     key.Binding
	PageUp       key.Binding
	HalfDown     key.Binding
	HalfUp       key.Binding
	Center       key.Binding
	MarkSet      key.Binding
	MarkJump     key.Binding
	SearchNext   key.Binding
	SearchPrev   key.Binding
	BlockUp      key.Binding
	BlockDown    key.Binding
	ScrollTop    key.Binding
	ScrollBot    key.Binding
	WordSearch   key.Binding
	WordSearchRev key.Binding
}

var NormalKeys = normalKeyMap{
	Up:          key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("k", "up")),
	Down:        key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("j", "down")),
	Left:        key.NewBinding(key.WithKeys("left", "h"), key.WithHelp("h", "collapse")),
	Right:       key.NewBinding(key.WithKeys("right", "l"), key.WithHelp("l", "expand")),
	Enter:       key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "open")),
	Back:        key.NewBinding(key.WithKeys("esc"), key.WithHelp("Esc", "back")),
	Quit:        key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
	Help:        key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
	Filter:      key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "filter")),
	Command:     key.NewBinding(key.WithKeys(":"), key.WithHelp(":", "command")),
	Jump:        key.NewBinding(key.WithKeys("f"), key.WithHelp("f", "jump")),
	Expand:      key.NewBinding(key.WithKeys(" "), key.WithHelp("space", "expand")),
	Related:     key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "links")),
	One:         key.NewBinding(key.WithKeys("1"), key.WithHelp("1", "dashboard")),
	Two:         key.NewBinding(key.WithKeys("2"), key.WithHelp("2", "tasks")),
	Three:       key.NewBinding(key.WithKeys("3"), key.WithHelp("3", "sprints")),
	OpenInEditor: key.NewBinding(key.WithKeys("o"), key.WithHelp("o", "open")),

	GotoTop:      key.NewBinding(key.WithKeys("g", "g"), key.WithHelp("gg", "top")),
	GotoBottom:   key.NewBinding(key.WithKeys("G"), key.WithHelp("G", "bottom")),
	PageDown:     key.NewBinding(key.WithKeys("ctrl+f"), key.WithHelp("C-f", "page down")),
	PageUp:       key.NewBinding(key.WithKeys("ctrl+b"), key.WithHelp("C-b", "page up")),
	HalfDown:     key.NewBinding(key.WithKeys("ctrl+d"), key.WithHelp("C-d", "half down")),
	HalfUp:       key.NewBinding(key.WithKeys("ctrl+u"), key.WithHelp("C-u", "half up")),
	Center:       key.NewBinding(key.WithKeys("z", "z"), key.WithHelp("zz", "center")),
	MarkSet:      key.NewBinding(key.WithKeys("m"), key.WithHelp("m", "mark")),
	MarkJump:     key.NewBinding(key.WithKeys("'"), key.WithHelp("'", "jump mark")),
	SearchNext:   key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "next")),
	SearchPrev:   key.NewBinding(key.WithKeys("N"), key.WithHelp("N", "prev")),
	BlockUp:      key.NewBinding(key.WithKeys("{"), key.WithHelp("{", "prev epic")),
	BlockDown:    key.NewBinding(key.WithKeys("}"), key.WithHelp("}", "next epic")),
	ScrollTop:    key.NewBinding(key.WithKeys("z", "t"), key.WithHelp("zt", "top")),
	ScrollBot:    key.NewBinding(key.WithKeys("z", "b"), key.WithHelp("zb", "bottom")),
	WordSearch:   key.NewBinding(key.WithKeys("*"), key.WithHelp("*", "search")),
	WordSearchRev: key.NewBinding(key.WithKeys("#"), key.WithHelp("#", "rsearch")),
}

