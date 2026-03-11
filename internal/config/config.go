package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Account represents a git user account profile.
type Account struct {
	Profile string
	Name    string
	Email   string
	SSHKey  string
}

// Config holds all loaded accounts.
type Config struct {
	Accounts []Account
	byName   map[string]*Account
}

// DefaultPath returns the default credentials file path (~/.ge/credentials).
func DefaultPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".ge", "credentials")
}

// Load reads and parses the credentials file.
func Load(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	cfg := &Config{byName: make(map[string]*Account)}
	scanner := bufio.NewScanner(f)

	var current *Account
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Section header
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			profile := strings.TrimSpace(line[1 : len(line)-1])
			current = &Account{Profile: profile}
			cfg.Accounts = append(cfg.Accounts, *current)
			cfg.byName[profile] = &cfg.Accounts[len(cfg.Accounts)-1]
			current = cfg.byName[profile]
			continue
		}

		// Key = Value
		if current == nil {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		val = expandHome(val)

		switch key {
		case "name":
			current.Name = val
		case "email":
			current.Email = val
		case "ssh_key":
			current.SSHKey = val
		}
	}

	return cfg, scanner.Err()
}

// Get returns an account by profile name.
func (c *Config) Get(profile string) (*Account, bool) {
	a, ok := c.byName[profile]
	return a, ok
}

// Save writes all accounts back to the credentials file.
func (c *Config) Save(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	var sb strings.Builder
	for i, a := range c.Accounts {
		if i > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(fmt.Sprintf("[%s]\n", a.Profile))
		sb.WriteString(fmt.Sprintf("name = %s\n", a.Name))
		sb.WriteString(fmt.Sprintf("email = %s\n", a.Email))
		if a.SSHKey != "" {
			sb.WriteString(fmt.Sprintf("ssh_key = %s\n", a.SSHKey))
		}
	}

	return os.WriteFile(path, []byte(sb.String()), 0644)
}

// RemoveAccount removes an account by profile name and returns true if found.
func (c *Config) RemoveAccount(profile string) bool {
	for i, a := range c.Accounts {
		if a.Profile == profile {
			c.Accounts = append(c.Accounts[:i], c.Accounts[i+1:]...)
			delete(c.byName, profile)
			return true
		}
	}
	return false
}

// AddAccount adds a new account and updates the internal index.
func (c *Config) AddAccount(a Account) {
	a.SSHKey = expandHome(a.SSHKey)
	c.Accounts = append(c.Accounts, a)
	c.byName[a.Profile] = &c.Accounts[len(c.Accounts)-1]
}

func expandHome(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[2:])
	}
	return path
}
