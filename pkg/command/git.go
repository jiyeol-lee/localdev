package command

import (
	"encoding/json"
	"os/exec"
	"strings"
)

type BranchSyncStatus struct {
	Behind int `json:"behind"`
	Ahead  int `json:"ahead"`
}

// GetCurrentBranch returns the current branch name of the git repository in the specified directory.
func GetCurrentBranch(dir string) (string, error) {
	cmd := exec.Command("git", "branch", "--show-current")
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	branch := strings.TrimSpace(string(output))
	return branch, nil
}

// GetBranchSyncStatus returns the sync status of the current branch with its remote counterpart.
func GetBranchSyncStatus(dir string) (*BranchSyncStatus, error) {
	cmd := exec.Command(
		"sh",
		"-c",
		`git rev-list --left-right --count origin/$(git rev-parse --abbrev-ref HEAD)...HEAD | awk '{printf("{ \"behind\": %s, \"ahead\": %s }\n", $1, $2)}'`,
	)
	cmd.Dir = dir
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	var status BranchSyncStatus
	err = json.Unmarshal(output, &status)
	if err != nil {
		return nil, err
	}
	return &status, nil
}
