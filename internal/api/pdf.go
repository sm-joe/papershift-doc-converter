package api

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/sm-joe/papershift-doc-converter/internal/jobs"
)

func (app *Application) mergePDFHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	if r.Method != http.MethodPost {
		http.Error(
			w,
			"method not allowed",
			http.StatusMethodNotAllowed,
		)
		return
	}

	if err := r.ParseMultipartForm(100 << 20); err != nil {
		http.Error(
			w,
			"invalid multipart form",
			http.StatusBadRequest,
		)
		return
	}

	files := r.MultipartForm.File["files"]

	if len(files) < 2 {
		http.Error(
			w,
			"at least two PDF files are required",
			http.StatusBadRequest,
		)
		return
	}

	jobID, err := newJobID()
	if err != nil {
		http.Error(
			w,
			"failed to create job id",
			http.StatusInternalServerError,
		)
		return
	}

	inputs := make([]jobs.MergePDFFile, 0, len(files))
	opened := make([]io.Closer, 0, len(files))

	for _, header := range files {
		file, err := header.Open()
		if err != nil {
			for _, openedFile := range opened {
				_ = openedFile.Close()
			}

			http.Error(
				w,
				"failed to open uploaded file",
				http.StatusBadRequest,
			)
			return
		}

		inputs = append(inputs, jobs.MergePDFFile{
			Filename: header.Filename,
			Input:    file,
		})

		opened = append(opened, file)
	}

	for _, file := range opened {
		defer file.Close()
	}

	response, err := app.Jobs.MergePDF(
		context.Background(),
		jobs.MergePDFRequest{
			JobID: jobID,
			Files: inputs,
		},
	)
	if err != nil {
		http.Error(
			w,
			err.Error(),
			http.StatusBadRequest,
		)
		return
	}

	app.Results.Set(jobs.Result{
		Job:       response.Job,
		Output:    response.Output,
		Filename:  response.Filename,
		ExpiresAt: time.Now().Add(10 * time.Minute),
	})

	writeJSON(w, http.StatusAccepted, map[string]interface{}{
		"id":     response.Job.ID,
		"status": response.Job.Status,
		"output": "/api/v1/conversions/" + response.Job.ID + "/download",
	})
}
