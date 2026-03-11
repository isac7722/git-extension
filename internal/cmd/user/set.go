package user

import (
	"fmt"

	"github.com/spf13/cobra"
)

var setCmd = &cobra.Command{
	Use:   "set <profile>",
	Short: "Set user profile for current repository (local config)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return err
		}

		account, ok := cfg.Get(args[0])
		if !ok {
			return fmt.Errorf("profile %q not found", args[0])
		}

		return doSwitch(account, true)
	},
}
