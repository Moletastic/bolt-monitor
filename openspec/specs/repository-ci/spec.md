## Purpose

Define repository CI expectations for pull requests and main branch updates across backend Go modules, dashboard code, and SST infrastructure.

## Requirements

### Requirement: Repository runs pull request and main branch CI
The repository SHALL define a GitHub Actions workflow that validates code changes on pull requests and pushes to `main`.

#### Scenario: Contributor opens a pull request
- **WHEN** a contributor opens or updates a pull request
- **THEN** GitHub Actions runs repository CI checks before merge

#### Scenario: Change lands on main
- **WHEN** commits are pushed to `main`
- **THEN** GitHub Actions runs repository CI checks against the updated branch state

### Requirement: Backend CI validates Go code
The repository CI SHALL validate backend Go modules for formatting, vet diagnostics, and tests.

#### Scenario: Backend Go job runs
- **WHEN** repository CI executes the backend job
- **THEN** it checks Go formatting without modifying files
- **AND** it runs Go vet across repository service and shared modules
- **AND** it runs the repository Go test suite

### Requirement: Dashboard CI validates Next.js code
The repository CI SHALL validate the dashboard with reproducible dependency installation and existing quality gates.

#### Scenario: Dashboard job runs
- **WHEN** repository CI executes the dashboard job
- **THEN** it installs dependencies from `apps/dashboard/pnpm-lock.yaml` without lockfile mutation
- **AND** it runs dashboard format check, lint, typecheck, and tests

### Requirement: Infrastructure CI validates SST TypeScript code
The repository CI SHALL validate SST infrastructure code with reproducible dependency installation and typechecking.

#### Scenario: Infrastructure job runs
- **WHEN** repository CI executes the infrastructure job
- **THEN** it installs dependencies from `infra/pnpm-lock.yaml` without lockfile mutation
- **AND** it runs the infrastructure TypeScript typecheck
- **AND** it does not deploy infrastructure or require AWS credentials

### Requirement: Workflow delegates repeated Go command logic to Makefile
The repository CI SHALL avoid duplicating module-level Go validation lists inside GitHub Actions YAML when that logic can live in the root `Makefile`.

#### Scenario: Backend job needs module-aware validation
- **WHEN** the backend CI job runs formatting, vet, and test checks
- **THEN** the workflow invokes Makefile targets for repeated Go command logic rather than embedding the full module list in workflow YAML
