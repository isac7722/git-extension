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

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Remove worktrees with merged or gone branches",
	RunE: func(cmd *cobra.Command, args []string) error {
		force, _ := cmd.Flags().GetBool("force")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		// Fetch and prune remote tracking refs
		fmt.Fprintln(os.Stderr, "Fetching remote changes...")
		_, _ = git.Run("fetch", "--prune")

		wts, err := git.ListWorktrees()
		if err != nil {
			return err
		}

		// Build sets of merged and gone branch names
		mergedSet := make(map[string]bool)
		goneSet := make(map[string]bool)

		mergedBranches, _ := mergedBranchesIncludingWorktrees()
		for _, b := range mergedBranches {
			mergedSet[b] = true
		}

		goneBranches, _ := goneBranchesIncludingWorktrees()
		for _, b := range goneBranches {
			goneSet[b] = true
		}

		// Find stale worktrees
		type staleWorktree struct {
			info  git.WorktreeInfo
			tag   string // "merged", "gone", "gone+merged"
			dirty bool
		}

		var stale []staleWorktree
		for _, wt := range wts {
			if wt.IsMain || wt.Branch == "" {
				continue
			}

			isMerged := mergedSet[wt.Branch]
			isGone := goneSet[wt.Branch]

			if !isMerged && !isGone {
				continue
			}

			tag := ""
			switch {
			case isGone && isMerged:
				tag = "gone+merged"
			case isGone:
				tag = "gone"
			case isMerged:
				tag = "merged"
			}

			dirty := strings.Contains(wt.Status, "*")
			stale = append(stale, staleWorktree{info: wt, tag: tag, dirty: dirty})
		}

		if len(stale) == 0 {
			fmt.Fprintln(os.Stderr, "No stale worktrees found.")
			return nil
		}

		// Dry-run: just print and exit
		if dryRun {
			fmt.Fprintf(os.Stderr, "Found %d stale worktree(s):\n\n", len(stale))
			for _, s := range stale {
				warning := ""
				if s.dirty {
					warning = " ⚠ dirty"
				}
				fmt.Fprintf(os.Stderr, "  %s  [%s]%s  %s\n",
					s.info.Branch, s.tag, warning, shortenPath(s.info.Path))
			}
			return nil
		}

		cwd, _ := os.Getwd()
		cwd, _ = filepath.Abs(cwd)

		// Build multi-selector items
		var items []tui.MultiItem
		for _, s := range stale {
			hint := fmt.Sprintf("[%s]  %s", s.tag, shortenPath(s.info.Path))
			if s.dirty {
				hint += "  ⚠ dirty"
			}
			if pathContains(cwd, s.info.Path) {
				hint += " (you are here)"
			}

			// Clean stale worktrees are pre-checked; dirty ones are not
			checked := !s.dirty

			items = append(items, tui.MultiItem{
				Label:   s.info.Branch,
				Value:   s.info.Path,
				Hint:    hint,
				Checked: checked,
			})
		}

		var targets []staleWorktree

		if force {
			targets = stale
		} else {
			indices, err := tui.RunMultiSelector(items, "Select stale worktrees to remove:")
			if err != nil {
				return err
			}
			if len(indices) == 0 {
				return nil
			}
			for _, i := range indices {
				targets = append(targets, stale[i])
			}

			// Confirm
			msg := fmt.Sprintf("Remove %d worktree(s)?", len(targets))
			if len(targets) == 1 {
				msg = fmt.Sprintf("Remove worktree %q?", targets[0].info.Branch)
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

		for _, s := range targets {
			isInside := pathContains(cwd, s.info.Path)
			forceRemove := force || s.dirty

			if err := git.RemoveWorktree(s.info.Path, forceRemove); err != nil {
				fmt.Fprintf(os.Stderr, "✘ Failed to remove %q: %s\n", s.info.Branch, err)
				failed++
				continue
			}

			_ = git.DeleteBranch(s.info.Branch, forceRemove)
			fmt.Fprintf(os.Stderr, "✔ Removed worktree %q [%s]\n", s.info.Branch, s.tag)
			removed++

			if isInside {
				needCdToMain = true
			}
		}

		fmt.Fprintf(os.Stderr, "\n%d removed, %d failed\n", removed, failed)

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
	cleanCmd.Flags().BoolP("force", "f", false, "Force removal without confirmation")
	cleanCmd.Flags().Bool("dry-run", false, "Preview stale worktrees without removing")
}

// mergedBranchesIncludingWorktrees returns names of branches merged into default branch,
// including those checked out in worktrees (unlike git.MergedBranches which excludes them).
func mergedBranchesIncludingWorktrees() ([]string, error) {
	def := git.DefaultBranch()
	out, err := git.Run("branch", "--merged", def)
	if err != nil {
		return nil, err
	}

	current, _ := git.CurrentBranch()
	var names []string
	for _, line := range strings.Split(out, "\n") {
		name := strings.TrimSpace(line)
		name = strings.TrimPrefix(name, "* ")
		name = strings.TrimPrefix(name, "+ ")
		if name == "" || name == current || git.IsProtected(name) {
			continue
		}
		names = append(names, name)
	}
	return names, nil
}

// goneBranchesIncludingWorktrees returns names of branches whose upstream is gone,
// including those checked out in worktrees (unlike git.GoneBranches which excludes them).
func goneBranchesIncludingWorktrees() ([]string, error) {
	out, err := git.Run("for-each-ref", "--format=%(refname:short) %(upstream:track)", "refs/heads/")
	if err != nil {
		return nil, err
	}

	current, _ := git.CurrentBranch()
	var names []string
	for _, line := range strings.Split(out, "\n") {
		if !strings.Contains(line, "[gone]") {
			continue
		}
		name := strings.Fields(line)[0]
		if name == current || git.IsProtected(name) {
			continue
		}
		names = append(names, name)
	}
	return names, nil
}
