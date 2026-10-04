PaperShift Architecture
System
```text
                 +-------------------+
                 |      Browser      |
                 +---------+---------+
                           |
                    HTTP :3000
                           |
                 +---------v---------+
                 |    Next.js Web    |
                 +---------+---------+
                           |
                    HTTP :8080
                           |
                 +---------v---------+
                 |      Go API       |
                 +---------+---------+
                           |
                 +---------v---------+
                 |   Job Service     |
                 +---------+---------+
                           |
                 +---------v---------+
                 | Converter Registry|
                 +---------+---------+
                           |
              +------------+------------+
              |            |            |
              v            v            v
         LibreOffice  ImageMagick   PDF tools
```
Repository
```text
apps/
├── api/
└── web/

engines/
├── imagemagick/
├── libreoffice/
└── pandoc/

internal/
├── api/
├── converter/
├── detection/
├── formats/
├── jobs/
├── security/
└── storage/

tests/
├── fixtures/
└── integration/

docker/
└── compose.yml
```
API
The Go API uses the standard HTTP library and exposes health, readiness, format, capability, conversion, and download endpoints.
The current MVP does not require a database.
Detection
Input detection combines filename information, binary signatures, text detection, and OOXML ZIP structure validation.
OOXML detection checks expected package entries rather than trusting extensions alone.
Conversion registry
Engines implement a common converter interface and declare supported input/output pairs.
The registry selects an engine for a requested conversion pair.
Job workspaces
Each conversion receives an isolated workspace:
```text
<papershift-work>/
└── papershift/
    └── <job-id>/
        ├── input/
        └── output/
```
Workspaces use restrictive permissions and are cleaned after processing.
Resource controls
Application-level controls include:
conversion timeout
maximum input size
maximum output size
empty-output validation
Container-level controls include CPU, memory, PID, and file-descriptor limits.
Containers
API
`ghcr.io/sm-joe/papershift-doc-converter`
Contains the Go API and conversion dependencies.
Web
`ghcr.io/sm-joe/papershift-doc-converter-web`
Contains the Next.js standalone production runtime.
Release flow
```text
Git tag
  ↓
GitHub Actions
  ↓
Build API + Web
  ↓
Trivy scan
  ↓
Publish both images to GHCR
  ↓
Compose pulls both images
  ↓
API + Web run together
```