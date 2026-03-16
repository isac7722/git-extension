package worktree

import (
	"fmt"
	"os"

	"github.com/isac7722/git-extension/internal/git"
	"github.com/isac7722/git-extension/internal/tui"
	"github.com/isac7722/git-extension/internal/worktreesetup"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Generate .ge-worktree.yaml from project detection",
	Long:  "Detect project setup requirements and create a .ge-worktree.yaml configuration file.",
	RunE: func(cmd *cobra.Command, args []string) error {
		force, _ := cmd.Flags().GetBool("force")

		mainPath, err := git.MainWorktreePath()
		if err != nil {
			return err
		}

		// Check if config already exists
		if !force {
			existing, err := worktreesetup.Load(mainPath)
			if err != nil {
				return err
			}
			if existing != nil {
				return fmt.Errorf("%s already exists (use --force to overwrite)", worktreesetup.ConfigFile)
			}
		}

		suggestions := worktreesetup.Detect(mainPath)
		if len(suggestions) == 0 {
			fmt.Fprintf(os.Stderr, "No setup actions detected.\n")
			return nil
		}

		cfg, err := selectAndBuildConfig(suggestions)
		if err != nil {
			return err
		}
		if cfg == nil {
			return nil // cancelled
		}

		if err := worktreesetup.Save(mainPath, cfg); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		fmt.Fprintf(os.Stderr, "✔ Created %s\n", worktreesetup.ConfigFile)
		fmt.Fprintf(os.Stderr, "  Tip: commit this file to share with your team\n")
		return nil
	},
}

func init() {
	initCmd.Flags().Bool("force", false, "Overwrite existing .ge-worktree.yaml")
}

// selectAndBuildConfig shows a multi-selector for suggestions and builds a Config.
// Returns nil if the user cancelled.
func selectAndBuildConfig(suggestions []worktreesetup.Suggestion) (*worktreesetup.Config, error) {
	items := make([]tui.MultiItem, len(suggestions))
	for i, s := range suggestions {
		items[i] = tui.MultiItem{
			Label:   s.Label,
			Value:   s.Value,
			Checked: true,
		}
	}

	indices, err := tui.RunMultiSelector(items, "Detected setup:")
	if err != nil {
		return nil, err
	}
	if indices == nil {
		return nil, nil // cancelled
	}

	cfg := &worktreesetup.Config{}
	for _, idx := range indices {
		s := suggestions[idx]
		switch s.Type {
		case "copy":
			cfg.Copy = append(cfg.Copy, s.Value)
		case "setup":
			cfg.Setup = append(cfg.Setup, s.Value)
		}
	}

	return cfg, nil
}
