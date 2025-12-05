package hook

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// PostEdit auto-commits changes if on a feature branch.
// Reads file path from stdin JSON.
func PostEdit(filePath string) error {
	if filePath == "" {
		return nil
	}

	repo, err := getRepo(filePath)
	if err != nil {
		return nil // Not in a repo, skip
	}

	if isMainBranch(repo) {
		return nil // Skip on main/master
	}

	return gitAddAndCommit(repo, filePath)
}

// PostWrite auto-commits new files if on a feature branch.
// Reads file path from stdin JSON.
func PostWrite(filePath string) error {
	if filePath == "" {
		return nil
	}

	repo, err := getRepo(filePath)
	if err != nil {
		return nil // Not in a repo, skip
	}

	if isMainBranch(repo) {
		return nil // Skip on main/master
	}

	return gitAddAndCommitNewFile(repo, filePath)
}

func getRepo(filePath string) (string, error) {
	// Find git repo root
	dir := filepath.Dir(filePath)
	for {
		gitDir := filepath.Join(dir, ".git")
		if info, err := os.Stat(gitDir); err == nil && info.IsDir() {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break // Reached root
		}
		dir = parent
	}
	return "", fmt.Errorf("not in a git repository")
}

func isMainBranch(repoRoot string) bool {
	cmd := exec.Command("git", "-C", repoRoot, "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	branch := strings.TrimSpace(string(output))
	return branch == "main" || branch == "master"
}

func gitAddAndCommit(repoRoot, filePath string) error {
	// Make path relative to repo root
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return err
	}
	relPath, err := filepath.Rel(repoRoot, absPath)
	if err != nil {
		return err
	}

	// Stage the file
	cmd := exec.Command("git", "-C", repoRoot, "add", relPath)
	if err := cmd.Run(); err != nil {
		return err
	}

	// Check if there are staged changes
	cmd = exec.Command("git", "-C", repoRoot, "diff", "--cached", "--quiet", relPath)
	if err := cmd.Run(); err == nil {
		// No changes to commit
		return nil
	}

	// Calculate lines changed (simplified)
	filename := filepath.Base(filePath)
	msg := fmt.Sprintf("Update %s: moderate changes", filename)

	// Create commit
	cmd = exec.Command("git", "-C", repoRoot, "commit", "-m", msg)
	return cmd.Run()
}

func gitAddAndCommitNewFile(repoRoot, filePath string) error {
	// Make path relative to repo root
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return err
	}
	relPath, err := filepath.Rel(repoRoot, absPath)
	if err != nil {
		return err
	}

	// Stage the file
	cmd := exec.Command("git", "-C", repoRoot, "add", relPath)
	if err := cmd.Run(); err != nil {
		return err
	}

	// Check if file is staged
	cmd = exec.Command("git", "-C", repoRoot, "diff", "--cached", "--quiet", relPath)
	if err := cmd.Run(); err == nil {
		// No changes to commit
		return nil
	}

	filename := filepath.Base(filePath)
	msg := fmt.Sprintf("Add new file: %s", filename)

	// Create commit
	cmd = exec.Command("git", "-C", repoRoot, "commit", "-m", msg)
	return cmd.Run()
}
