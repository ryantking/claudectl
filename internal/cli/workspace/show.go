package workspace

import (
	"fmt"

	"github.com/ryantking/agentctl/internal/output"
	"github.com/ryantking/agentctl/internal/ui"
	"github.com/ryantking/agentctl/internal/workspace"
	"github.com/spf13/cobra"
)

// NewWorkspaceShowCmd creates the workspace show command.
func NewWorkspaceShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show [branch]",
		Short: "Print workspace path (for shell integration)",
		Long:  "Outputs the absolute path to the workspace directory. Useful for shell functions and scripts that want to spawn Claude in a new terminal window. If no branch is provided, opens an interactive picker.",
		Args:  cobra.MaximumNArgs(1),
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

			ws, err := manager.GetWorkspace(branch)
			if err != nil {
				result := output.Error(err.Error())
				output.Output(result)
				return err
			}

			jsonOutput, _ := cmd.Root().PersistentFlags().GetBool("json")
			if jsonOutput {
				result := output.Success(map[string]interface{}{
					"path":   ws.Path,
					"branch": ws.Branch,
				})
				output.Output(result)
			} else {
				// Just print the path for easy shell integration
				fmt.Println(ws.Path)
			}

			return nil
		},
	}

	return cmd
}
