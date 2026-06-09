package model

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type TaskType string

const (
	TypeTask  TaskType = "task"
	TypeBug   TaskType = "bug"
	TypeStory TaskType = "story"
	TypeSpike TaskType = "spike"
	TypeChore TaskType = "chore"
	TypeEpic  TaskType = "epic"
)

type Status string

const (
	StatusTodo       Status = "todo"
	StatusInProgress Status = "in-progress"
	StatusReview     Status = "review"
	StatusOnHold     Status = "on-hold"
	StatusDone       Status = "done"
	StatusCancelled  Status = "cancelled"
)

type Priority string

const (
	PriorityCritical Priority = "critical"
	PriorityHigh     Priority = "high"
	PriorityMedium   Priority = "medium"
	PriorityLow      Priority = "low"
)

type Severity string

const (
	SeverityBlocker  Severity = "blocker"
	SeverityCritical Severity = "critical"
	SeverityMajor    Severity = "major"
	SeverityMinor    Severity = "minor"
	SeverityTrivial  Severity = "trivial"
)

type Task struct {
	ID           string   `yaml:"id"`
	Type         TaskType `yaml:"type"`
	Status       Status   `yaml:"status"`
	Priority     Priority `yaml:"priority"`
	Severity     Severity `yaml:"severity"`
	Assignee     string   `yaml:"assignee"`
	Reporter     string   `yaml:"reporter"`
	Labels       []string `yaml:"labels"`
	Components   []string `yaml:"components"`
	StoryPoints  *int     `yaml:"story_points"`
	Epic         string   `yaml:"epic"`
	Sprint       string   `yaml:"sprint"`
	FixVersion   string   `yaml:"fix_version"`
	Resolution   string   `yaml:"resolution"`
	Parent       string   `yaml:"parent"`
	Children     []string `yaml:"children"`
	DependsOn    []string `yaml:"depends_on"`
	Blocks       []string `yaml:"blocks"`
	RelatedTo    []string `yaml:"related_to"`
	DueDate      string   `yaml:"due_date"`
	Created      string   `yaml:"created"`
	Updated      string   `yaml:"updated"`

	Body     string `yaml:"-"` // markdown body after frontmatter
	Filename string `yaml:"-"` // source filename
	ProjectID string `yaml:"-"` // which project this belongs to
	Summary string `yaml:"-"` // extracted from ## Summary section
	Description string `yaml:"-"` // extracted from ## Description section
}

func ParseTask(path, projectID string) (*Task, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read task %s: %w", path, err)
	}

	parts := strings.SplitN(string(data), "---", 3)
	if len(parts) < 3 {
		return nil, fmt.Errorf("task %s: missing frontmatter", path)
	}

	var t Task
	if err := yaml.Unmarshal([]byte(parts[1]), &t); err != nil {
		return nil, fmt.Errorf("parse frontmatter %s: %w", path, err)
	}

	t.Body = strings.TrimSpace(parts[2])
	t.Filename = filepath.Base(path)
	t.ProjectID = projectID

	// Extract sections from body
	for _, line := range strings.Split(t.Body, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "## Summary") {
			continue
		}
		if strings.HasPrefix(trimmed, "## Description") {
			continue
		}
		if t.Summary == "" && trimmed != "" && !strings.HasPrefix(trimmed, "#") {
			t.Summary = trimmed
		}
		if t.Summary != "" && t.Description == "" && strings.HasPrefix(trimmed, "## ") {
			t.Description = ""
		}
	}

	// Extract summary from body: first non-header line after ## Summary
	bodyLines := strings.Split(t.Body, "\n")
	inSummary := false
	for _, line := range bodyLines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "## Summary" || trimmed == "## Description" {
			inSummary = true
			continue
		}
		if inSummary && strings.HasPrefix(trimmed, "## ") && trimmed != "## Summary" && trimmed != "## Description" {
			inSummary = false
			continue
		}
		if inSummary && trimmed != "" {
			if t.Summary == "" {
				t.Summary = trimmed
				inSummary = false
			}
		}
	}

	return &t, nil
}

func (t *Task) DisplayID() string {
	return t.ID
}

func (t *Task) FullID() string {
	return fmt.Sprintf("%s-%s", t.ProjectID, t.ID)
}

func (t *Task) IsExternal() bool {
	prefix := strings.Split(t.ID, "-")[0]
	return prefix != t.ProjectID
}

func (t *Task) PriorityScore() int {
	switch t.Priority {
	case PriorityCritical:
		return 0
	case PriorityHigh:
		return 1
	case PriorityMedium:
		return 2
	case PriorityLow:
		return 3
	default:
		return 99
	}
}

func (t *Task) StatusScore() int {
	switch t.Status {
	case StatusTodo:
		return 0
	case StatusInProgress:
		return 1
	case StatusReview:
		return 2
	case StatusOnHold:
		return 3
	case StatusDone:
		return 4
	case StatusCancelled:
		return 5
	default:
		return 99
	}
}

var Fibonacci = []int{1, 2, 3, 5, 8, 13}

func ValidStoryPoints(sp int) bool {
	for _, v := range Fibonacci {
		if v == sp {
			return true
		}
	}
	return false
}
