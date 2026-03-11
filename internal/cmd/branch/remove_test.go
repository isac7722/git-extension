package branch

import (
	"testing"
)

func TestParseRemoveArgs(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantF    bool
		wantB    []string
	}{
		{
			name:  "no args",
			args:  nil,
			wantF: false,
			wantB: nil,
		},
		{
			name:  "branches only",
			args:  []string{"feat1", "feat2"},
			wantF: false,
			wantB: []string{"feat1", "feat2"},
		},
		{
			name:  "force long flag",
			args:  []string{"--force", "feat1"},
			wantF: true,
			wantB: []string{"feat1"},
		},
		{
			name:  "force short flag",
			args:  []string{"-f", "feat1"},
			wantF: true,
			wantB: []string{"feat1"},
		},
		{
			name:  "force flag at end",
			args:  []string{"feat1", "--force"},
			wantF: true,
			wantB: []string{"feat1"},
		},
		{
			name:  "force only",
			args:  []string{"--force"},
			wantF: true,
			wantB: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			force, branches := parseRemoveArgs(tt.args)
			if force != tt.wantF {
				t.Errorf("force = %v, want %v", force, tt.wantF)
			}
			if len(branches) != len(tt.wantB) {
				t.Fatalf("branches = %v, want %v", branches, tt.wantB)
			}
			for i := range branches {
				if branches[i] != tt.wantB[i] {
					t.Errorf("branches[%d] = %q, want %q", i, branches[i], tt.wantB[i])
				}
			}
		})
	}
}

func TestLocationLabel(t *testing.T) {
	tests := []struct {
		local  bool
		remote bool
		want   string
	}{
		{true, true, "local + remote"},
		{false, true, "remote"},
		{true, false, "local"},
		{false, false, "local"},
	}

	for _, tt := range tests {
		got := locationLabel(tt.local, tt.remote)
		if got != tt.want {
			t.Errorf("locationLabel(%v, %v) = %q, want %q", tt.local, tt.remote, got, tt.want)
		}
	}
}
