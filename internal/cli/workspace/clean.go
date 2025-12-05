package workspace

import (
	"fmt"

	"github.com/ryantking/agentctl/internal/output"
	"github.com/ryantking/agentctl/internal/workspace"
	"github.com/spf13/cobra"
)

// NewWorkspaceCleanCmd creates the workspace clean command.
func NewWorkspaceCleanCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clean",
		Short: "Remove all clean workspaces",
		Long:  "Removes all workspaces that have no uncommitted changes. Useful for cleanup after completing work.",
		RunE: func(cmd *cobra.Command, args []string) error {
			manager, err := workspace.NewManager()
			if err != nil {
				result := output.Error(err.Error())
				output.Output(result)
				return err
			}

			removed, err := manager.CleanWorkspaces(true)
			if err != nil {
				result := output.Error(err.Error())
				output.Output(result)
				return err
			}

			if len(removed) == 0 {
				result := output.Success(nil)
				result.Message = "No clean workspaces to remove"
				output.Output(result)
			} else {
				result := output.Success(map[string]interface{}{
					"removed": removed,
				})
				result.Message = fmt.Sprintf("Removed %d workspace(s)", len(removed))
				output.Output(result)
			}

			return nil
		},
	}

	return cmd
}
