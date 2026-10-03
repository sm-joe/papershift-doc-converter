package detection

import (
	"archive/zip"
	"bytes"
	"testing"
)

func TestDetectPDF(t *testing.T) {
	detector := New()

	format, err := detector.Detect(
		"document.pdf",
		bytes.NewReader([]byte("%PDF-1.7")),
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if format.ID != "pdf" {
		t.Fatalf("expected pdf, got %s", format.ID)
	}
}

func TestDetectPNG(t *testing.T) {
	detector := New()

	data := []byte{
		0x89, 0x50, 0x4e, 0x47,
		0x0d, 0x0a, 0x1a, 0x0a,
	}

	format, err := detector.Detect(
		"image.png",
		bytes.NewReader(data),
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if format.ID != "png" {
		t.Fatalf("expected png, got %s", format.ID)
	}
}

func TestDetectJPEG(t *testing.T) {
	detector := New()

	format, err := detector.Detect(
		"image.jpg",
		bytes.NewReader([]byte{0xff, 0xd8, 0xff, 0xe0}),
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if format.ID != "jpg" {
		t.Fatalf("expected jpg, got %s", format.ID)
	}
}

func TestDetectDOCXContainer(t *testing.T) {
	data := createZIPFixture(t, map[string]string{
		"[Content_Types].xml": "content types",
		"word/document.xml":   "document",
		"word/styles.xml":     "styles",
	})

	detector := New()

	format, err := detector.Detect(
		"document.docx",
		bytes.NewReader(data),
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if format.ID != "docx" {
		t.Fatalf("expected docx, got %s", format.ID)
	}
}

func TestDetectXLSXContainer(t *testing.T) {
	data := createZIPFixture(t, map[string]string{
		"[Content_Types].xml": "content types",
		"xl/workbook.xml":     "workbook",
		"xl/styles.xml":       "styles",
	})

	detector := New()

	format, err := detector.Detect(
		"spreadsheet.xlsx",
		bytes.NewReader(data),
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if format.ID != "xlsx" {
		t.Fatalf("expected xlsx, got %s", format.ID)
	}
}

func TestDetectPPTXContainer(t *testing.T) {
	data := createZIPFixture(t, map[string]string{
		"[Content_Types].xml":   "content types",
		"ppt/presentation.xml":  "presentation",
		"ppt/slides/slide1.xml": "slide",
	})

	detector := New()

	format, err := detector.Detect(
		"presentation.pptx",
		bytes.NewReader(data),
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if format.ID != "pptx" {
		t.Fatalf("expected pptx, got %s", format.ID)
	}
}

func TestDetectMarkdown(t *testing.T) {
	detector := New()

	format, err := detector.Detect(
		"README.md",
		bytes.NewReader([]byte("# PaperShift")),
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if format.ID != "md" {
		t.Fatalf("expected md, got %s", format.ID)
	}
}

func TestUnknownBinary(t *testing.T) {
	detector := New()

	_, err := detector.Detect(
		"unknown.bin",
		bytes.NewReader([]byte{0x00, 0x01, 0x02, 0x03}),
	)

	if err != ErrUnknownFormat {
		t.Fatalf("expected ErrUnknownFormat, got %v", err)
	}
}

func createZIPFixture(t *testing.T, files map[string]string) []byte {
	t.Helper()

	var buffer bytes.Buffer

	writer := zip.NewWriter(&buffer)

	for name, content := range files {
		file, err := writer.Create(name)
		if err != nil {
			t.Fatalf("create ZIP entry %q: %v", name, err)
		}

		if _, err := file.Write([]byte(content)); err != nil {
			t.Fatalf("write ZIP entry %q: %v", name, err)
		}
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("close ZIP writer: %v", err)
	}

	return buffer.Bytes()
}
