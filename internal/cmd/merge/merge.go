package merge

import (
	"fmt"
	"os"

	"github.com/isac7722/ge-cli/internal/git"
	"github.com/isac7722/ge-cli/internal/tui"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "merge [from] [to]",
	Short: "Merge a branch into another branch",
	Long: `Merge a source branch into a target branch.

Usage:
  ge merge              Interactive selector for both from and to branches
  ge merge <from>       Merge <from> into the current branch
  ge merge <from> <to>  Merge <from> into <to>`,
	Args: cobra.MaximumNArgs(2),
	RunE: runMerge,
}

func runMerge(cmd *cobra.Command, args []string) error {
	if !git.IsInsideWorkTree() {
		return fmt.Errorf("not a git repository")
	}

	current, err := git.CurrentBranch()
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}

	var from, to string

	switch len(args) {
	case 0:
		from, to, err = selectBranches(current)
		if err != nil {
			return err
		}
	case 1:
		from = args[0]
		to = current
	case 2:
		from = args[0]
		to = args[1]
	}

	// Validate branches exist
	if !git.BranchExists(from) {
		return fmt.Errorf("branch %q does not exist", from)
	}
	if !git.BranchExists(to) {
		return fmt.Errorf("branch %q does not exist", to)
	}
	if from == to {
		return fmt.Errorf("cannot merge a branch into itself")
	}

	// Checkout to target branch if needed
	if current != to {
		fmt.Fprintf(os.Stderr, "Switching to %s...\n", to)
		if _, err := git.Run("checkout", to); err != nil {
			return fmt.Errorf("failed to checkout %s: %w", to, err)
		}
	}

	// Merge using passthrough so user sees git output directly
	if err := git.Passthrough("merge", from); err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "%s Merged %s → %s\n", tui.Green.Render("✔"), from, to)
	return nil
}

func selectBranches(current string) (string, string, error) {
	branches, err := git.LocalBranches()
	if err != nil {
		return "", "", fmt.Errorf("failed to list branches: %w", err)
	}
	if len(branches) < 2 {
		return "", "", fmt.Errorf("need at least 2 branches to merge")
	}

	// Select "from" branch (current branch as default)
	var fromItems []tui.SelectorItem
	for _, b := range branches {
		fromItems = append(fromItems, tui.SelectorItem{
			Label:    b,
			Value:    b,
			Selected: b == current,
		})
	}

	fromIdx, err := tui.RunSelector(fromItems, "Select source branch (from)")
	if err != nil {
		return "", "", err
	}
	if fromIdx < 0 {
		return "", "", fmt.Errorf("cancelled")
	}
	from := fromItems[fromIdx].Value

	// Select "to" branch (exclude from, default branch as default)
	defaultBranch := git.DefaultBranch()
	var toItems []tui.SelectorItem
	for _, b := range branches {
		if b == from {
			continue
		}
		toItems = append(toItems, tui.SelectorItem{
			Label:    b,
			Value:    b,
			Selected: b == defaultBranch,
		})
	}

	toIdx, err := tui.RunSelector(toItems, "Select target branch (to)")
	if err != nil {
		return "", "", err
	}
	if toIdx < 0 {
		return "", "", fmt.Errorf("cancelled")
	}
	to := toItems[toIdx].Value

	return from, to, nil
}
