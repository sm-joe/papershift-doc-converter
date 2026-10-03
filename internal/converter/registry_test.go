package converter

import (
	"context"
	"io"

	"testing"

	"github.com/sm-joe/papershift-doc-converter/internal/formats"
)

type testConverter struct {
	name   string
	input  string
	output string
}

func (c testConverter) Name() string {
	return c.name
}

func (c testConverter) Supports(input formats.Format, output formats.Format) bool {
	return input.ID == c.input && output.ID == c.output
}

func (c testConverter) Convert(
	_ context.Context,
	job Job,
) (Result, error) {
	_, _ = io.Copy(job.Output, job.Input)

	return Result{
		InputFormat:  job.InputFormat,
		OutputFormat: job.OutputFormat,
	}, nil
}

func TestRegistryFindsSupportedConverter(t *testing.T) {
	registry := NewRegistry()

	registry.Register(testConverter{
		name:   "test",
		input:  "docx",
		output: "pdf",
	})

	input, _ := formats.Get("docx")
	output, _ := formats.Get("pdf")

	converter, err := registry.Find(input, output)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if converter.Name() != "test" {
		t.Fatalf("expected test converter, got %s", converter.Name())
	}
}

func TestRegistryRejectsUnsupportedConversion(t *testing.T) {
	registry := NewRegistry()

	registry.Register(testConverter{
		name:   "test",
		input:  "docx",
		output: "pdf",
	})

	input, _ := formats.Get("png")
	output, _ := formats.Get("pdf")

	_, err := registry.Find(input, output)

	if err != ErrNoConverter {
		t.Fatalf("expected ErrNoConverter, got %v", err)
	}
}

func TestRegistryReturnsRegisteredConverters(t *testing.T) {
	registry := NewRegistry()

	registry.Register(testConverter{name: "one"})
	registry.Register(testConverter{name: "two"})

	converters := registry.All()

	if len(converters) != 2 {
		t.Fatalf("expected 2 converters, got %d", len(converters))
	}
}
