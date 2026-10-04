# Security Policy

## Supported Versions

Security fixes are currently provided for the latest released version of PaperShift.

| Version | Supported |
| --- | --- |
| 1.0.x | Yes |
| < 1.0 | No |

Users should upgrade to the latest release when security fixes are published.

## Reporting a Vulnerability

Please **do not report security vulnerabilities through public GitHub issues**.

If you discover a potential security vulnerability in PaperShift, report it privately through the repository's GitHub security reporting mechanism.

When reporting a vulnerability, please include:

- A clear description of the vulnerability.
- The affected PaperShift version or commit.
- The affected component or file, if known.
- Steps required to reproduce the issue.
- A proof of concept, if available.
- The potential security impact.
- Any suggested mitigation or remediation.

Please avoid including secrets, credentials, personal information, or unnecessary sensitive data in the report.

## What to Report

Examples of security issues include:

- Remote code execution.
- Command injection.
- Path traversal.
- Arbitrary file access.
- Container escape.
- Authentication or authorization bypasses.
- Unsafe handling of uploaded files.
- Malicious file processing vulnerabilities.
- Denial-of-service vulnerabilities caused by conversion processing.
- Resource exhaustion that bypasses configured limits.
- Dependency vulnerabilities with meaningful security impact.
- Improper isolation between conversion jobs.
- Vulnerabilities that allow access to another user's conversion data.
- Security weaknesses in Docker or deployment configuration.

## File Conversion Security

PaperShift processes potentially untrusted files.

Conversion engines such as LibreOffice, ImageMagick, and PDF-processing utilities operate on user-supplied content and are therefore treated as security-sensitive components.

PaperShift uses multiple controls to reduce the impact of malicious or malformed input, including:

- Isolated temporary conversion workspaces.
- Conversion execution timeouts.
- Maximum input size limits.
- Maximum output size limits.
- Output validation.
- Automatic workspace cleanup.
- Non-root container execution.
- Read-only container filesystems where applicable.
- Dropped Linux capabilities.
- `no-new-privileges`.
- Temporary filesystems for runtime temporary data.
- CPU limits.
- Memory limits.
- Process limits.
- File-descriptor limits.
- Container image vulnerability scanning.

These controls reduce risk but do not guarantee that arbitrary uploaded files are safe.

## Docker Security

The production containers are designed to run with reduced privileges.

The deployment configuration should preserve:

- Non-root container users.
- Read-only filesystems.
- Dropped Linux capabilities.
- `no-new-privileges`.
- Restricted temporary storage.
- CPU and memory limits.
- Process limits.
- Controlled writable workspaces.

Do not disable these controls in production without understanding the security implications.

## Dependency Security

PaperShift dependencies should be kept reasonably current.

The CI/CD pipeline performs container vulnerability scanning using Trivy and fails the release when applicable unfixed HIGH or CRITICAL vulnerabilities are detected.

Dependency and base-image updates should be reviewed for:

- Security impact.
- Compatibility.
- Runtime behavior.
- Conversion-engine compatibility.

## Secure Development

Contributors should:

- Treat uploaded files as untrusted.
- Avoid executing user-controlled input directly.
- Avoid shell command construction from untrusted values.
- Preserve conversion timeouts and resource limits.
- Validate generated output.
- Clean up temporary files after processing.
- Avoid logging sensitive file contents.
- Avoid committing credentials, tokens, or secrets.
- Keep container privileges as restricted as practical.
- Run the project's tests and security checks before submitting changes.

See `CONTRIBUTING.md` for contribution guidelines.

## Disclosure Process

After receiving a valid vulnerability report, maintainers will:

1. Review and reproduce the reported issue where possible.
2. Assess its severity and security impact.
3. Determine the affected versions and components.
4. Develop and test an appropriate fix.
5. Release the fix when practical.
6. Publish security-relevant release information when appropriate.

The project may coordinate disclosure timing with the reporter when the vulnerability requires users to upgrade before public disclosure.

## Security Best Practices for Deployments

Production deployments should:

- Use the latest supported PaperShift release.
- Keep Docker and the host operating system patched.
- Restrict network exposure to the required interfaces.
- Avoid exposing the conversion API directly to the public internet without appropriate network controls.
- Use a reverse proxy or gateway where appropriate.
- Apply appropriate authentication and authorization controls when PaperShift is deployed in a multi-user environment.
- Monitor container and host resources.
- Monitor application and container logs.
- Regularly review image and dependency vulnerabilities.
- Treat uploaded documents as untrusted data.

## Scope

This policy covers the PaperShift source code, official container images, conversion services, deployment configuration, and security-sensitive project infrastructure maintained by the project.

Third-party software used by PaperShift may have its own security policies and vulnerability-reporting processes.

## Acknowledgements

We appreciate responsible security researchers and contributors who help improve the security of PaperShift through coordinated vulnerability disclosure.