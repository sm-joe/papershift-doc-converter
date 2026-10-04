# Changelog

All notable changes to PaperShift are documented in this file.

The project follows [Semantic Versioning](https://semver.org/).

## [1.0.0] - 2026-10-04

### Added

- Initial production release of PaperShift.
- Docker-first document conversion platform.
- Go-based conversion API.
- Next.js web interface.
- Converter registry and format detection.
- Secure temporary workspace handling for conversion jobs.
- Conversion timeout enforcement.
- Input and output size limits.
- Output validation.
- Automatic cleanup of conversion workspaces.
- Health and readiness endpoints.
- Supported-format and capability discovery endpoints.
- LibreOffice-based document, spreadsheet, and presentation conversions.
- ImageMagick-based image conversions.
- PDF conversion engine.
- Support for PDF, DOC, DOCX, ODT, RTF, TXT, HTML, Markdown.
- Support for XLS, XLSX, ODS, CSV, TSV.
- Support for PPT, PPTX, ODP.
- Support for PNG, JPG, WebP, GIF, TIFF, BMP, SVG, AVIF.
- REST API integration tests covering supported conversion paths and negative cases.
- Hardened API container running as a non-root user.
- Hardened Web container running as a non-root user.
- Read-only container filesystems.
- Dropped Linux capabilities.
- `no-new-privileges` security configuration.
- Container resource limits.
- Trivy vulnerability scanning in CI/CD.
- GitHub Actions CI/CD workflows.
- GHCR publishing for API and Web container images.
- Versioned and `latest` container image tags.
- Production deployment through Docker Compose.
- Architecture, API, deployment, development, and conversion-engine documentation.

### Security

- Untrusted conversion input is processed inside isolated temporary workspaces.
- Conversion execution is bounded by configurable timeouts.
- Input and output sizes are explicitly limited.
- Containers run without root privileges.
- Container capabilities are dropped.
- Filesystems are mounted read-only where applicable.
- Temporary storage uses controlled temporary filesystems and dedicated work volumes.
- Container CPU, memory, process, and file-descriptor limits are configured.
- Container images are scanned for HIGH and CRITICAL vulnerabilities.
- Output files are validated before being returned to clients.

[1.0.0]: https://github.com/sm-joe/papershift-doc-converter/releases/tag/v1.0.0