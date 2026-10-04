package api

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestMergePDFHandler(t *testing.T) {
	app := NewApplication(t.TempDir())
	handler := NewHandler(app)

	pdf1 := createTestPDF(t, "First PDF")
	pdf2 := createTestPDF(t, "Second PDF")

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	for _, path := range []string{pdf1, pdf2} {
		file, err := os.Open(path)
		if err != nil {
			t.Fatalf("open test PDF: %v", err)
		}

		part, err := writer.CreateFormFile(
			"files",
			filepath.Base(path),
		)
		if err != nil {
			file.Close()
			t.Fatalf("create multipart file: %v", err)
		}

		if _, err := io.Copy(part, file); err != nil {
			file.Close()
			t.Fatalf("copy test PDF: %v", err)
		}

		file.Close()
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/pdf/merge",
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

func createTestPDF(t *testing.T, text string) string {
	t.Helper()

	path := filepath.Join(
		t.TempDir(),
		"test.pdf",
	)

	pdf := `%PDF-1.4
1 0 obj
<< /Type /Catalog /Pages 2 0 R >>
endobj
2 0 obj
<< /Type /Pages /Kids [3 0 R] /Count 1 >>
endobj
3 0 obj
<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Contents 4 0 R >>
endobj
4 0 obj
<< /Length 44 >>
stream
BT
/F1 24 Tf
100 700 Td
(` + text + `) Tj
ET
endstream
endobj
xref
0 5
0000000000 65535 f
0000000009 00000 n
0000000058 00000 n
0000000115 00000 n
0000000212 00000 n
trailer
<< /Size 5 /Root 1 0 R >>
startxref
306
%%EOF
`

	if err := os.WriteFile(path, []byte(pdf), 0600); err != nil {
		t.Fatalf("write test PDF: %v", err)
	}

	return path
}
