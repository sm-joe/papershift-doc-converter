package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"testing"
)

const apiURL = "http://localhost:8080"

const samplePDF = `%PDF-1.4
1 0 obj
<< /Type /Catalog /Pages 2 0 R >>
endobj
2 0 obj
<< /Type /Pages /Kids [3 0 R] /Count 1 >>
endobj
3 0 obj
<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Contents 4 0 R /Resources << /Font << /F1 5 0 R >> >> >>
endobj
4 0 obj
<< /Length 44 >>
stream
BT
/F1 24 Tf
72 720 Td
(PaperShift Integration Test) Tj
ET
endstream
endobj
5 0 obj
<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>
endobj
xref
0 6
0000000000 65535 f 
0000000009 00000 n 
0000000058 00000 n 
0000000115 00000 n 
0000000279 00000 n 
0000000373 00000 n 
trailer
<< /Size 6 /Root 1 0 R >>
startxref
453
%%EOF
`

type conversionResponse struct {
	ID       string `json:"id"`
	Status   string `json:"status"`
	Filename string `json:"filename"`
	Output   string `json:"output"`
}

func TestHealth(t *testing.T) {
	response, err := http.Get(apiURL + "/health")
	if err != nil {
		t.Fatalf("health request failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf("expected health status 200, got %d", response.StatusCode)
	}
}

func TestPDFConversions(t *testing.T) {
	tests := []struct {
		name       string
		output     string
		check      func(t *testing.T, data []byte)
	}{
		{
			name:   "PDF to TXT",
			output: "txt",
			check: func(t *testing.T, data []byte) {
				if !strings.Contains(string(data), "PaperShift Integration Test") {
					t.Fatalf("TXT output does not contain expected text")
				}
			},
		},
		{
			name:   "PDF to HTML",
			output: "html",
			check: func(t *testing.T, data []byte) {
				if len(data) == 0 {
					t.Fatal("HTML output is empty")
				}
			},
		},
		{
			name:   "PDF to DOCX",
			output: "docx",
			check: func(t *testing.T, data []byte) {
				assertZIP(t, data)
			},
		},
		{
			name:   "PDF to ODT",
			output: "odt",
			check: func(t *testing.T, data []byte) {
				assertZIP(t, data)
			},
		},
		{
			name:   "PDF to PNG",
			output: "png",
			check: func(t *testing.T, data []byte) {
				assertSignature(
					t,
					data,
					[]byte{0x89, 0x50, 0x4e, 0x47},
				)
			},
		},
		{
			name:   "PDF to JPG",
			output: "jpg",
			check: func(t *testing.T, data []byte) {
				assertSignature(
					t,
					data,
					[]byte{0xff, 0xd8, 0xff},
				)
			},
		},
		{
			name:   "PDF to WebP",
			output: "webp",
			check: func(t *testing.T, data []byte) {
				assertSignature(
					t,
					data,
					[]byte{'R', 'I', 'F', 'F'},
				)

				if len(data) < 12 ||
					string(data[8:12]) != "WEBP" {
					t.Fatal("output is not a valid WebP container")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := convertPDF(t, tt.output)

			if output.Status != "completed" {
				t.Fatalf(
					"conversion status = %q, want completed",
					output.Status,
				)
			}

			if output.ID == "" {
				t.Fatal("conversion response did not contain a job ID")
			}

			data := downloadOutput(t, output.Output)

			if len(data) == 0 {
				t.Fatal("conversion output is empty")
			}

			tt.check(t, data)
		})
	}
}

func convertPDF(t *testing.T, outputFormat string) conversionResponse {
	t.Helper()

	var body bytes.Buffer

	writer := multipart.NewWriter(&body)

	part, err := writer.CreateFormFile(
		"file",
		"sample.pdf",
	)
	if err != nil {
		t.Fatalf("create multipart file: %v", err)
	}

	if _, err := part.Write([]byte(samplePDF)); err != nil {
		t.Fatalf("write PDF fixture: %v", err)
	}

	if err := writer.WriteField(
		"output_format",
		outputFormat,
	); err != nil {
		t.Fatalf("write output format: %v", err)
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	request, err := http.NewRequest(
		http.MethodPost,
		apiURL+"/api/v1/conversions",
		&body,
	)
	if err != nil {
		t.Fatalf("create conversion request: %v", err)
	}

	request.Header.Set(
		"Content-Type",
		writer.FormDataContentType(),
	)

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatalf("conversion request failed: %v", err)
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read conversion response: %v", err)
	}

	if response.StatusCode != http.StatusAccepted &&
		response.StatusCode != http.StatusOK {
		t.Fatalf(
			"conversion returned HTTP %d: %s",
			response.StatusCode,
			string(responseBody),
		)
	}

	var result conversionResponse

	if err := json.Unmarshal(responseBody, &result); err != nil {
		t.Fatalf(
			"decode conversion response: %v: %s",
			err,
			string(responseBody),
		)
	}

	return result
}

func downloadOutput(t *testing.T, outputPath string) []byte {
	t.Helper()

	response, err := http.Get(apiURL + outputPath)
	if err != nil {
		t.Fatalf("download request failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(response.Body)

		t.Fatalf(
			"download returned HTTP %d: %s",
			response.StatusCode,
			string(body),
		)
	}

	data, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read download: %v", err)
	}

	return data
}

func assertZIP(t *testing.T, data []byte) {
	t.Helper()

	if len(data) < 4 ||
		data[0] != 'P' ||
		data[1] != 'K' ||
		data[2] != 0x03 ||
		data[3] != 0x04 {
		t.Fatal("output is not a ZIP-based document")
	}
}

func assertSignature(
	t *testing.T,
	data []byte,
	signature []byte,
) {
	t.Helper()

	if len(data) < len(signature) {
		t.Fatalf(
			"output too small: got %d bytes",
			len(data),
		)
	}

	for index, expected := range signature {
		if data[index] != expected {
			t.Fatalf(
				"invalid output signature at byte %d: got 0x%x, want 0x%x",
				index,
				data[index],
				expected,
			)
		}
	}
}

func TestMain(m *testing.M) {
	fmt.Println("PaperShift integration tests")
	os.Exit(m.Run())
}