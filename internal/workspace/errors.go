package workspace

import "fmt"

var (
	ErrWorkspaceExists   = fmt.Errorf("workspace already exists")
	ErrWorkspaceNotFound = fmt.Errorf("workspace not found")
	ErrBranchInUse       = fmt.Errorf("branch is already checked out")
	ErrNotInGitRepo      = fmt.Errorf("not in a git repository")
)

// WorkspaceError represents a workspace operation error.
type WorkspaceError struct {
	Workspace string
	Op        string
	Err       error
}

func (e *WorkspaceError) Error() string {
	return fmt.Sprintf("%s %s: %v", e.Op, e.Workspace, e.Err)
}

func (e *WorkspaceError) Unwrap() error {
	return e.Err
}
