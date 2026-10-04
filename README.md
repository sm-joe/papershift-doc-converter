PaperShift
PaperShift is a self-hosted, Docker-first document and file conversion service.
It provides a Next.js web application backed by a Go API and multiple conversion engines. Files are processed in ephemeral job workspaces; the current architecture does not require a database.
Architecture
```text
Browser
   |
   v
PaperShift Web :3000
   |
   v
PaperShift API :8080
   |
   +-- Converter Registry
       +-- LibreOffice
       +-- Pandoc
       +-- ImageMagick
       +-- PDF tools
```
Current release
`v1.0.0`
API image:
```text
ghcr.io/sm-joe/papershift-doc-converter:v1.0.0
```
Web image:
```text
ghcr.io/sm-joe/papershift-doc-converter-web:v1.0.0
```
Quick start
```powershell
docker compose -f docker/compose.yml pull
docker compose -f docker/compose.yml up -d
```
Open:
```text
http://localhost:3000
```
API:
```text
http://localhost:8080
```
Check services:
```powershell
docker compose -f docker/compose.yml ps
curl.exe http://localhost:8080/health
```
Stop:
```powershell
docker compose -f docker/compose.yml down
```
Testing
Backend:
```powershell
go vet ./apps/... ./engines/... ./internal/...
go test ./apps/... ./engines/... ./internal/...
```
Frontend:
```powershell
cd apps/web
npm ci
npm run lint
npm run build
```
Integration:
```powershell
go test ./tests/integration -v
```
CI/CD
GitHub Actions validates Go, the frontend, Docker builds, container integration tests, and Trivy security scans.
A semantic version tag such as `v1.0.0` triggers the release workflow, which publishes both API and Web images to GHCR.
Documentation
Deployment
Architecture
API
Development
Conversion Engines
Security
Contributing
Changelog
License
Apache License 2.0.