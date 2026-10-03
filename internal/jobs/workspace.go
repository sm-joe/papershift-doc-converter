package jobs

import (
	"fmt"
	"os"
	"path/filepath"
)

type Workspace struct {
	Path   string
	Input  string
	Output string
}

func CreateWorkspace(root string, jobID string) (Workspace, error) {
	if root == "" {
		root = os.TempDir()
	}

	path := filepath.Join(root, "papershift", jobID)

	input := filepath.Join(path, "input")
	output := filepath.Join(path, "output")

	if err := os.MkdirAll(input, 0700); err != nil {
		return Workspace{}, fmt.Errorf("create input workspace: %w", err)
	}

	if err := os.MkdirAll(output, 0700); err != nil {
		_ = os.RemoveAll(path)
		return Workspace{}, fmt.Errorf("create output workspace: %w", err)
	}

	return Workspace{
		Path:   path,
		Input:  input,
		Output: output,
	}, nil
}

func (w Workspace) Cleanup() error {
	return os.RemoveAll(w.Path)
}
