PaperShift API
Base URL
```text
http://localhost:8080
```
Health
`GET /health`
```powershell
curl.exe http://localhost:8080/health
```
Readiness
`GET /ready`
```powershell
curl.exe http://localhost:8080/ready
```
Formats
`GET /api/v1/formats`
Returns the registered formats and their metadata.
```powershell
curl.exe http://localhost:8080/api/v1/formats
```
Capabilities
`GET /api/v1/capabilities`
Returns supported conversion pairs exposed by the registered engines.
```powershell
curl.exe http://localhost:8080/api/v1/capabilities
```
Clients should use this endpoint to determine which output formats are available.
Create conversion
`POST /api/v1/conversions`
Submit a file and requested output format as multipart form data.
Example:
```powershell
curl.exe -X POST http://localhost:8080/api/v1/conversions `
  -F "file=@sample.docx" `
  -F "output_format=pdf"
```
Download
`GET /api/v1/conversions/{id}/download`
Example:
```powershell
curl.exe -o output.pdf http://localhost:8080/api/v1/conversions/<job-id>/download
```
Failure classes
Clients should handle:
unsupported input format
unknown output format
unsupported conversion pair
invalid input
conversion failure
conversion timeout
input size exceeded
output size exceeded
empty conversion output
Recommended client flow
```text
GET /api/v1/formats
        ↓
GET /api/v1/capabilities
        ↓
Select input/output
        ↓
POST /api/v1/conversions
        ↓
Download result
```