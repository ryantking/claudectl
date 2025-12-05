package workspace

import (
	"encoding/json"
	"os"

	"github.com/ryantking/agentctl/internal/output"
	"github.com/ryantking/agentctl/internal/ui"
	"github.com/ryantking/agentctl/internal/workspace"
	"github.com/spf13/cobra"
)

// NewWorkspaceListCmd creates the workspace list command.
func NewWorkspaceListCmd() *cobra.Command {
	var jsonFlag bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all managed workspaces",
		Long:  "Shows workspaces in ~/.claude/workspaces/ with their status.",
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

			// Handle --json flag on list command
			if jsonFlag {
				data := make([]map[string]interface{}, len(workspaces))
				for i, w := range workspaces {
					data[i] = w.ToMap()
				}
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(data)
			}

			jsonOutput, _ := cmd.Root().PersistentFlags().GetBool("json")
			if jsonOutput {
				result := output.Success(map[string]interface{}{
					"workspaces": workspacesToMaps(workspaces),
				})
				output.Output(result)
				return nil
			}

			// Use bubbletea table for nice output
			if err := ui.ShowWorkspaceTable(workspaces); err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&jsonFlag, "json", "j", false, "Output as JSON list of workspaces")

	return cmd
}

func workspacesToMaps(workspaces []workspace.Workspace) []map[string]interface{} {
	result := make([]map[string]interface{}, len(workspaces))
	for i, w := range workspaces {
		result[i] = w.ToMap()
	}
	return result
}
