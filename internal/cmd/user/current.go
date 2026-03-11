package user

import (
	"fmt"

	"github.com/isac7722/git-extension/internal/git"
	"github.com/spf13/cobra"
)

var currentCmd = &cobra.Command{
	Use:   "current",
	Short: "Show current git user",
	Run: func(cmd *cobra.Command, args []string) {
		name, email := git.CurrentUser()
		if name == "" && email == "" {
			fmt.Println("No git user configured")
			return
		}
		fmt.Printf("%s <%s>\n", name, email)
	},
}
