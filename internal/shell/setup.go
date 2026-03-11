package shell

import (
	"fmt"
	"os"
	"strings"
)

const markerStart = "# >>> ge-cli >>>"
const markerEnd = "# <<< ge-cli <<<"

// Setup injects the eval line into the user's RC file.
func Setup(dryRun bool) (string, error) {
	shell := DetectShell()
	rcPath := DetectRCFile(shell)

	if HasMarker(rcPath) {
		return "", fmt.Errorf("ge-cli is already configured in %s", rcPath)
	}

	block := fmt.Sprintf(`
%s
eval "$(command ge init %s)"
%s
`, markerStart, shell, markerEnd)

	if dryRun {
		return fmt.Sprintf("Would add to %s:\n%s", rcPath, block), nil
	}

	f, err := os.OpenFile(rcPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return "", fmt.Errorf("failed to open %s: %w", rcPath, err)
	}
	defer f.Close()

	if _, err := f.WriteString(block); err != nil {
		return "", fmt.Errorf("failed to write to %s: %w", rcPath, err)
	}

	return fmt.Sprintf("Added ge-cli initialization to %s\nRestart your shell or run: source %s", rcPath, rcPath), nil
}

// RemoveMarker removes the ge-cli block from an RC file.
func RemoveMarker(rcPath string) error {
	data, err := os.ReadFile(rcPath)
	if err != nil {
		return err
	}

	content := string(data)
	startIdx := strings.Index(content, markerStart)
	endIdx := strings.Index(content, markerEnd)
	if startIdx == -1 || endIdx == -1 {
		return nil
	}

	// Remove from start of marker to end of marker line (including newline)
	endIdx += len(markerEnd)
	if endIdx < len(content) && content[endIdx] == '\n' {
		endIdx++
	}

	newContent := content[:startIdx] + content[endIdx:]
	return os.WriteFile(rcPath, []byte(newContent), 0644)
}
