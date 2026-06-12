package app

import (
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/hantsaniala/backlog/model"
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

// taskFilePath returns the full filesystem path to a task's markdown file.
func taskFilePath(task *model.Task, backlogRoot string) string {
	return filepath.Join(backlogRoot, "tasks", task.Filename)
}

// openInEditorCmd returns a tea.Cmd that opens a file in the configured editor.
func openInEditorCmd(editorCmd, filePath string) tea.Cmd {
	return func() tea.Msg {
		cmd := exec.Command("sh", "-c", fmt.Sprintf("%s %s", editorCmd, filePath))
		if err := cmd.Start(); err != nil {
			return notificationMsg{text: fmt.Sprintf("Failed to open editor: %v", err)}
		}
		return nil
	}
}

// editorOpenMsg is sent when the user requests to open a task file in the editor.
type editorOpenMsg struct {
	task   *model.Task
	editor string
	root   string
}

// editorOpenCmd creates a command that sends an editorOpenMsg to the app model
// so it can handle opening the file in the configured editor.
func editorOpenCmd(task *model.Task, editorCmd, root string) tea.Cmd {
	return func() tea.Msg {
		return editorOpenMsg{task: task, editor: editorCmd, root: root}
	}
}
