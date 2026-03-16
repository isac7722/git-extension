package protection

import (
	"fmt"
	"os"
	"strings"

	"github.com/isac7722/git-extension/internal/git"
	"github.com/isac7722/git-extension/internal/tui"
	"github.com/spf13/cobra"
)

// Cmd is the protection command.
var Cmd = &cobra.Command{
	Use:   "protection",
	Short: "Manage protected branches",
	Long:  `List, add, or remove protected branches. Default protected branches: main, prod, stg, dev.`,
	RunE:  runList,
}

var addCmd = &cobra.Command{
	Use:   "add <branch> [<branch>...]",
	Short: "Add branches to the protection list",
	Args:  cobra.MinimumNArgs(1),
	RunE:  runAdd,
}

var removeCmd = &cobra.Command{
	Use:   "remove [<branch>...]",
	Short: "Remove branches from the protection list",
	RunE:  runRemove,
}

var globalFlag bool

func init() {
	addCmd.Flags().BoolVar(&globalFlag, "global", false, "Apply to global git config (default: local)")
	removeCmd.Flags().BoolVar(&globalFlag, "global", false, "Apply to global git config (default: local)")
	Cmd.AddCommand(addCmd)
	Cmd.AddCommand(removeCmd)
}

func runList(cmd *cobra.Command, args []string) error {
	allDefaults := git.AllDefaultProtectedBranches()
	excludedSet := make(map[string]bool)
	for _, b := range git.ExcludedDefaultBranches() {
		excludedSet[b] = true
	}
	custom := git.CustomProtectedBranches()

	fmt.Println("Protected branches:")
	for _, b := range allDefaults {
		if excludedSet[b] {
			fmt.Printf("  %s (default, excluded)\n", b)
		} else {
			fmt.Printf("  %s (default)\n", b)
		}
	}
	for _, b := range custom {
		fmt.Printf("  %s (custom)\n", b)
	}

	if len(custom) == 0 && len(allDefaults) > 0 {
		fmt.Println("\nNo custom protected branches configured.")
	}

	return nil
}

func runAdd(cmd *cobra.Command, args []string) error {
	existing := git.CustomProtectedBranches()
	existingSet := make(map[string]bool)
	for _, b := range existing {
		existingSet[b] = true
	}

	// Check defaults (active ones)
	defaultSet := make(map[string]bool)
	for _, b := range git.DefaultProtectedBranches() {
		defaultSet[b] = true
	}

	// Check excluded defaults
	excludedSet := make(map[string]bool)
	for _, b := range git.ExcludedDefaultBranches() {
		excludedSet[b] = true
	}

	var toAdd []string
	for _, name := range args {
		// If it's an excluded default, restore it
		if excludedSet[name] {
			if err := git.RestoreDefaultBranch(name, globalFlag); err != nil {
				return fmt.Errorf("failed to restore default branch '%s': %w", name, err)
			}
			fmt.Fprintf(os.Stderr, "✔ Restored '%s' to default protected branches\n", name)
			continue
		}
		if defaultSet[name] {
			fmt.Fprintf(os.Stderr, "ℹ '%s' is already a default protected branch\n", name)
			continue
		}
		if existingSet[name] {
			fmt.Fprintf(os.Stderr, "ℹ '%s' is already protected\n", name)
			continue
		}
		toAdd = append(toAdd, name)
	}

	if len(toAdd) == 0 {
		return nil
	}

	newList := append(existing, toAdd...)
	value := strings.Join(newList, ",")

	var err error
	if globalFlag {
		err = git.SetConfigGlobal("ge.protected-branches", value)
	} else {
		err = git.SetConfigLocal("ge.protected-branches", value)
	}
	if err != nil {
		return fmt.Errorf("failed to update config: %w", err)
	}

	for _, name := range toAdd {
		fmt.Fprintf(os.Stderr, "✔ Added '%s' to protected branches\n", name)
	}
	return nil
}

func runRemove(cmd *cobra.Command, args []string) error {
	existing := git.CustomProtectedBranches()
	defaults := git.DefaultProtectedBranches()

	// Interactive selection when no args given
	if len(args) == 0 {
		if len(existing) == 0 && len(defaults) == 0 {
			fmt.Println("No protected branches to remove.")
			return nil
		}

		var items []tui.MultiItem
		for _, b := range defaults {
			items = append(items, tui.MultiItem{
				Label:   fmt.Sprintf("%s (default)", b),
				Value:   b,
				Checked: false,
			})
		}
		for _, b := range existing {
			items = append(items, tui.MultiItem{
				Label:   fmt.Sprintf("%s (custom)", b),
				Value:   b,
				Checked: false,
			})
		}

		if len(items) == 0 {
			fmt.Println("No protected branches to remove.")
			return nil
		}

		indices, err := tui.RunMultiSelector(items, "Select branches to unprotect:")
		if err != nil {
			return err
		}
		if indices == nil {
			fmt.Println("Cancelled.")
			return nil
		}

		for _, i := range indices {
			args = append(args, items[i].Value)
		}
		if len(args) == 0 {
			fmt.Println("No branches selected.")
			return nil
		}
	}

	defaultSet := make(map[string]bool)
	for _, b := range git.AllDefaultProtectedBranches() {
		defaultSet[b] = true
	}

	existingSet := make(map[string]bool)
	for _, b := range existing {
		existingSet[b] = true
	}

	customRemoveSet := make(map[string]bool)
	for _, name := range args {
		if defaultSet[name] {
			// Confirm before excluding a default branch
			confirmed, err := tui.RunConfirm(fmt.Sprintf("⚠ '%s' is a default protected branch. Remove?", name))
			if err != nil {
				return err
			}
			if !confirmed {
				fmt.Fprintf(os.Stderr, "  Skipped '%s'\n", name)
				continue
			}
			if err := git.ExcludeDefaultBranch(name, globalFlag); err != nil {
				return fmt.Errorf("failed to exclude default branch '%s': %w", name, err)
			}
			fmt.Fprintf(os.Stderr, "✔ Excluded '%s' from default protected branches\n", name)
			continue
		}
		if !existingSet[name] {
			fmt.Fprintf(os.Stderr, "ℹ '%s' is not in the protection list\n", name)
			continue
		}
		customRemoveSet[name] = true
		fmt.Fprintf(os.Stderr, "✔ Removed '%s' from protected branches\n", name)
	}

	if len(customRemoveSet) == 0 {
		return nil
	}

	// Build remaining list excluding removed branches
	var remaining []string
	for _, b := range existing {
		if !customRemoveSet[b] {
			remaining = append(remaining, b)
		}
	}

	if len(remaining) == 0 {
		var err error
		if globalFlag {
			err = git.UnsetConfigGlobal("ge.protected-branches")
		} else {
			err = git.UnsetConfigLocal("ge.protected-branches")
		}
		if err != nil {
			return fmt.Errorf("failed to update config: %w", err)
		}
	} else {
		value := strings.Join(remaining, ",")
		var err error
		if globalFlag {
			err = git.SetConfigGlobal("ge.protected-branches", value)
		} else {
			err = git.SetConfigLocal("ge.protected-branches", value)
		}
		if err != nil {
			return fmt.Errorf("failed to update config: %w", err)
		}
	}

	return nil
}
