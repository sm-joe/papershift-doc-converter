package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/sm-joe/papershift-doc-converter/internal/jobs"
)

const (
	maxUploadSize = 100 << 20 // 100 MiB
	formFileField = "file"
	outputField   = "output_format"
)

func (app *Application) conversionsHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	r.Body = http.MaxBytesReader(
		w,
		r.Body,
		maxUploadSize,
	)

	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		http.Error(
			w,
			"invalid multipart request",
			http.StatusBadRequest,
		)
		return
	}

	file, header, err := r.FormFile(formFileField)
	if err != nil {
		http.Error(
			w,
			"file is required",
			http.StatusBadRequest,
		)
		return
	}
	defer file.Close()

	outputFormat := r.FormValue(outputField)

	if outputFormat == "" {
		http.Error(
			w,
			"output_format is required",
			http.StatusBadRequest,
		)
		return
	}

	jobID, err := newJobID()
	if err != nil {
		http.Error(
			w,
			"failed to create job",
			http.StatusInternalServerError,
		)
		return
	}

	response, err := app.Jobs.Convert(
		context.Background(),
		jobs.ConvertRequest{
			JobID:        jobID,
			Filename:     header.Filename,
			Input:        file,
			OutputFormat: outputFormat,
		},
	)
	if err != nil {
		http.Error(
			w,
			fmt.Sprintf("conversion failed: %v", err),
			http.StatusUnprocessableEntity,
		)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"job": response.Job,
	})
}
