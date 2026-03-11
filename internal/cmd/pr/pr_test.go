package pr

import (
	"testing"
)

func TestResolveHeadBase_TwoArgs(t *testing.T) {
	head, base, err := resolveHeadBase([]string{"feature", "main"})
	if err != nil {
		t.Fatal(err)
	}
	if head != "feature" {
		t.Errorf("expected head 'feature', got %q", head)
	}
	if base != "main" {
		t.Errorf("expected base 'main', got %q", base)
	}
}

func TestResolveHeadBase_OneArg(t *testing.T) {
	head, base, err := resolveHeadBase([]string{"feature"})
	if err != nil {
		t.Fatal(err)
	}
	if head != "feature" {
		t.Errorf("expected head 'feature', got %q", head)
	}
	// base should be auto-detected (DefaultBranch), not empty
	if base == "" {
		t.Error("base should not be empty")
	}
}

func TestResolveHeadBase_NoArgs(t *testing.T) {
	// This will call git.CurrentBranch() - only works inside a git repo
	// We're in the test repo, so this should work
	head, base, err := resolveHeadBase([]string{})
	if err != nil {
		t.Fatalf("resolveHeadBase with no args failed: %v", err)
	}
	if head == "" {
		t.Error("head should not be empty")
	}
	if base == "" {
		t.Error("base should not be empty")
	}
}

func TestCmd_MaxArgs(t *testing.T) {
	// Verify the cobra command is configured correctly
	if Cmd.Use != "pr [head] [base]" {
		t.Errorf("unexpected Use: %q", Cmd.Use)
	}
}
