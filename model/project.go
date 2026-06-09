package model

type ProjectConfig struct {
	ProjectID       string            `yaml:"project_id"`
	Name            string            `yaml:"name"`
	ExternalProjects map[string]string `yaml:"external_projects"`
}
