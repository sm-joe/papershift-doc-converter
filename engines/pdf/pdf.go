package pdf

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/sm-joe/papershift-doc-converter/internal/converter"
	"github.com/sm-joe/papershift-doc-converter/internal/formats"
)

var ErrInvalidOutput = errors.New("pdf converter produced no output")

type Converter struct {
	pdftotext string
	pdftohtml string
	pandoc    string
	workspace string
}

func New(pdftotext, pdftohtml, pandoc string) *Converter {
	return NewWithWorkspace(pdftotext, pdftohtml, pandoc, "")
}

func NewWithWorkspace(
	pdftotext string,
	pdftohtml string,
	pandoc string,
	workspace string,
) *Converter {
	return &Converter{
		pdftotext: pdftotext,
		pdftohtml: pdftohtml,
		pandoc:    pandoc,
		workspace: workspace,
	}
}

func (c *Converter) Name() string {
	return "pdf"
}

func (c *Converter) Supports(
	input formats.Format,
	output formats.Format,
) bool {
	if input.ID != "pdf" {
		return false
	}

	switch output.ID {
	case "txt", "docx", "odt", "html":
		return true
	default:
		return false
	}
}

func (c *Converter) Convert(
	ctx context.Context,
	job converter.Job,
) (converter.Result, error) {
	workDir, err := os.MkdirTemp(
		c.workspace,
		"papershift-pdf-*",
	)
	if err != nil {
		return converter.Result{}, fmt.Errorf(
			"create workspace: %w",
			err,
		)
	}
	defer os.RemoveAll(workDir)

	inputPath := filepath.Join(workDir, "input.pdf")

	if err := writeInput(inputPath, job.Input); err != nil {
		return converter.Result{}, err
	}

	switch job.OutputFormat.ID {
	case "txt":
		return c.convertText(ctx, inputPath, job.Output, job, workDir)

	case "html":
		return c.convertHTML(ctx, inputPath, job.Output, job)

	case "docx", "odt":
		return c.convertDocument(ctx, inputPath, job.Output, job, workDir)

	default:
		return converter.Result{}, fmt.Errorf(
			"unsupported PDF output format: %s",
			job.OutputFormat.ID,
		)
	}
}

func (c *Converter) convertText(
	ctx context.Context,
	inputPath string,
	output io.Writer,
	job converter.Job,
	workDir string,
) (converter.Result, error) {
	outputPath := filepath.Join(workDir, "output.txt")

	if err := runCommand(
		ctx,
		c.pdftotext,
		"-layout",
		inputPath,
		outputPath,
	); err != nil {
		return converter.Result{}, fmt.Errorf(
			"PDF text extraction failed: %w",
			err,
		)
	}

	if err := copyOutput(outputPath, output); err != nil {
		return converter.Result{}, err
	}

	return converter.Result{
		InputFormat:  job.InputFormat,
		OutputFormat: job.OutputFormat,
	}, nil
}

func (c *Converter) convertHTML(
	ctx context.Context,
	inputPath string,
	output io.Writer,
	job converter.Job,
) (converter.Result, error) {
	cmd := exec.CommandContext(
		ctx,
		c.pdftohtml,
		"-noframes",
		"-stdout",
		inputPath,
	)

	data, err := cmd.Output()
	if err != nil {
		return converter.Result{}, fmt.Errorf(
			"PDF HTML conversion failed: %w",
			err,
		)
	}

	if len(data) == 0 {
		return converter.Result{}, ErrInvalidOutput
	}

	if _, err := output.Write(data); err != nil {
		return converter.Result{}, fmt.Errorf(
			"write HTML output: %w",
			err,
		)
	}

	return converter.Result{
		InputFormat:  job.InputFormat,
		OutputFormat: job.OutputFormat,
	}, nil
}

func (c *Converter) convertDocument(
	ctx context.Context,
	inputPath string,
	output io.Writer,
	job converter.Job,
	workDir string,
) (converter.Result, error) {
	textPath := filepath.Join(workDir, "document.txt")
	outputPath := filepath.Join(
		workDir,
		"output"+job.OutputFormat.Extension,
	)

	if err := runCommand(
		ctx,
		c.pdftotext,
		"-layout",
		inputPath,
		textPath,
	); err != nil {
		return converter.Result{}, fmt.Errorf(
			"PDF text extraction failed: %w",
			err,
		)
	}

	if err := runCommand(
		ctx,
		c.pandoc,
		textPath,
		"-f",
		"markdown",
		"-t",
		pandocTarget(job.OutputFormat.ID),
		"-o",
		outputPath,
	); err != nil {
		return converter.Result{}, fmt.Errorf(
			"PDF document conversion failed: %w",
			err,
		)
	}

	if err := copyOutput(outputPath, output); err != nil {
		return converter.Result{}, err
	}

	return converter.Result{
		InputFormat:  job.InputFormat,
		OutputFormat: job.OutputFormat,
	}, nil
}

func pandocTarget(format string) string {
	switch format {
	case "docx":
		return "docx"

	case "odt":
		return "odt"

	default:
		return format
	}
}

func runCommand(
	ctx context.Context,
	binary string,
	args ...string,
) error {
	cmd := exec.CommandContext(ctx, binary, args...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		message := strings.TrimSpace(string(output))

		if message == "" {
			return err
		}

		return fmt.Errorf(
			"%w: %s",
			err,
			message,
		)
	}

	return nil
}

func writeInput(
	path string,
	input io.Reader,
) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf(
			"create PDF input: %w",
			err,
		)
	}

	defer file.Close()

	if _, err := io.Copy(file, input); err != nil {
		return fmt.Errorf(
			"write PDF input: %w",
			err,
		)
	}

	return nil
}

func copyOutput(
	path string,
	output io.Writer,
) error {
	file, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return ErrInvalidOutput
		}

		return fmt.Errorf(
			"open PDF output: %w",
			err,
		)
	}

	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return fmt.Errorf(
			"stat PDF output: %w",
			err,
		)
	}

	if info.Size() == 0 {
		return ErrInvalidOutput
	}

	if _, err := io.Copy(output, file); err != nil {
		return fmt.Errorf(
			"copy PDF output: %w",
			err,
		)
	}

	return nil
}
