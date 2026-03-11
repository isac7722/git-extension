package user

import (
	"fmt"

	"github.com/isac7722/git-extension/internal/config"
	"github.com/isac7722/git-extension/internal/git"
	"github.com/isac7722/git-extension/internal/tui"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Short:   "List all user profiles with interactive selector",
	Aliases: []string{"ls"},
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadConfig()
		if err != nil {
			return err
		}

		if len(cfg.Accounts) == 0 {
			fmt.Println("No accounts configured. Run 'ge user add' to create one.")
			return nil
		}

		items := accountsToSelectorItems(cfg)
		idx, err := tui.RunSelector(items, "Select account to switch to:")
		if err != nil {
			return err
		}
		if idx < 0 {
			return nil
		}

		account := cfg.Accounts[idx]
		return doSwitch(&account, false)
	},
}

func accountsToSelectorItems(cfg *config.Config) []tui.SelectorItem {
	currentName, currentEmail := git.CurrentUser()
	var items []tui.SelectorItem
	for _, a := range cfg.Accounts {
		selected := a.Name == currentName && a.Email == currentEmail
		items = append(items, tui.SelectorItem{
			Label:    a.Profile,
			Value:    a.Profile,
			Hint:     fmt.Sprintf("%s <%s>", a.Name, a.Email),
			Selected: selected,
		})
	}
	return items
}
