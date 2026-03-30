package worktree

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/isac7722/git-extension/internal/git"
	"github.com/isac7722/git-extension/internal/tui"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configure worktree settings",
	Long: `Configure the default base directory for new worktrees.

When set, 'ge wt add branch' creates the worktree at <dir>/branch.
Default is ../worktrees/branch.

Examples:
  ge wt config            Set worktree directory interactively
  ge wt config --unset    Reset to default (../worktrees)`,
	RunE: runConfig,
}

func init() {
	configCmd.Flags().Bool("unset", false, "Reset worktree directory to default")
	configCmd.Flags().Bool("global", false, "Apply to global git config")
}

func runConfig(cmd *cobra.Command, args []string) error {
	unset, _ := cmd.Flags().GetBool("unset")
	global, _ := cmd.Flags().GetBool("global")

	if unset {
		var err error
		if global {
			err = git.UnsetConfigGlobal("ge.worktree.dir")
		} else {
			err = git.UnsetConfigLocal("ge.worktree.dir")
		}
		if err != nil {
			return fmt.Errorf("failed to unset: %w", err)
		}
		fmt.Fprintf(os.Stderr, "✔ Worktree directory reset to default (../worktrees)\n")
		return nil
	}

	// Show current value and prompt for new one
	current, _ := git.GetConfig("ge.worktree.dir")
	if current == "" {
		current = filepath.Join("..", "worktrees")
	}

	fmt.Fprintf(os.Stderr, "  Current: %s", current)
	if v, _ := git.GetConfig("ge.worktree.dir"); v == "" {
		fmt.Fprintf(os.Stderr, " (default)")
	}
	fmt.Fprintln(os.Stderr)

	value, ok, err := tui.RunPromptWithValue("Worktree directory:", "../worktrees", current)
	if err != nil {
		return err
	}
	if !ok {
		return nil // cancelled
	}

	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}

	// Expand ~ and resolve to absolute path
	if strings.HasPrefix(value, "~/") {
		home, _ := os.UserHomeDir()
		value = filepath.Join(home, value[2:])
	}
	abs, err := filepath.Abs(value)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	if global {
		err = git.SetConfigGlobal("ge.worktree.dir", abs)
	} else {
		err = git.SetConfigLocal("ge.worktree.dir", abs)
	}
	if err != nil {
		return fmt.Errorf("failed to save: %w", err)
	}

	fmt.Fprintf(os.Stderr, "✔ Worktree directory set to %s\n", abs)
	return nil
}
