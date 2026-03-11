package user

import (
	"fmt"
	"os"

	"github.com/isac7722/git-extension/internal/config"
	"github.com/isac7722/git-extension/internal/git"
	"github.com/spf13/cobra"
)

var switchCmd = &cobra.Command{
	Use:    "switch <profile>",
	Short:  "Switch to a user profile (global)",
	Hidden: true, // Users invoke as "ge user <profile>" via cobra's arg handling
	Args:   cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return err
		}

		account, ok := cfg.Get(args[0])
		if !ok {
			return fmt.Errorf("profile %q not found. Available: %s", args[0], profileList(cfg))
		}

		return doSwitch(account, false)
	},
}

func init() {
	// Override the user command's Args handling to catch "ge user <profile>"
	parentRunE := Cmd.RunE
	Cmd.Args = cobra.ArbitraryArgs
	Cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return parentRunE(cmd, args)
		}

		// Check if first arg is a known subcommand
		for _, sub := range cmd.Commands() {
			if sub.Name() == args[0] {
				return nil // Let cobra handle it
			}
			for _, alias := range sub.Aliases {
				if alias == args[0] {
					return nil
				}
			}
		}

		// Treat as profile name
		cfg, err := loadConfig()
		if err != nil {
			return err
		}

		account, ok := cfg.Get(args[0])
		if !ok {
			return fmt.Errorf("unknown subcommand or profile: %q\nAvailable profiles: %s", args[0], profileList(cfg))
		}

		return doSwitch(account, false)
	}
}

// doSwitch performs the actual account switch.
func doSwitch(account *config.Account, local bool) error {
	// Verify SSH key exists
	if account.SSHKey != "" {
		if _, err := os.Stat(account.SSHKey); err != nil {
			fmt.Fprintf(os.Stderr, "⚠ SSH key not found: %s\n", account.SSHKey)
		}
	}

	// Set git config
	if err := git.SetUser(account.Name, account.Email, account.SSHKey, local); err != nil {
		return fmt.Errorf("failed to set user: %w", err)
	}

	scope := "global"
	if local {
		scope = "local"
	}

	// For global switch, output __GE_EVAL for shell wrapper
	if !local && account.SSHKey != "" {
		// SSH agent key add
		_ = git.SSHAddKey(account.SSHKey)

		// Output eval line for shell wrapper to export env var
		fmt.Printf("__GE_EVAL:export GIT_SSH_COMMAND=\"ssh -i %s\"\n", account.SSHKey)
	}

	fmt.Fprintf(os.Stderr, "✔ Switched to %s (%s) [%s]\n", account.Profile, account.Email, scope)
	return nil
}

func profileList(cfg *config.Config) string {
	var names []string
	for _, a := range cfg.Accounts {
		names = append(names, a.Profile)
	}
	result := ""
	for i, n := range names {
		if i > 0 {
			result += ", "
		}
		result += n
	}
	return result
}
