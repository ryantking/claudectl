package git

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// IsWorktreeClean checks if a worktree has uncommitted changes.
// Returns (isClean, statusMessage).
func IsWorktreeClean(worktreePath string) (bool, string) {
	// Check for staged and unstaged changes
	cmd := exec.Command("git", "-C", worktreePath, "status", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Sprintf("failed to check status: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return true, "clean"
	}

	// Count changes
	var staged, modified, untracked int
	for _, line := range lines {
		if line == "" {
			continue
		}
		status := line[:2]
		if status[0] != ' ' && status[0] != '?' {
			staged++
		}
		if status[1] != ' ' && status[1] != '?' {
			modified++
		}
		if status == "??" {
			untracked++
		}
	}

	var parts []string
	if staged > 0 {
		parts = append(parts, strconv.Itoa(staged)+" staged")
	}
	if modified > 0 {
		parts = append(parts, strconv.Itoa(modified)+" modified")
	}
	if untracked > 0 {
		parts = append(parts, strconv.Itoa(untracked)+" untracked")
	}

	if len(parts) == 0 {
		return true, "clean"
	}

	return false, strings.Join(parts, ", ")
}

// GetStatusSummary returns a brief git status summary.
func GetStatusSummary(repoRoot string) (string, error) {
	isClean, status := IsWorktreeClean(repoRoot)
	if isClean {
		return "clean", nil
	}
	return status, nil
}
