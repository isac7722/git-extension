package git

import (
	"fmt"
	"path/filepath"
	"strings"
)

// WorktreeInfo holds worktree metadata.
type WorktreeInfo struct {
	Path     string
	Branch   string
	IsMain   bool
	IsBare   bool
	Status   string // "✔", "*", or combined like "* ↑1 ↓2"
}

// ListWorktrees returns all worktrees with status info.
func ListWorktrees() ([]WorktreeInfo, error) {
	out, err := Run("worktree", "list", "--porcelain")
	if err != nil {
		return nil, err
	}

	var worktrees []WorktreeInfo
	var current WorktreeInfo

	for _, line := range strings.Split(out, "\n") {
		switch {
		case strings.HasPrefix(line, "worktree "):
			current = WorktreeInfo{Path: strings.TrimPrefix(line, "worktree ")}
		case strings.HasPrefix(line, "branch "):
			ref := strings.TrimPrefix(line, "branch ")
			current.Branch = strings.TrimPrefix(ref, "refs/heads/")
		case line == "bare":
			current.IsBare = true
		case line == "":
			if current.Path != "" {
				// Determine status
				current.Status = worktreeStatus(current.Path, current.Branch)
				worktrees = append(worktrees, current)
			}
			current = WorktreeInfo{}
		}
	}
	// Handle last entry without trailing newline
	if current.Path != "" {
		current.Status = worktreeStatus(current.Path, current.Branch)
		worktrees = append(worktrees, current)
	}

	// Mark main worktree
	if len(worktrees) > 0 {
		worktrees[0].IsMain = true
	}

	return worktrees, nil
}

func worktreeStatus(path, branch string) string {
	if branch == "" {
		return ""
	}

	var parts []string

	// Check dirty
	out, err := RunInDir(path, "status", "--porcelain")
	if err == nil && out != "" {
		parts = append(parts, "*")
	}

	// Check ahead/behind
	out, err = RunInDir(path, "rev-list", "--left-right", "--count", "@{upstream}...HEAD")
	if err == nil {
		fields := strings.Fields(out)
		if len(fields) == 2 {
			behind := fields[0]
			ahead := fields[1]
			if ahead != "0" {
				parts = append(parts, "↑"+ahead)
			}
			if behind != "0" {
				parts = append(parts, "↓"+behind)
			}
		}
	}

	if len(parts) == 0 {
		return "✔"
	}
	return strings.Join(parts, " ")
}

// AddWorktree creates a new worktree. Creates branch if it doesn't exist.
func AddWorktree(branch, dir string) (string, error) {
	if dir == "" {
		dir = filepath.Join("..", branch)
	}

	// Check if branch exists
	_, err := Run("rev-parse", "--verify", "refs/heads/"+branch)
	if err == nil {
		// Branch exists
		_, err = Run("worktree", "add", dir, branch)
	} else {
		// Check remote
		_, remoteErr := Run("rev-parse", "--verify", "refs/remotes/origin/"+branch)
		if remoteErr == nil {
			_, err = Run("worktree", "add", dir, branch)
		} else {
			_, err = Run("worktree", "add", "-b", branch, dir)
		}
	}
	if err != nil {
		return "", err
	}

	abs, _ := filepath.Abs(dir)
	return abs, nil
}

// RemoveWorktree removes a worktree by path.
func RemoveWorktree(path string, force bool) error {
	args := []string{"worktree", "remove", path}
	if force {
		args = append(args, "--force")
	}
	_, err := Run(args...)
	return err
}

// AvailableBranches returns branches not currently checked out in a worktree.
func AvailableBranches() (local []string, remote []string, err error) {
	// Get branches in worktrees
	wts, err := ListWorktrees()
	if err != nil {
		return nil, nil, err
	}
	inUse := make(map[string]bool)
	for _, wt := range wts {
		inUse[wt.Branch] = true
	}

	// Local branches
	out, err := Run("for-each-ref", "--format=%(refname:short)", "refs/heads/")
	if err != nil {
		return nil, nil, err
	}
	for _, b := range strings.Split(out, "\n") {
		b = strings.TrimSpace(b)
		if b != "" && !inUse[b] {
			local = append(local, b)
		}
	}

	// Remote branches
	out, err = Run("for-each-ref", "--format=%(refname:short)", "refs/remotes/origin/")
	if err == nil {
		for _, b := range strings.Split(out, "\n") {
			b = strings.TrimSpace(b)
			if b == "" || b == "origin/HEAD" {
				continue
			}
			shortName := strings.TrimPrefix(b, "origin/")
			if !inUse[shortName] {
				// Skip if there's already a local branch with same name
				alreadyLocal := false
				for _, lb := range local {
					if lb == shortName {
						alreadyLocal = true
						break
					}
				}
				if !alreadyLocal {
					remote = append(remote, b)
				}
			}
		}
	}

	return local, remote, nil
}

// MainWorktreePath returns the path of the main worktree.
func MainWorktreePath() (string, error) {
	wts, err := ListWorktrees()
	if err != nil {
		return "", err
	}
	if len(wts) == 0 {
		return "", fmt.Errorf("no worktrees found")
	}
	return wts[0].Path, nil
}
