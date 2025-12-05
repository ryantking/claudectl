package workspace

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/ryantking/agentctl/internal/git"
)

// Workspace represents a git worktree managed by agentctl.
type Workspace struct {
	Path     string
	Branch   string
	Commit   string
	IsMain   bool
	RepoRoot string
}

// IsManaged checks if this workspace is managed by agentctl.
// Managed workspaces live under ~/.claude/workspaces/<repo>/
func (w *Workspace) IsManaged() bool {
	return strings.Contains(w.Path, ".claude/workspaces")
}

// IsClean checks if workspace has uncommitted changes.
// Returns (isClean, statusMessage).
func (w *Workspace) IsClean() (bool, string) {
	return git.IsWorktreeClean(w.Path)
}

// ToMap converts workspace to a map for JSON output.
func (w *Workspace) ToMap() map[string]interface{} {
	isClean, status := w.IsClean()
	return map[string]interface{}{
		"path":      w.Path,
		"branch":    w.Branch,
		"commit":    w.Commit,
		"is_main":   w.IsMain,
		"is_managed": w.IsManaged(),
		"is_clean":   isClean,
		"status":     status,
	}
}


// DiscoverWorkspaces discovers all workspaces using git worktree list.
func DiscoverWorkspaces(repoRoot string) ([]Workspace, error) {
	worktrees, err := git.ListWorktrees(repoRoot)
	if err != nil {
		return nil, err
	}

	workspaces := make([]Workspace, len(worktrees))
	for i, wt := range worktrees {
		workspaces[i] = Workspace{
			Path:     wt.Path,
			Branch:   wt.Branch,
			Commit:   wt.Commit,
			IsMain:   i == 0,
			RepoRoot: repoRoot,
		}
	}
	return workspaces, nil
}

// FindWorkspaceByBranch finds a workspace by its branch name.
func FindWorkspaceByBranch(branch string, repoRoot string) (*Workspace, error) {
	workspaces, err := DiscoverWorkspaces(repoRoot)
	if err != nil {
		return nil, err
	}
	for i := range workspaces {
		if workspaces[i].Branch == branch {
			return &workspaces[i], nil
		}
	}
	return nil, nil
}

// SanitizeWorkspaceName generates a safe workspace directory name from branch name.
func SanitizeWorkspaceName(branchName string) string {
	// Replace slashes and special chars with hyphens
	name := strings.ReplaceAll(branchName, "/", "-")
	name = strings.ReplaceAll(name, "\\", "-")
	name = strings.ReplaceAll(name, "_", "-")
	name = strings.ReplaceAll(name, " ", "-")
	// Remove any other unsafe characters (keep alphanumeric, hyphens, dots)
	var result strings.Builder
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '.' {
			result.WriteRune(r)
		}
	}
	name = result.String()
	// Remove leading/trailing hyphens
	name = strings.Trim(name, "-")
	return name
}

// GetWorkspacesBasePath returns the base path for workspaces: ~/.claude/workspaces/<repo-name>.
func GetWorkspacesBasePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	repoName, err := git.GetRepoName()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".claude", "workspaces", repoName), nil
}

// GetWorkspacePath calculates expected workspace path for a branch.
func GetWorkspacePath(branchName string, repoRoot string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	repoName := filepath.Base(repoRoot)
	workspaceName := SanitizeWorkspaceName(branchName)
	return filepath.Join(home, ".claude", "workspaces", repoName, workspaceName), nil
}
