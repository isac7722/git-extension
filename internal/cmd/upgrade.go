package cmd

import (
	"fmt"
	"os"

	"github.com/isac7722/git-extension/internal/update"
	"github.com/spf13/cobra"
)

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade ge to the latest version",
	RunE: func(cmd *cobra.Command, args []string) error {
		current := versionStr
		if current == "dev" {
			fmt.Fprintln(os.Stderr, "Cannot upgrade a development build. Install a release version first.")
			return nil
		}

		fmt.Fprintf(os.Stderr, "Current version: v%s\n", current)
		fmt.Fprintf(os.Stderr, "Checking for updates...\n")

		latest, err := update.CheckLatest()
		if err != nil {
			return fmt.Errorf("failed to check for updates: %w", err)
		}

		if !update.IsNewer(current, latest) {
			fmt.Fprintf(os.Stderr, "Already up to date (v%s)\n", current)
			return nil
		}

		fmt.Fprintf(os.Stderr, "Upgrading to v%s...\n", latest)

		if update.IsHomebrew() {
			fmt.Fprintln(os.Stderr, "Detected Homebrew installation.")
		}

		if err := update.Upgrade(latest); err != nil {
			return fmt.Errorf("upgrade failed: %w", err)
		}

		fmt.Fprintf(os.Stderr, "✓ Successfully upgraded to v%s\n", latest)
		return nil
	},
}
