package cli

import (
	"github.com/ryantking/agentctl/internal/cli/workspace"
	"github.com/spf13/cobra"
)

// NewWorkspaceCmd creates the workspace command group.
func NewWorkspaceCmd() *cobra.Command {
	return workspace.NewWorkspaceCmd()
}
