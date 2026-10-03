package api

import (
	"context"
	"net/http"
	"time"

	"github.com/sm-joe/papershift-doc-converter/internal/jobs"
)

func (app *Application) conversionsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseMultipartForm(100 << 20); err != nil {
		http.Error(w, "invalid multipart form", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "missing file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	outputFormat := r.FormValue("output_format")
	if outputFormat == "" {
		http.Error(w, "missing output_format", http.StatusBadRequest)
		return
	}

	jobID, err := newJobID()
	if err != nil {
		http.Error(w, "failed to create job id", http.StatusInternalServerError)
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
		http.Error(w, err.Error(), http.StatusBadRequest)
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
