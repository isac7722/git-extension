package fetch

import (
	"github.com/isac7722/git-extension/internal/git"
	"github.com/spf13/cobra"
)

// Cmd is the `ge fetch` command.
var Cmd = &cobra.Command{
	Use:   "fetch [args...]",
	Short: "Fetch and prune remote tracking branches",
	DisableFlagParsing: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		gitArgs := append([]string{"fetch", "--prune"}, args...)
		return git.Passthrough(gitArgs...)
	},
}
