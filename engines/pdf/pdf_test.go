package pdf

import (
	"testing"

	"github.com/sm-joe/papershift-doc-converter/internal/formats"
)

func getFormat(t *testing.T, id string) formats.Format {
	t.Helper()

	format, ok := formats.Get(id)
	if !ok {
		t.Fatalf("format %q not found", id)
	}

	return format
}

func TestConverterName(t *testing.T) {
	c := New(
		"pdftotext",
		"pdftohtml",
		"pandoc",
		"pdftoppm",
		"convert",
	)

	if got := c.Name(); got != "pdf" {
		t.Fatalf("expected pdf, got %q", got)
	}
}

func TestConverterSupports(t *testing.T) {
	c := New(
		"pdftotext",
		"pdftohtml",
		"pandoc",
		"pdftoppm",
		"convert",
	)

	pdf := getFormat(t, "pdf")
	txt := getFormat(t, "txt")
	docx := getFormat(t, "docx")
	odt := getFormat(t, "odt")
	html := getFormat(t, "html")
	jpg := getFormat(t, "jpg")
	png := getFormat(t, "png")
	webp := getFormat(t, "webp")

	tests := []struct {
		name   string
		input  formats.Format
		output formats.Format
		want   bool
	}{
		{
			name:   "pdf to txt",
			input:  pdf,
			output: txt,
			want:   true,
		},
		{
			name:   "pdf to docx",
			input:  pdf,
			output: docx,
			want:   true,
		},
		{
			name:   "pdf to odt",
			input:  pdf,
			output: odt,
			want:   true,
		},
		{
			name:   "pdf to html",
			input:  pdf,
			output: html,
			want:   true,
		},
		{
			name:   "pdf to jpg",
			input:  pdf,
			output: jpg,
			want:   true,
		},
		{
			name:   "non-pdf input unsupported",
			input:  docx,
			output: txt,
			want:   false,
		},
		{
			name:   "pdf to png",
			input:  pdf,
			output: png,
			want:   true,
		},
		{
			name:   "pdf to jpg",
			input:  pdf,
			output: jpg,
			want:   true,
		},
		{
			name:   "pdf to webp",
			input:  pdf,
			output: webp,
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := c.Supports(tt.input, tt.output); got != tt.want {
				t.Fatalf(
					"Supports(%q -> %q) = %v, want %v",
					tt.input.ID,
					tt.output.ID,
					got,
					tt.want,
				)
			}
		})
	}
}
