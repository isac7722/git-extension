package user

import (
	"fmt"

	"github.com/isac7722/ge-cli/internal/migration"
	"github.com/spf13/cobra"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Migrate from legacy gituser config",
	RunE: func(cmd *cobra.Command, args []string) error {
		if !migration.HasLegacy() {
			fmt.Println("No legacy gituser config found. Nothing to migrate.")
			return nil
		}

		msg, err := migration.Migrate()
		if err != nil {
			return err
		}

		fmt.Println(msg)
		return nil
	},
}
