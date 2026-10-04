package api

import (
	"bytes"
	"encoding/base64"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"testing"
)

func TestCompressImageHandler(t *testing.T) {
	convertPath, err := exec.LookPath("convert")
	if err != nil {
		t.Skip("ImageMagick convert is not installed")
	}

	versionOutput, err := exec.Command(
		convertPath,
		"-version",
	).CombinedOutput()
	if err != nil {
		t.Skip("ImageMagick convert is not available")
	}

	if !bytes.Contains(
		versionOutput,
		[]byte("ImageMagick"),
	) {
		t.Skip("system convert is not ImageMagick")
	}

	app := NewApplication(t.TempDir())
	handler := NewHandler(app)

	// Minimal valid 1x1 PNG.
	imageData, err := base64.StdEncoding.DecodeString(
		"iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mNk+A8AAQUBAScY42YAAAAASUVORK5CYII=",
	)
	if err != nil {
		t.Fatalf("decode test image: %v", err)
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	part, err := writer.CreateFormFile(
		"file",
		"test.png",
	)
	if err != nil {
		t.Fatalf("create multipart file: %v", err)
	}

	if _, err := io.Copy(
		part,
		bytes.NewReader(imageData),
	); err != nil {
		t.Fatalf("copy test image: %v", err)
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/image/compress",
		&body,
	)

	request.Header.Set(
		"Content-Type",
		writer.FormDataContentType(),
	)

	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusAccepted {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusAccepted,
			response.Code,
			response.Body.String(),
		)
	}
}

func TestCompressImageHandlerRejectsUnsupportedFormat(t *testing.T) {
	app := NewApplication(t.TempDir())
	handler := NewHandler(app)

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	part, err := writer.CreateFormFile(
		"file",
		"test.txt",
	)
	if err != nil {
		t.Fatalf("create multipart file: %v", err)
	}

	if _, err := io.WriteString(
		part,
		"not an image",
	); err != nil {
		t.Fatalf("write test file: %v", err)
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/image/compress",
		&body,
	)

	request.Header.Set(
		"Content-Type",
		writer.FormDataContentType(),
	)

	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusBadRequest,
			response.Code,
			response.Body.String(),
		)
	}
}
