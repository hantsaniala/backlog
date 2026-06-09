package app

import (
	"log"

	"github.com/fsnotify/fsnotify"
	"github.com/hantsaniala/backlog/model"

	tea "github.com/charmbracelet/bubbletea"
)

type reloadMsg struct{}
type reloadErrorMsg struct{ err error }

func watchBacklogDirectories(b *model.Backlog) tea.Cmd {
	return func() tea.Msg {
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			return reloadErrorMsg{err}
		}

		dirs := collectWatchDirs(b)
		for _, d := range dirs {
			if err := watcher.Add(d); err != nil {
				log.Printf("watch error for %s: %v", d, err)
			}
		}

		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return nil
				}
				if event.Op&(fsnotify.Create|fsnotify.Write|fsnotify.Remove|fsnotify.Rename) != 0 {
					watcher.Close()
					return reloadMsg{}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return nil
				}
				log.Printf("fsnotify error: %v", err)
			}
		}
	}
}

func collectWatchDirs(b *model.Backlog) []string {
	dirs := []string{
		b.Current.Root,
		b.Current.Root + "/tasks",
		b.Current.Root + "/epics",
		b.Current.Root + "/sprints",
	}
	for _, ext := range b.Externals {
		dirs = append(dirs, ext.Root, ext.Root+"/tasks", ext.Root+"/epics", ext.Root+"/sprints")
	}
	return dirs
}
