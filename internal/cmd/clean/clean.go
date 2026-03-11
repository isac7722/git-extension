package clean

import (
	"fmt"
	"os"

	"github.com/isac7722/ge-cli/internal/git"
	"github.com/isac7722/ge-cli/internal/tui"
	"github.com/spf13/cobra"
)

// Cmd is the clean command.
var Cmd = &cobra.Command{
	Use:   "clean",
	Short: "Clean up stale branches",
	Long: `Remove branches that are gone from remote, merged, or local-only.
By default, shows all stale branches with an interactive selector.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		gone, _ := cmd.Flags().GetBool("gone")
		merged, _ := cmd.Flags().GetBool("merged")
		local, _ := cmd.Flags().GetBool("local")
		all, _ := cmd.Flags().GetBool("all")
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		force, _ := cmd.Flags().GetBool("force")

		// Default to all if no specific filter
		if !gone && !merged && !local {
			all = true
		}

		// Fetch and prune first
		fmt.Fprintf(os.Stderr, "Fetching and pruning...\n")
		if _, err := git.Run("fetch", "--prune"); err != nil {
			fmt.Fprintf(os.Stderr, "⚠ fetch --prune failed: %v\n", err)
		}

		// Collect branches
		var candidates []git.BranchInfo
		seen := make(map[string]bool)

		addBranches := func(branches []git.BranchInfo, err error) {
			if err != nil {
				fmt.Fprintf(os.Stderr, "⚠ %v\n", err)
				return
			}
			for _, b := range branches {
				if seen[b.Name] {
					// Merge tags
					for i, c := range candidates {
						if c.Name == b.Name {
							candidates[i].Tag = c.Tag + "+" + b.Tag
							break
						}
					}
					continue
				}
				seen[b.Name] = true
				candidates = append(candidates, b)
			}
		}

		if all || gone {
			addBranches(git.GoneBranches())
		}
		if all || merged {
			addBranches(git.MergedBranches())
		}
		if all || local {
			addBranches(git.LocalOnlyBranches())
		}

		if len(candidates) == 0 {
			fmt.Println("✔ No stale branches found.")
			return nil
		}

		// Dry run: just display
		if dryRun {
			fmt.Printf("Found %d stale branch(es):\n", len(candidates))
			for _, b := range candidates {
				date := ""
				if b.Date != "" {
					date = fmt.Sprintf("  (%s)", b.Date)
				}
				fmt.Printf("  %s  [%s]%s\n", b.Name, b.Tag, date)
			}
			return nil
		}

		// Determine which branches to delete
		var toDelete []git.BranchInfo

		if force {
			toDelete = candidates
		} else {
			// Interactive multi-select
			var items []tui.MultiItem
			for _, b := range candidates {
				hint := b.Tag
				if b.Date != "" {
					hint += "  " + b.Date
				}
				items = append(items, tui.MultiItem{
					Label:   b.Name,
					Value:   b.Name,
					Hint:    hint,
					Checked: true,
				})
			}

			indices, err := tui.RunMultiSelector(items)
			if err != nil {
				return err
			}
			if indices == nil {
				fmt.Println("Cancelled.")
				return nil
			}

			for _, i := range indices {
				toDelete = append(toDelete, candidates[i])
			}
		}

		if len(toDelete) == 0 {
			fmt.Println("No branches selected.")
			return nil
		}

		// Delete branches
		deleted := 0
		for _, b := range toDelete {
			// Use force delete for gone/local branches (they may not be merged)
			forceDelete := b.Tag != "merged"
			if err := git.DeleteBranch(b.Name, forceDelete); err != nil {
				fmt.Fprintf(os.Stderr, "✗ Failed to delete %s: %v\n", b.Name, err)
			} else {
				fmt.Fprintf(os.Stderr, "✔ Deleted %s\n", b.Name)
				deleted++
			}
		}

		fmt.Fprintf(os.Stderr, "\nDeleted %d branch(es).\n", deleted)
		return nil
	},
}

func init() {
	Cmd.Flags().Bool("gone", false, "Only branches whose remote tracking is gone")
	Cmd.Flags().Bool("merged", false, "Only branches merged into default branch")
	Cmd.Flags().Bool("local", false, "Only branches with no upstream")
	Cmd.Flags().BoolP("all", "a", false, "All stale branches (default)")
	Cmd.Flags().Bool("dry-run", false, "Preview without deleting")
	Cmd.Flags().BoolP("force", "f", false, "Skip confirmation, delete all")
}
