package worktree

import (
	"github.com/spf13/cobra"
)

// Cmd is the worktree parent command.
var Cmd = &cobra.Command{
	Use:   "worktree",
	Short: "Manage git worktrees",
	Long:  "Create, list, and remove git worktrees with enhanced features.",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Default: show interactive list
		return runWorktreeList()
	},
}

func init() {
	Cmd.AddCommand(addCmd)
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(removeCmd)
}
