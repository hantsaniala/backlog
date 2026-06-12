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
	editor := flag.String("editor", "", "editor command for opening files (default: $EDITOR or nvim)")
	wait := flag.Bool("wait", false, "wait for user input before exiting")
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

	editorCmd := *editor
	if editorCmd == "" {
		editorCmd = os.Getenv("EDITOR")
		if editorCmd == "" {
			editorCmd = "nvim"
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

	appModel := app.New(b)
	appModel.SetEditor(editorCmd)

	p := tea.NewProgram(
		appModel,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error running TUI: %v\n", err)
		os.Exit(1)
	}

	if *wait {
		fmt.Print("\nPress Enter to exit...")
		fmt.Scanln()
	}
}
