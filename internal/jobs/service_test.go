package jobs

import (
	"bytes"
	"context"
	"testing"

	"github.com/sm-joe/papershift-doc-converter/internal/converter"
	"github.com/sm-joe/papershift-doc-converter/internal/detection"
	"github.com/sm-joe/papershift-doc-converter/internal/formats"
)

type testConverter struct{}

func (testConverter) Name() string {
	return "test"
}

func (testConverter) Supports(
	input formats.Format,
	output formats.Format,
) bool {
	return input.ID == "txt" && output.ID == "txt"
}

func (testConverter) Convert(
	_ context.Context,
	job converter.Job,
) (converter.Result, error) {
	_, err := bytes.NewBufferString(
		"converted",
	).WriteTo(job.Output)

	if err != nil {
		return converter.Result{}, err
	}

	return converter.Result{
		InputFormat:  job.InputFormat,
		OutputFormat: job.OutputFormat,
	}, nil
}

func TestServiceConvert(t *testing.T) {
	registry := converter.NewRegistry()
	registry.Register(testConverter{})

	service := NewService(
		t.TempDir(),
		detection.New(),
		registry,
	)

	response, err := service.Convert(
		context.Background(),
		ConvertRequest{
			JobID:        "test-job",
			Filename:     "input.txt",
			Input:        bytes.NewBufferString("hello"),
			OutputFormat: "txt",
		},
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if response.Job.Status != StatusCompleted {
		t.Fatalf(
			"expected completed status, got %s",
			response.Job.Status,
		)
	}

	if response.Job.InputFormat.ID != "txt" {
		t.Fatalf(
			"expected txt input, got %s",
			response.Job.InputFormat.ID,
		)
	}

	if response.Job.OutputFormat.ID != "txt" {
		t.Fatalf(
			"expected txt output, got %s",
			response.Job.OutputFormat.ID,
		)
	}
}
