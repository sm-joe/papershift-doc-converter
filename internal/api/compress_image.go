package api

import (
	"context"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/sm-joe/papershift-doc-converter/internal/jobs"
)

func (app *Application) compressImageHandler(
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
			"exactly one image file is required",
			http.StatusBadRequest,
		)
		return
	}

	filename := fileHeader[0].Filename
	extension := strings.ToLower(
		strings.TrimPrefix(
			filepath.Ext(filename),
			".",
		),
	)

	switch extension {
	case "jpg", "jpeg", "png", "webp":
	default:
		http.Error(
			w,
			"only JPG, JPEG, PNG, and WebP images are supported",
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
	defer func() {
		_ = file.Close()
	}()

	jobID, err := newJobID()
	if err != nil {
		http.Error(
			w,
			"failed to create job id",
			http.StatusInternalServerError,
		)
		return
	}

	response, err := app.Jobs.CompressImage(
		context.Background(),
		jobs.CompressImageRequest{
			JobID:    jobID,
			Filename: filename,
			Input:    io.Reader(file),
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
		"output": "/api/v1/conversions/" +
			response.Job.ID +
			"/download",
	})
}
