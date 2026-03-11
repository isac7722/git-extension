package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/isac7722/ge-cli/internal/cmd/clean"
	"github.com/isac7722/ge-cli/internal/cmd/merge"
	"github.com/isac7722/ge-cli/internal/cmd/user"
	"github.com/isac7722/ge-cli/internal/cmd/worktree"
	"github.com/spf13/cobra"
)

var versionStr = "dev"

// SetVersionInfo sets version info from ldflags.
func SetVersionInfo(version, _, _ string) {
	versionStr = version
}

var rootCmd = &cobra.Command{
	Use:   "ge",
	Short: "Git Extension — multi-account, worktree, and branch management",
	Long: `ge is a lightweight CLI extending git with multi-account management,
enhanced worktree support, and branch cleanup utilities.`,
	// When running as "ge __run ...", strip __run and dispatch normally.
	// This is used by the shell wrapper for commands needing __GE_EVAL.
	SilenceUsage:  true,
	SilenceErrors: true,
	// If no subcommand matches, passthrough to git
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return cmd.Help()
		}
		return gitPassthrough(args)
	},
}

// Hidden __run command for shell wrapper protocol
var runCmd = &cobra.Command{
	Use:    "__run",
	Hidden: true,
	// Disable flag parsing so all args pass through
	DisableFlagParsing: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("__run requires arguments")
		}
		// Re-execute root with the provided arguments
		rootCmd.SetArgs(args)
		return rootCmd.Execute()
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(setupCmd)
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(user.Cmd)
	rootCmd.AddCommand(worktree.Cmd)
	rootCmd.AddCommand(clean.Cmd)
	rootCmd.AddCommand(merge.Cmd)

	// "wt" alias for worktree
	wtCmd := *worktree.Cmd
	wtCmd.Use = "wt"
	wtCmd.Hidden = true
	rootCmd.AddCommand(&wtCmd)
}

// Execute runs the root command.
func Execute() error {
	// Handle unknown subcommands as git passthrough
	rootCmd.FParseErrWhitelist.UnknownFlags = true

	// Check if the first arg is a known subcommand
	args := os.Args[1:]
	if len(args) > 0 && !isKnownCommand(args[0]) {
		return gitPassthrough(args)
	}

	return rootCmd.Execute()
}

func isKnownCommand(name string) bool {
	// Flags like --help, -h, --version should be handled by cobra
	if len(name) > 0 && name[0] == '-' {
		return true
	}
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == name {
			return true
		}
		for _, alias := range cmd.Aliases {
			if alias == name {
				return true
			}
		}
	}
	// Also check built-in cobra commands
	switch name {
	case "help", "completion":
		return true
	}
	return false
}

func gitPassthrough(args []string) error {
	cmd := exec.Command("git", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		return err
	}
	return nil
}
