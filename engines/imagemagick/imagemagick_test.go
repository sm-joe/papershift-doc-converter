package imagemagick

import (
	"testing"

	"github.com/sm-joe/papershift-doc-converter/internal/formats"
)

func TestSupports(t *testing.T) {
	input, _ := formats.Get("png")
	output, _ := formats.Get("jpg")

	converter := New("magick")

	if !converter.Supports(input, output) {
		t.Fatal("expected PNG to JPG to be supported")
	}
}

func TestRejectsNonImageFormats(t *testing.T) {
	input, _ := formats.Get("txt")
	output, _ := formats.Get("png")

	converter := New("magick")

	if converter.Supports(input, output) {
		t.Fatal("expected TXT to PNG to be rejected")
	}
}
