package agent

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	agentpkg "github.com/isac7722/git-extension/internal/agent"
	"github.com/isac7722/git-extension/internal/tui"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "agent [prompt]",
	Short: "AI-powered git agent using Claude",
	Long: `Start an AI agent that can execute git commands, read files, and manage
GitHub PRs on your behalf. Requires an Anthropic API key.

Without arguments, starts an interactive session.
With arguments, runs a single task and exits.`,
	Example: `  ge agent "create a PR for the current changes"
  ge agent "summarize recent commits"
  ge agent                    # interactive mode
  ge agent config set-key     # set API key
  ge agent config show        # show current key`,
	RunE: func(cmd *cobra.Command, args []string) error {
		a, err := agentpkg.New()
		if err != nil {
			var noKey *agentpkg.ErrNoAPIKey
			if errors.As(err, &noKey) {
				a, err = handleFirstTimeSetup()
				if err != nil {
					return err
				}
				if a == nil {
					return nil
				}
			} else {
				return err
			}
		}

		prompt := strings.Join(args, " ")
		return a.Run(context.Background(), prompt)
	},
}

func init() {
	Cmd.AddCommand(configCmd)
}

func handleFirstTimeSetup() (*agentpkg.Agent, error) {
	fmt.Fprintln(os.Stderr, "No Anthropic API key found. Let's set one up.")
	fmt.Fprintln(os.Stderr, "Get your key at https://console.anthropic.com/settings/keys")
	fmt.Fprintln(os.Stderr)

	key, ok, err := tui.RunSecretPrompt("Anthropic API Key:", "sk-ant-...")
	if err != nil {
		return nil, err
	}
	if !ok || key == "" {
		return nil, nil
	}

	if err := saveAndConfirmKey(key); err != nil {
		return nil, err
	}

	// Set env var so New() picks it up
	os.Setenv("ANTHROPIC_API_KEY", key)
	return agentpkg.New()
}
