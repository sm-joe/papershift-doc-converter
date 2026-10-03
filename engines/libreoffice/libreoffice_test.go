package libreoffice

import (
	"context"
	"strings"
	"testing"

	"github.com/sm-joe/papershift-doc-converter/internal/converter"
	"github.com/sm-joe/papershift-doc-converter/internal/formats"
)

func TestName(t *testing.T) {
	converter := New("libreoffice")

	if converter.Name() != "libreoffice" {
		t.Fatalf(
			"expected libreoffice, got %s",
			converter.Name(),
		)
	}
}

func TestSupportsDOCXToPDF(t *testing.T) {
	converter := New("libreoffice")

	input, _ := formats.Get("docx")
	output, _ := formats.Get("pdf")

	if !converter.Supports(input, output) {
		t.Fatal("expected DOCX to PDF to be supported")
	}
}

func TestRejectsPNGToPDF(t *testing.T) {
	converter := New("libreoffice")

	input, _ := formats.Get("png")
	output, _ := formats.Get("pdf")

	if converter.Supports(input, output) {
		t.Fatal("expected PNG to PDF to be unsupported")
	}
}

func TestReplaceExtension(t *testing.T) {
	result := replaceExtension(
		"document.docx",
		".pdf",
	)

	if result != "document.pdf" {
		t.Fatalf(
			"expected document.pdf, got %s",
			result,
		)
	}
}

func TestConvertFailsWithMissingBinary(t *testing.T) {
	converter := New("papershift-nonexistent-binary")

	input, _ := formats.Get("docx")
	output, _ := formats.Get("pdf")

	var result strings.Builder

	_, err := converter.Convert(
		context.Background(),
		converterJob(
			input,
			output,
			strings.NewReader("not a real document"),
			&result,
		),
	)

	if err == nil {
		t.Fatal("expected conversion to fail")
	}
}

func converterJob(
	input formats.Format,
	output formats.Format,
	reader *strings.Reader,
	writer *strings.Builder,
) converter.Job {
	return converter.Job{
		InputFormat:  input,
		OutputFormat: output,
		Input:        reader,
		Output:       writer,
	}
}
