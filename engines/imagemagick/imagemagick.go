package imagemagick

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

var ErrInvalidOutput = errors.New("imagemagick produced no output")

type Converter struct {
	binary    string
	workspace string
}

func New(binary string) *Converter {
	return NewWithWorkspace(binary, "")
}

func NewWithWorkspace(binary, workspace string) *Converter {
	return &Converter{
		binary:    binary,
		workspace: workspace,
	}
}

func (c *Converter) Name() string {
	return "imagemagick"
}

func (c *Converter) Supports(input, output formats.Format) bool {
	if input.Family != formats.FamilyImage || output.Family != formats.FamilyImage {
		return false
	}

	return supportedFormat(input.ID) && supportedFormat(output.ID)
}

func supportedFormat(id string) bool {
	switch id {
	case "png", "jpg", "webp", "gif", "tiff", "bmp", "svg", "avif":
		return true
	default:
		return false
	}
}

func (c *Converter) Convert(ctx context.Context, job converter.Job) (converter.Result, error) {
	workDir, err := os.MkdirTemp(c.workspace, "papershift-imagemagick-*")
	if err != nil {
		return converter.Result{}, fmt.Errorf("create workspace: %w", err)
	}
	defer os.RemoveAll(workDir)

	inputPath := filepath.Join(
		workDir,
		"input"+job.InputFormat.Extension,
	)

	outputPath := filepath.Join(
		workDir,
		"output"+job.OutputFormat.Extension,
	)

	if err := writeInput(inputPath, job.Input); err != nil {
		return converter.Result{}, err
	}

	cmd := exec.CommandContext(
		ctx,
		c.binary,
		inputPath,
		outputPath,
	)

	cmd.Dir = workDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return converter.Result{}, fmt.Errorf(
			"imagemagick conversion failed: %w: %s",
			err,
			strings.TrimSpace(string(output)),
		)
	}

	if err := copyOutput(outputPath, job.Output); err != nil {
		return converter.Result{}, err
	}

	return converter.Result{
		InputFormat:  job.InputFormat,
		OutputFormat: job.OutputFormat,
	}, nil
}

func writeInput(path string, input io.Reader) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create imagemagick input: %w", err)
	}

	defer file.Close()

	if _, err := io.Copy(file, input); err != nil {
		return fmt.Errorf("write imagemagick input: %w", err)
	}

	return nil
}

func copyOutput(path string, output io.Writer) error {
	file, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return ErrInvalidOutput
		}

		return fmt.Errorf("open imagemagick output: %w", err)
	}

	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return fmt.Errorf("stat imagemagick output: %w", err)
	}

	if info.Size() == 0 {
		return ErrInvalidOutput
	}

	if _, err := io.Copy(output, file); err != nil {
		return fmt.Errorf("copy imagemagick output: %w", err)
	}

	return nil
}