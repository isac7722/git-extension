package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// Run executes a git command and returns combined output.
func Run(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		errMsg := strings.TrimSpace(stderr.String())
		if errMsg == "" {
			errMsg = strings.TrimSpace(stdout.String())
		}
		return "", fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, errMsg)
	}
	return strings.TrimSpace(stdout.String()), nil
}

// RunInDir executes a git command in a specific directory.
func RunInDir(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		errMsg := strings.TrimSpace(stderr.String())
		if errMsg == "" {
			errMsg = strings.TrimSpace(stdout.String())
		}
		return "", fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, errMsg)
	}
	return strings.TrimSpace(stdout.String()), nil
}

// Passthrough executes a git command with stdin/stdout/stderr inherited.
func Passthrough(args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Stdin = nil
	cmd.Stdout = nil // inherits
	cmd.Stderr = nil // inherits
	// We need to set these to os.Std* explicitly
	cmd.Stdin = nil
	return cmd.Run()
}

// IsInsideWorkTree checks if the current directory is inside a git repo.
func IsInsideWorkTree() bool {
	_, err := Run("rev-parse", "--is-inside-work-tree")
	return err == nil
}

// TopLevel returns the top-level directory of the current git repo.
func TopLevel() (string, error) {
	return Run("rev-parse", "--show-toplevel")
}
