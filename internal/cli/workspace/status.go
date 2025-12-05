package workspace

import (
	"fmt"

	"github.com/ryantking/agentctl/internal/output"
	"github.com/ryantking/agentctl/internal/ui"
	"github.com/ryantking/agentctl/internal/workspace"
	"github.com/spf13/cobra"
)

// NewWorkspaceStatusCmd creates the workspace status command.
func NewWorkspaceStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status [branch]",
		Short: "Show detailed workspace status",
		Long:  "Displays status information including uncommitted changes, ahead/behind status relative to remote, and other details. If no branch is provided, opens an interactive picker.",
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

			statusInfo, err := manager.GetWorkspaceStatus(ws)
			if err != nil {
				result := output.Error(err.Error())
				output.Output(result)
				return err
			}

			jsonOutput, _ := cmd.Root().PersistentFlags().GetBool("json")
			if jsonOutput {
				result := output.Success(statusInfo)
				output.Output(result)
				return nil
			}

			// Human-readable output
			fmt.Printf("\nWorkspace: %v\n", statusInfo["branch"])
			fmt.Printf("Path:      %v\n", statusInfo["path"])
			fmt.Printf("Commit:    %v\n", statusInfo["commit"])
			fmt.Printf("Status:    %v\n", statusInfo["status"])

			if aheadBehind, ok := statusInfo["ahead_behind"].(map[string]int); ok {
				fmt.Printf("Sync:      %d ahead, %d behind origin\n", aheadBehind["ahead"], aheadBehind["behind"])
			}

			fmt.Println()

			return nil
		},
	}

	return cmd
}
