package config

import (
	"os"
	"path/filepath"
	"strings"
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

func TestGet_NotFound(t *testing.T) {
	cfg := &Config{byName: make(map[string]*Account)}
	_, ok := cfg.Get("nonexistent")
	if ok {
		t.Error("expected false for nonexistent profile")
	}
}

func TestRemoveAccount(t *testing.T) {
	cfg := &Config{byName: make(map[string]*Account)}
	cfg.AddAccount(Account{Profile: "work", Name: "John", Email: "john@co.com"})
	cfg.AddAccount(Account{Profile: "personal", Name: "Jane", Email: "jane@co.com"})

	removed := cfg.RemoveAccount("work")
	if !removed {
		t.Error("expected true for removing existing account")
	}
	if len(cfg.Accounts) != 1 {
		t.Fatalf("expected 1 account, got %d", len(cfg.Accounts))
	}
	if _, ok := cfg.Get("work"); ok {
		t.Error("work should not be found after removal")
	}
	if _, ok := cfg.Get("personal"); !ok {
		t.Error("personal should still exist")
	}
}

func TestRemoveAccount_NotFound(t *testing.T) {
	cfg := &Config{byName: make(map[string]*Account)}
	cfg.AddAccount(Account{Profile: "work", Name: "John", Email: "john@co.com"})

	removed := cfg.RemoveAccount("nonexistent")
	if removed {
		t.Error("expected false for removing nonexistent account")
	}
	if len(cfg.Accounts) != 1 {
		t.Error("accounts should not change")
	}
}

func TestUpdateAccount(t *testing.T) {
	cfg := &Config{byName: make(map[string]*Account)}
	cfg.AddAccount(Account{Profile: "work", Name: "John", Email: "john@co.com"})

	updated := cfg.UpdateAccount("work", Account{
		Profile: "work",
		Name:    "John Updated",
		Email:   "john.updated@co.com",
		SSHKey:  "/new/key",
	})
	if !updated {
		t.Error("expected true for updating existing account")
	}

	acct, ok := cfg.Get("work")
	if !ok {
		t.Fatal("work should still be found")
	}
	if acct.Name != "John Updated" {
		t.Errorf("expected 'John Updated', got %q", acct.Name)
	}
	if acct.Email != "john.updated@co.com" {
		t.Errorf("expected 'john.updated@co.com', got %q", acct.Email)
	}
}

func TestUpdateAccount_Rename(t *testing.T) {
	cfg := &Config{byName: make(map[string]*Account)}
	cfg.AddAccount(Account{Profile: "old", Name: "John", Email: "john@co.com"})

	updated := cfg.UpdateAccount("old", Account{
		Profile: "new",
		Name:    "John",
		Email:   "john@co.com",
	})
	if !updated {
		t.Error("expected true for renaming account")
	}

	if _, ok := cfg.Get("old"); ok {
		t.Error("old profile should not be found after rename")
	}
	if _, ok := cfg.Get("new"); !ok {
		t.Error("new profile should be found after rename")
	}
}

func TestUpdateAccount_NotFound(t *testing.T) {
	cfg := &Config{byName: make(map[string]*Account)}
	updated := cfg.UpdateAccount("nonexistent", Account{Profile: "new"})
	if updated {
		t.Error("expected false for updating nonexistent account")
	}
}

func TestAddAccount_ExpandsHome(t *testing.T) {
	cfg := &Config{byName: make(map[string]*Account)}
	home, _ := os.UserHomeDir()

	cfg.AddAccount(Account{
		Profile: "test",
		Name:    "Test",
		Email:   "test@co.com",
		SSHKey:  "~/.ssh/test_key",
	})

	acct, _ := cfg.Get("test")
	expected := filepath.Join(home, ".ssh/test_key")
	if acct.SSHKey != expected {
		t.Errorf("expected %q, got %q", expected, acct.SSHKey)
	}
}

func TestLoad_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "credentials")
	os.WriteFile(path, []byte(""), 0644)

	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Accounts) != 0 {
		t.Errorf("expected 0 accounts, got %d", len(cfg.Accounts))
	}
}

func TestLoad_WithComments(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "credentials")
	content := `# Main config
[work]
name = John
email = john@co.com
# ssh key below
ssh_key = /path/to/key
`
	os.WriteFile(path, []byte(content), 0644)

	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Accounts) != 1 {
		t.Fatalf("expected 1 account, got %d", len(cfg.Accounts))
	}
	acct, _ := cfg.Get("work")
	if acct.SSHKey != "/path/to/key" {
		t.Errorf("expected '/path/to/key', got %q", acct.SSHKey)
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := Load("/nonexistent/path/credentials")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestSave_WithoutSSHKey(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "credentials")

	cfg := &Config{byName: make(map[string]*Account)}
	cfg.AddAccount(Account{
		Profile: "minimal",
		Name:    "Min User",
		Email:   "min@co.com",
	})

	if err := cfg.Save(path); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(path)
	content := string(data)
	if strings.Contains(content, "ssh_key") {
		t.Error("saved file should not contain ssh_key when empty")
	}
}

func TestSave_MultipleAccounts(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "credentials")

	cfg := &Config{byName: make(map[string]*Account)}
	cfg.AddAccount(Account{Profile: "a", Name: "A", Email: "a@co.com"})
	cfg.AddAccount(Account{Profile: "b", Name: "B", Email: "b@co.com", SSHKey: "/key"})

	if err := cfg.Save(path); err != nil {
		t.Fatal(err)
	}

	cfg2, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg2.Accounts) != 2 {
		t.Fatalf("expected 2 accounts, got %d", len(cfg2.Accounts))
	}
	acctB, _ := cfg2.Get("b")
	if acctB.SSHKey != "/key" {
		t.Errorf("expected '/key', got %q", acctB.SSHKey)
	}
}

func TestSave_CreatesDirectory(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "dir", "credentials")

	cfg := &Config{byName: make(map[string]*Account)}
	cfg.AddAccount(Account{Profile: "test", Name: "Test", Email: "t@co.com"})

	if err := cfg.Save(path); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("file should exist after save")
	}
}

func TestParseConfigLine(t *testing.T) {
	acct := &Account{}

	parseConfigLine(acct, "name = John Doe")
	if acct.Name != "John Doe" {
		t.Errorf("expected 'John Doe', got %q", acct.Name)
	}

	parseConfigLine(acct, "email = john@co.com")
	if acct.Email != "john@co.com" {
		t.Errorf("expected 'john@co.com', got %q", acct.Email)
	}

	parseConfigLine(acct, "ssh_key = /path/to/key")
	if acct.SSHKey != "/path/to/key" {
		t.Errorf("expected '/path/to/key', got %q", acct.SSHKey)
	}

	// Invalid line (no =) should be ignored
	parseConfigLine(acct, "invalid line without equals")
	// No crash = pass

	// Unknown key should be ignored
	parseConfigLine(acct, "unknown_key = value")
	// No crash = pass
}

func TestDefaultPath(t *testing.T) {
	path := DefaultPath()
	if !strings.Contains(path, ".ge") {
		t.Errorf("expected path to contain '.ge', got %q", path)
	}
	if !strings.HasSuffix(path, "credentials") {
		t.Errorf("expected path to end with 'credentials', got %q", path)
	}
}
