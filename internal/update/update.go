package update

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	repoOwner    = "isac7722"
	repoName     = "git-extension"
	cacheTTL     = 24 * time.Hour
	cacheFile    = ".update_check"
	latestAPIURL = "https://api.github.com/repos/" + repoOwner + "/" + repoName + "/releases/latest"
)

type githubRelease struct {
	TagName string `json:"tag_name"`
}

// CheckLatest queries GitHub for the latest release version.
func CheckLatest() (string, error) {
	req, err := http.NewRequest("GET", latestAPIURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API returned %d", resp.StatusCode)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", err
	}
	return strings.TrimPrefix(release.TagName, "v"), nil
}

// IsNewer returns true if latest is a newer version than current.
func IsNewer(current, latest string) bool {
	if current == "dev" || current == "" {
		return false
	}
	c := parseVersion(current)
	l := parseVersion(latest)
	for i := 0; i < 3; i++ {
		if l[i] > c[i] {
			return true
		}
		if l[i] < c[i] {
			return false
		}
	}
	return false
}

func parseVersion(v string) [3]int {
	v = strings.TrimPrefix(v, "v")
	var parts [3]int
	if _, err := fmt.Sscanf(v, "%d.%d.%d", &parts[0], &parts[1], &parts[2]); err != nil {
		return parts
	}
	return parts
}

// IsHomebrew checks if ge was installed via Homebrew.
func IsHomebrew() bool {
	binPath, err := exec.LookPath("ge")
	if err != nil {
		return false
	}
	return strings.Contains(binPath, "Cellar") || strings.Contains(binPath, "homebrew")
}

// Upgrade performs the upgrade based on installation method.
func Upgrade(latest string) error {
	if IsHomebrew() {
		return upgradeHomebrew()
	}
	return upgradeBinary(latest)
}

func upgradeHomebrew() error {
	cmd := exec.Command("brew", "upgrade", "ge")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func upgradeBinary(latest string) error {
	binPath, err := exec.LookPath("ge")
	if err != nil {
		return fmt.Errorf("cannot find ge binary: %w", err)
	}

	archiveName := fmt.Sprintf("%s_%s_%s_%s.tar.gz", repoName, latest, runtime.GOOS, runtime.GOARCH)
	downloadURL := fmt.Sprintf("https://github.com/%s/%s/releases/download/v%s/%s",
		repoOwner, repoName, latest, archiveName)

	resp, err := http.Get(downloadURL)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed: HTTP %d", resp.StatusCode)
	}

	tmpDir, err := os.MkdirTemp("", "ge-upgrade-*")
	if err != nil {
		return err
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	if err := extractTarGz(resp.Body, tmpDir); err != nil {
		return fmt.Errorf("extract failed: %w", err)
	}

	newBin := filepath.Join(tmpDir, "ge")
	if _, err := os.Stat(newBin); err != nil {
		return fmt.Errorf("binary not found in archive")
	}

	// Replace existing binary
	if err := os.Rename(newBin, binPath); err != nil {
		// Cross-device rename: copy instead
		return copyFile(newBin, binPath)
	}
	return nil
}

func extractTarGz(r io.Reader, dest string) error {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer func() { _ = gz.Close() }()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if hdr.Typeflag != tar.TypeReg {
			continue
		}
		outPath := filepath.Join(dest, filepath.Base(hdr.Name))
		f, err := os.OpenFile(outPath, os.O_CREATE|os.O_WRONLY, os.FileMode(hdr.Mode))
		if err != nil {
			return err
		}
		if _, err := io.Copy(f, tr); err != nil {
			_ = f.Close()
			return err
		}
		if err := f.Close(); err != nil {
			return err
		}
	}
	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}

	if _, err := io.Copy(out, in); err != nil {
		_ = out.Close()
		return err
	}
	return out.Close()
}

// CachePath returns the path to the update check cache file.
func CachePath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".ge", cacheFile)
}

// CheckCached reads the cached update check result.
// Returns latest version and whether an update is available.
func CheckCached(current string) (latest string, hasUpdate bool) {
	data, err := os.ReadFile(CachePath())
	if err != nil {
		return "", false
	}
	lines := strings.SplitN(string(data), "\n", 2)
	if len(lines) < 2 {
		return "", false
	}
	cachedVersion := strings.TrimSpace(lines[0])
	cachedTime, err := time.Parse(time.RFC3339, strings.TrimSpace(lines[1]))
	if err != nil {
		return "", false
	}
	if time.Since(cachedTime) > cacheTTL {
		return "", false
	}
	return cachedVersion, IsNewer(current, cachedVersion)
}

// CheckAndCache checks for updates and writes the result to cache.
func CheckAndCache(current string) (latest string, hasUpdate bool) {
	latest, err := CheckLatest()
	if err != nil {
		return "", false
	}
	writeCache(latest)
	return latest, IsNewer(current, latest)
}

func writeCache(version string) {
	path := CachePath()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return
	}
	content := fmt.Sprintf("%s\n%s\n", version, time.Now().Format(time.RFC3339))
	_ = os.WriteFile(path, []byte(content), 0644)
}
