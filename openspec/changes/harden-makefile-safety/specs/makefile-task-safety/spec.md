## ADDED Requirements

### Requirement: Root Make tasks are safe and discoverable
The root Makefile SHALL use Bash with fail-fast, unset-variable, and pipeline-failure handling for recipes. It SHALL remove a file target when its recipe fails. A bare `make` invocation SHALL render a help listing of documented public targets and SHALL NOT install dependencies, synchronize workspaces, build artifacts, or mutate infrastructure.

#### Scenario: Bare Make invocation
- **WHEN** a developer runs `make` from repository root
- **THEN** Make prints available documented targets
- **AND** no setup, build, test, formatting, or infrastructure recipe runs

#### Scenario: Generated multi-command check fails
- **WHEN** a command in a generated sequence for a Make validation target fails
- **THEN** the Make target exits nonzero
- **AND** later commands in that recipe do not run

#### Scenario: File-target recipe fails
- **WHEN** a recipe fails after creating its file target
- **THEN** Make removes that target file

### Requirement: Go Lambda builds use the canonical service list
The `build-go` target SHALL build and package each service in `GO_SERVICES` exactly once. It SHALL produce the existing `handler` and `function.zip` artifacts in each service directory, and SHALL stop when a build or packaging command fails. The `clean` target SHALL remove those artifacts using the same service list.

#### Scenario: Go Lambda build succeeds
- **WHEN** a developer runs `make build-go`
- **THEN** every service in `GO_SERVICES` receives a Linux ARM64 `handler` and `function.zip`

#### Scenario: Go Lambda build fails
- **WHEN** building or packaging one service fails
- **THEN** `make build-go` exits nonzero
- **AND** later services are not built or packaged

### Requirement: Infrastructure target summary is readable
Infrastructure commands SHALL present the selected target as labeled, multi-line non-secret fields. `make infra-status` SHALL include target, stage, lifecycle class, owner, service, account, region, profile, and dashboard origin without credentials.

#### Scenario: Operator inspects target
- **WHEN** an operator runs `make infra-status` with a valid selected target
- **THEN** output includes a status heading followed by one labeled target field per line

### Requirement: Targeted formatting documents supported path input
The `format-dashboard-files` and `format-infra-files` help descriptions SHALL state that `FILES` accepts whitespace-delimited paths and does not support paths containing whitespace or single quotes.

#### Scenario: Developer inspects formatting help
- **WHEN** a developer runs `make help`
- **THEN** each targeted formatting command describes its `FILES` input constraint
