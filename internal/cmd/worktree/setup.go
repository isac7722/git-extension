package worktree

import (
	"fmt"
	"os"

	"github.com/isac7722/git-extension/internal/git"
	"github.com/isac7722/git-extension/internal/worktreesetup"
	"github.com/spf13/cobra"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Run worktree setup from .ge-worktree.yaml",
	Long:  "Re-run file copies and setup commands defined in .ge-worktree.yaml for the current worktree.",
	RunE: func(cmd *cobra.Command, args []string) error {
		force, _ := cmd.Flags().GetBool("force")

		mainPath, err := git.MainWorktreePath()
		if err != nil {
			return err
		}

		cwd, err := os.Getwd()
		if err != nil {
			return err
		}

		// Don't run setup in main worktree
		if cwd == mainPath {
			return fmt.Errorf("already in main worktree; setup is for secondary worktrees")
		}

		cfg, err := worktreesetup.Load(mainPath)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
		if cfg == nil {
			return fmt.Errorf("no %s found in %s", worktreesetup.ConfigFile, mainPath)
		}

		fmt.Fprintf(os.Stderr, "  Setting up worktree...\n")
		ok := worktreesetup.Run(cfg, mainPath, cwd, force)
		if ok {
			fmt.Fprintf(os.Stderr, "%s Setup complete!\n", "✔")
		} else {
			fmt.Fprintf(os.Stderr, "⚠ Setup partially completed.\n")
		}

		return nil
	},
}

func init() {
	setupCmd.Flags().Bool("force", false, "Overwrite existing files")
}
