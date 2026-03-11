package worktree

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/isac7722/ge-cli/internal/git"
	"github.com/isac7722/ge-cli/internal/tui"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:     "remove [branch-or-path]",
	Short:   "Remove a worktree",
	Aliases: []string{"rm"},
	RunE: func(cmd *cobra.Command, args []string) error {
		force, _ := cmd.Flags().GetBool("force")

		wts, err := git.ListWorktrees()
		if err != nil {
			return err
		}

		// Filter removable (exclude main worktree)
		var removable []git.WorktreeInfo
		for _, wt := range wts {
			if !wt.IsMain {
				removable = append(removable, wt)
			}
		}

		if len(removable) == 0 {
			fmt.Println("No removable worktrees found.")
			return nil
		}

		var target git.WorktreeInfo

		if len(args) > 0 {
			// Find by branch name or path
			found := false
			for _, wt := range removable {
				if wt.Branch == args[0] || wt.Path == args[0] {
					target = wt
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("worktree %q not found", args[0])
			}
		} else {
			// Interactive selector
			cwd, _ := os.Getwd()
			cwd, _ = filepath.Abs(cwd)

			var items []tui.SelectorItem
			for _, wt := range removable {
				hint := shortenPath(wt.Path)
				if pathContains(cwd, wt.Path) {
					hint += " (you are here)"
				}
				items = append(items, tui.SelectorItem{
					Label:    wt.Branch,
					Value:    wt.Path,
					Hint:     hint,
					Selected: pathContains(cwd, wt.Path),
				})
			}

			idx, err := tui.RunSelector(items, "Select worktree to remove:")
			if err != nil {
				return err
			}
			if idx < 0 {
				return nil
			}
			target = removable[idx]
		}

		// Confirm
		confirmed, err := tui.RunConfirm(fmt.Sprintf("Remove worktree %q?", target.Branch))
		if err != nil {
			return err
		}
		if !confirmed {
			fmt.Println("Cancelled.")
			return nil
		}

		// Check if we're inside the worktree being removed
		cwd, _ := os.Getwd()
		cwd, _ = filepath.Abs(cwd)
		isInside := pathContains(cwd, target.Path)

		// Remove worktree
		if err := git.RemoveWorktree(target.Path, force); err != nil {
			return fmt.Errorf("failed to remove worktree: %w", err)
		}

		// Delete the branch
		_ = git.DeleteBranch(target.Branch, force)

		fmt.Fprintf(os.Stderr, "✔ Removed worktree %q\n", target.Branch)

		// If we were inside, cd to main worktree
		if isInside {
			mainPath, err := git.MainWorktreePath()
			if err == nil {
				fmt.Printf("__GE_EVAL:cd %q\n", mainPath)
			}
		}

		return nil
	},
}

func init() {
	removeCmd.Flags().BoolP("force", "f", false, "Force removal even if dirty")
}
