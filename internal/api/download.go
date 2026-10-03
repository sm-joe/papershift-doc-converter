package api

import (
	"net/http"
	"strings"

	"github.com/sm-joe/papershift-doc-converter/internal/jobs"
)

func (app *Application) downloadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	const prefix = "/api/v1/conversions/"
	if !strings.HasPrefix(r.URL.Path, prefix) {
		http.NotFound(w, r)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, prefix)
	const suffix = "/download"

	if !strings.HasSuffix(path, suffix) {
		http.NotFound(w, r)
		return
	}

	jobID := strings.TrimSuffix(path, suffix)
	if jobID == "" {
		http.NotFound(w, r)
		return
	}

	result, err := app.Results.Get(jobID)
	if err != nil {
		if err == jobs.ErrJobNotFound {
			http.NotFound(w, r)
			return
		}

		http.Error(w, "failed to retrieve conversion result", http.StatusInternalServerError)
		return
	}

	contentType := "application/octet-stream"

	switch result.Job.OutputFormat.ID {
	case "pdf":
		contentType = "application/pdf"
	case "txt":
		contentType = "text/plain; charset=utf-8"
	case "html":
		contentType = "text/html; charset=utf-8"
	case "md":
		contentType = "text/markdown; charset=utf-8"
	case "csv", "tsv":
		contentType = "text/csv; charset=utf-8"
	case "png":
		contentType = "image/png"
	case "jpg":
		contentType = "image/jpeg"
	case "webp":
		contentType = "image/webp"
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set(
		"Content-Disposition",
		`attachment; filename="`+strings.ReplaceAll(result.Filename, `"`, "")+`"`,
	)
	w.Header().Set("Content-Length", stringSize(len(result.Output)))

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(result.Output)
}

func stringSize(size int) string {
	if size == 0 {
		return "0"
	}

	digits := ""
	for size > 0 {
		digits = string(rune('0'+size%10)) + digits
		size /= 10
	}

	return digits
}
