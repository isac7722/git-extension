package worktree

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/isac7722/git-extension/internal/git"
	"github.com/isac7722/git-extension/internal/tui"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:     "remove [branch-or-path...]",
	Short:   "Remove worktrees",
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

		cwd, _ := os.Getwd()
		cwd, _ = filepath.Abs(cwd)

		var targets []git.WorktreeInfo

		if len(args) > 0 {
			// Find by branch name or path
			for _, arg := range args {
				found := false
				for _, wt := range removable {
					if wt.Branch == arg || wt.Path == arg {
						targets = append(targets, wt)
						found = true
						break
					}
				}
				if !found {
					return fmt.Errorf("worktree %q not found", arg)
				}
			}
		} else {
			// Interactive multi-selector
			var items []tui.MultiItem
			for _, wt := range removable {
				hint := shortenPath(wt.Path)
				if pathContains(cwd, wt.Path) {
					hint += " (you are here)"
				}
				items = append(items, tui.MultiItem{
					Label:   wt.Branch,
					Value:   wt.Path,
					Hint:    hint,
					Checked: false,
				})
			}

			indices, err := tui.RunMultiSelector(items, "Select worktrees to remove:")
			if err != nil {
				return err
			}
			if len(indices) == 0 {
				return nil
			}
			for _, i := range indices {
				targets = append(targets, removable[i])
			}
		}

		// Confirm
		if !force {
			msg := fmt.Sprintf("Remove %d worktree(s)?", len(targets))
			if len(targets) == 1 {
				msg = fmt.Sprintf("Remove worktree %q?", targets[0].Branch)
			}
			confirmed, err := tui.RunConfirm(msg)
			if err != nil {
				return err
			}
			if !confirmed {
				fmt.Println("Cancelled.")
				return nil
			}
		}

		// Remove worktrees
		removed, failed := 0, 0
		needCdToMain := false

		for _, target := range targets {
			isInside := pathContains(cwd, target.Path)

			if err := git.RemoveWorktree(target.Path, force); err != nil {
				fmt.Fprintf(os.Stderr, "✘ Failed to remove %q: %s\n", target.Branch, err)
				failed++
				continue
			}

			_ = git.DeleteBranch(target.Branch, force)
			fmt.Fprintf(os.Stderr, "✔ Removed worktree %q\n", target.Branch)
			removed++

			if isInside {
				needCdToMain = true
			}
		}

		// Summary for multi-remove
		if len(targets) > 1 {
			fmt.Fprintf(os.Stderr, "\n%d removed, %d failed\n", removed, failed)
		}

		// If we were inside a removed worktree, cd to main
		if needCdToMain {
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
