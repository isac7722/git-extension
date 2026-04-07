package cmd

import (
	"fmt"
	"os"

	"github.com/isac7722/git-extension/internal/tui"
	"github.com/isac7722/git-extension/internal/uninstall"
	"github.com/spf13/cobra"
)

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Remove ge and all its configuration from this system",
	Long:  `Discovers and removes shell integration, config files, git config keys, and legacy artifacts.`,
	RunE:  runUninstall,
}

func init() {
	uninstallCmd.Flags().Bool("dry-run", false, "Preview what would be removed without removing anything")
}

func runUninstall(cmd *cobra.Command, args []string) error {
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	plan := uninstall.Discover()

	if len(plan.Targets) == 0 && plan.BinaryPath == "" {
		fmt.Fprintln(os.Stderr, "Nothing to uninstall.")
		return nil
	}

	fmt.Fprint(os.Stderr, uninstall.FormatPlan(plan))

	if dryRun {
		fmt.Fprintln(os.Stderr, "\n(dry-run: no changes made)")
		return nil
	}

	if len(plan.Targets) > 0 {
		ok, err := tui.RunConfirm("Proceed with uninstall?")
		if err != nil {
			return err
		}
		if !ok {
			fmt.Fprintln(os.Stderr, "Cancelled.")
			return nil
		}

		fmt.Fprintln(os.Stderr)
		results := uninstall.Execute(plan)
		for _, r := range results {
			fmt.Fprintln(os.Stderr, r)
		}
	}

	if plan.BinaryPath != "" {
		fmt.Fprintln(os.Stderr)
		if plan.IsHomebrew {
			fmt.Fprintln(os.Stderr, "To complete uninstall:")
			fmt.Fprintln(os.Stderr, "  brew uninstall ge")
		} else {
			fmt.Fprintln(os.Stderr, "To complete uninstall, remove the binary:")
			fmt.Fprintf(os.Stderr, "  sudo rm %s\n", plan.BinaryPath)
		}
	}

	fmt.Fprintln(os.Stderr, "\nRestart your shell to apply changes.")
	return nil
}
