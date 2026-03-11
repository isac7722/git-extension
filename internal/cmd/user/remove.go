package user

import (
	"fmt"

	"github.com/isac7722/ge-cli/internal/config"
	"github.com/isac7722/ge-cli/internal/tui"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:     "remove [profile]",
	Short:   "Remove a user profile",
	Aliases: []string{"rm"},
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return err
		}

		if len(cfg.Accounts) == 0 {
			fmt.Println("No accounts configured.")
			return nil
		}

		var profile string
		if len(args) == 1 {
			profile = args[0]
			if _, ok := cfg.Get(profile); !ok {
				return fmt.Errorf("profile %q not found", profile)
			}
		} else {
			items := accountsToSelectorItems(cfg)
			idx, err := tui.RunSelector(items, "Select account to remove:")
			if err != nil {
				return err
			}
			if idx < 0 {
				return nil
			}
			profile = cfg.Accounts[idx].Profile
		}

		confirmed, err := tui.RunConfirm(fmt.Sprintf("Remove profile %q?", profile))
		if err != nil {
			return err
		}
		if !confirmed {
			fmt.Println("Cancelled.")
			return nil
		}

		cfg.RemoveAccount(profile)
		if err := cfg.Save(config.DefaultPath()); err != nil {
			return fmt.Errorf("failed to save credentials: %w", err)
		}

		fmt.Printf("Profile %q removed.\n", profile)
		return nil
	},
}
