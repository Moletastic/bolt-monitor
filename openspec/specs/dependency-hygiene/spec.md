# dependency-hygiene Specification

## Purpose
TBD - created by archiving change review-and-prune-dependencies. Update Purpose after archive.
## Requirements
### Requirement: Dependency removal preserves owning-root integrity
The repository SHALL remove `devicons-react` only from the dashboard package root and SHALL update the corresponding pnpm lockfile deterministically.

#### Scenario: Unused dashboard dependency is removed
- **WHEN** review proves a dashboard dependency has no source, test, tool, or runtime use
- **THEN** the dashboard manifest and committed pnpm lockfile remove it together
- **AND** dashboard typecheck, tests, and production build pass

#### Scenario: Stale Go dependency is removed
- **WHEN** review proves a Go module requirement is stale or unreachable
- **THEN** the owning module is reconciled with Go module tooling
- **AND** its source import graph, tests, and repository Go checks pass

### Requirement: Dependency removal preserves dashboard behavior
The repository SHALL verify the dashboard with its frozen pnpm install, lint, type check, tests, and production build after removing `devicons-react`.

#### Scenario: Dashboard checks run after removal
- **WHEN** `devicons-react` has been removed
- **THEN** the frozen dashboard install, lint, type check, tests, and production build pass
