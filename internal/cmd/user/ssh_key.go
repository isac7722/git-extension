package user

import (
	"fmt"

	"github.com/isac7722/ge-cli/internal/config"
	"github.com/spf13/cobra"
)

var sshKeyCmd = &cobra.Command{
	Use:   "ssh-key <profile> [new_path]",
	Short: "View or update SSH key for a profile",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := config.DefaultPath()
		cfg, err := config.Load(path)
		if err != nil {
			return err
		}

		account, ok := cfg.Get(args[0])
		if !ok {
			return fmt.Errorf("profile %q not found", args[0])
		}

		if len(args) == 1 {
			// View
			if account.SSHKey == "" {
				fmt.Printf("No SSH key configured for %q\n", args[0])
			} else {
				fmt.Println(account.SSHKey)
			}
			return nil
		}

		// Update
		account.SSHKey = args[1]
		if err := cfg.Save(path); err != nil {
			return fmt.Errorf("failed to save: %w", err)
		}
		fmt.Printf("✔ Updated SSH key for %q to %s\n", args[0], args[1])
		return nil
	},
}
