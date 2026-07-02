## Why

The repository has verified local commands for Go services, the Next.js dashboard, and SST infrastructure, but no GitHub Actions workflow to enforce them before changes land. Contributors can miss formatting, type, lint, or test failures until after review, and CI command logic would be easy to duplicate incorrectly across workflow YAML.

## What Changes

- Add a GitHub Actions CI workflow that runs on pull requests and pushes to `main`.
- Validate backend Go code with formatting, vet, and tests through Makefile targets.
- Validate the dashboard with frozen pnpm installs, format check, lint, typecheck, and tests.
- Validate SST infrastructure with frozen pnpm installs and typecheck.
- Keep repeated module-level Go command lists in `Makefile` rather than duplicating them in workflow YAML.

## Capabilities

### New Capabilities
- `repository-ci`: GitHub Actions CI for backend, dashboard, and infrastructure validation.

### Modified Capabilities

(None)

## Impact

- Adds `.github/workflows/ci.yml`.
- Extends root `Makefile` with CI-oriented Go targets.
- Uses existing package-manager lockfiles in `apps/dashboard` and `infra`.
- Does not deploy infrastructure or require AWS credentials.
