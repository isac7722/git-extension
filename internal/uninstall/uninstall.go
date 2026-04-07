package uninstall

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/isac7722/git-extension/internal/git"
	"github.com/isac7722/git-extension/internal/shell"
)

// Target represents a single cleanup item.
type Target struct {
	Category    string // "shell", "config", "git-config", "legacy"
	Description string
	Path        string
	Exists      bool
}

// Plan holds all discovered cleanup targets.
type Plan struct {
	Targets    []Target
	IsHomebrew bool
	BinaryPath string
}

// Discover scans the system for ge-related artifacts.
func Discover() *Plan {
	plan := &Plan{}
	home, _ := os.UserHomeDir()

	// 1. Binary
	if binPath, err := exec.LookPath("ge"); err == nil {
		plan.BinaryPath = binPath
		plan.IsHomebrew = strings.Contains(binPath, "Cellar") || strings.Contains(binPath, "homebrew")
	}

	// 2. Shell RC files
	rcFiles := []string{
		filepath.Join(home, ".zshrc"),
		filepath.Join(home, ".bashrc"),
		filepath.Join(home, ".bash_profile"),
	}
	for _, rc := range rcFiles {
		if shell.HasAnyMarker(rc) {
			plan.Targets = append(plan.Targets, Target{
				Category:    "shell",
				Description: fmt.Sprintf("%s — shell integration block", rc),
				Path:        rc,
				Exists:      true,
			})
		}
	}

	// 3. Config directory
	geDir := filepath.Join(home, ".ge")
	if info, err := os.Stat(geDir); err == nil && info.IsDir() {
		plan.Targets = append(plan.Targets, Target{
			Category:    "config",
			Description: "~/.ge/ (credentials, agent config)",
			Path:        geDir,
			Exists:      true,
		})
	}

	// 4. Git config keys (global)
	globalKeys := []string{
		"ge.protected-branches",
		"ge.protected-branches.exclude-defaults",
		"ge.worktree.dir",
		"ge.clean.protected",
	}
	for _, key := range globalKeys {
		if val, err := git.Run("config", "--global", "--get", key); err == nil && val != "" {
			plan.Targets = append(plan.Targets, Target{
				Category:    "git-config",
				Description: fmt.Sprintf("%s (global)", key),
				Path:        key,
				Exists:      true,
			})
		}
	}

	// 4b. Git config keys (local, only if in a repo)
	if git.IsInsideWorkTree() {
		for _, key := range globalKeys {
			if val, err := git.Run("config", "--local", "--get", key); err == nil && val != "" {
				plan.Targets = append(plan.Targets, Target{
					Category:    "git-config",
					Description: fmt.Sprintf("%s (local)", key),
					Path:        key + ":local",
					Exists:      true,
				})
			}
		}
	}

	// 5. Legacy paths
	legacyDir := filepath.Join(home, ".config", "gituser")
	if _, err := os.Stat(legacyDir); err == nil {
		plan.Targets = append(plan.Targets, Target{
			Category:    "legacy",
			Description: "~/.config/gituser/",
			Path:        legacyDir,
			Exists:      true,
		})
	}
	if matches, _ := filepath.Glob(filepath.Join(home, ".config", "gituser.bak.*")); len(matches) > 0 {
		for _, m := range matches {
			plan.Targets = append(plan.Targets, Target{
				Category:    "legacy",
				Description: fmt.Sprintf("~/.config/%s", filepath.Base(m)),
				Path:        m,
				Exists:      true,
			})
		}
	}

	return plan
}

// FormatPlan returns a human-readable preview of the plan.
func FormatPlan(plan *Plan) string {
	if len(plan.Targets) == 0 && plan.BinaryPath == "" {
		return "Nothing to uninstall."
	}

	var sb strings.Builder
	sb.WriteString("The following will be removed:\n")

	categories := []struct {
		key   string
		label string
	}{
		{"shell", "Shell integration"},
		{"config", "Config files"},
		{"git-config", "Git config"},
		{"legacy", "Legacy files"},
	}

	for _, cat := range categories {
		var items []Target
		for _, t := range plan.Targets {
			if t.Category == cat.key {
				items = append(items, t)
			}
		}
		if len(items) == 0 {
			continue
		}
		sb.WriteString(fmt.Sprintf("\n  %s:\n", cat.label))
		for _, t := range items {
			sb.WriteString(fmt.Sprintf("    • %s\n", t.Description))
		}
	}

	if plan.BinaryPath != "" {
		sb.WriteString(fmt.Sprintf("\n  Binary:\n    • %s (manual removal required)\n", plan.BinaryPath))
	}

	return sb.String()
}

// Execute removes all targets in the plan. Returns result messages.
func Execute(plan *Plan) []string {
	var results []string

	for _, t := range plan.Targets {
		if !t.Exists {
			continue
		}
		switch t.Category {
		case "shell":
			if err := shell.RemoveAllMarkers(t.Path); err != nil {
				results = append(results, fmt.Sprintf("✗ Failed to remove shell block from %s: %v", t.Path, err))
			} else {
				results = append(results, fmt.Sprintf("✔ Removed shell block from %s", t.Path))
			}
		case "config":
			if err := os.RemoveAll(t.Path); err != nil {
				results = append(results, fmt.Sprintf("✗ Failed to remove %s: %v", t.Description, err))
			} else {
				results = append(results, fmt.Sprintf("✔ Removed %s", t.Description))
			}
		case "git-config":
			key := t.Path
			scope := "--global"
			if strings.HasSuffix(key, ":local") {
				key = strings.TrimSuffix(key, ":local")
				scope = "--local"
			}
			if _, err := git.Run("config", scope, "--unset", key); err != nil {
				results = append(results, fmt.Sprintf("✗ Failed to remove git config %s: %v", t.Description, err))
			} else {
				results = append(results, fmt.Sprintf("✔ Removed git config: %s", t.Description))
			}
		case "legacy":
			if err := os.RemoveAll(t.Path); err != nil {
				results = append(results, fmt.Sprintf("✗ Failed to remove %s: %v", t.Description, err))
			} else {
				results = append(results, fmt.Sprintf("✔ Removed %s", t.Description))
			}
		}
	}

	return results
}
