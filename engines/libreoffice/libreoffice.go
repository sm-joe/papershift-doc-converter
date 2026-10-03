package libreoffice

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/sm-joe/papershift-doc-converter/internal/converter"
	"github.com/sm-joe/papershift-doc-converter/internal/formats"
)

var ErrOutputMissing = errors.New("converter produced no output")

type Converter struct {
	binary string
}

func New(binary string) *Converter {
	if binary == "" {
		binary = "libreoffice"
	}

	return &Converter{
		binary: binary,
	}
}

func (c *Converter) Name() string {
	return "libreoffice"
}

func (c *Converter) Supports(
	input formats.Format,
	output formats.Format,
) bool {
	if input.Family != formats.FamilyDocument &&
		input.Family != formats.FamilySpreadsheet &&
		input.Family != formats.FamilyPresentation {
		return false
	}

	switch input.ID {
	case "doc", "docx", "odt", "rtf":
	default:
		return false
	}

	return output.ID == "pdf"
}

func (c *Converter) Convert(
	ctx context.Context,
	job converter.Job,
) (converter.Result, error) {
	workDir, err := os.MkdirTemp("", "papershift-lo-*")
	if err != nil {
		return converter.Result{}, fmt.Errorf("create workspace: %w", err)
	}
	defer os.RemoveAll(workDir)

	inputPath := filepath.Join(
		workDir,
		"input"+job.InputFormat.Extension,
	)

	outputDir := filepath.Join(workDir, "output")

	if err := os.MkdirAll(outputDir, 0700); err != nil {
		return converter.Result{}, fmt.Errorf("create output directory: %w", err)
	}

	if err := writeInput(inputPath, job.Input); err != nil {
		return converter.Result{}, err
	}

	cmd := exec.CommandContext(
		ctx,
		c.binary,
		"--headless",
		"--convert-to",
		job.OutputFormat.Extension[1:],
		"--outdir",
		outputDir,
		inputPath,
	)

	cmd.Dir = workDir

	cmd.Env = append(
		os.Environ(),
		"HOME="+workDir,
		"TMPDIR="+workDir,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		if ctx.Err() != nil {
			return converter.Result{}, fmt.Errorf(
				"libreoffice conversion cancelled or timed out: %w",
				ctx.Err(),
			)
		}

		return converter.Result{}, fmt.Errorf(
			"libreoffice conversion failed: %w: %s",
			err,
			string(output),
		)
	}

	outputPath := filepath.Join(
		outputDir,
		replaceExtension(
			filepath.Base(inputPath),
			job.OutputFormat.Extension,
		),
	)

	if _, err := os.Stat(outputPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return converter.Result{}, ErrOutputMissing
		}

		return converter.Result{}, fmt.Errorf(
			"inspect conversion output: %w",
			err,
		)
	}

	if err := copyFile(outputPath, job.Output); err != nil {
		return converter.Result{}, fmt.Errorf(
			"copy conversion output: %w",
			err,
		)
	}

	return converter.Result{
		InputFormat:  job.InputFormat,
		OutputFormat: job.OutputFormat,
	}, nil
}
