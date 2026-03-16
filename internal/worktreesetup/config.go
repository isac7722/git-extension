package worktreesetup

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const ConfigFile = ".ge-worktree.yaml"

// Config represents the worktree setup configuration.
type Config struct {
	Copy  []string `yaml:"copy,omitempty"`
	Setup []string `yaml:"setup,omitempty"`
}

// Load reads .ge-worktree.yaml from the given directory.
// Returns nil, nil if the file does not exist.
func Load(dir string) (*Config, error) {
	data, err := os.ReadFile(filepath.Join(dir, ConfigFile))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// Save writes the config to .ge-worktree.yaml in the given directory.
func Save(dir string, cfg *Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, ConfigFile), data, 0644)
}
