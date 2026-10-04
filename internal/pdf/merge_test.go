package pdf

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestMerge(t *testing.T) {
	dir := t.TempDir()

	input1 := filepath.Join(dir, "input-1.pdf")
	input2 := filepath.Join(dir, "input-2.pdf")
	output := filepath.Join(dir, "merged.pdf")

	if err := writeTestPDF(input1); err != nil {
		t.Fatalf("create first PDF: %v", err)
	}

	if err := writeTestPDF(input2); err != nil {
		t.Fatalf("create second PDF: %v", err)
	}

	if err := Merge(
		context.Background(),
		[]string{input1, input2},
		output,
	); err != nil {
		t.Fatalf("merge PDFs: %v", err)
	}

	info, err := os.Stat(output)
	if err != nil {
		t.Fatalf("stat merged PDF: %v", err)
	}

	if info.Size() == 0 {
		t.Fatal("merged PDF is empty")
	}
}

func TestMergeRequiresMultipleInputs(t *testing.T) {
	err := Merge(
		context.Background(),
		[]string{"input.pdf"},
		"output.pdf",
	)

	if err == nil {
		t.Fatal("expected error for a single input")
	}
}

func TestMergeHonorsCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := Merge(
		ctx,
		[]string{"input-1.pdf", "input-2.pdf"},
		"output.pdf",
	)

	if err == nil {
		t.Fatal("expected context cancellation error")
	}
}

func writeTestPDF(path string) error {
	const pdf = `%PDF-1.4
1 0 obj
<< /Type /Catalog /Pages 2 0 R >>
endobj
2 0 obj
<< /Type /Pages /Kids [3 0 R] /Count 1 >>
endobj
3 0 obj
<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Contents 4 0 R >>
endobj
4 0 obj
<< /Length 44 >>
stream
BT
/F1 24 Tf
100 700 Td
(PaperShift Test) Tj
ET
endstream
endobj
xref
0 5
0000000000 65535 f
0000000009 00000 n
0000000058 00000 n
0000000115 00000 n
0000000212 00000 n
trailer
<< /Size 5 /Root 1 0 R >>
startxref
306
%%EOF
`

	return os.WriteFile(
		path,
		[]byte(pdf),
		0600,
	)
}