package worktreesetup

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/isac7722/git-extension/internal/tui"
)

// CopyResult holds the result of a file copy operation.
type CopyResult struct {
	Source       string
	Target       string
	Copied       bool
	Skipped      bool
	Error        error
	IsDir        bool
	FilesCopied  int
	FilesSkipped int
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
func CopyFiles(files []CopySpec, srcDir, dstDir string, force bool) []CopyResult {
	var results []CopyResult
	for _, spec := range files {
		r := copyFile(spec, srcDir, dstDir, force)
		results = append(results, r)
		label := copyLabel(r.Source, r.Target)

		switch {
		case r.Error != nil:
			fmt.Fprintf(os.Stderr, "    %s %s\n",
				tui.Red.Render("✗"),
				fmt.Sprintf("Copy %s: %s", label, r.Error))
		case r.IsDir:
			msg := fmt.Sprintf("Copied %s/ (%d files", label, r.FilesCopied)
			if r.FilesSkipped > 0 {
				msg += fmt.Sprintf(", %d skipped", r.FilesSkipped)
			}
			msg += ")"
			if r.FilesCopied == 0 && r.FilesSkipped > 0 {
				fmt.Fprintf(os.Stderr, "    %s %s\n",
					tui.Dim.Render("-"),
					tui.Dim.Render(fmt.Sprintf("%s/ (all %d exist, skip)", label, r.FilesSkipped)))
			} else {
				fmt.Fprintf(os.Stderr, "    %s %s\n", tui.Green.Render("✔"), msg)
			}
		case r.Skipped:
			fmt.Fprintf(os.Stderr, "    %s %s\n",
				tui.Dim.Render("-"),
				tui.Dim.Render(fmt.Sprintf("%s (exists, skip)", label)))
		default:
			fmt.Fprintf(os.Stderr, "    %s Copied %s\n",
				tui.Green.Render("✔"), label)
		}
	}
	return results
}

func copyFile(spec CopySpec, srcDir, dstDir string, force bool) CopyResult {
	srcRel, err := cleanCopyPath(spec.From)
	if err != nil {
		return CopyResult{Source: spec.From, Target: spec.To, Error: err}
	}
	target := spec.To
	if target == "" {
		target = spec.From
	}
	dstRel, err := cleanCopyPath(target)
	if err != nil {
		return CopyResult{Source: srcRel, Target: target, Error: err}
	}
	result := CopyResult{Source: srcRel, Target: dstRel}

	srcPath := filepath.Join(srcDir, srcRel)
	dstPath := filepath.Join(dstDir, dstRel)

	srcInfo, err := os.Stat(srcPath)
	if err != nil {
		result.Error = fmt.Errorf("source not found")
		return result
	}

	if srcInfo.IsDir() {
		return copyDir(result, srcPath, dstPath, force)
	}

	if !force {
		if _, err := os.Stat(dstPath); err == nil {
			result.Skipped = true
			return result
		}
	}

	if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
		result.Error = err
		return result
	}

	if err := writeFile(srcPath, dstPath, srcInfo.Mode()); err != nil {
		result.Error = err
		return result
	}
	result.Copied = true
	return result
}

func writeFile(srcPath, dstPath string, mode os.FileMode) error {
	src, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer func() { _ = src.Close() }()

	dst, err := os.OpenFile(dstPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}

	if _, err := io.Copy(dst, src); err != nil {
		_ = dst.Close()
		return err
	}
	return dst.Close()
}

// copyDir recursively copies srcRoot to dstRoot, honoring the source's
// .gitignore (top-level only) and always excluding .git. Existing files
// at the destination are skipped unless force is true.
func copyDir(result CopyResult, srcRoot, dstRoot string, force bool) CopyResult {
	ignored := loadGitignore(srcRoot)
	result.IsDir = true

	walkErr := filepath.WalkDir(srcRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == srcRoot {
			return nil
		}
		rel, err := filepath.Rel(srcRoot, path)
		if err != nil {
			return err
		}
		relSlash := filepath.ToSlash(rel)
		topLevel := strings.SplitN(relSlash, "/", 2)[0]
		if topLevel == ".git" {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if ignored(rel, d.IsDir()) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		dstPath := filepath.Join(dstRoot, rel)

		if d.IsDir() {
			info, err := d.Info()
			if err != nil {
				return err
			}
			return os.MkdirAll(dstPath, info.Mode())
		}

		if !force {
			if _, err := os.Stat(dstPath); err == nil {
				result.FilesSkipped++
				return nil
			}
		}
		if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
			return err
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		if err := writeFile(path, dstPath, info.Mode()); err != nil {
			return err
		}
		result.FilesCopied++
		return nil
	})

	if walkErr != nil {
		result.Error = walkErr
		return result
	}
	if result.FilesCopied > 0 {
		result.Copied = true
	} else if result.FilesSkipped > 0 {
		result.Skipped = true
	}
	return result
}

func copyLabel(source, target string) string {
	if source == target || target == "" {
		return source
	}
	return source + " -> " + target
}

func cleanCopyPath(path string) (string, error) {
	if strings.TrimSpace(path) == "" {
		return "", fmt.Errorf("path is required")
	}
	cleaned := filepath.Clean(path)
	cleaned = strings.TrimLeft(cleaned, string(filepath.Separator))
	if cleaned == "" {
		return "", fmt.Errorf("path is required")
	}
	if filepath.IsAbs(cleaned) {
		return "", fmt.Errorf("absolute paths are not supported")
	}
	if cleaned == "." || cleaned == ".." || strings.HasPrefix(cleaned, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("path must stay inside the worktree")
	}
	return cleaned, nil
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
