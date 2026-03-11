package cmd

import (
	"fmt"

	"github.com/isac7722/git-extension/internal/shell"
	"github.com/spf13/cobra"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Configure shell integration automatically",
	Long:  "Detects your shell and adds ge initialization to your RC file.",
	RunE: func(cmd *cobra.Command, args []string) error {
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		msg, err := shell.Setup(dryRun)
		if err != nil {
			return err
		}

		fmt.Println(msg)
		return nil
	},
}

func init() {
	setupCmd.Flags().Bool("dry-run", false, "Preview changes without modifying files")
}
