# Contributing to PaperShift

Thank you for your interest in contributing to PaperShift.

PaperShift is a Docker-first, security-focused document conversion platform. Contributions should preserve the project's reliability, security, portability, and developer experience.

## Before You Start

Please read:

- `README.md`
- `docs/ARCHITECTURE.md`
- `docs/DEVELOPMENT.md`
- `docs/API.md`
- `docs/CONVERSION-ENGINES.md`
- `SECURITY.md`
- `CODE_OF_CONDUCT.md`

For larger changes, open an issue first so the proposed direction can be discussed before significant implementation work begins.

## Development Principles

Contributions should follow these principles:

- Keep changes focused and reviewable.
- Prefer small, isolated commits.
- Avoid unnecessary architectural changes.
- Preserve existing API behavior unless a breaking change is intentional and documented.
- Treat uploaded files as untrusted input.
- Keep conversion engines isolated behind the converter abstraction.
- Avoid introducing unnecessary runtime dependencies.
- Prefer standard library functionality where practical.
- Keep containers minimal and hardened.
- Maintain compatibility with the supported Docker-based deployment model.

## Repository Structure

```text
papershift-doc-converter/
├── apps/
│   ├── api/
│   └── web/
├── engines/
├── internal/
├── tests/
├── docker/
├── docs/
└── .github/
    └── workflows/
```

See `docs/ARCHITECTURE.md` for the detailed architecture.

## Local Development

### API

From the repository root:

```bash
go test ./apps/... ./engines/... ./internal/...
go vet ./apps/... ./engines/... ./internal/...
```

Run the API locally with the required conversion dependencies installed on the development machine.

### Web

From `apps/web`:

```bash
npm ci
npm run lint
npm run build
```

The development web application uses:

```text
NEXT_PUBLIC_API_URL=http://localhost:8080
```

## Testing

Before submitting a change, run the relevant tests locally.

At minimum, changes affecting the backend should pass:

```bash
go vet ./apps/... ./engines/... ./internal/...
go test ./apps/... ./engines/... ./internal/...
```

Frontend changes should pass:

```bash
npm run lint
npm run build
```

Changes affecting conversion behavior should also include or update integration coverage where appropriate.

The CI pipeline performs backend validation, frontend validation, Docker image building, and container vulnerability scanning.

## Adding a New Format

When adding a new format:

1. Add the format definition to the format registry.
2. Define its MIME types and extension.
3. Assign the appropriate format family.
4. Update format detection if required.
5. Add converter support through the converter abstraction.
6. Add integration coverage for supported conversions.
7. Update the API/capability documentation where necessary.
8. Update `docs/CONVERSION-ENGINES.md` if a new engine or capability is introduced.

Do not bypass the converter registry with format-specific logic directly inside HTTP handlers.

## Adding or Changing Conversion Engines

Conversion engines execute against untrusted input and must be treated as security-sensitive components.

Changes should consider:

- Process execution boundaries.
- Context cancellation and timeouts.
- Temporary workspace isolation.
- Input and output limits.
- Exit-code handling.
- Output validation.
- Cleanup on success and failure.
- Dependency availability inside the Docker image.

See `docs/CONVERSION-ENGINES.md` and `SECURITY.md`.

## Frontend Contributions

Keep the existing PaperShift visual language consistent:

- Warm editorial/paper aesthetic.
- Clear conversion workflow.
- Strong typography hierarchy.
- Minimal unnecessary UI complexity.
- Responsive behavior.
- Accessible controls and meaningful labels.

Avoid introducing unrelated UI redesigns as part of functional changes.

## Docker Changes

When modifying Dockerfiles or Compose configuration:

- Preserve non-root execution.
- Preserve read-only filesystem hardening where applicable.
- Do not add unnecessary Linux capabilities.
- Keep `no-new-privileges` enabled.
- Keep resource limits appropriate.
- Avoid unnecessary packages.
- Run the relevant image security scans.

## Pull Requests

A pull request should clearly describe:

- What changed.
- Why the change was needed.
- Any user-visible behavior changes.
- Any API changes.
- Any security implications.
- How the change was tested.

Keep pull requests focused. Avoid mixing unrelated refactoring with feature or bug-fix changes.

## Commit Messages

Use concise, descriptive commit messages.

Examples:

```text
feat: add SVG conversion support
fix: enforce conversion output limits
docs: update deployment guide
test: add PDF conversion coverage
refactor: simplify converter registry
security: harden conversion workspace handling
```

## Security Issues

Do not disclose security vulnerabilities through public GitHub issues.

Follow the reporting guidance in `SECURITY.md`.

## Documentation

Documentation is part of the project.

If a change affects:

- API behavior
- Supported formats
- Deployment
- Architecture
- Security controls
- Development workflows
- Conversion engines

update the relevant documentation as part of the same change.

## License

By contributing to PaperShift, you agree that your contributions will be licensed under the same license as the project.