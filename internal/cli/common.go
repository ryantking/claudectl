package cli

import "github.com/spf13/cobra"

// GetJSONOutput returns whether JSON output is enabled for a command.
func GetJSONOutput(cmd *cobra.Command) bool {
	jsonFlag, _ := cmd.Root().PersistentFlags().GetBool("json")
	return jsonFlag
}
