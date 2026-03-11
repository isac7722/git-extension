package branch

import (
	"fmt"
	"os"

	"github.com/isac7722/git-extension/internal/git"
	"github.com/isac7722/git-extension/internal/tui"
)

func runRemove(args []string) error {
	force, branches := parseRemoveArgs(args)

	if len(branches) == 0 {
		selected, err := selectBranchesToRemove()
		if err != nil {
			return err
		}
		if len(selected) == 0 {
			return nil
		}
		branches = selected
	}

	valid := filterValidBranches(branches)
	if len(valid) == 0 {
		fmt.Fprintln(os.Stderr, "No branches to delete.")
		return nil
	}

	if !force {
		if err := confirmRemoteDeletion(valid); err != nil {
			if err == errCancelled {
				return nil
			}
			return err
		}
	}

	deleteBranches(valid)
	return nil
}

func parseRemoveArgs(args []string) (bool, []string) {
	force := false
	var branches []string
	for _, a := range args {
		if a == "--force" || a == "-f" {
			force = true
		} else {
			branches = append(branches, a)
		}
	}
	return force, branches
}

func filterValidBranches(branches []string) []string {
	current, _ := git.CurrentBranch()
	wtBranches := worktreeBranchSet()
	var valid []string
	for _, name := range branches {
		if name == current {
			fmt.Fprintf(os.Stderr, "✘ Skipping '%s': current branch\n", name)
		} else if git.IsProtected(name) {
			fmt.Fprintf(os.Stderr, "✘ Skipping '%s': protected branch\n", name)
		} else if wtBranches[name] {
			fmt.Fprintf(os.Stderr, "✘ Skipping '%s': checked out in another worktree\n", name)
		} else {
			valid = append(valid, name)
		}
	}
	return valid
}

func confirmRemoteDeletion(valid []string) error {
	hasRemote := false
	for _, name := range valid {
		if git.HasRemoteBranch(name) {
			hasRemote = true
			break
		}
	}
	if !hasRemote {
		return nil
	}
	msg := fmt.Sprintf("Delete %d branch(es) including remote?", len(valid))
	ok, err := tui.RunConfirm(msg)
	if err != nil {
		return err
	}
	if !ok {
		return errCancelled
	}
	return nil
}

var errCancelled = fmt.Errorf("cancelled")

func deleteBranches(valid []string) {
	for _, name := range valid {
		local, remote, err := git.DeleteBranchFull(name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "✘ %s: %s\n", name, err)
			continue
		}
		scope := locationLabel(local, remote)
		fmt.Fprintf(os.Stderr, "✔ Deleted %s (%s)\n", name, scope)
	}
}

func selectBranchesToRemove() ([]string, error) {
	_, _ = git.Run("fetch", "--prune")
	allBranches, err := git.AllBranches()
	if err != nil {
		return nil, err
	}

	items, branchNames := buildBranchItems(allBranches)
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

func buildBranchItems(allBranches []git.BranchEntry) ([]tui.MultiItem, []string) {
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
	return items, branchNames
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
