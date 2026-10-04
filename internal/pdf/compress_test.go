package pdf

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestCompress(t *testing.T) {
	dir := t.TempDir()

	input := filepath.Join(dir, "input.pdf")
	output := filepath.Join(dir, "output.pdf")

	pdf := []byte("%PDF-1.4\n1 0 obj\n<< /Type /Catalog /Pages 2 0 R >>\nendobj\n2 0 obj\n<< /Type /Pages /Kids [] /Count 0 >>\nendobj\nxref\n0 3\n0000000000 65535 f \n0000000009 00000 n \n0000000058 00000 n \ntrailer\n<< /Root 1 0 R /Size 3 >>\nstartxref\n115\n%%EOF\n")

	if err := os.WriteFile(input, pdf, 0600); err != nil {
		t.Fatalf("write input PDF: %v", err)
	}

	if _, err := exec.LookPath("gs"); err != nil {
		t.Skip("Ghostscript is not installed; compression integration test requires Ghostscript")
	}

	if err := Compress(context.Background(), input, output); err != nil {
		t.Fatalf("Compress() error = %v", err)
	}

	info, err := os.Stat(output)
	if err != nil {
		t.Fatalf("stat output PDF: %v", err)
	}

	if info.Size() == 0 {
		t.Fatal("compressed PDF is empty")
	}
}

func TestCompressCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := Compress(
		ctx,
		"input.pdf",
		"output.pdf",
	)
	if err == nil {
		t.Fatal("expected canceled context error")
	}
}
