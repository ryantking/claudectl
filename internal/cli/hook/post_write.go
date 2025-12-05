package hook

import (
	"os"

	"github.com/ryantking/agentctl/internal/hook"
	"github.com/spf13/cobra"
)

// NewHookPostWriteCmd creates the hook post-write command.
func NewHookPostWriteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "post-write",
		Short: "PostToolUse hook for Write tool (new files)",
		Long:  "Auto-commits new files if on a feature branch. Reads file path and session ID from stdin JSON.",
		RunE: func(cmd *cobra.Command, args []string) error {
			input, _ := hook.GetStdinData()
			filePath := hook.GetFilePath(input)
			_ = hook.PostWrite(filePath)
			os.Exit(0)
			return nil
		},
	}

	return cmd
}
