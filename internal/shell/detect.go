package shell

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// DetectShell returns the user's current shell name.
func DetectShell() string {
	shell := os.Getenv("SHELL")
	if shell == "" {
		if runtime.GOOS == "darwin" {
			return "zsh"
		}
		return "bash"
	}
	base := filepath.Base(shell)
	switch base {
	case "zsh":
		return "zsh"
	case "bash":
		return "bash"
	default:
		return base
	}
}

// DetectRCFile returns the path to the shell's RC file.
func DetectRCFile(shell string) string {
	home, _ := os.UserHomeDir()
	switch shell {
	case "zsh":
		return filepath.Join(home, ".zshrc")
	case "bash":
		// Prefer .bashrc, check .bash_profile too
		bashrc := filepath.Join(home, ".bashrc")
		if _, err := os.Stat(bashrc); err == nil {
			return bashrc
		}
		profile := filepath.Join(home, ".bash_profile")
		if _, err := os.Stat(profile); err == nil {
			return profile
		}
		return bashrc
	default:
		return filepath.Join(home, ".profile")
	}
}

// HasMarker checks if the RC file already contains the ge-cli marker.
func HasMarker(rcPath string) bool {
	data, err := os.ReadFile(rcPath)
	if err != nil {
		return false
	}
	return strings.Contains(string(data), "# >>> ge-cli >>>") ||
		strings.Contains(string(data), "# >>> git-extension >>>")
}
