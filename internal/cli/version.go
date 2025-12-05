package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	versionInfo = struct {
		version string
		commit  string
		date    string
	}{
		version: "dev",
		commit:  "none",
		date:    "unknown",
	}
)

// SetVersion sets the version information.
func SetVersion(version, commit, date string) {
	versionInfo.version = version
	versionInfo.commit = commit
	versionInfo.date = date
}

// NewVersionCmd creates the version command.
func NewVersionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Show the current version",
		RunE: func(cmd *cobra.Command, args []string) error {
			if jsonOutput {
				data := map[string]string{
					"version": versionInfo.version,
					"commit":  versionInfo.commit,
					"date":    versionInfo.date,
				}
				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				return encoder.Encode(data)
			}
			fmt.Printf("agentctl %s\n", versionInfo.version)
			return nil
		},
	}
	return cmd
}
