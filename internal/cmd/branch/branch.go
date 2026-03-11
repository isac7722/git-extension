package branch

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/isac7722/ge-cli/internal/git"
	"github.com/isac7722/ge-cli/internal/tui"
	"github.com/spf13/cobra"
)

// Cmd is the branch command.
var Cmd = &cobra.Command{
	Use:                "branch",
	Short:              "Interactive branch switcher",
	DisableFlagParsing: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 {
			return gitBranchPassthrough(args)
		}
		return runInteractiveBranch()
	},
}

func runInteractiveBranch() error {
	branches, err := git.AllBranches()
	if err != nil {
		return err
	}
	if len(branches) == 0 {
		fmt.Fprintln(os.Stderr, "No branches found.")
		return nil
	}

	items := make([]tui.SelectorItem, len(branches))
	for i, b := range branches {
		hint := formatHint(b)
		items[i] = tui.SelectorItem{
			Label:         b.Name,
			Value:         b.Name,
			FormattedHint: hint,
			Selected:      b.IsCurrent,
		}
	}

	idx, err := tui.RunSelector(items, "Switch branch:")
	if err != nil {
		return err
	}
	if idx < 0 {
		return nil
	}

	selected := branches[idx]
	if selected.IsCurrent {
		fmt.Fprintf(os.Stderr, "Already on '%s'\n", selected.Name)
		return nil
	}

	return switchToBranch(selected)
}

func formatHint(b git.BranchEntry) string {
	var tag string
	switch {
	case b.IsLocal && b.IsRemote:
		tag = tui.Green.Render("remote")
	case !b.IsLocal && b.IsRemote:
		tag = tui.Dim.Render("remote")
	case b.IsLocal && !b.IsRemote:
		tag = tui.Dim.Render("local")
	}

	if b.IsWorktree {
		tag += "  " + tui.Yellow.Render("worktree")
	}

	if b.Date != "" {
		return tag + "  " + tui.Dim.Render(b.Date)
	}
	return tag
}

func switchToBranch(b git.BranchEntry) error {
	if b.IsWorktree {
		fmt.Fprintf(os.Stderr, "branch '%s' is checked out in another worktree (use 'ge worktree list' to see worktrees)\n", b.Name)
		return nil
	}

	// Try git switch first (works for both local and remote-tracking branches in git 2.23+)
	_, err := git.Run("switch", b.Name)
	if err != nil && !b.IsLocal && b.IsRemote {
		// Fallback for older git: create tracking branch explicitly
		_, err = git.Run("checkout", "-b", b.Name, "origin/"+b.Name)
	}
	if err != nil {
		return fmt.Errorf("failed to switch to branch '%s': %w", b.Name, err)
	}
	fmt.Fprintf(os.Stderr, "✔ Switched to branch '%s'\n", b.Name)
	return nil
}

func gitBranchPassthrough(args []string) error {
	cmd := exec.Command("git", append([]string{"branch"}, args...)...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		return err
	}
	return nil
}
