package app

import tea "github.com/charmbracelet/bubbletea"

// catchMsg synchronously gets a msg from a cmd.
func catchMsg(cmd tea.Cmd) <-chan tea.Msg {
	ch := make(chan tea.Msg, 1)
	if cmd == nil {
		close(ch)
		return ch
	}
	go func() {
		ch <- cmd()
		close(ch)
	}()
	return ch
}
