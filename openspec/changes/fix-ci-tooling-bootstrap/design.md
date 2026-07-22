## Context

`make ci-go`, `make lint-go`, and `make build-go` depend on `setup`, which
installs both JavaScript workspaces before synchronizing Go modules. The backend
CI job installs Go only. Infrastructure CI independently installs pnpm but then
calls a script removed from `infra/package.json`.

## Goals / Non-Goals

**Goals:**
- Make Go release targets depend only on Go tooling and the Go workspace.
- Keep JavaScript installation explicit in dashboard and infrastructure jobs.
- Remove CI invocation of a nonexistent script.

**Non-Goals:**
- Change pnpm versions, lockfiles, or dependency policies.
- Add SST provider generation or alter deployment behavior.

## Decisions

- Add a Go-only bootstrap target for `go work sync`, used by Go test, vet, lint,
  and build targets. This removes Node/pnpm from Go CI requirements.
- Remove `install:providers` step rather than add an empty script. No current
  release gate consumes generated provider artifacts.

## Risks / Trade-offs

- [A future Go target needs JS-generated input] → Declare that dependency on the
  specific target rather than restoring global setup coupling.
- [SST later requires provider generation] → Add a real package script and a
  consuming release-gate test in a dedicated change.
