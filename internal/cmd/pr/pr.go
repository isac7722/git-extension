package pr

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/isac7722/git-extension/internal/gh"
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

func init() {
	Cmd.Flags().StringSlice("reviewer", nil, "Request reviewers (GitHub usernames)")
	Cmd.Flags().String("assignee", "@me", "PR assignee (default: @me)")
	Cmd.Flags().Bool("no-assign", false, "Skip automatic self-assignment")
	Cmd.Flags().Bool("draft", false, "Create as draft PR")
}

type prOptions struct {
	Head, Base, Title, Body string
	Reviewers               []string
	Assignee                string
	Draft                   bool
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

	// Resolve reviewers
	reviewers, err := resolveReviewers(cmd)
	if err != nil {
		return err
	}

	// Resolve assignee
	assignee := resolveAssignee(cmd)

	// Resolve draft
	draft, _ := cmd.Flags().GetBool("draft")

	opts := prOptions{
		Head:      head,
		Base:      base,
		Title:     title,
		Body:      body,
		Reviewers: reviewers,
		Assignee:  assignee,
		Draft:     draft,
	}

	// Show summary and confirm
	if !showSummary(opts) {
		fmt.Fprintln(os.Stderr, "Cancelled.")
		return nil
	}

	return createPR(opts)
}

func resolveReviewers(cmd *cobra.Command) ([]string, error) {
	// If --reviewer flag provided, use it
	if cmd.Flags().Changed("reviewer") {
		reviewers, _ := cmd.Flags().GetStringSlice("reviewer")
		return reviewers, nil
	}

	// Ask if user wants to add reviewers
	ok, err := tui.RunConfirm("Add reviewers?")
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, nil
	}

	// Fetch collaborators
	collaborators, err := gh.RepoCollaborators()
	if err != nil {
		fmt.Fprintf(os.Stderr, "⚠ Could not fetch collaborators: %v\n", err)
		return nil, nil
	}

	// Exclude current user
	currentUser, _ := gh.CurrentUsername()
	var items []tui.MultiItem
	for _, login := range collaborators {
		if login == currentUser {
			continue
		}
		items = append(items, tui.MultiItem{
			Label: login,
			Value: login,
		})
	}

	if len(items) == 0 {
		fmt.Fprintln(os.Stderr, "No reviewers available.")
		return nil, nil
	}

	indices, err := tui.RunMultiSelector(items, "Select reviewers:")
	if err != nil {
		return nil, err
	}

	var selected []string
	for _, i := range indices {
		selected = append(selected, items[i].Value)
	}
	return selected, nil
}

func resolveAssignee(cmd *cobra.Command) string {
	noAssign, _ := cmd.Flags().GetBool("no-assign")
	if noAssign {
		return ""
	}
	assignee, _ := cmd.Flags().GetString("assignee")
	return assignee
}

func showSummary(opts prOptions) bool {
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, tui.Bold.Render("── PR Summary ──────────────"))

	fmt.Fprintf(os.Stderr, "  %s → %s\n", opts.Head, opts.Base)
	fmt.Fprintf(os.Stderr, "  Title:     %s\n", opts.Title)

	draftLabel := "No"
	if opts.Draft {
		draftLabel = "Yes"
	}
	fmt.Fprintf(os.Stderr, "  Draft:     %s\n", draftLabel)

	if len(opts.Reviewers) > 0 {
		fmt.Fprintf(os.Stderr, "  Reviewers: %s\n", strings.Join(opts.Reviewers, ", "))
	} else {
		fmt.Fprintf(os.Stderr, "  Reviewers: %s\n", tui.Dim.Render("(none)"))
	}

	if opts.Assignee != "" {
		fmt.Fprintf(os.Stderr, "  Assignee:  %s\n", opts.Assignee)
	} else {
		fmt.Fprintf(os.Stderr, "  Assignee:  %s\n", tui.Dim.Render("(none)"))
	}

	fmt.Fprintln(os.Stderr, tui.Bold.Render("────────────────────────────"))
	fmt.Fprintln(os.Stderr)

	ok, err := tui.RunConfirm("Create this pull request?")
	if err != nil {
		return false
	}
	return ok
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
		selectedHead, err := selectHeadBranch()
		if err != nil {
			return "", "", err
		}
		head = selectedHead
	}

	if base == "" {
		selectedBase, err := selectBaseBranch(head)
		if err != nil {
			return "", "", err
		}
		base = selectedBase
	}
	return head, base, nil
}

func selectHeadBranch() (string, error) {
	branches, err := git.AllBranches()
	if err != nil {
		return git.CurrentBranch()
	}

	currentBranch, _ := git.CurrentBranch()

	var items []tui.SelectorItem
	for _, b := range branches {
		items = append(items, tui.SelectorItem{
			Label:    b.Name,
			Value:    b.Name,
			Hint:     b.Date,
			Selected: b.Name == currentBranch,
		})
	}

	if len(items) == 0 {
		return git.CurrentBranch()
	}

	idx, err := tui.RunSelector(items, "Select head branch (source):")
	if err != nil {
		return "", err
	}
	if idx < 0 {
		return "", fmt.Errorf("cancelled")
	}
	return items[idx].Value, nil
}

func selectBaseBranch(head string) (string, error) {
	branches, err := git.AllBranches()
	if err != nil {
		// fallback: 브랜치 목록 못 가져오면 DefaultBranch 사용
		return git.DefaultBranch(), nil
	}

	defaultBranch := git.DefaultBranch()

	var items []tui.SelectorItem
	for _, b := range branches {
		if b.Name == head {
			continue // head 브랜치는 base 후보에서 제외
		}
		items = append(items, tui.SelectorItem{
			Label:    b.Name,
			Value:    b.Name,
			Hint:     b.Date,
			Selected: b.Name == defaultBranch,
		})
	}

	if len(items) == 0 {
		return defaultBranch, nil
	}

	idx, err := tui.RunSelector(items, "Select base branch:")
	if err != nil {
		return "", err
	}
	if idx < 0 {
		return "", fmt.Errorf("cancelled")
	}
	return items[idx].Value, nil
}

func createPR(opts prOptions) error {
	ghArgs := []string{
		"pr", "create",
		"--head", opts.Head,
		"--base", opts.Base,
		"--title", opts.Title,
		"--body", opts.Body,
	}

	if len(opts.Reviewers) > 0 {
		ghArgs = append(ghArgs, "--reviewer", strings.Join(opts.Reviewers, ","))
	}
	if opts.Assignee != "" {
		ghArgs = append(ghArgs, "--assignee", opts.Assignee)
	}
	if opts.Draft {
		ghArgs = append(ghArgs, "--draft")
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
	_, err := git.Run("rev-parse", "--verify", "refs/remotes/origin/"+branch)
	if err == nil {
		return nil
	}

	fmt.Fprintf(os.Stderr, "Branch %q not found on remote. Pushing...\n", branch)
	_, err = git.Run("push", "-u", "origin", branch)
	if err != nil {
		return fmt.Errorf("failed to push branch: %w", err)
	}
	fmt.Fprintf(os.Stderr, "✔ Pushed %s to origin\n", branch)
	return nil
}
