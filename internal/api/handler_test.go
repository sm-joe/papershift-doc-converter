package api

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/sm-joe/papershift-doc-converter/internal/converter"
	"github.com/sm-joe/papershift-doc-converter/internal/detection"
	"github.com/sm-joe/papershift-doc-converter/internal/formats"
	"github.com/sm-joe/papershift-doc-converter/internal/jobs"
)

type integrationConverter struct{}

func (integrationConverter) Name() string {
	return "integration-test"
}

func (integrationConverter) Supports(input formats.Format, output formats.Format) bool {
	return input.ID == "txt" && output.ID == "txt"
}

func (integrationConverter) Convert(
	_ context.Context,
	job converter.Job,
) (converter.Result, error) {
	data, err := io.ReadAll(job.Input)
	if err != nil {
		return converter.Result{}, err
	}

	if _, err := job.Output.Write(bytes.ToUpper(data)); err != nil {
		return converter.Result{}, err
	}

	return converter.Result{
		InputFormat:  job.InputFormat,
		OutputFormat: job.OutputFormat,
	}, nil
}

func newIntegrationApplication(t *testing.T) *Application {
	t.Helper()

	registry := converter.NewRegistry()
	registry.Register(integrationConverter{})

	return &Application{
		Formats:    formats.All(),
		Converters: registry,
		Jobs: jobs.NewService(
			t.TempDir(),
			&detection.Detector{},
			registry,
		),
		Results: jobs.NewStore(),
	}
}

func TestConversionDownloadLifecycle(t *testing.T) {
	app := newIntegrationApplication(t)
	server := httptest.NewServer(NewHandler(app))
	defer server.Close()

	var body bytes.Buffer

	writer := multipart.NewWriter(&body)

	fileWriter, err := writer.CreateFormFile("file", "example.txt")
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}

	if _, err := fileWriter.Write([]byte("hello papershift")); err != nil {
		t.Fatalf("write input: %v", err)
	}

	if err := writer.WriteField("output_format", "txt"); err != nil {
		t.Fatalf("write output format: %v", err)
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	request, err := http.NewRequest(
		http.MethodPost,
		server.URL+"/api/v1/conversions",
		&body,
	)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}

	request.Header.Set("Content-Type", writer.FormDataContentType())

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatalf("conversion request failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusAccepted {
		data, _ := io.ReadAll(response.Body)

		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusAccepted,
			response.StatusCode,
			string(data),
		)
	}

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read conversion response: %v", err)
	}

	responseText := string(responseBody)

	if !strings.Contains(responseText, `"id"`) {
		t.Fatalf("expected job id in response: %s", responseText)
	}

	if !strings.Contains(responseText, `"status":"completed"`) {
		t.Fatalf("expected completed status: %s", responseText)
	}

	start := strings.Index(responseText, `"id":"`)
	if start == -1 {
		t.Fatalf("job id not found in response: %s", responseText)
	}

	start += len(`"id":"`)

	end := strings.Index(responseText[start:], `"`)
	if end == -1 {
		t.Fatalf("job id terminator not found: %s", responseText)
	}

	jobID := responseText[start : start+end]

	downloadURL := server.URL +
		"/api/v1/conversions/" +
		jobID +
		"/download"

	downloadResponse, err := http.Get(downloadURL)
	if err != nil {
		t.Fatalf("download request failed: %v", err)
	}
	defer downloadResponse.Body.Close()

	if downloadResponse.StatusCode != http.StatusOK {
		t.Fatalf(
			"expected download status %d, got %d",
			http.StatusOK,
			downloadResponse.StatusCode,
		)
	}

	downloaded, err := io.ReadAll(downloadResponse.Body)
	if err != nil {
		t.Fatalf("read downloaded result: %v", err)
	}

	if string(downloaded) != "HELLO PAPERSHIFT" {
		t.Fatalf(
			"unexpected downloaded content: %q",
			string(downloaded),
		)
	}

	contentDisposition := downloadResponse.Header.Get("Content-Disposition")

	if !strings.Contains(contentDisposition, "example.txt") {
		t.Fatalf(
			"expected output filename in Content-Disposition, got %q",
			contentDisposition,
		)
	}
}

func TestDownloadMissingResult(t *testing.T) {
	app := newIntegrationApplication(t)
	server := httptest.NewServer(NewHandler(app))
	defer server.Close()

	response, err := http.Get(
		server.URL + "/api/v1/conversions/nonexistent/download",
	)
	if err != nil {
		t.Fatalf("download request failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusNotFound {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusNotFound,
			response.StatusCode,
		)
	}
}

func TestDownloadExpiredResult(t *testing.T) {
	app := newIntegrationApplication(t)
	server := httptest.NewServer(NewHandler(app))
	defer server.Close()

	app.Results.Set(jobs.Result{
		Job: jobs.Job{
			ID:     "expired-job",
			Status: jobs.StatusCompleted,
			OutputFormat: formats.Format{
				ID:        "txt",
				Extension: "txt",
			},
		},
		Output:    []byte("expired"),
		Filename:  "expired.txt",
		ExpiresAt: time.Now().Add(-time.Minute),
	})

	response, err := http.Get(
		server.URL + "/api/v1/conversions/expired-job/download",
	)
	if err != nil {
		t.Fatalf("download request failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusNotFound {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusNotFound,
			response.StatusCode,
		)
	}
}
