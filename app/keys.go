package app

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	Up         key.Binding
	Down       key.Binding
	Top        key.Binding
	Bottom     key.Binding
	Enter      key.Binding
	Back       key.Binding
	Quit       key.Binding
	Help       key.Binding
	Filter     key.Binding
	FilterMode key.Binding
	Sort       key.Binding
	Reload     key.Binding
	TabLeft    key.Binding
	TabRight   key.Binding
	ToggleProj key.Binding
	Left       key.Binding
	Right      key.Binding
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
	Top: key.NewBinding(
		key.WithKeys("g"),
		key.WithHelp("g", "top"),
	),
	Bottom: key.NewBinding(
		key.WithKeys("G"),
		key.WithHelp("G", "bottom"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "open"),
	),
	Back: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "back"),
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
	FilterMode: key.NewBinding(
		key.WithKeys("f"),
		key.WithHelp("f", "filter by"),
	),
	Sort: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "sort"),
	),
	Reload: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "reload"),
	),
	TabLeft: key.NewBinding(
		key.WithKeys("shift+tab"),
		key.WithHelp("S-tab", "prev tab"),
	),
	TabRight: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "next tab"),
	),
	ToggleProj: key.NewBinding(
		key.WithKeys("p"),
		key.WithHelp("p", "toggle projects"),
	),
	Left: key.NewBinding(
		key.WithKeys("left", "h"),
		key.WithHelp("\u2190/h", "prev"),
	),
	Right: key.NewBinding(
		key.WithKeys("right", "l"),
		key.WithHelp("\u2192/l", "next"),
	),
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Enter, k.Back, k.Help, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Top, k.Bottom},
		{k.Enter, k.Back, k.Quit},
		{k.Filter, k.FilterMode, k.Sort, k.Reload},
		{k.TabLeft, k.TabRight, k.ToggleProj},
	}
}
