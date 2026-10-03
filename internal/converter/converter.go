package converter

import (
	"context"
	"io"

	"github.com/sm-joe/papershift-doc-converter/internal/formats"
)

type Job struct {
	InputFormat  formats.Format
	OutputFormat formats.Format
	Input        io.Reader
	Output       io.Writer
}

type Result struct {
	InputFormat  formats.Format
	OutputFormat formats.Format
}

type Converter interface {
	Name() string

	Supports(input formats.Format, output formats.Format) bool

	Convert(ctx context.Context, job Job) (Result, error)
}
