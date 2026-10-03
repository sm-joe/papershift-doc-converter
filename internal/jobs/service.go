package jobs

import (
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
	"strconv"

	"github.com/sm-joe/papershift-doc-converter/internal/converter"
	"github.com/sm-joe/papershift-doc-converter/internal/detection"
	"github.com/sm-joe/papershift-doc-converter/internal/formats"
)

type Service struct {
	workspaceRoot string
	detector      *detection.Detector
	converters    *converter.Registry
	conversionTimeout time.Duration
	maxInputSize int64
	maxOutputSize int64
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
		conversionTimeout: time.Duration(
			envInt64("PAPERSHIFT_CONVERSION_TIMEOUT_SECONDS", 120),
		) * time.Second,
		maxInputSize: envInt64(
			"PAPERSHIFT_MAX_INPUT_SIZE_BYTES",
			50*1024*1024,
		),
		maxOutputSize: envInt64(
			"PAPERSHIFT_MAX_OUTPUT_SIZE_BYTES",
			100*1024*1024,
		),
	}
}

type ConvertRequest struct {
	JobID        string
	Filename     string
	Input        io.Reader
	OutputFormat string
}

type ConvertResponse struct {
	Job      Job
	Filename string
	Output   []byte
}

func (s *Service) Convert(ctx context.Context, request ConvertRequest) (ConvertResponse, error) {
	conversionCtx, cancel := context.WithTimeout(
		ctx,
		s.conversionTimeout,
	)
	defer cancel()
	now := time.Now()

	job := Job{
		ID:        request.JobID,
		Status:    StatusPending,
		CreatedAt: now,
	}

	workspace, err := CreateWorkspace(s.workspaceRoot, job.ID)
	if err != nil {
		return ConvertResponse{}, fmt.Errorf("create workspace: %w", err)
	}
	defer workspace.Cleanup()

	inputPath := filepath.Join(
		workspace.Input,
		safeFilename(request.Filename),
	)

	if err := writeInputLimited(
		inputPath,
		request.Input,
		s.maxInputSize,
	); err != nil {
		job.Status = StatusFailed
		job.Error = err.Error()

		return ConvertResponse{
			Job: job,
		}, err
	}

	inputFile, err := os.Open(inputPath)
	if err != nil {
		job.Status = StatusFailed
		job.Error = err.Error()

		return ConvertResponse{
			Job: job,
		}, fmt.Errorf("open input: %w", err)
	}
	defer inputFile.Close()

	inputFormat, err := s.detector.Detect(request.Filename, inputFile)
	if err != nil {
		job.Status = StatusFailed
		job.Error = err.Error()

		return ConvertResponse{
			Job: job,
		}, fmt.Errorf("detect input format: %w", err)
	}

	outputFormat, ok := formats.Get(request.OutputFormat)
	if !ok {
		err := fmt.Errorf("unknown output format: %s", request.OutputFormat)

		job.Status = StatusFailed
		job.Error = err.Error()

		return ConvertResponse{
			Job: job,
		}, err
	}

	job.InputFormat = inputFormat
	job.OutputFormat = outputFormat

	selectedConverter, err := s.converters.Find(inputFormat, outputFormat)
	if err != nil {
		job.Status = StatusFailed
		job.Error = err.Error()

		return ConvertResponse{
			Job: job,
		}, err
	}

	if _, err := inputFile.Seek(0, io.SeekStart); err != nil {
		job.Status = StatusFailed
		job.Error = err.Error()

		return ConvertResponse{
			Job: job,
		}, fmt.Errorf("rewind input: %w", err)
	}

	job.Status = StatusProcessing

	startedAt := time.Now()
	job.StartedAt = &startedAt

	var output bytes.Buffer

	limitedOutput := &limitedWriter{
		writer: &output,
		limit:  s.maxOutputSize,
	}

	_, err = selectedConverter.Convert(conversionCtx, converter.Job{
		InputFormat:  inputFormat,
		OutputFormat: outputFormat,
		Input:        inputFile,
		Output:       limitedOutput,
	})
	if err != nil {
		job.Status = StatusFailed
		job.Error = err.Error()

		return ConvertResponse{
			Job: job,
		}, fmt.Errorf("conversion failed: %w", err)
	}

	if output.Len() == 0 {
	err := fmt.Errorf("conversion produced empty output")

	job.Status = StatusFailed
	job.Error = err.Error()

	return ConvertResponse{
		Job: job,
	}, err
}

	completedAt := time.Now()
	job.CompletedAt = &completedAt
	job.Status = StatusCompleted

	outputFilename := replaceExtension(
		safeFilename(request.Filename),
		outputFormat.Extension,
	)

	return ConvertResponse{
		Job:      job,
		Filename: outputFilename,
		Output:   output.Bytes(),
	}, nil
}

func writeInput(path string, input io.Reader) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create input file: %w", err)
	}
	defer file.Close()

	if _, err := io.Copy(file, input); err != nil {
		return fmt.Errorf("write input file: %w", err)
	}

	return nil
}

func writeInputLimited(
	path string,
	input io.Reader,
	maxSize int64,
) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create input file: %w", err)
	}
	defer file.Close()

	reader := io.LimitReader(input, maxSize+1)

	written, err := io.Copy(file, reader)
	if err != nil {
		return fmt.Errorf("write input file: %w", err)
	}

	if written > maxSize {
		return fmt.Errorf(
			"input file exceeds maximum size of %d bytes",
			maxSize,
		)
	}

	return nil
}

func safeFilename(name string) string {
	return filepath.Base(name)
}

func replaceExtension(filename, extension string) string {
	if extension == "" {
		return filename
	}

	if !strings.HasPrefix(extension, ".") {
		extension = "." + extension
	}

	ext := filepath.Ext(filename)
	if ext == "" {
		return filename + extension
	}

	return filename[:len(filename)-len(ext)] + extension
}

func newJobID() (string, error) {
	var bytesID [16]byte

	if _, err := rand.Read(bytesID[:]); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", bytesID), nil
}

func envInt64(name string, defaultValue int64) int64 {
	value := os.Getenv(name)
	if value == "" {
		return defaultValue
	}

	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil || parsed <= 0 {
		return defaultValue
	}

	return parsed
}

type limitedWriter struct {
	writer  io.Writer
	limit   int64
	written int64
}

func (w *limitedWriter) Write(p []byte) (int, error) {
	remaining := w.limit - w.written

	if remaining <= 0 {
		return 0, fmt.Errorf(
			"output exceeds maximum size of %d bytes",
			w.limit,
		)
	}

	if int64(len(p)) > remaining {
		n, _ := w.writer.Write(p[:remaining])
		w.written += int64(n)

		return n, fmt.Errorf(
			"output exceeds maximum size of %d bytes",
			w.limit,
		)
	}

	n, err := w.writer.Write(p)
	w.written += int64(n)

	return n, err
}
