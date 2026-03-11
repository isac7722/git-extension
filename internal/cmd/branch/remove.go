package branch

import (
	"fmt"
	"os"

	"github.com/isac7722/ge-cli/internal/git"
	"github.com/isac7722/ge-cli/internal/tui"
)

func runRemove(args []string) error {
	// Parse --force/-f flag
	force := false
	var branches []string
	for _, a := range args {
		if a == "--force" || a == "-f" {
			force = true
		} else {
			branches = append(branches, a)
		}
	}

	if len(branches) == 0 {
		// Interactive mode
		selected, err := selectBranchesToRemove()
		if err != nil {
			return err
		}
		if len(selected) == 0 {
			return nil
		}
		branches = selected
	}

	// Validate and filter
	current, _ := git.CurrentBranch()
	wtBranches := worktreeBranchSet()
	var valid []string
	for _, name := range branches {
		if name == current {
			fmt.Fprintf(os.Stderr, "✘ Skipping '%s': current branch\n", name)
			continue
		}
		if git.IsProtected(name) {
			fmt.Fprintf(os.Stderr, "✘ Skipping '%s': protected branch\n", name)
			continue
		}
		if wtBranches[name] {
			fmt.Fprintf(os.Stderr, "✘ Skipping '%s': checked out in another worktree\n", name)
			continue
		}
		valid = append(valid, name)
	}

	if len(valid) == 0 {
		fmt.Fprintln(os.Stderr, "No branches to delete.")
		return nil
	}

	// Check if any have remotes — prompt confirmation if so
	if !force {
		hasRemote := false
		for _, name := range valid {
			if git.HasRemoteBranch(name) {
				hasRemote = true
				break
			}
		}
		if hasRemote {
			msg := fmt.Sprintf("Delete %d branch(es) including remote?", len(valid))
			ok, err := tui.RunConfirm(msg)
			if err != nil {
				return err
			}
			if !ok {
				return nil
			}
		}
	}

	// Delete
	for _, name := range valid {
		local, remote, err := git.DeleteBranchFull(name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "✘ %s: %s\n", name, err)
			continue
		}
		scope := locationLabel(local, remote)
		fmt.Fprintf(os.Stderr, "✔ Deleted %s (%s)\n", name, scope)
	}

	return nil
}

func selectBranchesToRemove() ([]string, error) {
	git.Run("fetch", "--prune")
	allBranches, err := git.AllBranches()
	if err != nil {
		return nil, err
	}

	current, _ := git.CurrentBranch()
	wtBranches := worktreeBranchSet()

	var items []tui.MultiItem
	var branchNames []string
	for _, b := range allBranches {
		if b.Name == current || git.IsProtected(b.Name) || wtBranches[b.Name] {
			continue
		}
		hint := locationLabel(b.IsLocal, b.IsRemote)
		if b.Date != "" {
			hint += "  " + b.Date
		}
		items = append(items, tui.MultiItem{
			Label:   b.Name,
			Value:   b.Name,
			Hint:    hint,
			Checked: false,
		})
		branchNames = append(branchNames, b.Name)
	}

	if len(items) == 0 {
		fmt.Fprintln(os.Stderr, "No deletable branches found.")
		return nil, nil
	}

	indices, err := tui.RunMultiSelector(items)
	if err != nil {
		return nil, err
	}

	var selected []string
	for _, i := range indices {
		selected = append(selected, branchNames[i])
	}
	return selected, nil
}

func locationLabel(local, remote bool) string {
	switch {
	case local && remote:
		return "local + remote"
	case remote:
		return "remote"
	default:
		return "local"
	}
}

func worktreeBranchSet() map[string]bool {
	wts, err := git.ListWorktrees()
	if err != nil {
		return nil
	}
	m := make(map[string]bool)
	for _, wt := range wts {
		if wt.Branch != "" {
			m[wt.Branch] = true
		}
	}
	return m
}
