package git

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

var (
	ErrNotInGitRepo = fmt.Errorf("not in a git repository")
)

// GetRepoRoot returns the root directory of the current git repository.
// Correctly handles worktrees by finding the actual repository root
// instead of the worktree directory.
func GetRepoRoot() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	output, err := cmd.Output()
	if err != nil {
		return "", ErrNotInGitRepo
	}
	root := strings.TrimSpace(string(output))
	if root == "" {
		return "", ErrNotInGitRepo
	}
	return root, nil
}

// GetRepoName returns the name of the current git repository.
func GetRepoName() (string, error) {
	root, err := GetRepoRoot()
	if err != nil {
		return "", err
	}
	return filepath.Base(root), nil
}

// GetCurrentBranch returns the name of the current branch.
// Returns empty string if in detached HEAD state.
func GetCurrentBranch(repoRoot string) (string, error) {
	cmd := exec.Command("git", "-C", repoRoot, "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	branch := strings.TrimSpace(string(output))
	// Detached HEAD returns "HEAD"
	if branch == "HEAD" {
		return "", nil
	}
	return branch, nil
}

// BranchExists checks if a branch exists locally or remotely.
func BranchExists(repoRoot, branchName string) (bool, error) {
	// Check local branches
	cmd := exec.Command("git", "-C", repoRoot, "branch", "--list", branchName)
	output, err := cmd.Output()
	if err != nil {
		return false, err
	}
	if strings.TrimSpace(string(output)) != "" {
		return true, nil
	}

	// Check remote branches
	cmd = exec.Command("git", "-C", repoRoot, "branch", "-r", "--list", fmt.Sprintf("origin/%s", branchName))
	output, err = cmd.Output()
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(string(output)) != "", nil
}
