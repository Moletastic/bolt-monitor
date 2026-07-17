## MODIFIED Requirements

### Requirement: Dashboard CI validates Next.js code
The repository CI SHALL validate the dashboard with reproducible dependency installation and all verified dashboard quality gates, including a production build.

#### Scenario: Dashboard job runs
- **WHEN** repository CI executes the dashboard job
- **THEN** it installs dependencies from `apps/dashboard/pnpm-lock.yaml` without lockfile mutation
- **AND** it runs dashboard format check, lint, typecheck, tests, and production build

## ADDED Requirements

### Requirement: Repository CI runs the verified release-gate command set
Before the authentication security cutover, repository CI SHALL run the repository-owned Go, dashboard, infrastructure, Bruno, API-contract, health-contract/documentation, and portable configuration validation commands required to establish release confidence.

#### Scenario: Pull request or main branch validation runs
- **WHEN** repository CI runs for a pull request or a push to `main`
- **THEN** it invokes root Makefile targets for the verified Go, dashboard, and infrastructure checks
- **AND** it runs `make check-bruno` and the API-contract drift gate
- **AND** the drift gate compares SST, OpenAPI, Bruno, and the statically enumerable monitor-handler routes
- **AND** it validates the public health envelope and corresponding documentation
- **AND** it tests portable profile/stage selection and production-stage rejection where deployment configuration is consumed
- **AND** it does not require AWS deployment credentials

#### Scenario: Authentication route-helper refactor is ready to begin
- **WHEN** the authentication change is ready to refactor SST route registration through a protected-route helper
- **THEN** the complete local release-gate command set passes
- **AND** every known handler/SST route discrepancy has been explicitly reconciled

#### Scenario: Verified command changes
- **WHEN** a repository-owned validation command is updated
- **THEN** CI delegates to the Makefile command surface instead of maintaining a divergent module or command list in workflow YAML

### Requirement: CI preserves dependency installation trust controls
Repository CI SHALL use immutable JavaScript installs and SHALL preserve the explicit install-script trust policy for each in-scope package root.

#### Scenario: CI installs dashboard or infrastructure dependencies
- **WHEN** a CI job installs dependencies for `apps/dashboard` or `infra`
- **THEN** it uses the committed pnpm lockfile without mutation
- **AND** only dependencies approved by that package root's install-script allowlist can execute install or build scripts

### Requirement: CI failures are actionable and cost-bounded
Repository CI SHALL report the failing validation surface clearly and SHALL avoid unnecessary deployments or unbounded duplicate work in ordinary pull-request and `main` validation.

#### Scenario: A release gate fails
- **WHEN** a Go, dashboard, infrastructure, Bruno, or API-contract check detects a violation
- **THEN** the workflow identifies the failed command or job
- **AND** the validator reports the affected file or normalized route and the expected correction where available

#### Scenario: Ordinary repository CI runs
- **WHEN** CI runs for a pull request or a push to `main`
- **THEN** it performs only local deterministic build and validation work
- **AND** it does not deploy SST resources
- **AND** repeated setup or validation is grouped or cached where doing so does not weaken isolation or reproducibility

### Requirement: Untrusted pull requests have no cloud credential path
Repository CI SHALL keep AWS credentials and credentialed staging operations unavailable to untrusted pull-request execution.

#### Scenario: Untrusted pull-request validation runs
- **WHEN** CI processes code from an untrusted pull request
- **THEN** no job receives AWS credentials or staging API tokens
- **AND** no job deploys, updates, or removes an SST stage
- **AND** no production deployment automation is reachable from that event
