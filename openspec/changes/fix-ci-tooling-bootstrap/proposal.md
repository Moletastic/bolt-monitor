## Why

The backend CI job invokes Go Make targets that also install JavaScript
workspaces, but it does not provision pnpm. The infrastructure job invokes a
removed `install:providers` script. Both failures block otherwise valid CI.

## What Changes

- Decouple Go test, vet, lint, and build targets from JavaScript dependency
  installation while retaining Go workspace synchronization.
- Remove the obsolete infrastructure provider-generation CI step.
- Keep dashboard and infrastructure dependency installation in their own jobs.

## Capabilities

### New Capabilities
- `ci-tooling-bootstrap`: CI jobs provision only tools and dependencies needed
  by their release gates.

### Modified Capabilities

- None.

## Impact

- `.github/workflows/ci.yml`
- `Makefile`
- GitHub Actions backend and infrastructure jobs
