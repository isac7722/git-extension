package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	anthropic "github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/toolrunner"
)

type GitCommandInput struct {
	Command string `json:"command" jsonschema:"required,description=Git subcommand and arguments (e.g. 'status --short' or 'log --oneline -10'). Do not include 'git' prefix."`
}

type ReadFileInput struct {
	Path string `json:"path" jsonschema:"required,description=File path to read (relative to current directory)"`
}

type GhCommandInput struct {
	Command string `json:"command" jsonschema:"required,description=GitHub CLI subcommand and arguments (e.g. 'pr create --title ...' or 'pr list'). Do not include 'gh' prefix."`
}

type ShellCommandInput struct {
	Command string `json:"command" jsonschema:"required,description=Shell command to execute. Use for non-git operations like ls or find."`
}

func textResult(text string) anthropic.BetaToolResultBlockParamContentUnion {
	return anthropic.BetaToolResultBlockParamContentUnion{
		OfText: &anthropic.BetaTextBlockParam{Text: text},
	}
}

func buildTools() ([]anthropic.BetaTool, error) {
	var tools []anthropic.BetaTool

	// 1. git command tool
	gitTool, err := toolrunner.NewBetaToolFromJSONSchema(
		"git",
		"Execute a git command. The 'git' prefix is added automatically.",
		func(ctx context.Context, input GitCommandInput) (anthropic.BetaToolResultBlockParamContentUnion, error) {
			if strings.TrimSpace(input.Command) == "" {
				return textResult("Error: empty command"), nil
			}

			// Block dangerous commands without explicit subcommands
			dangerous := []string{"push --force", "push -f", "reset --hard", "clean -f"}
			for _, d := range dangerous {
				if strings.Contains(input.Command, d) {
					return textResult(fmt.Sprintf("⚠️ Dangerous command detected: git %s\nThis command was executed but please confirm with the user first in future.", input.Command)), nil
				}
			}

			cmd := exec.CommandContext(ctx, "sh", "-c", "git "+input.Command)
			output, err := cmd.CombinedOutput()
			result := strings.TrimSpace(string(output))
			if err != nil {
				return textResult(fmt.Sprintf("Error (exit %v):\n%s", err, result)), nil
			}
			if result == "" {
				result = "(no output)"
			}
			return textResult(result), nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("git tool: %w", err)
	}
	tools = append(tools, gitTool)

	// 2. read file tool
	readFileTool, err := toolrunner.NewBetaToolFromJSONSchema(
		"read_file",
		"Read the contents of a file. Returns the file content as text.",
		func(ctx context.Context, input ReadFileInput) (anthropic.BetaToolResultBlockParamContentUnion, error) {
			data, err := os.ReadFile(input.Path)
			if err != nil {
				return textResult(fmt.Sprintf("Error reading file: %v", err)), nil
			}
			content := string(data)
			// Truncate large files
			const maxSize = 30000
			if len(content) > maxSize {
				content = content[:maxSize] + fmt.Sprintf("\n\n... (truncated, total %d bytes)", len(data))
			}
			return textResult(content), nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("read_file tool: %w", err)
	}
	tools = append(tools, readFileTool)

	// 3. gh (GitHub CLI) tool
	ghTool, err := toolrunner.NewBetaToolFromJSONSchema(
		"gh",
		"Execute a GitHub CLI (gh) command. The 'gh' prefix is added automatically. Use for PR creation, issue management, etc.",
		func(ctx context.Context, input GhCommandInput) (anthropic.BetaToolResultBlockParamContentUnion, error) {
			if strings.TrimSpace(input.Command) == "" {
				return textResult("Error: empty command"), nil
			}
			cmd := exec.CommandContext(ctx, "sh", "-c", "gh "+input.Command)
			output, err := cmd.CombinedOutput()
			result := strings.TrimSpace(string(output))
			if err != nil {
				return textResult(fmt.Sprintf("Error (exit %v):\n%s", err, result)), nil
			}
			if result == "" {
				result = "(no output)"
			}
			return textResult(result), nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("gh tool: %w", err)
	}
	tools = append(tools, ghTool)

	// 4. shell command tool (limited)
	shellTool, err := toolrunner.NewBetaToolFromJSONSchema(
		"shell",
		"Execute a shell command. Use for non-git operations like listing files, searching, etc.",
		func(ctx context.Context, input ShellCommandInput) (anthropic.BetaToolResultBlockParamContentUnion, error) {
			cmd := exec.CommandContext(ctx, "sh", "-c", input.Command)
			output, err := cmd.CombinedOutput()
			result := strings.TrimSpace(string(output))
			if err != nil {
				return textResult(fmt.Sprintf("Error (exit %v):\n%s", err, result)), nil
			}
			if result == "" {
				result = "(no output)"
			}
			// Truncate large output
			const maxSize = 30000
			if len(result) > maxSize {
				result = result[:maxSize] + "\n\n... (truncated)"
			}
			return textResult(result), nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("shell tool: %w", err)
	}
	tools = append(tools, shellTool)

	return tools, nil
}

// Suppress unused import warning
var _ = json.Marshal
