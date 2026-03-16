package worktreesetup

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Suggestion represents a detected setup action.
type Suggestion struct {
	Label string // display label (e.g., "Copy .env", "uv sync")
	Type  string // "copy" or "setup"
	Value string // actual value (file path or command)
}

// Detect scans the directory for project markers and returns suggestions.
func Detect(dir string) []Suggestion {
	var suggestions []Suggestion

	// Detect gitignored .env* files
	suggestions = append(suggestions, detectEnvFiles(dir)...)

	// Detect setup commands from marker files
	suggestions = append(suggestions, detectSetupCommands(dir)...)

	return suggestions
}

func detectEnvFiles(dir string) []Suggestion {
	// Use git to find ignored files matching .env*
	cmd := exec.Command("git", "ls-files", "--others", "--ignored", "--exclude-standard")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return nil
	}

	var suggestions []Suggestion
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		name := filepath.Base(line)
		if strings.HasPrefix(name, ".env") {
			suggestions = append(suggestions, Suggestion{
				Label: "Copy " + line,
				Type:  "copy",
				Value: line,
			})
		}
	}
	return suggestions
}

func detectSetupCommands(dir string) []Suggestion {
	var suggestions []Suggestion

	has := func(name string) bool {
		_, err := os.Stat(filepath.Join(dir, name))
		return err == nil
	}

	// Python: uv
	if has("pyproject.toml") && has("uv.lock") {
		suggestions = append(suggestions, Suggestion{
			Label: "uv sync",
			Type:  "setup",
			Value: "uv sync",
		})
	} else if has("pyproject.toml") && has("requirements.txt") {
		suggestions = append(suggestions, Suggestion{
			Label: "pip install -r requirements.txt",
			Type:  "setup",
			Value: "pip install -r requirements.txt",
		})
	}

	// Node.js
	if has("package-lock.json") {
		suggestions = append(suggestions, Suggestion{
			Label: "npm install",
			Type:  "setup",
			Value: "npm install",
		})
	} else if has("yarn.lock") {
		suggestions = append(suggestions, Suggestion{
			Label: "yarn install",
			Type:  "setup",
			Value: "yarn install",
		})
	} else if has("pnpm-lock.yaml") {
		suggestions = append(suggestions, Suggestion{
			Label: "pnpm install",
			Type:  "setup",
			Value: "pnpm install",
		})
	}

	// Go
	if has("go.mod") {
		suggestions = append(suggestions, Suggestion{
			Label: "go mod download",
			Type:  "setup",
			Value: "go mod download",
		})
	}

	return suggestions
}
