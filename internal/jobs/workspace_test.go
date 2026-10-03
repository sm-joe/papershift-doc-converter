package jobs

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCreateWorkspace(t *testing.T) {
	root := t.TempDir()

	workspace, err := CreateWorkspace(root, "test-job")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if workspace.Path == "" {
		t.Fatal("expected workspace path")
	}

	if _, err := os.Stat(workspace.Input); err != nil {
		t.Fatalf("input directory was not created: %v", err)
	}

	if _, err := os.Stat(workspace.Output); err != nil {
		t.Fatalf("output directory was not created: %v", err)
	}
}

func TestWorkspaceCleanup(t *testing.T) {
	root := t.TempDir()

	workspace, err := CreateWorkspace(root, "cleanup-job")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	testFile := filepath.Join(workspace.Input, "test.txt")

	if err := os.WriteFile(testFile, []byte("test"), 0600); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	if err := workspace.Cleanup(); err != nil {
		t.Fatalf("cleanup failed: %v", err)
	}

	if _, err := os.Stat(workspace.Path); !os.IsNotExist(err) {
		t.Fatal("expected workspace to be removed")
	}
}
