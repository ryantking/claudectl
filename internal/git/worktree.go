package git

import (
	"bufio"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// Worktree represents a git worktree.
type Worktree struct {
	Path   string
	Branch string
	Commit string
}

// ListWorktrees returns all worktrees for a repository.
func ListWorktrees(repoRoot string) ([]Worktree, error) {
	cmd := exec.Command("git", "-C", repoRoot, "worktree", "list", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list worktrees: %w", err)
	}
	return parseWorktreeList(string(output)), nil
}

// AddWorktree creates a new worktree.
// If createBranch is true, creates a new branch from baseBranch (or HEAD if baseBranch is empty).
// If createBranch is false, checks out the existing branch.
func AddWorktree(repoRoot, path, branch string, createBranch bool, baseBranch string) error {
	args := []string{"-C", repoRoot, "worktree", "add"}
	if createBranch {
		args = append(args, "-b", branch)
		if baseBranch == "" {
			baseBranch = "HEAD"
		}
		args = append(args, path, baseBranch)
	} else {
		args = append(args, path, branch)
	}

	cmd := exec.Command("git", args...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to add worktree: %w", err)
	}
	return nil
}

// RemoveWorktree removes a worktree.
func RemoveWorktree(repoRoot, path string, force bool) error {
	args := []string{"-C", repoRoot, "worktree", "remove"}
	if force {
		args = append(args, "--force")
	}
	args = append(args, path)

	cmd := exec.Command("git", args...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to remove worktree: %w", err)
	}
	return nil
}

// parseWorktreeList parses git worktree list --porcelain output.
func parseWorktreeList(output string) []Worktree {
	var worktrees []Worktree
	var current Worktree
	inWorktree := false

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "worktree ") {
			if inWorktree {
				worktrees = append(worktrees, current)
			}
			current = Worktree{
				Path: strings.TrimPrefix(line, "worktree "),
			}
			inWorktree = true
		} else if strings.HasPrefix(line, "HEAD ") {
			commit := strings.TrimPrefix(line, "HEAD ")
			if len(commit) > 8 {
				commit = commit[:8]
			}
			current.Commit = commit
		} else if strings.HasPrefix(line, "branch ") {
			ref := strings.TrimPrefix(line, "branch ")
			if strings.HasPrefix(ref, "refs/heads/") {
				current.Branch = strings.TrimPrefix(ref, "refs/heads/")
			} else {
				current.Branch = ref
			}
		} else if line == "" && inWorktree {
			worktrees = append(worktrees, current)
			current = Worktree{}
			inWorktree = false
		}
	}

	// Handle last entry if no trailing newline
	if inWorktree {
		worktrees = append(worktrees, current)
	}

	return worktrees
}

// GetWorktreePath returns the absolute path of a worktree.
func GetWorktreePath(repoRoot, path string) (string, error) {
	// If path is relative, make it absolute relative to repo root
	if !filepath.IsAbs(path) {
		path = filepath.Join(repoRoot, path)
	}
	return filepath.Abs(path)
}
