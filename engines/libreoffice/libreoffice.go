package libreoffice

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/sm-joe/papershift-doc-converter/internal/converter"
	"github.com/sm-joe/papershift-doc-converter/internal/formats"
)

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
	return "libreoffice"
}

func (c *Converter) Supports(input, output formats.Format) bool {
	if output.ID != "pdf" {
		return false
	}

	switch input.ID {
	case "doc", "docx", "odt", "rtf", "txt", "html", "md":
		return true
	default:
		return false
	}
}

func (c *Converter) Convert(ctx context.Context, job converter.Job) (converter.Result, error) {
	if !c.Supports(job.InputFormat, job.OutputFormat) {
		return converter.Result{}, converter.ErrNoConverter
	}

	baseDir := c.workspace
	if baseDir == "" {
		baseDir = os.TempDir()
	}

	workDir, err := os.MkdirTemp(baseDir, "papershift-lo-*")
	if err != nil {
		return converter.Result{}, fmt.Errorf("create libreoffice workspace: %w", err)
	}
	defer os.RemoveAll(workDir)

	inputPath := filepath.Join(
		workDir,
		"input"+job.InputFormat.Extension,
	)

	outputDir := filepath.Join(workDir, "output")
	profileDir := filepath.Join(workDir, "profile")
	tempDir := filepath.Join(workDir, "tmp")
	cacheDir := filepath.Join(workDir, "cache")
	configDir := filepath.Join(workDir, "config")
	dataDir := filepath.Join(workDir, "data")

	for _, dir := range []string{
		outputDir,
		profileDir,
		tempDir,
		cacheDir,
		configDir,
		dataDir,
	} {
		if err := os.MkdirAll(dir, 0700); err != nil {
			return converter.Result{}, fmt.Errorf("create libreoffice directory: %w", err)
		}
	}

	if err := writeInput(inputPath, job.Input); err != nil {
		return converter.Result{}, err
	}

	userInstallation := "file://" + filepath.ToSlash(profileDir)

	cmd := exec.CommandContext(
		ctx,
		c.binary,
		"--headless",
		"--invisible",
		"--nologo",
		"--nodefault",
		"--nofirststartwizard",
		"--nolockcheck",
		"--norestore",
		"-env:UserInstallation="+userInstallation,
		"--convert-to",
		strings.TrimPrefix(job.OutputFormat.Extension, "."),
		"--outdir",
		outputDir,
		inputPath,
	)

	cmd.Dir = workDir

	cmd.Env = replaceEnvironment(os.Environ(), map[string]string{
		"HOME":            workDir,
		"TMPDIR":          tempDir,
		"XDG_CACHE_HOME":  cacheDir,
		"XDG_CONFIG_HOME": configDir,
		"XDG_DATA_HOME":   dataDir,
	})

	combinedOutput, err := cmd.CombinedOutput()
	if err != nil {
		return converter.Result{}, fmt.Errorf(
			"libreoffice conversion failed: %w: %s",
			err,
			strings.TrimSpace(string(combinedOutput)),
		)
	}

	outputPath := filepath.Join(
		outputDir,
		replaceExtension(filepath.Base(inputPath), job.OutputFormat.Extension),
	)

	if _, err := os.Stat(outputPath); err != nil {
		return converter.Result{}, fmt.Errorf(
			"libreoffice output missing: %w",
			err,
		)
	}

	if err := copyFile(outputPath, job.Output); err != nil {
		return converter.Result{}, err
	}

	return converter.Result{
		InputFormat:  job.InputFormat,
		OutputFormat: job.OutputFormat,
	}, nil
}

func replaceEnvironment(
	environment []string,
	replacements map[string]string,
) []string {
	result := make([]string, 0, len(environment)+len(replacements))

	for _, entry := range environment {
		key, _, found := strings.Cut(entry, "=")
		if found {
			if _, replace := replacements[key]; replace {
				continue
			}
		}

		result = append(result, entry)
	}

	for key, value := range replacements {
		result = append(result, key+"="+value)
	}

	return result
}
