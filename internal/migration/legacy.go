package migration

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/isac7722/ge-cli/internal/config"
)

// LegacyPath returns the path to the old gituser config directory.
func LegacyPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "gituser")
}

// HasLegacy checks if legacy gituser config exists.
func HasLegacy() bool {
	_, err := os.Stat(filepath.Join(LegacyPath(), "accounts"))
	return err == nil
}

// Migrate converts legacy gituser accounts to ge credentials format.
func Migrate() (string, error) {
	legacyDir := LegacyPath()
	accountsFile := filepath.Join(legacyDir, "accounts")

	f, err := os.Open(accountsFile)
	if err != nil {
		return "", fmt.Errorf("no legacy accounts found at %s", accountsFile)
	}
	defer func() { _ = f.Close() }()

	cfg := &config.Config{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Format: aliases:name:email:ssh_key
		// aliases can be comma-separated (e.g., "work,w")
		parts := strings.SplitN(line, ":", 4)
		if len(parts) < 3 {
			continue
		}

		aliases := strings.Split(parts[0], ",")
		profile := strings.TrimSpace(aliases[0])
		name := strings.TrimSpace(parts[1])
		email := strings.TrimSpace(parts[2])
		sshKey := ""
		if len(parts) > 3 {
			sshKey = strings.TrimSpace(parts[3])
		}

		cfg.AddAccount(config.Account{
			Profile: profile,
			Name:    name,
			Email:   email,
			SSHKey:  sshKey,
		})
	}

	if len(cfg.Accounts) == 0 {
		return "", fmt.Errorf("no accounts found in legacy config")
	}

	// Save to new location
	credPath := config.DefaultPath()
	if err := cfg.Save(credPath); err != nil {
		return "", fmt.Errorf("failed to save credentials: %w", err)
	}

	// Backup legacy directory
	backupName := fmt.Sprintf("%s.bak.%s", legacyDir, time.Now().Format("20060102"))
	if err := os.Rename(legacyDir, backupName); err != nil {
		// Non-fatal: just warn
		return fmt.Sprintf("Migrated %d accounts to %s\nWarning: could not backup legacy dir: %v",
			len(cfg.Accounts), credPath, err), nil
	}

	return fmt.Sprintf("Migrated %d accounts to %s\nLegacy config backed up to %s",
		len(cfg.Accounts), credPath, backupName), nil
}
