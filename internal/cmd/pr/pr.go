package pr

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/isac7722/git-extension/internal/git"
	"github.com/isac7722/git-extension/internal/tui"
	"github.com/spf13/cobra"
)

// Cmd is the `ge pr` command.
var Cmd = &cobra.Command{
	Use:   "pr [head] [base]",
	Short: "Create a GitHub pull request",
	Long: `Create a GitHub pull request interactively.

Arguments:
  head  Source branch (default: current branch)
  base  Target branch (default: repository default branch)

If arguments are omitted, an interactive selector is shown.`,
	Args: cobra.MaximumNArgs(2),
	RunE: runPR,
}

func runPR(cmd *cobra.Command, args []string) error {
	head, base, err := resolveHeadBase(args)
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "Creating PR: %s → %s\n", head, base)

	if err := ensurePushed(head); err != nil {
		return err
	}

	// Prompt for title
	title, ok, err := tui.RunPrompt("PR title:", "")
	if err != nil {
		return err
	}
	if !ok || title == "" {
		fmt.Fprintln(os.Stderr, "Cancelled.")
		return nil
	}

	// Prompt for description
	body, ok, err := tui.RunPrompt("PR description:", "")
	if err != nil {
		return err
	}
	if !ok {
		fmt.Fprintln(os.Stderr, "Cancelled.")
		return nil
	}

	return createPR(head, base, title, body)
}

func resolveHeadBase(args []string) (string, string, error) {
	var head, base string
	switch len(args) {
	case 2:
		head = args[0]
		base = args[1]
	case 1:
		head = args[0]
	case 0:
		cur, err := git.CurrentBranch()
		if err != nil {
			return "", "", fmt.Errorf("failed to detect current branch: %w", err)
		}
		head = cur
	}
	if base == "" {
		base = git.DefaultBranch()
	}
	return head, base, nil
}

func createPR(head, base, title, body string) error {
	ghArgs := []string{
		"pr", "create",
		"--head", head,
		"--base", base,
		"--title", title,
		"--body", body,
	}

	ghCmd := exec.Command("gh", ghArgs...)
	var stdout, stderr strings.Builder
	ghCmd.Stdout = &stdout
	ghCmd.Stderr = &stderr

	if err := ghCmd.Run(); err != nil {
		errMsg := strings.TrimSpace(stderr.String())
		if errMsg == "" {
			errMsg = err.Error()
		}
		return fmt.Errorf("gh pr create failed: %s", errMsg)
	}

	url := strings.TrimSpace(stdout.String())
	if url != "" {
		fmt.Fprintf(os.Stderr, "✔ Pull request created: %s\n", url)
	}
	return nil
}

// ensurePushed checks if the branch exists on the remote, and pushes it if not.
func ensurePushed(branch string) error {
	// Check if remote tracking branch exists
	_, err := git.Run("rev-parse", "--verify", "refs/remotes/origin/"+branch)
	if err == nil {
		return nil // already pushed
	}

	fmt.Fprintf(os.Stderr, "Branch %q not found on remote. Pushing...\n", branch)
	_, err = git.Run("push", "-u", "origin", branch)
	if err != nil {
		return fmt.Errorf("failed to push branch: %w", err)
	}
	fmt.Fprintf(os.Stderr, "✔ Pushed %s to origin\n", branch)
	return nil
}
