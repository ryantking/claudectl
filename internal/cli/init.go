package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ryantking/agentctl/internal/git"
	"github.com/ryantking/agentctl/internal/setup"
	"github.com/ryantking/agentctl/internal/output"
	"github.com/spf13/cobra"
)

// NewInitCmd creates the init command.
func NewInitCmd() *cobra.Command {
	var globalInstall, force, noIndex bool

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize Claude Code configuration",
		Long: `Initialize Claude Code configuration. Installs CLAUDE.md, agents, skills, and settings from the bundled templates directory.
By default, skips existing files.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var target string
			var err error

			if globalInstall {
				home, err := os.UserHomeDir()
				if err != nil {
					result := output.Error(fmt.Sprintf("failed to get home directory: %v", err))
					output.Output(result)
					return err
				}
				target = filepath.Join(home, ".claude")
			} else {
				target, err = git.GetRepoRoot()
				if err != nil {
					msg := fmt.Sprintf("%v\n\nRun from inside a git repository or use --global", err)
					result := output.Error(msg)
					output.Output(result)
					return err
				}
			}

			manager, err := setup.NewManager(target)
			if err != nil {
				result := output.Error(err.Error())
				output.Output(result)
				return err
			}

			if err := manager.Install(force, noIndex || globalInstall); err != nil {
				result := output.Error(err.Error())
				output.Output(result)
				return err
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&globalInstall, "global", "g", false, "Install to $HOME/.claude instead of current repository")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Overwrite existing files")
	cmd.Flags().BoolVar(&noIndex, "no-index", false, "Skip Claude CLI repository indexing")

	return cmd
}
