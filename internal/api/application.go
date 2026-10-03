package api

import (
	"github.com/sm-joe/papershift-doc-converter/engines/libreoffice"
	"github.com/sm-joe/papershift-doc-converter/internal/converter"
	"github.com/sm-joe/papershift-doc-converter/internal/detection"
	"github.com/sm-joe/papershift-doc-converter/internal/formats"
	"github.com/sm-joe/papershift-doc-converter/internal/jobs"
)

type Application struct {
	Formats    []formats.Format
	Converters *converter.Registry
	Jobs       *jobs.Service
}

func NewApplication(workspaceRoot string) *Application {
	converterRegistry := converter.NewRegistry()

	converterRegistry.Register(
		libreoffice.New("libreoffice"),
	)

	return &Application{
		Formats:    formats.All(),
		Converters: converterRegistry,
		Jobs: jobs.NewService(
			workspaceRoot,
			detection.New(),
			converterRegistry,
		),
	}
}
