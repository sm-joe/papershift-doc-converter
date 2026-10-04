package api

import (
	"context"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/sm-joe/papershift-doc-converter/internal/jobs"
)

func (app *Application) rotatePDFHandler(
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

	fileHeader := r.MultipartForm.File["file"]

	if len(fileHeader) != 1 {
		http.Error(
			w,
			"exactly one PDF file is required",
			http.StatusBadRequest,
		)
		return
	}

	rotation, err := strconv.Atoi(
		r.FormValue("rotation"),
	)
	if err != nil {
		http.Error(
			w,
			"rotation must be 90, 180, or 270 degrees",
			http.StatusBadRequest,
		)
		return
	}

	file, err := fileHeader[0].Open()
	if err != nil {
		http.Error(
			w,
			"failed to open uploaded file",
			http.StatusBadRequest,
		)
		return
	}
	defer file.Close()

	jobID, err := newJobID()
	if err != nil {
		http.Error(
			w,
			"failed to create job id",
			http.StatusInternalServerError,
		)
		return
	}

	response, err := app.Jobs.RotatePDF(
		context.Background(),
		jobs.RotatePDFRequest{
			JobID:    jobID,
			Filename: fileHeader[0].Filename,
			Input:    io.Reader(file),
			Rotation: rotation,
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