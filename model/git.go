package model

import (
	"os/exec"
	"strings"
)

type GitState struct {
	Branch string
	Dirty  bool
}

func GetGitState(root string) *GitState {
	gs := &GitState{
		Branch: "unknown",
		Dirty:  false,
	}

	branch, err := exec.Command("git", "-C", root, "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err == nil {
		gs.Branch = strings.TrimSpace(string(branch))
	}

	status, err := exec.Command("git", "-C", root, "status", "--porcelain").Output()
	if err == nil {
		gs.Dirty = strings.TrimSpace(string(status)) != ""
	}

	return gs
}

func GitCommit(root, prefix string) error {
	add := exec.Command("git", "-C", root, "add", ".backlog/")
	if err := add.Run(); err != nil {
		return err
	}

	commit := exec.Command("git", "-C", root, "commit", "-m", prefix+" update tasks")
	return commit.Run()
}

func GitPush(root string) error {
	push := exec.Command("git", "-C", root, "push")
	return push.Run()
}

func GitLog(root string) (string, error) {
	out, err := exec.Command("git", "-C", root, "log", "--oneline", "-20").Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}
