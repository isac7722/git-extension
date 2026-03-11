package cmd

import (
	"fmt"

	"github.com/isac7722/git-extension/internal/shell"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init [shell]",
	Short: "Output shell integration script",
	Long:  `Outputs shell integration code. Add 'eval "$(ge init zsh)"' to your .zshrc.`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		shellName := shell.DetectShell()
		if len(args) > 0 {
			shellName = args[0]
		}

		script, err := shell.InitScript(shellName)
		if err != nil {
			return err
		}

		fmt.Print(script)
		return nil
	},
}
