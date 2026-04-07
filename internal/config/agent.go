package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// AgentConfig holds agent-specific configuration.
type AgentConfig struct {
	AnthropicAPIKey string
	Model           string
}

// AgentConfigPath returns the default path for agent credentials.
func AgentConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".ge", "agent_credentials")
}

// LoadAgentConfig reads agent configuration from the default path.
func LoadAgentConfig() (*AgentConfig, error) {
	return LoadAgentConfigFrom(AgentConfigPath())
}

// LoadAgentConfigFrom reads agent configuration from the given path.
func LoadAgentConfigFrom(path string) (*AgentConfig, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	cfg := &AgentConfig{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "[") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		switch key {
		case "anthropic_api_key":
			cfg.AnthropicAPIKey = value
		case "model":
			cfg.Model = value
		}
	}
	return cfg, scanner.Err()
}

// SaveAgentConfig writes agent configuration to the default path with 0600 permissions.
func SaveAgentConfig(cfg *AgentConfig) error {
	return SaveAgentConfigTo(AgentConfigPath(), cfg)
}

// SaveAgentConfigTo writes agent configuration to the given path.
func SaveAgentConfigTo(path string, cfg *AgentConfig) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	content := fmt.Sprintf("[default]\nanthropic_api_key = %s\n", cfg.AnthropicAPIKey)
	if cfg.Model != "" {
		content += fmt.Sprintf("model = %s\n", cfg.Model)
	}
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	// Ensure permissions are correct even if file existed
	return os.Chmod(path, 0600)
}

// RemoveAgentConfig deletes the agent credentials file.
func RemoveAgentConfig() error {
	path := AgentConfigPath()
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
