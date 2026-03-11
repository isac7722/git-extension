package user

import (
	"fmt"

	"github.com/isac7722/ge-cli/internal/config"
	"github.com/isac7722/ge-cli/internal/tui"
	"github.com/spf13/cobra"
)

// Cmd is the user parent command.
var Cmd = &cobra.Command{
	Use:   "user",
	Short: "Manage git user accounts",
	Long:  "Switch between git user accounts, manage SSH keys, and configure profiles.",
	RunE: func(cmd *cobra.Command, args []string) error {
		// No subcommand: show interactive selector
		return runInteractiveSelect()
	},
}

func init() {
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(currentCmd)
	Cmd.AddCommand(addCmd)
	Cmd.AddCommand(setCmd)
	Cmd.AddCommand(sshKeyCmd)
	Cmd.AddCommand(switchCmd)
	Cmd.AddCommand(migrateCmd)
	Cmd.AddCommand(removeCmd)
}

func loadConfig() (*config.Config, error) {
	path := config.DefaultPath()
	cfg, err := config.Load(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load credentials from %s: %w\nRun 'ge user add' to create your first profile", path, err)
	}
	return cfg, nil
}

func runInteractiveSelect() error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	if len(cfg.Accounts) == 0 {
		fmt.Println("No accounts configured. Run 'ge user add' to create one.")
		return nil
	}

	items := accountsToSelectorItems(cfg)
	idx, err := tui.RunSelector(items, "Select account to switch to:")
	if err != nil {
		return err
	}
	if idx < 0 {
		return nil
	}

	account := cfg.Accounts[idx]
	return doSwitch(&account, false)
}
