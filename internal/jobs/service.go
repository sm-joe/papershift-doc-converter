package jobs

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/sm-joe/papershift-doc-converter/internal/converter"
	"github.com/sm-joe/papershift-doc-converter/internal/detection"
	"github.com/sm-joe/papershift-doc-converter/internal/formats"
)

type Service struct {
	workspaceRoot string
	detector      *detection.Detector
	converters    *converter.Registry
}

func NewService(
	workspaceRoot string,
	detector *detection.Detector,
	converters *converter.Registry,
) *Service {
	return &Service{
		workspaceRoot: workspaceRoot,
		detector:      detector,
		converters:    converters,
	}
}

type ConvertRequest struct {
	JobID        string
	Filename     string
	Input        io.Reader
	OutputFormat string
}

type ConvertResponse struct {
	Job    Job
	Output string
}

func (s *Service) Convert(
	ctx context.Context,
	request ConvertRequest,
) (ConvertResponse, error) {
	job := Job{
		ID:        request.JobID,
		Status:    StatusPending,
		CreatedAt: time.Now().UTC(),
	}

	workspace, err := CreateWorkspace(
		s.workspaceRoot,
		request.JobID,
	)
	if err != nil {
		return ConvertResponse{}, err
	}

	defer workspace.Cleanup()

	inputPath := filepath.Join(
		workspace.Input,
		safeFilename(request.Filename),
	)

	if err := writeInput(inputPath, request.Input); err != nil {
		return ConvertResponse{}, fmt.Errorf(
			"write input: %w",
			err,
		)
	}

	inputFile, err := os.Open(inputPath)
	if err != nil {
		return ConvertResponse{}, fmt.Errorf(
			"open input: %w",
			err,
		)
	}

	defer inputFile.Close()

	inputFormat, err := s.detector.Detect(
		request.Filename,
		inputFile,
	)
	if err != nil {
		return ConvertResponse{}, fmt.Errorf(
			"detect input format: %w",
			err,
		)
	}

	outputFormat, ok := formats.Get(request.OutputFormat)
	if !ok {
		return ConvertResponse{}, fmt.Errorf(
			"unsupported output format: %s",
			request.OutputFormat,
		)
	}

	job.InputFormat = inputFormat
	job.OutputFormat = outputFormat
	job.Status = StatusProcessing

	selectedConverter, err := s.converters.Find(
		inputFormat,
		outputFormat,
	)
	if err != nil {
		return ConvertResponse{}, fmt.Errorf(
			"find converter: %w",
			err,
		)
	}

	if _, err := inputFile.Seek(0, io.SeekStart); err != nil {
		return ConvertResponse{}, fmt.Errorf(
			"rewind input: %w",
			err,
		)
	}

	outputName := safeFilename(
		replaceExtension(
			request.Filename,
			outputFormat.Extension,
		),
	)

	outputPath := filepath.Join(
		workspace.Output,
		outputName,
	)

	outputFile, err := os.OpenFile(
		outputPath,
		os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
		0600,
	)
	if err != nil {
		return ConvertResponse{}, fmt.Errorf(
			"create output: %w",
			err,
		)
	}

	defer outputFile.Close()

	_, err = selectedConverter.Convert(
		ctx,
		converter.Job{
			InputFormat:  inputFormat,
			OutputFormat: outputFormat,
			Input:        inputFile,
			Output:       outputFile,
		},
	)
	if err != nil {
		return ConvertResponse{}, fmt.Errorf(
			"conversion failed: %w",
			err,
		)
	}

	job.Status = StatusCompleted

	now := time.Now().UTC()
	job.CompletedAt = &now

	return ConvertResponse{
		Job:    job,
		Output: outputPath,
	}, nil
}

func writeInput(path string, input io.Reader) error {
	file, err := os.OpenFile(
		path,
		os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
		0600,
	)
	if err != nil {
		return err
	}

	defer file.Close()

	_, err = io.Copy(file, input)

	return err
}

func safeFilename(name string) string {
	return filepath.Base(name)
}

func replaceExtension(name, extension string) string {
	base := name

	if ext := filepath.Ext(name); ext != "" {
		base = name[:len(name)-len(ext)]
	}

	return base + extension
}
