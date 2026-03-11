package user

import (
	"fmt"
	"os"

	"github.com/isac7722/git-extension/internal/config"
	"github.com/isac7722/git-extension/internal/tui"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new user profile interactively",
	RunE: func(cmd *cobra.Command, args []string) error {
		profile, ok, err := tui.RunPrompt("Profile name:", "e.g., work")
		if err != nil {
			return err
		}
		if !ok || profile == "" {
			return nil
		}

		name, ok, err := tui.RunPrompt("Full name:", "e.g., John Doe")
		if err != nil {
			return err
		}
		if !ok || name == "" {
			return nil
		}

		email, ok, err := tui.RunPrompt("Email:", "e.g., john@company.com")
		if err != nil {
			return err
		}
		if !ok || email == "" {
			return nil
		}

		sshKey, _, err := tui.RunPrompt("SSH key path (optional):", "e.g., ~/.ssh/id_ed25519")
		if err != nil {
			return err
		}

		// Load existing or create new config
		path := config.DefaultPath()
		cfg, err := config.Load(path)
		if err != nil {
			cfg = &config.Config{}
		}

		// Check for duplicate
		if _, exists := cfg.Get(profile); exists {
			return fmt.Errorf("profile %q already exists", profile)
		}

		cfg.AddAccount(config.Account{
			Profile: profile,
			Name:    name,
			Email:   email,
			SSHKey:  sshKey,
		})

		if err := cfg.Save(path); err != nil {
			return fmt.Errorf("failed to save: %w", err)
		}

		fmt.Fprintf(os.Stderr, "✔ Added profile %q (%s <%s>)\n", profile, name, email)
		return nil
	},
}
