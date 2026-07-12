## ADDED Requirements

### Requirement: Repository provides local Git hook configuration
The repository SHALL define a committed Lefthook configuration that installs local Git hooks for repository contributors.

#### Scenario: Contributor installs local hooks
- **WHEN** a contributor runs the documented hook setup command
- **THEN** Lefthook installs the repository Git hooks from committed configuration

### Requirement: Pre-commit hooks format staged JavaScript package files
The local pre-commit hook SHALL format staged dashboard and infrastructure files with the repository's existing Prettier configuration before the commit completes.

#### Scenario: Dashboard file needs formatting
- **WHEN** a staged file under `apps/dashboard` matches the configured Prettier file globs and needs formatting
- **THEN** the pre-commit hook runs the dashboard formatting command through the Makefile
- **AND** the formatted file is restaged before the commit continues

#### Scenario: Infrastructure file needs formatting
- **WHEN** a staged file under `infra` matches the configured Prettier file globs and needs formatting
- **THEN** the pre-commit hook runs the infrastructure formatting command through the Makefile
- **AND** the formatted file is restaged before the commit continues

### Requirement: Hook commands delegate reusable logic to Makefile
The local hook configuration SHALL delegate reusable formatting and validation commands to the root `Makefile` rather than duplicating package-level command logic in hook configuration.

#### Scenario: Hook formats package files
- **WHEN** the local pre-commit hook formats staged dashboard or infrastructure files
- **THEN** the hook invokes a Makefile target for the package-specific formatting operation

### Requirement: Commit message hook validates Conventional Commits
The local commit message hook SHALL validate commit messages using the repository's existing commitlint configuration.

#### Scenario: Commit message is invalid
- **WHEN** a contributor creates a commit with a message that violates commitlint rules
- **THEN** the commit message hook fails the commit before it is created

#### Scenario: Commit message is valid
- **WHEN** a contributor creates a commit with a valid Conventional Commit message
- **THEN** the commit message hook allows the commit to continue

### Requirement: CI remains authoritative when hooks are bypassed
The local hook setup SHALL NOT replace repository CI validation.

#### Scenario: Contributor bypasses local hooks
- **WHEN** a contributor bypasses local Git hooks for a commit
- **THEN** repository CI still runs the configured format and validation checks for pull requests and main branch updates
