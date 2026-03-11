package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print ge version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("ge %s\n", versionStr)
	},
}
