package user

import (
	"fmt"

	"github.com/isac7722/ge-cli/internal/config"
	"github.com/isac7722/ge-cli/internal/tui"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:     "update [profile]",
	Short:   "Update a user profile",
	Aliases: []string{"edit"},
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return err
		}

		if len(cfg.Accounts) == 0 {
			fmt.Println("No accounts configured. Run 'ge user add' to create one.")
			return nil
		}

		var profile string
		var account *config.Account
		if len(args) == 1 {
			profile = args[0]
			a, ok := cfg.Get(profile)
			if !ok {
				return fmt.Errorf("profile %q not found", profile)
			}
			account = a
		} else {
			items := accountsToSelectorItems(cfg)
			idx, err := tui.RunSelector(items, "Select account to update:")
			if err != nil {
				return err
			}
			if idx < 0 {
				return nil
			}
			profile = cfg.Accounts[idx].Profile
			account = &cfg.Accounts[idx]
		}

		newProfile, ok, err := tui.RunPromptWithValue("Profile:", "e.g., work", profile)
		if err != nil {
			return err
		}
		if !ok {
			fmt.Println("Cancelled.")
			return nil
		}
		if newProfile == "" {
			newProfile = profile
		}
		if newProfile != profile {
			if _, exists := cfg.Get(newProfile); exists {
				return fmt.Errorf("profile %q already exists", newProfile)
			}
		}

		name, ok, err := tui.RunPromptWithValue("Name:", "e.g., John Doe", account.Name)
		if err != nil {
			return err
		}
		if !ok {
			fmt.Println("Cancelled.")
			return nil
		}
		if name == "" {
			name = account.Name
		}

		email, ok, err := tui.RunPromptWithValue("Email:", "e.g., john@company.com", account.Email)
		if err != nil {
			return err
		}
		if !ok {
			fmt.Println("Cancelled.")
			return nil
		}
		if email == "" {
			email = account.Email
		}

		sshKey, ok, err := tui.RunPromptWithValue("SSH key path:", "e.g., ~/.ssh/id_ed25519", account.SSHKey)
		if err != nil {
			return err
		}
		if !ok {
			fmt.Println("Cancelled.")
			return nil
		}
		if sshKey == "" {
			sshKey = account.SSHKey
		}

		updated := config.Account{
			Profile: newProfile,
			Name:    name,
			Email:   email,
			SSHKey:  sshKey,
		}

		cfg.UpdateAccount(profile, updated)
		if err := cfg.Save(config.DefaultPath()); err != nil {
			return fmt.Errorf("failed to save credentials: %w", err)
		}

		fmt.Printf("✔ Updated profile %q (%s <%s>)\n", newProfile, name, email)
		return doSwitch(&updated, true)
	},
}
