## Purpose

Define the repository-level policy for installing JavaScript dependencies in
in-scope package roots. The policy covers the package manager used,
lockfile discipline, and the explicit trust decisions required for
dependencies that need install or build scripts. It exists to reduce
exposure to dependency hijacks and unexpected lifecycle-script execution
without changing the contribution surface for normal code work.

## Requirements

### Requirement: Repository defines a primary JavaScript package-manager workflow

Repository SHALL define one primary package-manager workflow for in-scope
JavaScript projects and document it as the default contributor path.

#### Scenario: Contributor follows repository setup instructions

- **WHEN** a contributor reads repository documentation for `infra/` or
  `apps/dashboard`
- **THEN** they find one consistent package-manager workflow rather than
  mixed `npm` and `pnpm` guidance

### Requirement: Repository uses committed lockfile state for in-scope JavaScript installs

Repository SHALL keep committed lockfile state for in-scope JavaScript
package roots so dependency resolution remains reproducible and reviewable.

#### Scenario: Contributor installs dependencies in an in-scope package root

- **WHEN** a contributor installs dependencies without intentionally
  changing package versions
- **THEN** the package manager resolves dependencies from committed
  lockfile state rather than silently drifting to a new graph

### Requirement: Repository makes install-script trust an explicit decision

Repository SHALL require explicit review of dependencies that need
install or build scripts as part of the in-scope JavaScript dependency
workflow.

#### Scenario: In-scope dependency requires install-time script execution

- **WHEN** a dependency in an in-scope JavaScript package root requires
  install or build script execution
- **THEN** the repository workflow records or enforces that trust
  decision explicitly rather than inheriting it silently from
  package-manager defaults

### Requirement: Repository documents scope boundaries for hardened install policy

Repository SHALL document which JavaScript package roots are covered by
the hardened package-manager policy and which are intentionally excluded.

#### Scenario: Contributor inspects package-manager policy scope

- **WHEN** a contributor reviews dependency workflow documentation
- **THEN** they can tell whether `infra/`, `apps/dashboard`, and
  `openapi/` are governed by the same install-security policy or different
  ones
