PaperShift Development
Prerequisites
Go 1.27+
Node.js 24+
npm
Docker
Docker Compose v2
API
From the repository root:
```powershell
go run ./apps/api
```
Default address:
```text
:8080
```
Web
```powershell
cd apps/web
npm ci
npm run dev
```
The development UI runs on:
```text
http://localhost:3000
```
The current browser API endpoint is:
```text
http://localhost:8080
```
Validation
```powershell
go vet ./apps/... ./engines/... ./internal/...
go test ./apps/... ./engines/... ./internal/...
```
Frontend:
```powershell
cd apps/web
npm run lint
npm run build
```
Integration:
```powershell
go test ./tests/integration -v
```
Docker
API image:
```powershell
docker build -t papershift-api:local .
```
Web image:
```powershell
docker build -f apps/web/Dockerfile -t papershift-web:local apps/web
```
Compose
Validate:
```powershell
docker compose -f docker/compose.yml config
```
Run:
```powershell
docker compose -f docker/compose.yml pull
docker compose -f docker/compose.yml up -d
```
Stop:
```powershell
docker compose -f docker/compose.yml down
```
Adding a format
Register the format.
Update detection if required.
Add or extend an engine.
Add unit tests.
Add an integration fixture where appropriate.
Update public documentation.
Run the full validation suite.
Do not advertise a conversion pair until the engine supports and tests it.
Adding an engine
Implement the converter interface and register the engine in the application.
Engines should:
declare supported pairs
use context-aware process execution
operate inside isolated workspaces
validate output
propagate meaningful errors
honor cancellation and timeouts
avoid writing outside the assigned workspace
Release
Use semantic version tags:
```text
vMAJOR.MINOR.PATCH
```
Example:
```powershell
git tag -a v1.0.0 -m "PaperShift v1.0.0"
git push origin v1.0.0
```
The release workflow builds, scans, and publishes both API and Web images.