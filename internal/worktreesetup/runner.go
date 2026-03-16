package worktreesetup

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/isac7722/git-extension/internal/tui"
)

// CopyResult holds the result of a file copy operation.
type CopyResult struct {
	File    string
	Copied  bool
	Skipped bool
	Error   error
}

// CmdResult holds the result of a command execution.
type CmdResult struct {
	Command string
	Error   error
}

// Run executes the full setup: copy files then run commands.
// Returns true if all steps succeeded, false if any failed.
func Run(cfg *Config, srcDir, dstDir string, force bool) bool {
	allOk := true

	if len(cfg.Copy) > 0 {
		results := CopyFiles(cfg.Copy, srcDir, dstDir, force)
		for _, r := range results {
			if r.Error != nil {
				allOk = false
			}
		}
	}

	if len(cfg.Setup) > 0 {
		results := RunCommands(cfg.Setup, dstDir)
		for _, r := range results {
			if r.Error != nil {
				allOk = false
			}
		}
	}

	return allOk
}

// CopyFiles copies files from srcDir to dstDir.
// If force is false, existing files are skipped.
func CopyFiles(files []string, srcDir, dstDir string, force bool) []CopyResult {
	var results []CopyResult
	for _, f := range files {
		r := copyFile(f, srcDir, dstDir, force)
		results = append(results, r)

		switch {
		case r.Error != nil:
			fmt.Fprintf(os.Stderr, "    %s %s\n",
				tui.Red.Render("✗"),
				fmt.Sprintf("Copy %s: %s", f, r.Error))
		case r.Skipped:
			fmt.Fprintf(os.Stderr, "    %s %s\n",
				tui.Dim.Render("-"),
				tui.Dim.Render(fmt.Sprintf("%s (exists, skip)", f)))
		default:
			fmt.Fprintf(os.Stderr, "    %s Copied %s\n",
				tui.Green.Render("✔"), f)
		}
	}
	return results
}

func copyFile(name, srcDir, dstDir string, force bool) CopyResult {
	srcPath := filepath.Join(srcDir, name)
	dstPath := filepath.Join(dstDir, name)

	// Check source exists
	srcInfo, err := os.Stat(srcPath)
	if err != nil {
		return CopyResult{File: name, Error: fmt.Errorf("source not found")}
	}

	// Skip if destination exists and not force
	if !force {
		if _, err := os.Stat(dstPath); err == nil {
			return CopyResult{File: name, Skipped: true}
		}
	}

	// Ensure destination directory exists
	if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
		return CopyResult{File: name, Error: err}
	}

	src, err := os.Open(srcPath)
	if err != nil {
		return CopyResult{File: name, Error: err}
	}
	defer src.Close()

	dst, err := os.OpenFile(dstPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return CopyResult{File: name, Error: err}
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return CopyResult{File: name, Error: err}
	}

	return CopyResult{File: name, Copied: true}
}

// RunCommands executes commands in the given working directory.
func RunCommands(commands []string, workdir string) []CmdResult {
	var results []CmdResult
	for _, cmdStr := range commands {
		fmt.Fprintf(os.Stderr, "    %s %s ... ",
			tui.Dim.Render("▸"), cmdStr)

		r := runCommand(cmdStr, workdir)
		results = append(results, r)

		if r.Error != nil {
			fmt.Fprintf(os.Stderr, "%s\n", tui.Red.Render("failed"))
			fmt.Fprintf(os.Stderr, "      %s\n", tui.Dim.Render(r.Error.Error()))
		} else {
			fmt.Fprintf(os.Stderr, "%s\n", tui.Green.Render("done"))
		}
	}
	return results
}

func runCommand(cmdStr, workdir string) CmdResult {
	parts := strings.Fields(cmdStr)
	if len(parts) == 0 {
		return CmdResult{Command: cmdStr, Error: fmt.Errorf("empty command")}
	}

	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Dir = workdir
	cmd.Env = os.Environ()

	output, err := cmd.CombinedOutput()
	if err != nil {
		errMsg := strings.TrimSpace(string(output))
		if errMsg != "" {
			return CmdResult{Command: cmdStr, Error: fmt.Errorf("%w: %s", err, errMsg)}
		}
		return CmdResult{Command: cmdStr, Error: err}
	}

	return CmdResult{Command: cmdStr}
}
