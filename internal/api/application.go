package api

import (
	"github.com/sm-joe/papershift-doc-converter/engines/imagemagick"
	"github.com/sm-joe/papershift-doc-converter/engines/libreoffice"
	"github.com/sm-joe/papershift-doc-converter/engines/pdf"
	"github.com/sm-joe/papershift-doc-converter/internal/converter"
	"github.com/sm-joe/papershift-doc-converter/internal/detection"
	"github.com/sm-joe/papershift-doc-converter/internal/formats"
	"github.com/sm-joe/papershift-doc-converter/internal/jobs"
)

type Application struct {
	Formats    []formats.Format
	Converters *converter.Registry
	Jobs       *jobs.Service
	Results    *jobs.Store
}

func NewApplication(workspaceRoot string) *Application {
	converterRegistry := converter.NewRegistry()

	converterRegistry.Register(
		libreoffice.NewWithWorkspace(
			"/usr/lib/libreoffice/program/soffice",
			workspaceRoot,
		),
	)

	converterRegistry.Register(
		imagemagick.NewWithWorkspace(
			"/usr/bin/convert",
			workspaceRoot,
		),
	)

	converterRegistry.Register(
		pdf.NewWithWorkspace(
			"/usr/bin/pdftotext",
			"/usr/bin/pdftohtml",
			"/usr/bin/pandoc",
			"/usr/bin/pdftoppm",
			"/usr/bin/convert",
			workspaceRoot,
		),
	)

	detector := &detection.Detector{}

	return &Application{
		Formats:    formats.All(),
		Converters: converterRegistry,
		Jobs: jobs.NewService(
			workspaceRoot,
			detector,
			converterRegistry,
		),
		Results: jobs.NewStore(),
	}
}
