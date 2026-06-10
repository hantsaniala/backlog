package model

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

type ProjectSnapshot struct {
	Config  ProjectConfig
	Root    string
	Tasks   map[string]*Task
	Epics   map[string]*Epic
	Sprints []*Sprint
}

type Backlog struct {
	Current  *ProjectSnapshot
	Externals []*ProjectSnapshot
	AllTasks  []*Task
	AllEpics  map[string]*Epic
}

func FindBacklogRoot(start string) (string, error) {
	dir := start
	for {
		info, err := os.Stat(filepath.Join(dir, ".backlog"))
		if err == nil && info.IsDir() {
			return filepath.Join(dir, ".backlog"), nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("no .backlog/ found from %s to root", start)
		}
		dir = parent
	}
}

func LoadBacklog(backlogRoot string) (*Backlog, error) {
	visited := make(map[string]bool)
	current, err := loadProject(backlogRoot, visited)
	if err != nil {
		return nil, fmt.Errorf("load current project: %w", err)
	}

	b := &Backlog{
		Current:   current,
		Externals: make([]*ProjectSnapshot, 0),
		AllEpics:  make(map[string]*Epic),
	}

	// Load external projects
	for prefix, relPath := range current.Config.ExternalProjects {
		absPath := resolvePath(filepath.Dir(backlogRoot), relPath)
		if visited[absPath] {
			continue
		}
		snap, err := loadProject(absPath, visited)
		if err != nil {
			// Silently skip unloadable external projects
			continue
		}
		if snap.Config.ProjectID != prefix {
			continue
		}
		b.Externals = append(b.Externals, snap)
	}

	// Collect all tasks and epics
	b.AllTasks = make([]*Task, 0)
	for _, t := range current.Tasks {
		t.ProjectID = current.Config.ProjectID
		b.AllTasks = append(b.AllTasks, t)
	}
	for _, snap := range b.Externals {
		for _, t := range snap.Tasks {
			t.ProjectID = snap.Config.ProjectID
			b.AllTasks = append(b.AllTasks, t)
		}
	}

	sort.Slice(b.AllTasks, func(i, j int) bool {
		return b.AllTasks[i].ID < b.AllTasks[j].ID
	})

	for id, ep := range current.Epics {
		ep.ProjectID = current.Config.ProjectID
		b.AllEpics[id] = ep
	}
	for _, snap := range b.Externals {
		for id, ep := range snap.Epics {
			ep.ProjectID = snap.Config.ProjectID
			b.AllEpics[id] = ep
		}
	}

	// Link epics to their children
	b.linkEpicChildren()

	// Add epic-derived tasks to AllTasks (for health checks, search, etc.)
	for _, ep := range b.AllEpics {
		b.AllTasks = append(b.AllTasks, EpicToTask(ep))
	}
	sort.Slice(b.AllTasks, func(i, j int) bool {
		return b.AllTasks[i].ID < b.AllTasks[j].ID
	})

	return b, nil
}

func loadProject(backlogRoot string, visited map[string]bool) (*ProjectSnapshot, error) {
	absRoot, _ := filepath.Abs(backlogRoot)
	visited[absRoot] = true

	configPath := filepath.Join(backlogRoot, "project.md")
	config, err := loadProjectConfig(configPath)
	if err != nil {
		return nil, err
	}

	snap := &ProjectSnapshot{
		Config:  *config,
		Root:    absRoot,
		Tasks:   make(map[string]*Task),
		Epics:   make(map[string]*Epic),
		Sprints: make([]*Sprint, 0),
	}

	taskDir := filepath.Join(backlogRoot, "tasks")
	epicDir := filepath.Join(backlogRoot, "epics")
	sprintDir := filepath.Join(backlogRoot, "sprints")

	// Load tasks
	entries, err := os.ReadDir(taskDir)
	if err == nil {
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
				continue
			}
			task, err := ParseTask(filepath.Join(taskDir, e.Name()), config.ProjectID)
			if err != nil {
				continue
			}
			snap.Tasks[task.ID] = task
		}
	}

	// Load epics
	entries, err = os.ReadDir(epicDir)
	if err == nil {
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
				continue
			}
			data, err := os.ReadFile(filepath.Join(epicDir, e.Name()))
			if err != nil {
				continue
			}
			parts := strings.SplitN(string(data), "---", 3)
			if len(parts) < 3 {
				continue
			}
			var ep Epic
			if err := yaml.Unmarshal([]byte(parts[1]), &ep); err != nil {
				continue
			}
			ep.Body = strings.TrimSpace(parts[2])
			ep.Filename = e.Name()
			ep.ProjectID = config.ProjectID
			snap.Epics[ep.ID] = &ep
		}
	}

	// Load sprints
	entries, err = os.ReadDir(sprintDir)
	if err == nil {
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
				continue
			}
			data, err := os.ReadFile(filepath.Join(sprintDir, e.Name()))
			if err != nil {
				continue
			}
			parts := strings.SplitN(string(data), "---", 3)
			if len(parts) < 3 {
				continue
			}
			var sp Sprint
			if err := yaml.Unmarshal([]byte(parts[1]), &sp); err != nil {
				continue
			}
			sp.Filename = e.Name()
			sp.ProjectID = config.ProjectID
			snap.Sprints = append(snap.Sprints, &sp)
		}
	}

	return snap, nil
}

func loadProjectConfig(path string) (*ProjectConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read project config %s: %w", path, err)
	}
	parts := strings.SplitN(string(data), "---", 3)
	if len(parts) < 3 {
		return nil, fmt.Errorf("missing frontmatter in %s", path)
	}
	var cfg ProjectConfig
	if err := yaml.Unmarshal([]byte(parts[1]), &cfg); err != nil {
		return nil, fmt.Errorf("parse project config %s: %w", path, err)
	}
	if cfg.ExternalProjects == nil {
		cfg.ExternalProjects = make(map[string]string)
	}
	return &cfg, nil
}

func resolvePath(base, rel string) string {
	if filepath.IsAbs(rel) {
		return rel
	}
	return filepath.Clean(filepath.Join(base, rel))
}

func (b *Backlog) linkEpicChildren() {
	for _, ep := range b.AllEpics {
		for _, t := range b.AllTasks {
			if t.Epic == ep.ID {
				ep.Children = append(ep.Children, t)
			}
		}
	}
}

func (b *Backlog) SprintByName(name string) *Sprint {
	for _, s := range b.Current.Sprints {
		if s.Name == name || s.ID == name {
			s.Tasks = b.TasksBySprint(s.Name)
			return s
		}
	}
	for _, ext := range b.Externals {
		for _, s := range ext.Sprints {
			if s.Name == name || s.ID == name {
				s.Tasks = b.TasksBySprint(s.Name)
				return s
			}
		}
	}
	return nil
}

func (b *Backlog) TasksBySprint(sprintName string) []*Task {
	var result []*Task
	for _, t := range b.AllTasks {
		if t.Sprint == sprintName {
			result = append(result, t)
		}
	}
	return result
}

func (b *Backlog) TasksByStatus(status Status) []*Task {
	var result []*Task
	for _, t := range b.AllTasks {
		if t.Status == status {
			result = append(result, t)
		}
	}
	return result
}

func (b *Backlog) TaskByID(id string) *Task {
	for _, t := range b.AllTasks {
		if t.ID == id {
			return t
		}
	}
	for _, ep := range b.AllEpics {
		if ep.ID == id {
			return &Task{ID: ep.ID, Type: TypeEpic, ProjectID: ep.ProjectID}
		}
	}
	return nil
}

func EpicToTask(ep *Epic) *Task {
	summary := ep.Name
	if summary == "" && ep.Body != "" {
		for _, line := range strings.Split(ep.Body, "\n") {
			trimmed := strings.TrimSpace(line)
			if trimmed == "## Summary" {
				continue
			}
			if strings.HasPrefix(trimmed, "## ") {
				break
			}
			if trimmed != "" && summary == "" {
				summary = trimmed
			}
		}
	}
	return &Task{
		ID:        ep.ID,
		Type:      TypeEpic,
		Status:    ep.Status,
		Priority:  ep.Priority,
		Labels:    ep.Labels,
		Created:   ep.Created,
		Updated:   ep.Updated,
		Summary:   summary,
		Body:      ep.Body,
		ProjectID: ep.ProjectID,
	}
}

func (b *Backlog) TaskByFullID(fullID string) *Task {
	parts := strings.SplitN(fullID, "-", 2)
	if len(parts) != 2 {
		return nil
	}
	if b.Current.Config.ProjectID == parts[0] {
		if t, ok := b.Current.Tasks[fullID]; ok {
			return t
		}
	}
	for _, ext := range b.Externals {
		if ext.Config.ProjectID == parts[0] {
			if t, ok := ext.Tasks[fullID]; ok {
				return t
			}
		}
	}
	return nil
}

func (b *Backlog) StatusCounts(projectID string) map[Status]int {
	counts := map[Status]int{
		StatusTodo: 0, StatusInProgress: 0, StatusReview: 0,
		StatusOnHold: 0, StatusDone: 0, StatusCancelled: 0,
	}
	for _, t := range b.AllTasks {
		if projectID == "" || t.ProjectID == projectID {
			counts[t.Status]++
		}
	}
	return counts
}

func (b *Backlog) ProjectNames() map[string]string {
	names := map[string]string{
		b.Current.Config.ProjectID: b.Current.Config.Name,
	}
	for _, ext := range b.Externals {
		names[ext.Config.ProjectID] = ext.Config.Name
	}
	return names
}

func (b *Backlog) Reload() error {
	newBacklog, err := LoadBacklog(b.Current.Root)
	if err != nil {
		return err
	}

	b.Current = newBacklog.Current
	b.Externals = newBacklog.Externals
	b.AllTasks = newBacklog.AllTasks
	b.AllEpics = newBacklog.AllEpics

	return nil
}
