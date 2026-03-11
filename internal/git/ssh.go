package git

import (
	"fmt"
	"os/exec"
	"runtime"
)

// SSHAddKey adds an SSH key to the agent.
func SSHAddKey(keyPath string) error {
	// Start ssh-agent if not running
	if err := ensureSSHAgent(); err != nil {
		return err
	}

	// On macOS, try --apple-use-keychain first
	if runtime.GOOS == "darwin" {
		cmd := exec.Command("ssh-add", "--apple-use-keychain", keyPath)
		if err := cmd.Run(); err == nil {
			return nil
		}
	}

	cmd := exec.Command("ssh-add", keyPath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ssh-add %s: %w", keyPath, err)
	}
	return nil
}

func ensureSSHAgent() error {
	// Check if SSH_AUTH_SOCK is set (agent already running)
	cmd := exec.Command("ssh-add", "-l")
	if err := cmd.Run(); err == nil {
		return nil
	}
	// Agent might not be running, but we can't start it from a subprocess
	// in a way that affects the parent shell. The shell wrapper handles this.
	return nil
}
