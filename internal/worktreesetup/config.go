package worktreesetup

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const ConfigFile = ".ge-worktree.yaml"

// Config represents the worktree setup configuration.
type Config struct {
	Copy  []CopySpec `yaml:"copy,omitempty"`
	Setup []string   `yaml:"setup,omitempty"`
}

// CopySpec describes a copy from the main worktree to a target worktree.
type CopySpec struct {
	From string `yaml:"from"`
	To   string `yaml:"to"`
}

// UnmarshalYAML supports both the current object form and the legacy string
// shorthand where the source and destination paths are identical.
func (c *CopySpec) UnmarshalYAML(value *yaml.Node) error {
	switch value.Kind {
	case yaml.ScalarNode:
		var path string
		if err := value.Decode(&path); err != nil {
			return err
		}
		c.From = path
		c.To = path
		return nil
	case yaml.MappingNode:
		type copySpec CopySpec
		var spec copySpec
		if err := value.Decode(&spec); err != nil {
			return err
		}
		if spec.From == "" {
			return fmt.Errorf("copy.from is required")
		}
		if spec.To == "" {
			spec.To = spec.From
		}
		*c = CopySpec(spec)
		return nil
	default:
		return fmt.Errorf("copy entries must be a path string or from/to mapping")
	}
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
