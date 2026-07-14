## ADDED Requirements

### Requirement: Bruno collection covers exposed API routes
The repository SHALL maintain at least one Bruno HTTP request for every method-and-path route declared in `infra/stacks/bootstrap.ts`, across all Bruno collections.

#### Scenario: Wired route has matching request
- **WHEN** the local Bruno guard reads a route declaration and a Bruno request with the same method and normalized path exists
- **THEN** the route passes coverage validation

#### Scenario: Wired route lacks matching request
- **WHEN** the local Bruno guard reads a route declaration with no matching Bruno request
- **THEN** the guard reports the missing method-and-path route and fails

#### Scenario: Bruno request references removed route
- **WHEN** the local Bruno guard reads a Bruno request whose method and normalized path do not exist in bootstrap
- **THEN** the guard reports the stale request and fails

### Requirement: Bruno requests follow collection conventions
Each covered Bruno request SHALL use a verb-and-resource name, exact bootstrap route parameter names, one `domain:<domain>` tag, one `operation:<operation>` tag, and docs describing purpose, setup, and expected result.

#### Scenario: Request metadata is complete
- **WHEN** the guard validates a covered request
- **THEN** it accepts the request only when required name, tags, route variables, and documentation are present

#### Scenario: Request metadata is invalid
- **WHEN** a covered request lacks a required tag, uses a mismatched route variable, or lacks required documentation sections
- **THEN** the guard reports the metadata violation and fails

### Requirement: Repository exposes local Bruno validation
The repository SHALL expose `make check-bruno` as a local, deterministic validation command.

#### Scenario: Developer runs Bruno check
- **WHEN** developer runs `make check-bruno`
- **THEN** validator checks all Bruno collections against bootstrap routes and exits successfully only when coverage and conventions pass

#### Scenario: OpenSpec route is not wired
- **WHEN** an OpenSpec route requirement has no corresponding bootstrap route
- **THEN** validation reports the infrastructure/spec mismatch separately from Bruno coverage and does not treat it as a missing Bruno request

### Requirement: Bruno governance is documented
The repository SHALL document the high-level collection governance principle in `CONSTITUTION.md` and the operational conventions and maintenance command in `AGENTS.md`.

#### Scenario: Contributor adds or changes an API route
- **WHEN** contributor changes an SST API route
- **THEN** repository guidance directs contributor to update Bruno coverage, metadata, and run `make check-bruno`
