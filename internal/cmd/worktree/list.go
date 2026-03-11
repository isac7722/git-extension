package worktree

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/isac7722/git-extension/internal/git"
	"github.com/isac7722/git-extension/internal/tui"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Short:   "List worktrees with interactive selector",
	Aliases: []string{"ls"},
	RunE: func(cmd *cobra.Command, args []string) error {
		return runWorktreeList()
	},
}

func runWorktreeList() error {
	wts, err := git.ListWorktrees()
	if err != nil {
		return fmt.Errorf("failed to list worktrees: %w", err)
	}

	if len(wts) == 0 {
		fmt.Println("No worktrees found.")
		return nil
	}

	cwd, _ := os.Getwd()
	cwd, _ = filepath.Abs(cwd)

	var items []tui.SelectorItem
	for _, wt := range wts {
		label := wt.Branch
		if label == "" {
			label = "(bare)"
		}

		hint := fmt.Sprintf("%s  %s", shortenPath(wt.Path), wt.Status)
		selected := pathContains(cwd, wt.Path)

		items = append(items, tui.SelectorItem{
			Label:    label,
			Value:    wt.Path,
			Hint:     hint,
			Selected: selected,
		})
	}

	idx, err := tui.RunSelector(items, "Select worktree:")
	if err != nil {
		return err
	}
	if idx < 0 {
		return nil
	}

	// Output cd command for shell wrapper
	fmt.Printf("__GE_EVAL:cd %q\n", wts[idx].Path)
	return nil
}

func shortenPath(p string) string {
	home, _ := os.UserHomeDir()
	if rel, err := filepath.Rel(home, p); err == nil && len(rel) < len(p) {
		return "~/" + rel
	}
	return p
}

func pathContains(cwd, wtPath string) bool {
	absWt, _ := filepath.Abs(wtPath)
	return cwd == absWt
}
