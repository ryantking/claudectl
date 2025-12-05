package workspace

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/ryantking/agentctl/internal/git"
)

// WorkspaceManager manages workspace lifecycle operations.
type WorkspaceManager struct {
	repoRoot string
}

// NewManager creates a new WorkspaceManager.
func NewManager() (*WorkspaceManager, error) {
	repoRoot, err := git.GetRepoRoot()
	if err != nil {
		return nil, ErrNotInGitRepo
	}
	return &WorkspaceManager{repoRoot: repoRoot}, nil
}

// NewManagerAt creates a new WorkspaceManager at a specific repository root.
func NewManagerAt(repoRoot string) (*WorkspaceManager, error) {
	return &WorkspaceManager{repoRoot: repoRoot}, nil
}

// ListWorkspaces lists all workspaces.
func (m *WorkspaceManager) ListWorkspaces(managedOnly bool) ([]Workspace, error) {
	workspaces, err := DiscoverWorkspaces(m.repoRoot)
	if err != nil {
		return nil, err
	}
	if managedOnly {
		var managed []Workspace
		for _, w := range workspaces {
			if w.IsManaged() && !w.IsMain {
				managed = append(managed, w)
			}
		}
		return managed, nil
	}
	return workspaces, nil
}

// GetWorkspace finds workspace by branch name.
func (m *WorkspaceManager) GetWorkspace(branch string) (*Workspace, error) {
	workspace, err := FindWorkspaceByBranch(branch, m.repoRoot)
	if err != nil {
		return nil, err
	}
	if workspace == nil {
		return nil, fmt.Errorf("%w: %s", ErrWorkspaceNotFound, branch)
	}
	return workspace, nil
}

// CreateWorkspace creates a new workspace with worktree.
func (m *WorkspaceManager) CreateWorkspace(branch string, baseBranch string) (*Workspace, error) {
	workspacePath, err := GetWorkspacePath(branch, m.repoRoot)
	if err != nil {
		return nil, err
	}

	// Check if workspace directory already exists
	if _, err := os.Stat(workspacePath); err == nil {
		return nil, fmt.Errorf("%w: %s", ErrWorkspaceExists, workspacePath)
	}

	// Check if branch is already checked out in another worktree
	existing, err := FindWorkspaceByBranch(branch, m.repoRoot)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("%w: %s", ErrBranchInUse, existing.Path)
	}

	// Create parent directory
	if err := os.MkdirAll(filepath.Dir(workspacePath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create workspace directory: %w", err)
	}

	// Check if branch exists
	branchExists, err := git.BranchExists(m.repoRoot, branch)
	if err != nil {
		return nil, err
	}

	if branchExists {
		// Branch exists, just create worktree
		if err := git.AddWorktree(m.repoRoot, workspacePath, branch, false, ""); err != nil {
			return nil, fmt.Errorf("failed to create worktree: %w", err)
		}
	} else {
		// Create new branch from base
		if baseBranch == "" {
			baseBranch, err = git.GetCurrentBranch(m.repoRoot)
			if err != nil || baseBranch == "" {
				baseBranch = "HEAD"
			}
		}
		if err := git.AddWorktree(m.repoRoot, workspacePath, branch, true, baseBranch); err != nil {
			return nil, fmt.Errorf("failed to create worktree: %w", err)
		}
	}

	// Return the newly created workspace
	workspace, err := FindWorkspaceByBranch(branch, m.repoRoot)
	if err != nil {
		return nil, err
	}
	if workspace == nil {
		return nil, fmt.Errorf("workspace created but could not be found")
	}
	return workspace, nil
}

// DeleteWorkspace removes a workspace.
func (m *WorkspaceManager) DeleteWorkspace(branch string, force bool) error {
	workspace, err := m.GetWorkspace(branch)
	if err != nil {
		return err
	}

	// Check if clean
	if !force {
		isClean, status := workspace.IsClean()
		if !isClean {
			return fmt.Errorf("workspace has uncommitted changes (%s). Use --force to remove anyway", status)
		}
	}

	if err := git.RemoveWorktree(m.repoRoot, workspace.Path, force); err != nil {
		return fmt.Errorf("failed to remove worktree: %w", err)
	}

	// Clean up empty parent directories
	parent := filepath.Dir(workspace.Path)
	for {
		if parent == filepath.Dir(parent) {
			break // Reached root
		}
		dir, err := os.ReadDir(parent)
		if err != nil {
			break
		}
		if len(dir) > 0 {
			break // Directory not empty
		}
		if err := os.Remove(parent); err != nil {
			break
		}
		parent = filepath.Dir(parent)
	}

	return nil
}

// CleanWorkspaces removes clean/merged workspaces.
func (m *WorkspaceManager) CleanWorkspaces(checkMerged bool) ([]string, error) {
	var removed []string
	workspaces, err := m.ListWorkspaces(true)
	if err != nil {
		return nil, err
	}

	for _, workspace := range workspaces {
		if workspace.IsMain {
			continue
		}

		isClean, _ := workspace.IsClean()
		if !checkMerged || isClean {
			if workspace.Branch != "" {
				if err := m.DeleteWorkspace(workspace.Branch, !checkMerged); err != nil {
					// Skip workspaces that can't be deleted
					continue
				}
				removed = append(removed, workspace.Branch)
			}
		}
	}

	return removed, nil
}

// GetWorkspaceStatus gets detailed workspace status.
func (m *WorkspaceManager) GetWorkspaceStatus(workspace *Workspace) (map[string]interface{}, error) {
	isClean, status := workspace.IsClean()

	result := map[string]interface{}{
		"path":      workspace.Path,
		"branch":    workspace.Branch,
		"commit":    workspace.Commit,
		"is_clean":  isClean,
		"status":    status,
	}

	// Get ahead/behind information
	if workspace.Branch != "" {
		cmd := exec.Command("git", "-C", workspace.Path, "rev-list", "--left-right", "--count", fmt.Sprintf("origin/%s...HEAD", workspace.Branch))
		output, err := cmd.Output()
		if err == nil {
			var behind, ahead int
			if _, err := fmt.Sscanf(string(output), "%d %d", &behind, &ahead); err == nil {
				result["ahead_behind"] = map[string]int{
					"ahead":  ahead,
					"behind": behind,
				}
			}
		}
	}

	return result, nil
}

// GetWorkspaceDiff gets git diff from workspace to target branch.
func (m *WorkspaceManager) GetWorkspaceDiff(workspace *Workspace, targetBranch string) (string, error) {
	cmd := exec.Command("git", "-C", workspace.Path, "diff", fmt.Sprintf("%s...HEAD", targetBranch))
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get diff: %w", err)
	}
	return string(output), nil
}
