package worktree

import (
	"fmt"
	"os"

	"github.com/isac7722/ge-cli/internal/git"
	"github.com/isac7722/ge-cli/internal/tui"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add [branch] [directory]",
	Short: "Create a new worktree",
	Long:  "Create a new worktree. Auto-creates the branch if it doesn't exist.",
	RunE: func(cmd *cobra.Command, args []string) error {
		var branch, dir string

		if len(args) >= 1 {
			branch = args[0]
		}
		if len(args) >= 2 {
			dir = args[1]
		}

		// Interactive branch selection if no branch specified
		if branch == "" {
			var err error
			branch, err = selectBranch()
			if err != nil {
				return err
			}
			if branch == "" {
				return nil // Cancelled
			}
		}

		absPath, err := git.AddWorktree(branch, dir)
		if err != nil {
			return fmt.Errorf("failed to add worktree: %w", err)
		}

		fmt.Fprintf(os.Stderr, "✔ Created worktree at %s\n", absPath)
		// cd into new worktree via shell wrapper
		fmt.Printf("__GE_EVAL:cd %q\n", absPath)
		return nil
	},
}

func selectBranch() (string, error) {
	local, remote, err := git.AvailableBranches()
	if err != nil {
		return "", err
	}

	var items []tui.SelectorItem

	for _, b := range local {
		items = append(items, tui.SelectorItem{
			Label: b,
			Value: b,
			Hint:  "(local)",
		})
	}
	for _, b := range remote {
		items = append(items, tui.SelectorItem{
			Label: b,
			Value: b,
			Hint:  "(remote)",
		})
	}

	if len(items) == 0 {
		// Allow creating a new branch by prompting
		name, ok, err := tui.RunPrompt("New branch name:", "feature/my-feature")
		if err != nil {
			return "", err
		}
		if !ok {
			return "", nil
		}
		return name, nil
	}

	// Add "create new branch" option at the top
	items = append([]tui.SelectorItem{{
		Label: "+ Create new branch",
		Value: "__new__",
		Hint:  "",
	}}, items...)

	idx, err := tui.RunSelector(items, "Select branch for worktree:")
	if err != nil {
		return "", err
	}
	if idx < 0 {
		return "", nil
	}

	if items[idx].Value == "__new__" {
		name, ok, err := tui.RunPrompt("New branch name:", "feature/my-feature")
		if err != nil {
			return "", err
		}
		if !ok {
			return "", nil
		}
		return name, nil
	}

	return items[idx].Value, nil
}
