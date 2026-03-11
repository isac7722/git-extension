package migration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/isac7722/git-extension/internal/config"
)

func TestParseLegacyAccount(t *testing.T) {
	tests := []struct {
		name    string
		line    string
		wantOK  bool
		wantAcc config.Account
	}{
		{
			name:   "full format with ssh key",
			line:   "work:John Doe:john@company.com:~/.ssh/work_key",
			wantOK: true,
			wantAcc: config.Account{
				Profile: "work",
				Name:    "John Doe",
				Email:   "john@company.com",
				SSHKey:  "~/.ssh/work_key",
			},
		},
		{
			name:   "without ssh key",
			line:   "personal:Jane Doe:jane@gmail.com",
			wantOK: true,
			wantAcc: config.Account{
				Profile: "personal",
				Name:    "Jane Doe",
				Email:   "jane@gmail.com",
				SSHKey:  "",
			},
		},
		{
			name:   "multiple aliases uses first",
			line:   "work,office:John:john@co.com:~/.ssh/key",
			wantOK: true,
			wantAcc: config.Account{
				Profile: "work",
				Name:    "John",
				Email:   "john@co.com",
				SSHKey:  "~/.ssh/key",
			},
		},
		{
			name:   "too few fields",
			line:   "onlyname:email",
			wantOK: false,
		},
		{
			name:   "empty line",
			line:   "",
			wantOK: false,
		},
		{
			name:   "with spaces",
			line:   " work : John Doe : john@co.com : ~/.ssh/key ",
			wantOK: true,
			wantAcc: config.Account{
				Profile: "work",
				Name:    "John Doe",
				Email:   "john@co.com",
				SSHKey:  "~/.ssh/key",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			acct, ok := parseLegacyAccount(tt.line)
			if ok != tt.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tt.wantOK)
			}
			if !ok {
				return
			}
			if acct.Profile != tt.wantAcc.Profile {
				t.Errorf("Profile = %q, want %q", acct.Profile, tt.wantAcc.Profile)
			}
			if acct.Name != tt.wantAcc.Name {
				t.Errorf("Name = %q, want %q", acct.Name, tt.wantAcc.Name)
			}
			if acct.Email != tt.wantAcc.Email {
				t.Errorf("Email = %q, want %q", acct.Email, tt.wantAcc.Email)
			}
			if acct.SSHKey != tt.wantAcc.SSHKey {
				t.Errorf("SSHKey = %q, want %q", acct.SSHKey, tt.wantAcc.SSHKey)
			}
		})
	}
}

func TestLoadLegacyAccounts(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "accounts")

	content := `# Legacy accounts file
work:John Doe:john@company.com:~/.ssh/work_key
personal:Jane Doe:jane@gmail.com

# This is a comment
team,office:Team Lead:lead@company.com:~/.ssh/team_key
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := loadLegacyAccounts(path)
	if err != nil {
		t.Fatal(err)
	}

	if len(cfg.Accounts) != 3 {
		t.Fatalf("expected 3 accounts, got %d", len(cfg.Accounts))
	}

	// Verify first account
	if cfg.Accounts[0].Profile != "work" {
		t.Errorf("expected 'work', got %q", cfg.Accounts[0].Profile)
	}
	if cfg.Accounts[0].Name != "John Doe" {
		t.Errorf("expected 'John Doe', got %q", cfg.Accounts[0].Name)
	}

	// Verify second account (no ssh key)
	if cfg.Accounts[1].SSHKey != "" {
		t.Errorf("expected empty ssh_key, got %q", cfg.Accounts[1].SSHKey)
	}
}

func TestLoadLegacyAccounts_FileNotFound(t *testing.T) {
	_, err := loadLegacyAccounts("/nonexistent/path/accounts")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
	if !strings.Contains(err.Error(), "no legacy accounts found") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestLoadLegacyAccounts_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "accounts")
	os.WriteFile(path, []byte(""), 0644)

	cfg, err := loadLegacyAccounts(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Accounts) != 0 {
		t.Errorf("expected 0 accounts, got %d", len(cfg.Accounts))
	}
}

func TestLoadLegacyAccounts_CommentsOnly(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "accounts")
	os.WriteFile(path, []byte("# comment 1\n# comment 2\n"), 0644)

	cfg, err := loadLegacyAccounts(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Accounts) != 0 {
		t.Errorf("expected 0 accounts, got %d", len(cfg.Accounts))
	}
}

func TestLoadLegacyAccounts_SkipsInvalidLines(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "accounts")
	content := "work:John:john@co.com\ninvalid\npersonal:Jane:jane@co.com"
	os.WriteFile(path, []byte(content), 0644)

	cfg, err := loadLegacyAccounts(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.Accounts) != 2 {
		t.Fatalf("expected 2 accounts (skipping invalid), got %d", len(cfg.Accounts))
	}
}

func TestHasLegacy(t *testing.T) {
	// Since HasLegacy checks ~/.config/gituser/accounts, we just verify it returns false
	// in test environment (the file shouldn't exist)
	// This is mainly a smoke test
	_ = HasLegacy() // just verify it doesn't panic
}

func TestLegacyPath(t *testing.T) {
	path := LegacyPath()
	if !strings.Contains(path, "gituser") {
		t.Errorf("expected path to contain 'gituser', got %q", path)
	}
	if !strings.Contains(path, ".config") {
		t.Errorf("expected path to contain '.config', got %q", path)
	}
}
