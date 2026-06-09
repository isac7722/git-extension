package worktreesetup

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCopyFilesMappedFileWithRootAlias(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()
	writeTestFile(t, filepath.Join(srcDir, ".env"), []byte("TOKEN=main\n"))

	results := CopyFiles([]CopySpec{{From: "/.env", To: "/server-main/.env"}}, srcDir, dstDir, false)
	if len(results) != 1 {
		t.Fatalf("len(results) = %d, want 1", len(results))
	}
	if results[0].Error != nil {
		t.Fatalf("CopyFiles() error = %v", results[0].Error)
	}
	if !results[0].Copied {
		t.Fatalf("Copied = false, result = %+v", results[0])
	}
	assertFileContent(t, filepath.Join(dstDir, "server-main", ".env"), "TOKEN=main\n")
}

func TestCopyFilesMappedDirectoryWithRootAlias(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()
	writeTestFile(t, filepath.Join(srcDir, ".vscode", "settings.json"), []byte("{}\n"))

	results := CopyFiles([]CopySpec{{From: "/.vscode", To: "/.vscode"}}, srcDir, dstDir, false)
	if len(results) != 1 {
		t.Fatalf("len(results) = %d, want 1", len(results))
	}
	if results[0].Error != nil {
		t.Fatalf("CopyFiles() error = %v", results[0].Error)
	}
	if !results[0].Copied || !results[0].IsDir || results[0].FilesCopied != 1 {
		t.Fatalf("result = %+v, want copied directory with 1 file", results[0])
	}
	assertFileContent(t, filepath.Join(dstDir, ".vscode", "settings.json"), "{}\n")
}

func TestCopyFilesSkipAndForceOverwrite(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()
	writeTestFile(t, filepath.Join(srcDir, ".env"), []byte("TOKEN=new\n"))
	writeTestFile(t, filepath.Join(dstDir, ".env"), []byte("TOKEN=old\n"))

	results := CopyFiles([]CopySpec{{From: ".env", To: ".env"}}, srcDir, dstDir, false)
	if results[0].Error != nil {
		t.Fatalf("CopyFiles(force=false) error = %v", results[0].Error)
	}
	if !results[0].Skipped {
		t.Fatalf("Skipped = false, result = %+v", results[0])
	}
	assertFileContent(t, filepath.Join(dstDir, ".env"), "TOKEN=old\n")

	results = CopyFiles([]CopySpec{{From: ".env", To: ".env"}}, srcDir, dstDir, true)
	if results[0].Error != nil {
		t.Fatalf("CopyFiles(force=true) error = %v", results[0].Error)
	}
	if !results[0].Copied {
		t.Fatalf("Copied = false, result = %+v", results[0])
	}
	assertFileContent(t, filepath.Join(dstDir, ".env"), "TOKEN=new\n")
}

func TestCopyFilesMappedDirectory(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()
	writeTestFile(t, filepath.Join(srcDir, "config", "app.env"), []byte("APP=api\n"))
	writeTestFile(t, filepath.Join(srcDir, "config", "nested", "worker.env"), []byte("WORKER=1\n"))

	results := CopyFiles([]CopySpec{{From: "config", To: "server/config"}}, srcDir, dstDir, false)
	if results[0].Error != nil {
		t.Fatalf("CopyFiles() error = %v", results[0].Error)
	}
	if !results[0].Copied || !results[0].IsDir || results[0].FilesCopied != 2 {
		t.Fatalf("result = %+v, want copied directory with 2 files", results[0])
	}
	assertFileContent(t, filepath.Join(dstDir, "server", "config", "app.env"), "APP=api\n")
	assertFileContent(t, filepath.Join(dstDir, "server", "config", "nested", "worker.env"), "WORKER=1\n")
}

func TestCopyFilesRejectsPathsOutsideWorktree(t *testing.T) {
	srcDir := t.TempDir()
	dstDir := t.TempDir()
	writeTestFile(t, filepath.Join(srcDir, ".env"), []byte("TOKEN=main\n"))

	tests := []CopySpec{
		{From: "../.env", To: ".env"},
		{From: ".env", To: "../.env"},
	}

	for _, spec := range tests {
		results := CopyFiles([]CopySpec{spec}, srcDir, dstDir, false)
		if len(results) != 1 {
			t.Fatalf("len(results) = %d, want 1", len(results))
		}
		if results[0].Error == nil {
			t.Fatalf("CopyFiles(%+v) error = nil, want error", spec)
		}
		if !strings.Contains(results[0].Error.Error(), "inside the worktree") {
			t.Fatalf("error = %v, want inside the worktree", results[0].Error)
		}
	}
}

func assertFileContent(t *testing.T, path, want string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", path, err)
	}
	if string(data) != want {
		t.Fatalf("ReadFile(%q) = %q, want %q", path, string(data), want)
	}
}
