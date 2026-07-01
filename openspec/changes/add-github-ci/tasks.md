## 1. Go CI targets

- [x] 1.1 Add a reusable Go module list to the root `Makefile` that covers services and shared modules under `go.work`.
- [x] 1.2 Add a Go format-check target that fails when `gofmt` would change files.
- [x] 1.3 Add a Go vet target that runs vet across repository Go modules.
- [x] 1.4 Add a Go CI aggregate target that runs format check, vet, and tests.

## 2. GitHub Actions workflow

- [x] 2.1 Add `.github/workflows/ci.yml` triggered by pull requests and pushes to `main`.
- [x] 2.2 Add a backend job that sets up Go and runs the Go CI Makefile target.
- [x] 2.3 Add a dashboard job that installs pnpm dependencies with the committed lockfile and runs format check, lint, typecheck, and tests.
- [x] 2.4 Add an infra job that installs pnpm dependencies with the committed lockfile and runs typecheck.

## 3. Verification

- [x] 3.1 Run the new Go CI target locally.
- [x] 3.2 Run dashboard validation locally with existing Makefile targets.
- [x] 3.3 Run infra validation locally with existing Makefile targets.
- [x] 3.4 Validate the workflow YAML syntax.
