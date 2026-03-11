package migration

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/isac7722/git-extension/internal/config"
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

	cfg, err := loadLegacyAccounts(accountsFile)
	if err != nil {
		return "", err
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
		return fmt.Sprintf("Migrated %d accounts to %s\nWarning: could not backup legacy dir: %v",
			len(cfg.Accounts), credPath, err), nil
	}

	return fmt.Sprintf("Migrated %d accounts to %s\nLegacy config backed up to %s",
		len(cfg.Accounts), credPath, backupName), nil
}

func loadLegacyAccounts(path string) (*config.Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("no legacy accounts found at %s", path)
	}
	defer func() { _ = f.Close() }()

	cfg := config.NewConfig()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if acct, ok := parseLegacyAccount(line); ok {
			cfg.AddAccount(acct)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading legacy config: %w", err)
	}
	return cfg, nil
}

func parseLegacyAccount(line string) (config.Account, bool) {
	// Format: aliases:name:email:ssh_key
	parts := strings.SplitN(line, ":", 4)
	if len(parts) < 3 {
		return config.Account{}, false
	}

	aliases := strings.Split(parts[0], ",")
	sshKey := ""
	if len(parts) > 3 {
		sshKey = strings.TrimSpace(parts[3])
	}

	return config.Account{
		Profile: strings.TrimSpace(aliases[0]),
		Name:    strings.TrimSpace(parts[1]),
		Email:   strings.TrimSpace(parts[2]),
		SSHKey:  sshKey,
	}, true
}
