package scaffold

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func Create(root string) error {
	backlogDir := filepath.Join(root, ".backlog")
	dirs := []string{
		backlogDir,
		filepath.Join(backlogDir, "tasks"),
		filepath.Join(backlogDir, "epics"),
		filepath.Join(backlogDir, "sprints"),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			return fmt.Errorf("create %s: %w", d, err)
		}
	}

	projectID := deriveProjectID(root)
	projectName := filepath.Base(root)

	// project.md
	projectMD := fmt.Sprintf(`---
project_id: %q
name: %q
external_projects: {}
---
`, projectID, projectName)
	if err := os.WriteFile(filepath.Join(backlogDir, "project.md"), []byte(projectMD), 0644); err != nil {
		return err
	}

	// backlog.md
	backlogMD := `# Backlog

**Todo: 0** | **In Progress: 0** | **Review: 0** | **On Hold: 0** | **Done: 0**

## Todo

## In Progress

## Review

## On Hold

## Done

## External Dependencies
`
	if err := os.WriteFile(filepath.Join(backlogDir, "backlog.md"), []byte(backlogMD), 0644); err != nil {
		return err
	}

	fmt.Printf("Created .backlog/ skeleton in %s\n", backlogDir)
	fmt.Printf("  Project ID: %s\n", projectID)
	fmt.Printf("  Project name: %s\n", projectName)
	fmt.Println("Edit .backlog/project.md to configure external projects.")

	return nil
}

func deriveProjectID(root string) string {
	name := filepath.Base(root)
	name = strings.ToUpper(name)
	name = strings.Map(func(r rune) rune {
		if r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, name)
	if len(name) > 4 {
		name = name[:4]
	}
	if name == "" {
		return "PROJ"
	}
	return name
}

func Exists(root string) bool {
	info, err := os.Stat(filepath.Join(root, ".backlog"))
	return err == nil && info.IsDir()
}
