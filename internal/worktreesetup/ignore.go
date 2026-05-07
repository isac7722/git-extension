package worktreesetup

import (
	"os"
	"path/filepath"
	"strings"
)

type gitignorePattern struct {
	pattern  string
	negate   bool
	dirOnly  bool
	anchored bool
}

// loadGitignore reads .gitignore at root and returns a matcher function
// reporting whether a relative path (slash-separated, relative to root) is ignored.
// If .gitignore does not exist, the matcher always returns false.
func loadGitignore(root string) func(relPath string, isDir bool) bool {
	data, err := os.ReadFile(filepath.Join(root, ".gitignore"))
	if err != nil {
		return func(string, bool) bool { return false }
	}

	var patterns []gitignorePattern
	for _, raw := range strings.Split(string(data), "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		p := gitignorePattern{}
		if strings.HasPrefix(line, "!") {
			p.negate = true
			line = line[1:]
		}
		if strings.HasSuffix(line, "/") {
			p.dirOnly = true
			line = strings.TrimSuffix(line, "/")
		}
		if strings.HasPrefix(line, "/") {
			p.anchored = true
			line = strings.TrimPrefix(line, "/")
		} else if strings.Contains(line, "/") {
			p.anchored = true
		}
		p.pattern = line
		patterns = append(patterns, p)
	}

	return func(relPath string, isDir bool) bool {
		relPath = filepath.ToSlash(relPath)
		ignored := false
		for _, p := range patterns {
			if p.dirOnly && !isDir {
				continue
			}
			if matchIgnorePattern(p, relPath) {
				ignored = !p.negate
			}
		}
		return ignored
	}
}

func matchIgnorePattern(p gitignorePattern, relPath string) bool {
	if p.anchored {
		if ok, _ := filepath.Match(p.pattern, relPath); ok {
			return true
		}
		// Also match any prefix directory of relPath against the anchored pattern
		// so files inside an ignored directory are also ignored.
		for dir := relPath; dir != "." && dir != "/"; dir = filepath.ToSlash(filepath.Dir(dir)) {
			if ok, _ := filepath.Match(p.pattern, dir); ok {
				return true
			}
		}
		return false
	}
	// Unanchored: match against basename of any path component.
	for _, part := range strings.Split(relPath, "/") {
		if ok, _ := filepath.Match(p.pattern, part); ok {
			return true
		}
	}
	return false
}
