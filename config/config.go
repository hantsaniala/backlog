package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type ThemeConfig struct {
	Primary     string `yaml:"primary"`
	Secondary   string `yaml:"secondary"`
	Success     string `yaml:"success"`
	Warning     string `yaml:"warning"`
	Error       string `yaml:"error"`
	Info        string `yaml:"info"`
	Accent      string `yaml:"accent"`
	Background  string `yaml:"background"`
	Surface     string `yaml:"surface"`
	Text        string `yaml:"text"`
	TextDim     string `yaml:"text_dim"`
	TextBright  string `yaml:"text_bright"`
	Border      string `yaml:"border"`
}

type GitConfig struct {
	AutoCommit   bool   `yaml:"auto_commit"`
	CommitPrefix string `yaml:"commit_prefix"`
}

type EditorConfig struct {
	Command string   `yaml:"command"`
	Args    []string `yaml:"args"`
}

type Config struct {
	Theme  ThemeConfig  `yaml:"theme"`
	Git    GitConfig    `yaml:"git"`
	Editor EditorConfig `yaml:"editor"`
}

func Default() *Config {
	return &Config{
		Theme: ThemeConfig{
			Primary:    "#7C3AED",
			Secondary:  "#06B6D4",
			Success:    "#10B981",
			Warning:    "#F59E0B",
			Error:      "#EF4444",
			Info:       "#3B82F6",
			Accent:     "#A78BFA",
			Background: "#1E1E2E",
			Surface:    "#2D2D44",
			Text:       "#E2E8F0",
			TextDim:    "#64748B",
			TextBright: "#F8FAFC",
			Border:     "#3D3D5C",
		},
		Git: GitConfig{
			AutoCommit:   false,
			CommitPrefix: "feat(backlog):",
		},
		Editor: EditorConfig{
			Command: "",
		},
	}
}

func Load() *Config {
	cfg := Default()

	home, err := os.UserHomeDir()
	if err != nil {
		return cfg
	}

	path := filepath.Join(home, ".config", "backlog", "config.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		return cfg
	}

	var custom Config
	if err := yaml.Unmarshal(data, &custom); err != nil {
		return cfg
	}

	merge(&cfg.Theme, &custom.Theme)
	merge(&cfg.Git, &custom.Git)
	merge(&cfg.Editor, &custom.Editor)

	return cfg
}

func merge[T any](defaults, overrides *T) {
	if overrides != nil {
		*defaults = *overrides
	}
}
