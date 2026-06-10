package app

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

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

// notificationMsg is sent from child models to app.Model to show a notification.
type notificationMsg struct {
	text string
}

// jumpNotificationCmd returns a tea.Cmd that sends a notification about a jump.
func jumpNotificationCmd(taskID string) tea.Cmd {
	return func() tea.Msg {
		return notificationMsg{text: fmt.Sprintf("Jumped to %s", taskID)}
	}
}
