package git

import (
	"strings"
)

// BranchEntry holds branch metadata for the branch switcher.
type BranchEntry struct {
	Name       string
	IsLocal    bool   // exists in refs/heads/
	IsRemote   bool   // exists in refs/remotes/origin/
	IsCurrent  bool
	IsWorktree bool   // checked out in another worktree
	Date       string // relative date (e.g., "3 days ago")
}

// AllBranches returns all local and remote branches sorted for the interactive switcher.
// Order: current → local+remote → local-only → remote-only (each group by committerdate).
func AllBranches() ([]BranchEntry, error) {
	current, _ := CurrentBranch()
	wtBranches := worktreeBranches()

	// Local branches with upstream info
	localOut, err := Run("for-each-ref", "--format=%(refname:short)\t%(upstream)\t%(committerdate:relative)", "--sort=-committerdate", "refs/heads/")
	if err != nil {
		return nil, err
	}

	localSet := make(map[string]bool)
	var currentEntry *BranchEntry
	var localRemote, localOnly []BranchEntry

	for _, line := range strings.Split(localOut, "\n") {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 3)
		name := parts[0]
		upstream := ""
		date := ""
		if len(parts) > 1 {
			upstream = parts[1]
		}
		if len(parts) > 2 {
			date = parts[2]
		}
		localSet[name] = true

		entry := BranchEntry{
			Name:       name,
			IsLocal:    true,
			IsRemote:   upstream != "",
			IsCurrent:  name == current,
			IsWorktree: wtBranches[name] && name != current,
			Date:       date,
		}

		if entry.IsCurrent {
			currentEntry = &entry
		} else if entry.IsRemote {
			localRemote = append(localRemote, entry)
		} else {
			localOnly = append(localOnly, entry)
		}
	}

	// Remote branches (only those not already local)
	remoteOut, _ := Run("for-each-ref", "--format=%(refname:short)\t%(committerdate:relative)", "--sort=-committerdate", "refs/remotes/origin/")
	var remoteOnly []BranchEntry
	for _, line := range strings.Split(remoteOut, "\n") {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 2)
		fullName := parts[0]
		date := ""
		if len(parts) > 1 {
			date = parts[1]
		}
		// Strip "origin/" prefix
		name := strings.TrimPrefix(fullName, "origin/")
		if name == "HEAD" || name == fullName || localSet[name] {
			continue
		}
		remoteOnly = append(remoteOnly, BranchEntry{
			Name:     name,
			IsRemote: true,
			Date:     date,
		})
	}

	// Assemble: current → local+remote → local-only → remote-only
	var result []BranchEntry
	if currentEntry != nil {
		result = append(result, *currentEntry)
	}
	result = append(result, localRemote...)
	result = append(result, localOnly...)
	result = append(result, remoteOnly...)

	return result, nil
}

// BranchInfo holds branch metadata for clean command.
type BranchInfo struct {
	Name   string
	Tag    string // "gone", "merged", "local", "gone+merged"
	Date   string // last commit date
}

// CurrentBranch returns the current branch name.
func CurrentBranch() (string, error) {
	return Run("rev-parse", "--abbrev-ref", "HEAD")
}

// DefaultBranch detects the default branch (main/master).
func DefaultBranch() string {
	// Try origin/HEAD
	out, err := Run("symbolic-ref", "refs/remotes/origin/HEAD")
	if err == nil {
		parts := strings.Split(out, "/")
		return parts[len(parts)-1]
	}
	// Fallback: check if main exists
	if _, err := Run("rev-parse", "--verify", "refs/heads/main"); err == nil {
		return "main"
	}
	return "master"
}

// ProtectedBranches returns the list of protected branch names.
func ProtectedBranches() []string {
	protected := []string{"main", "master", "develop", "dev", "staging", "production", "release"}

	// Check custom protected branches from git config
	custom, err := GetConfig("ge.clean.protected")
	if err == nil && custom != "" {
		for _, b := range strings.Split(custom, ",") {
			b = strings.TrimSpace(b)
			if b != "" {
				protected = append(protected, b)
			}
		}
	}

	return protected
}

// IsProtected checks if a branch name is in the protected list.
func IsProtected(branch string) bool {
	for _, p := range ProtectedBranches() {
		if branch == p {
			return true
		}
	}
	return false
}

// GoneBranches returns branches whose upstream tracking is gone.
func GoneBranches() ([]BranchInfo, error) {
	out, err := Run("for-each-ref", "--format=%(refname:short) %(upstream:track)", "refs/heads/")
	if err != nil {
		return nil, err
	}

	current, _ := CurrentBranch()
	wtBranches := worktreeBranches()
	var branches []BranchInfo
	for _, line := range strings.Split(out, "\n") {
		if !strings.Contains(line, "[gone]") {
			continue
		}
		name := strings.Fields(line)[0]
		if name == current || IsProtected(name) {
			continue
		}
		if wtBranches[name] {
			continue
		}
		date := branchDate(name)
		branches = append(branches, BranchInfo{Name: name, Tag: "gone", Date: date})
	}
	return branches, nil
}

// MergedBranches returns branches merged into the default branch.
func MergedBranches() ([]BranchInfo, error) {
	def := DefaultBranch()
	out, err := Run("branch", "--merged", def)
	if err != nil {
		return nil, err
	}

	current, _ := CurrentBranch()
	wtBranches := worktreeBranches()
	var branches []BranchInfo
	for _, line := range strings.Split(out, "\n") {
		name := strings.TrimSpace(line)
		name = strings.TrimPrefix(name, "* ")
		name = strings.TrimPrefix(name, "+ ")
		if name == "" || name == current || IsProtected(name) {
			continue
		}
		if wtBranches[name] {
			continue
		}
		date := branchDate(name)
		branches = append(branches, BranchInfo{Name: name, Tag: "merged", Date: date})
	}
	return branches, nil
}

// LocalOnlyBranches returns branches with no upstream tracking.
func LocalOnlyBranches() ([]BranchInfo, error) {
	out, err := Run("for-each-ref", "--format=%(refname:short) %(upstream)", "refs/heads/")
	if err != nil {
		return nil, err
	}

	current, _ := CurrentBranch()
	wtBranches := worktreeBranches()
	var branches []BranchInfo
	for _, line := range strings.Split(out, "\n") {
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		name := fields[0]
		hasUpstream := len(fields) > 1 && fields[1] != ""
		if hasUpstream || name == current || IsProtected(name) {
			continue
		}
		if wtBranches[name] {
			continue
		}
		date := branchDate(name)
		branches = append(branches, BranchInfo{Name: name, Tag: "local", Date: date})
	}
	return branches, nil
}

// DeleteBranch deletes a local branch. Use force for unmerged branches.
func DeleteBranch(name string, force bool) error {
	flag := "-d"
	if force {
		flag = "-D"
	}
	_, err := Run("branch", flag, name)
	return err
}

// worktreeBranches returns a set of branch names checked out in worktrees.
func worktreeBranches() map[string]bool {
	wts, err := ListWorktrees()
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

func branchDate(name string) string {
	out, err := Run("log", "-1", "--format=%cr", name)
	if err != nil {
		return ""
	}
	return out
}
