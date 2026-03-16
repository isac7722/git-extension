package worktree

import (
	"fmt"
	"os"

	"github.com/isac7722/git-extension/internal/git"
	"github.com/isac7722/git-extension/internal/tui"
	"github.com/isac7722/git-extension/internal/worktreesetup"
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

		noSetup, _ := cmd.Flags().GetBool("no-setup")
		if !noSetup {
			runWorktreeSetup(absPath)
		}

		// cd into new worktree via shell wrapper
		fmt.Printf("__GE_EVAL:cd %q\n", absPath)
		return nil
	},
}

func init() {
	addCmd.Flags().Bool("no-setup", false, "Skip worktree environment setup")
}

func runWorktreeSetup(dstPath string) {
	mainPath, err := git.MainWorktreePath()
	if err != nil {
		return
	}

	cfg, err := worktreesetup.Load(mainPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "⚠ Failed to load %s: %s\n", worktreesetup.ConfigFile, err)
		return
	}

	if cfg != nil {
		// Scenario A: config exists, run automatically
		fmt.Fprintf(os.Stderr, "  Setting up worktree...\n")
		ok := worktreesetup.Run(cfg, mainPath, dstPath, false)
		if ok {
			fmt.Fprintf(os.Stderr, "✔ Worktree ready!\n")
		} else {
			fmt.Fprintf(os.Stderr, "⚠ Setup partially completed. Run: ge wt setup\n")
		}
		return
	}

	// Scenario B: no config, auto-detect
	suggestions := worktreesetup.Detect(mainPath)
	if len(suggestions) == 0 {
		return
	}

	fmt.Fprintf(os.Stderr, "\nNo %s found. Detected setup:\n", worktreesetup.ConfigFile)

	selectedCfg, err := selectAndBuildConfig(suggestions)
	if err != nil || selectedCfg == nil {
		return // cancelled or error
	}

	fmt.Fprintf(os.Stderr, "  Setting up worktree...\n")
	ok := worktreesetup.Run(selectedCfg, mainPath, dstPath, false)
	if ok {
		fmt.Fprintf(os.Stderr, "✔ Setup complete\n")
	} else {
		fmt.Fprintf(os.Stderr, "⚠ Setup partially completed. Run: ge wt setup\n")
	}

	// Offer to save config
	save, err := tui.RunConfirm("Save to " + worktreesetup.ConfigFile + "?")
	if err != nil || !save {
		return
	}

	if err := worktreesetup.Save(mainPath, selectedCfg); err != nil {
		fmt.Fprintf(os.Stderr, "⚠ Failed to save config: %s\n", err)
		return
	}

	fmt.Fprintf(os.Stderr, "✔ Saved %s\n", worktreesetup.ConfigFile)
	fmt.Fprintf(os.Stderr, "  Tip: commit this file to share with your team\n")
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
		return promptNewBranch()
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
		return promptNewBranch()
	}

	return items[idx].Value, nil
}

func promptNewBranch() (string, error) {
	name, ok, err := tui.RunPrompt("New branch name:", "feature/my-feature")
	if err != nil {
		return "", err
	}
	if !ok {
		return "", nil
	}
	return name, nil
}
