package agent

import (
	"context"
	"strings"

	agentpkg "github.com/isac7722/git-extension/internal/agent"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "agent [prompt]",
	Short: "AI-powered git agent using Claude",
	Long: `Start an AI agent that can execute git commands, read files, and manage
GitHub PRs on your behalf. Requires ANTHROPIC_API_KEY environment variable.

Without arguments, starts an interactive session.
With arguments, runs a single task and exits.`,
	Example: `  ge agent "create a PR for the current changes"
  ge agent "summarize recent commits"
  ge agent  # interactive mode`,
	RunE: func(cmd *cobra.Command, args []string) error {
		a, err := agentpkg.New()
		if err != nil {
			return err
		}

		prompt := strings.Join(args, " ")
		return a.Run(context.Background(), prompt)
	},
}
