package jobs

import (
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/sm-joe/papershift-doc-converter/internal/converter"
	"github.com/sm-joe/papershift-doc-converter/internal/detection"
	"github.com/sm-joe/papershift-doc-converter/internal/docx"
	"github.com/sm-joe/papershift-doc-converter/internal/formats"
	"github.com/sm-joe/papershift-doc-converter/internal/image"
	"github.com/sm-joe/papershift-doc-converter/internal/pdf"
)

type Service struct {
	workspaceRoot     string
	detector          *detection.Detector
	converters        *converter.Registry
	conversionTimeout time.Duration
	maxInputSize      int64
	maxOutputSize     int64
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

type MergePDFRequest struct {
	JobID string
	Files []MergePDFFile
}

type MergePDFFile struct {
	Filename string
	Input    io.Reader
}

type MergePDFResponse struct {
	Job      Job
	Filename string
	Output   []byte
}

type RotatePDFRequest struct {
	JobID    string
	Filename string
	Input    io.Reader
	Rotation int
}

type RotatePDFResponse struct {
	Job      Job
	Filename string
	Output   []byte
}

type CompressPDFRequest struct {
	JobID    string
	Filename string
	Input    io.Reader
}

type CompressPDFResponse struct {
	Job      Job
	Filename string
	Output   []byte
}

type CompressImageRequest struct {
	JobID    string
	Filename string
	Input    io.Reader
}

type CompressImageResponse struct {
	Job      Job
	Filename string
	Output   []byte
}

type CompressDOCXRequest struct {
	JobID    string
	Filename string
	Input    io.Reader
}

type CompressDOCXResponse struct {
	Job      Job
	Filename string
	Output   []byte
}

func (s *Service) Convert(ctx context.Context, request ConvertRequest) (ConvertResponse, error) {
	if request.Input == nil {
		return ConvertResponse{}, fmt.Errorf("input is required")
	}

	if request.JobID == "" {
		return ConvertResponse{}, fmt.Errorf("job ID is required")
	}

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
	defer func() {
		_ = workspace.Cleanup()
	}()

	inputPath := filepath.Join(
		workspace.Input,
		safeFilename(request.Filename),
	)

	if err := writeInputLimited(
		conversionCtx,
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

	if err := conversionCtx.Err(); err != nil {
		job.Status = StatusFailed
		job.Error = err.Error()

		return ConvertResponse{Job: job}, err
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

func (s *Service) MergePDF(
	ctx context.Context,
	request MergePDFRequest,
) (MergePDFResponse, error) {
	if len(request.Files) < 2 {
		return MergePDFResponse{}, fmt.Errorf(
			"merge requires at least two PDF files",
		)
	}

	mergeCtx, cancel := context.WithTimeout(
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

	workspace, err := CreateWorkspace(
		s.workspaceRoot,
		job.ID,
	)
	if err != nil {
		return MergePDFResponse{}, fmt.Errorf(
			"create workspace: %w",
			err,
		)
	}
	defer func() {
		_ = workspace.Cleanup()
	}()

	inputPaths := make([]string, 0, len(request.Files))

	for index, file := range request.Files {
		if file.Input == nil {
			return MergePDFResponse{}, fmt.Errorf("merge input %d is required", index+1)
		}

		filename := safeFilename(file.Filename)

		if filename == "." || filename == "" {
			filename = fmt.Sprintf("input-%d.pdf", index+1)
		}

		inputPath := filepath.Join(
			workspace.Input,
			fmt.Sprintf("%d-%s", index+1, filename),
		)

		if err := writeInputLimited(
			mergeCtx,
			inputPath,
			file.Input,
			s.maxInputSize,
		); err != nil {
			job.Status = StatusFailed
			job.Error = err.Error()

			return MergePDFResponse{
				Job: job,
			}, err
		}

		inputPaths = append(inputPaths, inputPath)
	}

	outputPath := filepath.Join(
		workspace.Output,
		"merged.pdf",
	)

	job.Status = StatusProcessing

	startedAt := time.Now()
	job.StartedAt = &startedAt

	if err := pdf.Merge(
		mergeCtx,
		inputPaths,
		outputPath,
	); err != nil {
		job.Status = StatusFailed
		job.Error = err.Error()

		return MergePDFResponse{
			Job: job,
		}, fmt.Errorf("merge PDFs: %w", err)
	}

	outputFile, err := os.Open(outputPath)
	if err != nil {
		job.Status = StatusFailed
		job.Error = err.Error()

		return MergePDFResponse{
			Job: job,
		}, fmt.Errorf("open merged PDF: %w", err)
	}
	defer func() {
		_ = outputFile.Close()
	}()

	info, err := outputFile.Stat()
	if err != nil {
		job.Status = StatusFailed
		job.Error = err.Error()

		return MergePDFResponse{
			Job: job,
		}, fmt.Errorf("stat merged PDF: %w", err)
	}

	if info.Size() == 0 {
		err := fmt.Errorf("merge produced empty output")

		job.Status = StatusFailed
		job.Error = err.Error()

		return MergePDFResponse{
			Job: job,
		}, err
	}

	var output bytes.Buffer

	limitedOutput := &limitedWriter{
		writer: &output,
		limit:  s.maxOutputSize,
	}

	if _, err := io.Copy(limitedOutput, outputFile); err != nil {
		job.Status = StatusFailed
		job.Error = err.Error()

		return MergePDFResponse{
			Job: job,
		}, fmt.Errorf("read merged output: %w", err)
	}

	if err := mergeCtx.Err(); err != nil {
		job.Status = StatusFailed
		job.Error = err.Error()

		return MergePDFResponse{Job: job}, err
	}

	completedAt := time.Now()
	job.CompletedAt = &completedAt
	job.Status = StatusCompleted

	return MergePDFResponse{
		Job:      job,
		Filename: "merged.pdf",
		Output:   output.Bytes(),
	}, nil
}

func (s *Service) RotatePDF(
	ctx context.Context,
	request RotatePDFRequest,
) (RotatePDFResponse, error) {
	if request.Input == nil {
		return RotatePDFResponse{}, fmt.Errorf("PDF input is required")
	}

	if request.JobID == "" {
		return RotatePDFResponse{}, fmt.Errorf("job ID is required")
	}

	if request.Rotation != 90 &&
		request.Rotation != 180 &&
		request.Rotation != 270 {
		return RotatePDFResponse{}, fmt.Errorf(
			"rotation must be 90, 180, or 270 degrees",
		)
	}

	rotateCtx, cancel := context.WithTimeout(
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

	workspace, err := CreateWorkspace(
		s.workspaceRoot,
		job.ID,
	)
	if err != nil {
		return RotatePDFResponse{}, fmt.Errorf(
			"create workspace: %w",
			err,
		)
	}
	defer func() {
		_ = workspace.Cleanup()
	}()

	filename := safeFilename(request.Filename)

	if filename == "." || filename == "" {
		filename = "input.pdf"
	}

	inputPath := filepath.Join(
		workspace.Input,
		filename,
	)

	if err := writeInputLimited(
		rotateCtx,
		inputPath,
		request.Input,
		s.maxInputSize,
	); err != nil {
		job.Status = StatusFailed
		job.Error = err.Error()

		return RotatePDFResponse{
			Job: job,
		}, err
	}

	outputPath := filepath.Join(
		workspace.Output,
		"rotated.pdf",
	)

	job.Status = StatusProcessing

	startedAt := time.Now()
	job.StartedAt = &startedAt

	if err := pdf.Rotate(
		rotateCtx,
		inputPath,
		outputPath,
		request.Rotation,
	); err != nil {
		job.Status = StatusFailed
		job.Error = err.Error()

		return RotatePDFResponse{
			Job: job,
		}, fmt.Errorf("rotate PDF: %w", err)
	}

	outputFile, err := os.Open(outputPath)
	if err != nil {
		job.Status = StatusFailed
		job.Error = err.Error()

		return RotatePDFResponse{
			Job: job,
		}, fmt.Errorf("open rotated PDF: %w", err)
	}
	defer func() {
		_ = outputFile.Close()
	}()

	info, err := outputFile.Stat()
	if err != nil {
		job.Status = StatusFailed
		job.Error = err.Error()

		return RotatePDFResponse{
			Job: job,
		}, fmt.Errorf("stat rotated PDF: %w", err)
	}

	if info.Size() == 0 {
		err := fmt.Errorf("rotation produced empty output")

		job.Status = StatusFailed
		job.Error = err.Error()

		return RotatePDFResponse{
			Job: job,
		}, err
	}

	var output bytes.Buffer

	limitedOutput := &limitedWriter{
		writer: &output,
		limit:  s.maxOutputSize,
	}

	if _, err := io.Copy(limitedOutput, outputFile); err != nil {
		job.Status = StatusFailed
		job.Error = err.Error()

		return RotatePDFResponse{
			Job: job,
		}, fmt.Errorf("read rotated output: %w", err)
	}

	if err := rotateCtx.Err(); err != nil {
		job.Status = StatusFailed
		job.Error = err.Error()

		return RotatePDFResponse{Job: job}, err
	}

	completedAt := time.Now()
	job.CompletedAt = &completedAt
	job.Status = StatusCompleted

	return RotatePDFResponse{
		Job:      job,
		Filename: "rotated.pdf",
		Output:   output.Bytes(),
	}, nil
}

func (s *Service) CompressPDF(
	ctx context.Context,
	req CompressPDFRequest,
) (CompressPDFResponse, error) {
	if req.Input == nil {
		return CompressPDFResponse{}, fmt.Errorf("PDF input is required")
	}

	if req.JobID == "" {
		return CompressPDFResponse{}, fmt.Errorf("job ID is required")
	}

	compressCtx, cancel := context.WithTimeout(
		ctx,
		s.conversionTimeout,
	)
	defer cancel()

	now := time.Now()

	job := Job{
		ID:        req.JobID,
		Status:    StatusPending,
		CreatedAt: now,
	}

	workspace, err := CreateWorkspace(
		s.workspaceRoot,
		job.ID,
	)
	if err != nil {
		return CompressPDFResponse{}, fmt.Errorf(
			"create workspace: %w",
			err,
		)
	}
	defer func() {
		_ = workspace.Cleanup()
	}()

	filename := safeFilename(req.Filename)
	if filename == "." || filename == "" {
		filename = "input.pdf"
	}

	inputPath := filepath.Join(
		workspace.Input,
		filename,
	)

	if err := writeInputLimited(
		compressCtx,
		inputPath,
		req.Input,
		s.maxInputSize,
	); err != nil {
		job.Status = StatusFailed
		job.Error = err.Error()

		return CompressPDFResponse{
			Job: job,
		}, err
	}

	outputPath := filepath.Join(
		workspace.Output,
		"compressed.pdf",
	)

	job.Status = StatusProcessing

	startedAt := time.Now()
	job.StartedAt = &startedAt

	if err := pdf.Compress(
		compressCtx,
		inputPath,
		outputPath,
	); err != nil {
		job.Status = StatusFailed
		job.Error = err.Error()

		return CompressPDFResponse{
			Job: job,
		}, fmt.Errorf("compress PDF: %w", err)
	}

	outputFile, err := os.Open(outputPath)
	if err != nil {
		job.Status = StatusFailed
		job.Error = err.Error()

		return CompressPDFResponse{
			Job: job,
		}, fmt.Errorf("open compressed PDF: %w", err)
	}
	defer func() {
		_ = outputFile.Close()
	}()

	info, err := outputFile.Stat()
	if err != nil {
		job.Status = StatusFailed
		job.Error = err.Error()

		return CompressPDFResponse{
			Job: job,
		}, fmt.Errorf("stat compressed PDF: %w", err)
	}

	if info.Size() == 0 {
		err := fmt.Errorf("compression produced empty output")
		job.Status = StatusFailed
		job.Error = err.Error()

		return CompressPDFResponse{
			Job: job,
		}, err
	}

	var output bytes.Buffer

	limitedOutput := &limitedWriter{
		writer: &output,
		limit:  s.maxOutputSize,
	}

	if _, err := io.Copy(limitedOutput, outputFile); err != nil {
		job.Status = StatusFailed
		job.Error = err.Error()

		return CompressPDFResponse{
			Job: job,
		}, fmt.Errorf("read compressed output: %w", err)
	}

	if err := compressCtx.Err(); err != nil {
		job.Status = StatusFailed
		job.Error = err.Error()

		return CompressPDFResponse{Job: job}, err
	}

	completedAt := time.Now()
	job.CompletedAt = &completedAt
	job.Status = StatusCompleted

	return CompressPDFResponse{
		Job:      job,
		Filename: "compressed.pdf",
		Output:   output.Bytes(),
	}, nil
}

func (s *Service) CompressImage(
	ctx context.Context,
	request CompressImageRequest,
) (CompressImageResponse, error) {
	if request.Input == nil {
		return CompressImageResponse{}, fmt.Errorf(
			"image input is required",
		)
	}

	if request.JobID == "" {
		return CompressImageResponse{}, fmt.Errorf(
			"job ID is required",
		)
	}

	compressCtx, cancel := context.WithTimeout(
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

	workspace, err := CreateWorkspace(
		s.workspaceRoot,
		job.ID,
	)
	if err != nil {
		return CompressImageResponse{}, fmt.Errorf(
			"create workspace: %w",
			err,
		)
	}
	defer func() {
		_ = workspace.Cleanup()
	}()

	filename := safeFilename(request.Filename)

	if filename == "." || filename == "" {
		filename = "input"
	}

	inputPath := filepath.Join(
		workspace.Input,
		filename,
	)

	if err := writeInputLimited(
		compressCtx,
		inputPath,
		request.Input,
		s.maxInputSize,
	); err != nil {
		job.Status = StatusFailed
		job.Error = err.Error()

		return CompressImageResponse{
			Job: job,
		}, err
	}

	outputPath := filepath.Join(
		workspace.Output,
		"compressed"+filepath.Ext(filename),
	)

	job.Status = StatusProcessing

	startedAt := time.Now()
	job.StartedAt = &startedAt

	if err := image.Compress(
		compressCtx,
		inputPath,
		outputPath,
	); err != nil {
		job.Status = StatusFailed
		job.Error = err.Error()

		return CompressImageResponse{
			Job: job,
		}, fmt.Errorf(
			"compress image: %w",
			err,
		)
	}

	outputFile, err := os.Open(outputPath)
	if err != nil {
		job.Status = StatusFailed
		job.Error = err.Error()

		return CompressImageResponse{
			Job: job,
		}, fmt.Errorf(
			"open compressed image: %w",
			err,
		)
	}
	defer func() {
		_ = outputFile.Close()
	}()

	info, err := outputFile.Stat()
	if err != nil {
		job.Status = StatusFailed
		job.Error = err.Error()

		return CompressImageResponse{
			Job: job,
		}, fmt.Errorf(
			"stat compressed image: %w",
			err,
		)
	}

	if info.Size() == 0 {
		err := fmt.Errorf(
			"compression produced empty output",
		)

		job.Status = StatusFailed
		job.Error = err.Error()

		return CompressImageResponse{
			Job: job,
		}, err
	}

	var output bytes.Buffer

	limitedOutput := &limitedWriter{
		writer: &output,
		limit:  s.maxOutputSize,
	}

	if _, err := io.Copy(
		limitedOutput,
		outputFile,
	); err != nil {
		job.Status = StatusFailed
		job.Error = err.Error()

		return CompressImageResponse{
			Job: job,
		}, fmt.Errorf(
			"read compressed output: %w",
			err,
		)
	}

	if err := compressCtx.Err(); err != nil {
		job.Status = StatusFailed
		job.Error = err.Error()

		return CompressImageResponse{Job: job}, err
	}

	completedAt := time.Now()
	job.CompletedAt = &completedAt
	job.Status = StatusCompleted

	return CompressImageResponse{
		Job:      job,
		Filename: "compressed" + filepath.Ext(filename),
		Output:   output.Bytes(),
	}, nil
}

func (s *Service) CompressDOCX(
	ctx context.Context,
	req CompressDOCXRequest,
) (CompressDOCXResponse, error) {
	var response CompressDOCXResponse

	if req.Input == nil {
		return response, fmt.Errorf("DOCX input is required")
	}

	if req.JobID == "" {
		return response, fmt.Errorf("job ID is required")
	}

	ctx, cancel := context.WithTimeout(
		ctx,
		s.conversionTimeout,
	)
	defer cancel()

	job := Job{
		ID:     req.JobID,
		Status: StatusPending,
	}

	workspace, err := CreateWorkspace(
		s.workspaceRoot,
		job.ID,
	)
	if err != nil {
		return response, fmt.Errorf(
			"create DOCX workspace: %w",
			err,
		)
	}
	defer func() {
		_ = workspace.Cleanup()
	}()

	inputName := safeFilename(req.Filename)

	inputPath := filepath.Join(
		workspace.Input,
		inputName,
	)

	outputPath := filepath.Join(
		workspace.Output,
		"compressed.docx",
	)

	if err := writeInputLimited(
		ctx,
		inputPath,
		req.Input,
		s.maxInputSize,
	); err != nil {
		return response, fmt.Errorf(
			"write DOCX input: %w",
			err,
		)
	}

	if err := docx.Compress(
		ctx,
		inputPath,
		outputPath,
	); err != nil {
		return response, err
	}

	outputFile, err := os.Open(outputPath)
	if err != nil {
		return response, fmt.Errorf(
			"open compressed DOCX: %w",
			err,
		)
	}
	defer func() {
		_ = outputFile.Close()
	}()

	info, err := outputFile.Stat()
	if err != nil {
		return response, fmt.Errorf(
			"stat compressed DOCX: %w",
			err,
		)
	}

	if info.Size() == 0 {
		return response, fmt.Errorf(
			"compressed DOCX is empty",
		)
	}

	var output bytes.Buffer

	writer := limitedWriter{
		writer: &output,
		limit:  s.maxOutputSize,
	}

	if _, err := io.Copy(
		&writer,
		outputFile,
	); err != nil {
		return response, fmt.Errorf(
			"read compressed DOCX: %w",
			err,
		)
	}

	if err := ctx.Err(); err != nil {
		return response, err
	}

	job.Status = StatusCompleted

	response = CompressDOCXResponse{
		Job:      job,
		Filename: "compressed.docx",
		Output:   output.Bytes(),
	}

	return response, nil
}

func writeInputLimited(
	ctx context.Context,
	path string,
	input io.Reader,
	maxSize int64,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create input file: %w", err)
	}
	defer file.Close()

	reader := io.LimitReader(
		&contextReader{
			ctx:    ctx,
			reader: input,
		},
		maxSize+1,
	)

	written, err := io.Copy(file, reader)
	if err != nil {
		_ = file.Close()
		_ = os.Remove(path)

		if ctx.Err() != nil {
			return ctx.Err()
		}

		return fmt.Errorf("write input file: %w", err)
	}

	if written > maxSize {
		_ = file.Close()
		_ = os.Remove(path)

		return fmt.Errorf(
			"input file exceeds maximum size of %d bytes",
			maxSize,
		)
	}

	if err := ctx.Err(); err != nil {
		_ = file.Close()
		_ = os.Remove(path)

		return err

	}

	return nil
}

type contextReader struct {
	ctx    context.Context
	reader io.Reader
}

func (r *contextReader) Read(p []byte) (int, error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}

	return r.reader.Read(p)
}

func safeFilename(name string) string {
	name = strings.ReplaceAll(name, `\\`, "/")
	name = filepath.Base(name)

	if name == "." || name == ".." || name == "" {
		return "input"
	}

	return name
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
		n, err := w.writer.Write(p[:remaining])
		w.written += int64(n)

		if err != nil {
			return n, fmt.Errorf("write output: %w", err)
		}

		return n, fmt.Errorf(
			"output exceeds maximum size of %d bytes",
			w.limit,
		)
	}

	n, err := w.writer.Write(p)
	w.written += int64(n)

	return n, err
}
