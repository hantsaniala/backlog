package app

import (
	"fmt"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
)

// copyToClipboard copies a string to the system clipboard using available tools.
func copyToClipboard(text string) {
	// Try xclip first, then xsel, fallback silently
	if cmd := exec.Command("xclip", "-selection", "clipboard"); cmd != nil {
		w, err := cmd.StdinPipe()
		if err == nil {
			go func() {
				w.Write([]byte(text))
				w.Close()
				cmd.Run()
			}()
			return
		}
	}
	if cmd := exec.Command("xsel", "-ib"); cmd != nil {
		w, err := cmd.StdinPipe()
		if err == nil {
			go func() {
				w.Write([]byte(text))
				w.Close()
				cmd.Run()
			}()
		}
	}
}

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

// notifyCmd returns a tea.Cmd that sends a notification message.
func notifyCmd(msg string) tea.Cmd {
	return func() tea.Msg {
		return notificationMsg{text: msg}
	}
}
