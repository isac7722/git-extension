package shell

import (
	_ "embed"
	"fmt"
)

//go:embed ge_wrapper.zsh
var zshWrapper string

//go:embed ge_wrapper.bash
var bashWrapper string

// InitScript returns the shell initialization script for the given shell.
func InitScript(shell string) (string, error) {
	switch shell {
	case "zsh":
		return zshWrapper, nil
	case "bash":
		return bashWrapper, nil
	default:
		return "", fmt.Errorf("unsupported shell: %s (use 'zsh' or 'bash')", shell)
	}
}
