package workspace

import (
	"github.com/spf13/cobra"
)

// NewWorkspaceCmd creates the workspace command group.
func NewWorkspaceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "workspace",
		Short: "Manage Claude workspaces (git worktrees)",
		Long:  "Commands for managing Claude workspaces (git worktrees) with proper separation from terminal multiplexing.",
	}

	cmd.AddCommand(
		NewWorkspaceCreateCmd(),
		NewWorkspaceListCmd(),
		NewWorkspaceShowCmd(),
		NewWorkspaceStatusCmd(),
		NewWorkspaceDeleteCmd(),
		NewWorkspaceCleanCmd(),
	)

	return cmd
}
