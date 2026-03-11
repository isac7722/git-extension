package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestFilterRemoteBranches(t *testing.T) {
	output := "origin/main\norigin/feature\norigin/HEAD\norigin/dev"
	inUse := map[string]bool{"main": true}
	localSet := map[string]bool{"dev": true}

	result := filterRemoteBranches(output, inUse, localSet)

	if len(result) != 1 {
		t.Fatalf("expected 1 branch, got %d", len(result))
	}
	if result[0] != "origin/feature" {
		t.Errorf("expected origin/feature, got %s", result[0])
	}
}

func TestFilterRemoteBranches_Empty(t *testing.T) {
	result := filterRemoteBranches("", map[string]bool{}, map[string]bool{})
	if result != nil {
		t.Error("expected nil for empty input")
	}
}

func TestFilterRemoteBranches_AllFiltered(t *testing.T) {
	output := "origin/main\norigin/HEAD"
	inUse := map[string]bool{"main": true}
	localSet := map[string]bool{}

	result := filterRemoteBranches(output, inUse, localSet)
	if result != nil {
		t.Error("expected nil when all branches filtered")
	}
}

func TestListWorktrees_Integration(t *testing.T) {
	dir := initTestRepo(t)
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	wts, err := ListWorktrees()
	if err != nil {
		t.Fatal(err)
	}
	if len(wts) != 1 {
		t.Fatalf("expected 1 worktree, got %d", len(wts))
	}
	if !wts[0].IsMain {
		t.Error("first worktree should be marked as main")
	}
	if wts[0].Branch != "main" {
		t.Errorf("expected branch 'main', got %q", wts[0].Branch)
	}
}

func TestAddAndRemoveWorktree_Integration(t *testing.T) {
	dir := initTestRepo(t)
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	wtDir := filepath.Join(dir, "wt-feature")
	absPath, err := AddWorktree("feature", wtDir)
	if err != nil {
		t.Fatalf("AddWorktree failed: %v", err)
	}

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		t.Error("worktree directory should exist")
	}

	// Verify it shows up in list
	wts, _ := ListWorktrees()
	if len(wts) != 2 {
		t.Fatalf("expected 2 worktrees, got %d", len(wts))
	}

	found := false
	for _, wt := range wts {
		if wt.Branch == "feature" {
			found = true
			break
		}
	}
	if !found {
		t.Error("feature worktree not found in list")
	}

	// Remove
	err = RemoveWorktree(wtDir, false)
	if err != nil {
		t.Fatalf("RemoveWorktree failed: %v", err)
	}

	wts, _ = ListWorktrees()
	if len(wts) != 1 {
		t.Fatalf("expected 1 worktree after removal, got %d", len(wts))
	}
}

func TestAddWorktree_ExistingBranch(t *testing.T) {
	dir := initTestRepo(t)
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	// Create a branch first
	cmd := exec.Command("git", "branch", "existing")
	cmd.Dir = dir
	cmd.CombinedOutput()

	wtDir := filepath.Join(dir, "wt-existing")
	_, err := AddWorktree("existing", wtDir)
	if err != nil {
		t.Fatalf("AddWorktree with existing branch failed: %v", err)
	}
	defer RemoveWorktree(wtDir, true)

	wts, _ := ListWorktrees()
	found := false
	for _, wt := range wts {
		if wt.Branch == "existing" {
			found = true
		}
	}
	if !found {
		t.Error("existing branch worktree not found")
	}
}

func TestAddWorktree_DefaultDir(t *testing.T) {
	dir := initTestRepo(t)
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	// With empty dir, should use ../branch-name
	absPath, err := AddWorktree("auto-dir", "")
	if err != nil {
		t.Fatalf("AddWorktree with empty dir failed: %v", err)
	}
	defer RemoveWorktree(absPath, true)

	expected := filepath.Join("..", "auto-dir")
	absExpected, _ := filepath.Abs(expected)
	if absPath != absExpected {
		t.Errorf("expected path %q, got %q", absExpected, absPath)
	}
}

func TestMainWorktreePath_Integration(t *testing.T) {
	dir := initTestRepo(t)
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	path, err := MainWorktreePath()
	if err != nil {
		t.Fatal(err)
	}
	// Resolve symlinks (macOS /var -> /private/var)
	resolvedDir, _ := filepath.EvalSymlinks(dir)
	resolvedPath, _ := filepath.EvalSymlinks(path)
	if resolvedPath != resolvedDir {
		t.Errorf("expected %q, got %q", resolvedDir, resolvedPath)
	}
}

func TestWorktreeStatus_Clean(t *testing.T) {
	dir := initTestRepo(t)

	status := worktreeStatus(dir, "main")
	// Clean repo should show "✔" (no dirty files, no upstream so rev-list fails)
	if status != "✔" {
		t.Errorf("expected '✔' for clean repo, got %q", status)
	}
}

func TestWorktreeStatus_EmptyBranch(t *testing.T) {
	status := worktreeStatus("/tmp", "")
	if status != "" {
		t.Errorf("expected empty status for empty branch, got %q", status)
	}
}

func TestWorktreeStatus_Dirty(t *testing.T) {
	dir := initTestRepo(t)

	// Create an untracked file to make it dirty
	os.WriteFile(filepath.Join(dir, "dirty.txt"), []byte("dirty"), 0644)

	status := worktreeStatus(dir, "main")
	if status != "*" {
		t.Errorf("expected '*' for dirty repo, got %q", status)
	}
}
