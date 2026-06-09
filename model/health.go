package model

import "fmt"

type HealthIssue struct {
	Severity string // "error", "warning", "info"
	Message  string
	TaskID   string
}

type HealthReport struct {
	Issues []HealthIssue
}

func (b *Backlog) CheckHealth() *HealthReport {
	report := &HealthReport{}

	// 1. Orphan tasks
	for _, t := range b.AllTasks {
		if t.Parent == "" {
			continue
		}
		parent := b.TaskByID(t.Parent)
		if parent == nil {
			report.Issues = append(report.Issues, HealthIssue{
				Severity: "error",
				Message:  "orphan task: parent " + t.Parent + " not found",
				TaskID:   t.ID,
			})
		}
	}

	// 2. Incomplete frontmatter on tasks
	for _, t := range b.AllTasks {
		if t.Priority == "" {
			report.Issues = append(report.Issues, HealthIssue{
				Severity: "warning",
				Message:  "missing priority",
				TaskID:   t.ID,
			})
		}
		if t.StoryPoints == nil && t.Type != TypeChore && t.Type != TypeEpic {
			report.Issues = append(report.Issues, HealthIssue{
				Severity: "warning",
				Message:  "missing story points",
				TaskID:   t.ID,
			})
		}
		if t.Type == TypeBug && t.Severity == "" {
			report.Issues = append(report.Issues, HealthIssue{
				Severity: "warning",
				Message:  "bug without severity",
				TaskID:   t.ID,
			})
		}
	}

	// 3. Story points not Fibonacci
	for _, t := range b.AllTasks {
		if t.StoryPoints != nil && !ValidStoryPoints(*t.StoryPoints) {
			report.Issues = append(report.Issues, HealthIssue{
				Severity: "warning",
				Message:  "story points not Fibonacci",
				TaskID:   t.ID,
			})
		}
	}

	// 4. Circular dependencies check
	graph := make(map[string][]string)
	for _, t := range b.AllTasks {
		for _, dep := range t.DependsOn {
			graph[t.ID] = append(graph[t.ID], dep)
		}
	}
	for _, t := range b.AllTasks {
		if hasCycle(t.ID, graph, make(map[string]bool)) {
			report.Issues = append(report.Issues, HealthIssue{
				Severity: "error",
				Message:  "circular dependency detected",
				TaskID:   t.ID,
			})
		}
	}

	// 5. Stale tasks (created more than 30 days ago, still todo)
	// TODO: implement date-based stale detection

	// 6. WIP limit check
	wipCount := len(b.TasksByStatus(StatusInProgress))
	if wipCount > 3 {
		report.Issues = append(report.Issues, HealthIssue{
			Severity: "warning",
			Message:  fmt.Sprintf("WIP limit exceeded: %d items in progress (max 3)", wipCount),
			TaskID:   "",
		})
	}

	return report
}

func hasCycle(node string, graph map[string][]string, visited map[string]bool) bool {
	if visited[node] {
		return true
	}
	visited[node] = true
	for _, dep := range graph[node] {
		if hasCycle(dep, graph, visited) {
			return true
		}
	}
	visited[node] = false
	return false
}
