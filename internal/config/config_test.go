package config

import (
	"os"
	"path/filepath"
	"testing"
)

const testCredentials = `[work]
name = John Work
email = john@company.com
ssh_key = ~/.ssh/work_ed25519

# personal account
[personal]
name = John Personal
email = john@gmail.com
ssh_key = ~/.ssh/personal_ed25519
`

func TestLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "credentials")
	if err := os.WriteFile(path, []byte(testCredentials), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}

	if len(cfg.Accounts) != 2 {
		t.Fatalf("expected 2 accounts, got %d", len(cfg.Accounts))
	}

	work, ok := cfg.Get("work")
	if !ok {
		t.Fatal("work profile not found")
	}
	if work.Name != "John Work" {
		t.Errorf("expected 'John Work', got %q", work.Name)
	}
	if work.Email != "john@company.com" {
		t.Errorf("expected 'john@company.com', got %q", work.Email)
	}

	personal, ok := cfg.Get("personal")
	if !ok {
		t.Fatal("personal profile not found")
	}
	if personal.Name != "John Personal" {
		t.Errorf("expected 'John Personal', got %q", personal.Name)
	}
}

func TestSaveAndReload(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "credentials")

	cfg := &Config{byName: make(map[string]*Account)}
	cfg.AddAccount(Account{
		Profile: "test",
		Name:    "Test User",
		Email:   "test@example.com",
		SSHKey:  "/home/test/.ssh/id_ed25519",
	})

	if err := cfg.Save(path); err != nil {
		t.Fatal(err)
	}

	cfg2, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}

	if len(cfg2.Accounts) != 1 {
		t.Fatalf("expected 1 account, got %d", len(cfg2.Accounts))
	}

	a, ok := cfg2.Get("test")
	if !ok {
		t.Fatal("test profile not found after reload")
	}
	if a.Name != "Test User" || a.Email != "test@example.com" {
		t.Errorf("unexpected values: %+v", a)
	}
}

func TestExpandHome(t *testing.T) {
	home, _ := os.UserHomeDir()
	result := expandHome("~/test/path")
	expected := filepath.Join(home, "test/path")
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}

	result = expandHome("/absolute/path")
	if result != "/absolute/path" {
		t.Errorf("expected '/absolute/path', got %q", result)
	}
}
