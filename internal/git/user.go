package git

import "fmt"

// GetConfig reads a git config value.
func GetConfig(key string) (string, error) {
	return Run("config", "--get", key)
}

// SetConfigGlobal sets a git config value globally.
func SetConfigGlobal(key, value string) error {
	_, err := Run("config", "--global", key, value)
	return err
}

// SetConfigLocal sets a git config value for the current repo.
func SetConfigLocal(key, value string) error {
	_, err := Run("config", "--local", key, value)
	return err
}

// UnsetConfigLocal removes a git config key from the local repo.
func UnsetConfigLocal(key string) error {
	_, err := Run("config", "--local", "--unset", key)
	return err
}

// UnsetConfigGlobal removes a git config key globally.
func UnsetConfigGlobal(key string) error {
	_, err := Run("config", "--global", "--unset", key)
	return err
}

// CurrentUser returns the current git user name and email.
func CurrentUser() (name, email string) {
	name, _ = GetConfig("user.name")
	email, _ = GetConfig("user.email")
	return
}

// SetUser sets git user name, email, and optionally sshCommand.
func SetUser(name, email, sshKey string, local bool) error {
	set := SetConfigGlobal
	if local {
		set = SetConfigLocal
	}

	if err := set("user.name", name); err != nil {
		return fmt.Errorf("setting user.name: %w", err)
	}
	if err := set("user.email", email); err != nil {
		return fmt.Errorf("setting user.email: %w", err)
	}
	if sshKey != "" {
		sshCmd := fmt.Sprintf("ssh -i %s", sshKey)
		if err := set("core.sshCommand", sshCmd); err != nil {
			return fmt.Errorf("setting core.sshCommand: %w", err)
		}
	}
	return nil
}
