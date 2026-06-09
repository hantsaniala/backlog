package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/hantsaniala/backlog/app"
	"github.com/hantsaniala/backlog/model"
	"github.com/hantsaniala/backlog/scaffold"
)

func main() {
	path := flag.String("path", "", "path to project root (default: current directory)")
	flag.Parse()

	root := *path
	if root == "" {
		var err error
		root, err = os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error getting current directory: %v\n", err)
			os.Exit(1)
		}
	}

	backlogRoot, err := model.FindBacklogRoot(root)
	if err != nil {
		fmt.Printf("No .backlog/ found in %s or parent directories.\n", root)
		fmt.Print("Create one? [Y/n]: ")
		var answer string
		fmt.Scanln(&answer)
		if answer == "n" || answer == "N" || answer == "no" {
			os.Exit(0)
		}
		if err := scaffold.Create(root); err != nil {
			fmt.Fprintf(os.Stderr, "error creating .backlog/: %v\n", err)
			os.Exit(1)
		}
		backlogRoot = root + "/.backlog"
	}

	b, err := model.LoadBacklog(backlogRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading backlog: %v\n", err)
		os.Exit(1)
	}

	p := tea.NewProgram(
		app.New(b),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error running TUI: %v\n", err)
		os.Exit(1)
	}
}
