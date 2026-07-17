## MODIFIED Requirements

### Requirement: System maintains a source-controlled OpenAPI contract for the HTTP API
The system SHALL store a source-controlled OpenAPI document that describes the currently deployed HTTP API routes, operations, authentication requirements, standard response envelopes, and JSON payloads.

#### Scenario: Developer reviews API contract
- **WHEN** a developer opens the repository documentation assets
- **THEN** they can find one checked-in OpenAPI document that defines the current API surface

#### Scenario: OpenAPI contract covers current API routes
- **WHEN** a developer or automated release gate reads the OpenAPI document
- **THEN** every method-and-path route deployed by SST has a matching OpenAPI operation
- **AND** OpenAPI contains no operation absent from the deployed SST route surface

#### Scenario: OpenAPI contract describes route authentication
- **WHEN** deployed routes expose public or protected authentication metadata
- **THEN** matching OpenAPI operations describe equivalent security requirements
- **AND** `GET /api/health` remains explicitly public

## ADDED Requirements

### Requirement: Repository documentation reflects deployed architecture and health contract
Repository documentation SHALL describe the currently deployed SST resources, runtime data flow, API authentication boundary when present, and standard health response envelope without contradicting source-controlled infrastructure.

#### Scenario: Contributor reviews architecture guidance
- **WHEN** a contributor reads the repository architecture and deployment sections
- **THEN** the described API, dashboard, scheduled runtime, queue, notification runtime, and persistence relationships match `infra/stacks/bootstrap.ts`

#### Scenario: Contributor verifies health
- **WHEN** a contributor follows the documented health verification workflow
- **THEN** the expected response uses the standard success envelope rather than a legacy raw `{ "ok": true }` body

### Requirement: Repository documents API contract maintenance
Repository guidance SHALL direct contributors changing routes, authentication metadata, or response contracts to update SST wiring, handler route inventory, Bruno, OpenAPI, and relevant OpenSpec capabilities and to run the deterministic contract gates.

#### Scenario: Contributor changes an API route
- **WHEN** a contributor consults route-maintenance guidance
- **THEN** the guidance identifies the API artifacts that must remain synchronized
- **AND** it provides the Makefile commands that validate synchronization

#### Scenario: Contributor prepares authentication route registration changes
- **WHEN** a contributor follows the authentication cutover guidance
- **THEN** it requires dashboard production build, `make check-bruno`, full SST/OpenAPI/Bruno/handler route drift validation, health envelope/documentation validation, and applicable portable profile/stage checks to pass first

### Requirement: Repository documents completed OpenSpec archival as maintenance
Repository guidance SHALL describe archiving a completed OpenSpec change as a post-implementation maintenance step and SHALL NOT represent archive state as application behavior or a release-gate substitute.

#### Scenario: Change implementation is complete
- **WHEN** a contributor follows the OpenSpec completion guidance
- **THEN** the guidance directs them to validate implementation and then archive the completed change through the OpenSpec workflow
- **AND** it distinguishes archival bookkeeping from runtime, API, build, and test verification
