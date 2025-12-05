package workspace

import (
	"fmt"

	"github.com/ryantking/agentctl/internal/output"
	"github.com/ryantking/agentctl/internal/ui"
	"github.com/ryantking/agentctl/internal/workspace"
	"github.com/spf13/cobra"
)

// NewWorkspaceDeleteCmd creates the workspace delete command.
func NewWorkspaceDeleteCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete [branch]",
		Short: "Delete a workspace",
		Long: `By default, only deletes workspaces with no uncommitted changes.
Use --force to delete even with changes (WARNING: data loss). If no branch is provided, opens an interactive picker.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			manager, err := workspace.NewManager()
			if err != nil {
				result := output.Error(err.Error())
				output.Output(result)
				return err
			}

			workspaces, err := manager.ListWorkspaces(true)
			if err != nil {
				result := output.Error(err.Error())
				output.Output(result)
				return err
			}

			branch, err := ui.GetWorkspaceArg(args, workspaces)
			if err != nil {
				result := output.Error(err.Error())
				output.Output(result)
				return err
			}

			if err := manager.DeleteWorkspace(branch, force); err != nil {
				result := output.Error(err.Error())
				output.Output(result)
				return err
			}

			result := output.Success(map[string]interface{}{
				"branch": branch,
			})
			result.Message = fmt.Sprintf("Deleted workspace for branch: %s", branch)
			output.Output(result)

			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Force deletion even if workspace has uncommitted changes")

	return cmd
}
