package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestParseLocalBranches(t *testing.T) {
	output := "main\trefs/remotes/origin/main\t3 days ago\nfeature\t\t1 day ago\ndev\trefs/remotes/origin/dev\t2 hours ago"
	current := "main"
	wtBranches := map[string]bool{"dev": true}

	localSet, currentEntry, localRemote, localOnly := parseLocalBranches(output, current, wtBranches)

	// localSet should have all 3 branches
	if len(localSet) != 3 {
		t.Fatalf("expected 3 in localSet, got %d", len(localSet))
	}
	if !localSet["main"] || !localSet["feature"] || !localSet["dev"] {
		t.Error("localSet missing expected branches")
	}

	// current entry
	if currentEntry == nil {
		t.Fatal("currentEntry should not be nil")
	}
	if currentEntry.Name != "main" || !currentEntry.IsCurrent {
		t.Errorf("unexpected currentEntry: %+v", currentEntry)
	}
	if !currentEntry.IsRemote {
		t.Error("main should be marked as remote (has upstream)")
	}

	// localRemote: dev has upstream but is in worktree
	if len(localRemote) != 1 {
		t.Fatalf("expected 1 localRemote, got %d", len(localRemote))
	}
	if localRemote[0].Name != "dev" {
		t.Errorf("expected dev in localRemote, got %s", localRemote[0].Name)
	}
	if !localRemote[0].IsWorktree {
		t.Error("dev should be marked as worktree")
	}

	// localOnly: feature has no upstream
	if len(localOnly) != 1 {
		t.Fatalf("expected 1 localOnly, got %d", len(localOnly))
	}
	if localOnly[0].Name != "feature" {
		t.Errorf("expected feature in localOnly, got %s", localOnly[0].Name)
	}
}

func TestParseLocalBranches_EmptyOutput(t *testing.T) {
	localSet, currentEntry, localRemote, localOnly := parseLocalBranches("", "main", nil)
	if len(localSet) != 0 {
		t.Error("expected empty localSet")
	}
	if currentEntry != nil {
		t.Error("expected nil currentEntry")
	}
	if localRemote != nil {
		t.Error("expected nil localRemote")
	}
	if localOnly != nil {
		t.Error("expected nil localOnly")
	}
}

func TestParseRemoteBranches(t *testing.T) {
	output := "origin/main\t3 days ago\norigin/feature\t1 day ago\norigin/HEAD\t"
	localSet := map[string]bool{"main": true}

	result := parseRemoteBranches(output, localSet)

	if len(result) != 1 {
		t.Fatalf("expected 1 remote-only branch, got %d", len(result))
	}
	if result[0].Name != "feature" {
		t.Errorf("expected feature, got %s", result[0].Name)
	}
	if !result[0].IsRemote {
		t.Error("should be marked as remote")
	}
	if result[0].Date != "1 day ago" {
		t.Errorf("expected '1 day ago', got %q", result[0].Date)
	}
}

func TestParseRemoteBranches_EmptyOutput(t *testing.T) {
	result := parseRemoteBranches("", map[string]bool{})
	if result != nil {
		t.Error("expected nil for empty output")
	}
}

func TestParseRemoteBranches_SkipsNonOrigin(t *testing.T) {
	// Lines without "origin/" prefix should be skipped (name == fullName check)
	output := "upstream/feature\t1 day ago"
	result := parseRemoteBranches(output, map[string]bool{})
	if result != nil {
		t.Error("expected nil for non-origin branches")
	}
}

func TestIsProtected(t *testing.T) {
	tests := []struct {
		branch string
		want   bool
	}{
		{"main", true},
		{"master", true},
		{"develop", true},
		{"dev", true},
		{"staging", true},
		{"production", true},
		{"release", true},
		{"feature/foo", false},
		{"bugfix/bar", false},
		{"", false},
	}
	for _, tt := range tests {
		if got := IsProtected(tt.branch); got != tt.want {
			t.Errorf("IsProtected(%q) = %v, want %v", tt.branch, got, tt.want)
		}
	}
}

// initTestRepo creates a bare-bones git repo in a temp dir for integration tests.
func initTestRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	cmds := [][]string{
		{"git", "init", "--initial-branch=main"},
		{"git", "config", "user.name", "Test"},
		{"git", "config", "user.email", "test@test.com"},
	}
	for _, args := range cmds {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("%v failed: %s", args, out)
		}
	}

	// Create initial commit
	f, _ := os.Create(filepath.Join(dir, "README.md"))
	f.Close()
	cmd := exec.Command("git", "add", ".")
	cmd.Dir = dir
	cmd.CombinedOutput()
	cmd = exec.Command("git", "commit", "-m", "init")
	cmd.Dir = dir
	cmd.CombinedOutput()

	return dir
}

func TestCurrentBranch_Integration(t *testing.T) {
	dir := initTestRepo(t)
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	branch, err := CurrentBranch()
	if err != nil {
		t.Fatal(err)
	}
	if branch != "main" {
		t.Errorf("expected 'main', got %q", branch)
	}
}

func TestDefaultBranch_Integration(t *testing.T) {
	dir := initTestRepo(t)
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	// No remote, so should fallback to checking refs/heads/main
	branch := DefaultBranch()
	if branch != "main" {
		t.Errorf("expected 'main', got %q", branch)
	}
}

func TestDeleteBranch_Integration(t *testing.T) {
	dir := initTestRepo(t)
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	// Create a branch and switch back to main
	cmd := exec.Command("git", "branch", "to-delete")
	cmd.Dir = dir
	cmd.CombinedOutput()

	err := DeleteBranch("to-delete", false)
	if err != nil {
		t.Fatalf("DeleteBranch failed: %v", err)
	}

	// Verify branch is gone
	cmd = exec.Command("git", "rev-parse", "--verify", "refs/heads/to-delete")
	cmd.Dir = dir
	if err := cmd.Run(); err == nil {
		t.Error("branch should have been deleted")
	}
}

func TestDeleteBranch_ForceUnmerged(t *testing.T) {
	dir := initTestRepo(t)
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	// Create a branch with a divergent commit
	cmd := exec.Command("git", "checkout", "-b", "unmerged")
	cmd.Dir = dir
	cmd.CombinedOutput()
	f, _ := os.Create(filepath.Join(dir, "unmerged.txt"))
	f.Close()
	cmd = exec.Command("git", "add", ".")
	cmd.Dir = dir
	cmd.CombinedOutput()
	cmd = exec.Command("git", "commit", "-m", "unmerged commit")
	cmd.Dir = dir
	cmd.CombinedOutput()
	cmd = exec.Command("git", "checkout", "main")
	cmd.Dir = dir
	cmd.CombinedOutput()

	// Non-force should fail
	err := DeleteBranch("unmerged", false)
	if err == nil {
		t.Error("expected error deleting unmerged branch without force")
	}

	// Force should succeed
	err = DeleteBranch("unmerged", true)
	if err != nil {
		t.Fatalf("force delete should succeed: %v", err)
	}
}

func TestBranchDate(t *testing.T) {
	dir := initTestRepo(t)
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	date := branchDate("main")
	if date == "" {
		t.Error("expected non-empty date for main branch")
	}
}

func TestBranchDate_InvalidBranch(t *testing.T) {
	dir := initTestRepo(t)
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	date := branchDate("nonexistent")
	if date != "" {
		t.Errorf("expected empty date for nonexistent branch, got %q", date)
	}
}

func TestHasRemoteBranch_NoRemote(t *testing.T) {
	dir := initTestRepo(t)
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	if HasRemoteBranch("main") {
		t.Error("should return false when no remote exists")
	}
}
