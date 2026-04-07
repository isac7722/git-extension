package agent

import (
	"fmt"
	"os"
	"strings"

	"github.com/isac7722/git-extension/internal/config"
	"github.com/isac7722/git-extension/internal/tui"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage agent configuration",
}

var setKeyCmd = &cobra.Command{
	Use:   "set-key",
	Short: "Set Anthropic API key",
	RunE: func(cmd *cobra.Command, args []string) error {
		key, ok, err := tui.RunSecretPrompt("Anthropic API Key:", "sk-ant-...")
		if err != nil {
			return err
		}
		if !ok || key == "" {
			return nil
		}
		return saveAndConfirmKey(key)
	},
}

var showCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current agent configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		envKey := os.Getenv("ANTHROPIC_API_KEY")
		if envKey != "" {
			fmt.Fprintf(os.Stderr, "Anthropic API Key: %s [env]\n", maskKey(envKey))
			return nil
		}

		cfg, err := config.LoadAgentConfig()
		if err != nil || cfg.AnthropicAPIKey == "" {
			fmt.Fprintf(os.Stderr, "No API key configured.\n")
			fmt.Fprintf(os.Stderr, "Run: ge agent config set-key\n")
			return nil
		}

		fmt.Fprintf(os.Stderr, "Anthropic API Key: %s [%s]\n", maskKey(cfg.AnthropicAPIKey), config.AgentConfigPath())
		if cfg.Model != "" {
			fmt.Fprintf(os.Stderr, "Model: %s\n", cfg.Model)
		} else {
			fmt.Fprintf(os.Stderr, "Model: claude-sonnet-4-6 (default)\n")
		}
		return nil
	},
}

var removeKeyCmd = &cobra.Command{
	Use:   "remove-key",
	Short: "Remove stored API key",
	RunE: func(cmd *cobra.Command, args []string) error {
		ok, err := tui.RunConfirm("Remove stored API key?")
		if err != nil {
			return err
		}
		if !ok {
			return nil
		}
		if err := config.RemoveAgentConfig(); err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "API key removed.\n")
		return nil
	},
}

func init() {
	configCmd.AddCommand(setKeyCmd)
	configCmd.AddCommand(showCmd)
	configCmd.AddCommand(removeKeyCmd)
}

func maskKey(key string) string {
	if len(key) <= 11 {
		return strings.Repeat("*", len(key))
	}
	return key[:7] + "..." + key[len(key)-4:]
}

func saveAndConfirmKey(key string) error {
	cfg, _ := config.LoadAgentConfig()
	if cfg == nil {
		cfg = &config.AgentConfig{}
	}
	cfg.AnthropicAPIKey = key
	if err := config.SaveAgentConfig(cfg); err != nil {
		return fmt.Errorf("failed to save API key: %w", err)
	}
	fmt.Fprintf(os.Stderr, "✓ API key saved to %s\n", config.AgentConfigPath())
	return nil
}
