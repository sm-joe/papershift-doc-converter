PaperShift Deployment
Services
Service	Container	Port	Image
Web	`papershift-web`	3000	`ghcr.io/sm-joe/papershift-doc-converter-web`
API	`papershift`	8080	`ghcr.io/sm-joe/papershift-doc-converter`
Prerequisites
Docker Engine
Docker Compose v2
Access to the PaperShift GHCR packages
Authenticate to GHCR if the packages are private.
Start
The repository uses one Compose definition for the application stack:
```powershell
docker compose -f docker/compose.yml pull
docker compose -f docker/compose.yml up -d
```
Check status:
```powershell
docker compose -f docker/compose.yml ps
```
Open the application:
```text
http://localhost:3000
```
Check API health:
```powershell
curl.exe http://localhost:8080/health
```
Check readiness:
```powershell
curl.exe http://localhost:8080/ready
```
Logs
All services:
```powershell
docker compose -f docker/compose.yml logs -f
```
API only:
```powershell
docker compose -f docker/compose.yml logs -f papershift
```
Web only:
```powershell
docker compose -f docker/compose.yml logs -f web
```
Stop
```powershell
docker compose -f docker/compose.yml down
```
Updating
Update the pinned image version in the Compose configuration, then:
```powershell
docker compose -f docker/compose.yml pull
docker compose -f docker/compose.yml up -d
```
Both API and Web should use the same release version.
Release images
Each release publishes both:
```text
ghcr.io/sm-joe/papershift-doc-converter:<version>
ghcr.io/sm-joe/papershift-doc-converter-web:<version>
```
The release workflow also publishes `latest`.
For reproducible deployments, prefer a pinned version.
Runtime hardening
The Compose deployment uses read-only filesystems, dropped capabilities, `no-new-privileges`, tmpfs for temporary data, and resource limits.
Troubleshooting
Check service state:
```powershell
docker compose -f docker/compose.yml ps
```
Check logs:
```powershell
docker compose -f docker/compose.yml logs --tail=100
```
Test API:
```powershell
curl.exe http://localhost:8080/health
```
Test Web:
```powershell
curl.exe -I http://localhost:3000
```
The browser-side API URL is configured at Next.js build time. The current local configuration uses `http://localhost:8080`; a Docker-internal hostname such as `papershift:8080` must not be used for browser requests.