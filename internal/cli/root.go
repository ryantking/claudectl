package cli

import (
	"github.com/spf13/cobra"
)

var jsonOutput bool

// Execute runs the CLI application.
func Execute() error {
	return NewRootCmd().Execute()
}

// NewRootCmd creates the root command.
func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "agentctl",
		Short: "CLI for managing Claude Code configurations and workspaces",
		Long:  "A CLI tool for managing Claude Code configurations, hooks, and isolated workspaces using git worktrees.",
	}

	cmd.PersistentFlags().BoolVarP(&jsonOutput, "json", "j", false, "Output result as JSON")

	cmd.AddCommand(
		NewVersionCmd(),
		NewStatusCmd(),
		NewWorkspaceCmd(),
		NewHookCmd(),
		NewInitCmd(),
	)

	return cmd
}
