package app

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	Up          key.Binding
	Down        key.Binding
	Left        key.Binding
	Right       key.Binding
	Tab         key.Binding
	Enter       key.Binding
	Back        key.Binding
	Quit        key.Binding
	Help        key.Binding
	Filter      key.Binding
	Sort        key.Binding
	Reload      key.Binding
	One         key.Binding
	Two         key.Binding
	Three       key.Binding
	Four        key.Binding
	CycleStatus key.Binding
	MoveRight   key.Binding
	MoveLeft    key.Binding
	Expand      key.Binding
	Related     key.Binding
	Confirm     key.Binding
}

var Keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "down"),
	),
	Left: key.NewBinding(
		key.WithKeys("left", "h"),
		key.WithHelp("←/h", "prev / collapse"),
	),
	Right: key.NewBinding(
		key.WithKeys("right", "l"),
		key.WithHelp("→/l", "next / expand"),
	),
	Tab: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "switch panel"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "open / confirm"),
	),
	Back: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "back / cancel"),
	),
	Quit: key.NewBinding(
		key.WithKeys("ctrl+c", "q"),
		key.WithHelp("q", "quit"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
	),
	Filter: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "search"),
	),
	Sort: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "cycle status"),
	),
	Reload: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "related items"),
	),
	One: key.NewBinding(
		key.WithKeys("1"),
		key.WithHelp("1", "dashboard"),
	),
	Two: key.NewBinding(
		key.WithKeys("2"),
		key.WithHelp("2", "tasks"),
	),
	Three: key.NewBinding(
		key.WithKeys("3"),
		key.WithHelp("3", "epics"),
	),
	Four: key.NewBinding(
		key.WithKeys("4"),
		key.WithHelp("4", "sprints"),
	),
	CycleStatus: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "cycle status"),
	),
	MoveRight: key.NewBinding(
		key.WithKeys(">", "."),
		key.WithHelp(">", "move to sprint"),
	),
	MoveLeft: key.NewBinding(
		key.WithKeys("<", ","),
		key.WithHelp("<", "remove from sprint"),
	),
	Expand: key.NewBinding(
		key.WithKeys(" "),
		key.WithHelp("space", "expand/collapse"),
	),
	Related: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "related items"),
	),
	Confirm: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "confirm"),
	),
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Enter, k.Quit, k.Help}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Left, k.Right, k.Tab},
		{k.One, k.Two, k.Three, k.Four},
		{k.Enter, k.Back, k.Quit},
		{k.Filter, k.Sort, k.Reload},
	}
}
