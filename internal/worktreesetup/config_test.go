package worktreesetup

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadCopyLegacyString(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, ConfigFile), []byte("copy:\n  - .env\nsetup:\n  - go mod download\n"))

	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg == nil {
		t.Fatal("Load() returned nil config")
	}
	if len(cfg.Copy) != 1 {
		t.Fatalf("len(Copy) = %d, want 1", len(cfg.Copy))
	}
	if cfg.Copy[0].From != ".env" || cfg.Copy[0].To != ".env" {
		t.Fatalf("Copy[0] = %+v, want From and To .env", cfg.Copy[0])
	}
}

func TestLoadCopyObject(t *testing.T) {
	dir := t.TempDir()
	writeTestFile(t, filepath.Join(dir, ConfigFile), []byte("copy:\n  - from: .env\n    to: /server-main/.env\n"))

	cfg, err := Load(dir)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(cfg.Copy) != 1 {
		t.Fatalf("len(Copy) = %d, want 1", len(cfg.Copy))
	}
	if cfg.Copy[0].From != ".env" || cfg.Copy[0].To != "/server-main/.env" {
		t.Fatalf("Copy[0] = %+v, want mapped source and target", cfg.Copy[0])
	}
}

func TestSaveCopyUsesObjectForm(t *testing.T) {
	dir := t.TempDir()
	cfg := &Config{
		Copy: []CopySpec{{From: ".env", To: ".env"}},
	}

	if err := Save(dir, cfg); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, ConfigFile))
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	got := string(data)
	if !strings.Contains(got, "from: .env") || !strings.Contains(got, "to: .env") {
		t.Fatalf("saved config = %q, want from/to object form", got)
	}
}

func writeTestFile(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
}
