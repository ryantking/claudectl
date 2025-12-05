package cli

import (
	"github.com/ryantking/agentctl/internal/cli/hook"
	"github.com/spf13/cobra"
)

// NewHookCmd creates the hook command group.
func NewHookCmd() *cobra.Command {
	return hook.NewHookCmd()
}
