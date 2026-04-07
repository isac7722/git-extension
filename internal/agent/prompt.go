package agent

import (
	"fmt"

	"github.com/isac7722/git-extension/internal/git"
)

func buildSystemPrompt() string {
	var repoContext string
	if git.IsInsideWorkTree() {
		branch, _ := git.Run("branch", "--show-current")
		defaultBranch, _ := git.Run("symbolic-ref", "refs/remotes/origin/HEAD", "--short")
		status, _ := git.Run("status", "--short")
		recentLog, _ := git.Run("log", "--oneline", "-5")
		userName, _ := git.Run("config", "user.name")

		if status == "" {
			status = "(clean)"
		}
		repoContext = fmt.Sprintf(`
## Current Repository Context
- Branch: %s
- Default branch: %s
- Git user: %s
- Status:
%s
- Recent commits:
%s
`, branch, defaultBranch, userName, status, recentLog)
	} else {
		repoContext = "\n## No git repository detected in current directory.\n"
	}

	return fmt.Sprintf(`You are a Git automation agent. You help users manage their git repositories by executing git commands and GitHub CLI operations.

%s
## Guidelines
- Always check the current state before making changes (git status, git diff, etc.)
- Never force push to main/master without explicit confirmation
- Prefer creating new commits over amending existing ones
- Do not commit files that may contain secrets (.env, credentials, etc.)
- When creating commits, follow conventional commit style if the repo uses it
- When creating PRs, write clear titles and descriptions
- For destructive operations (reset --hard, branch -D, etc.), explain what will happen before executing

## Available Tools
You have access to git commands, file reading, and GitHub CLI. Use them to fulfill the user's request.
Respond in the same language the user uses.`, repoContext)
}
