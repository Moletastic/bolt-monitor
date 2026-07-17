## ADDED Requirements

### Requirement: Repository validates deployed route coverage across API assets
The repository SHALL provide a deterministic validation command that compares normalized HTTP method-and-path routes declared by SST with the Bruno collection and checked-in OpenAPI document.

#### Scenario: Route is consistently represented
- **WHEN** an SST route has a matching normalized method and path in Bruno and OpenAPI
- **THEN** the route passes coverage validation

#### Scenario: Deployed route is missing from an API asset
- **WHEN** an SST route has no matching Bruno request or OpenAPI operation
- **THEN** validation fails and identifies the missing source, HTTP method, and normalized path

#### Scenario: API asset contains a stale route
- **WHEN** Bruno or OpenAPI declares a method-and-path operation that is not deployed by SST
- **THEN** validation fails and identifies the stale file or operation

### Requirement: Repository detects statically identifiable handler and infrastructure route drift
The repository SHALL compare the monitor API handler's statically identifiable route inventory with SST routes so handler behavior is not silently unreachable and deployed routes are not silently unhandled. Known discrepancies SHALL fail validation until explicitly reconciled and SHALL be resolved before authentication refactors SST route registration through a protected-route helper.

#### Scenario: Handler behavior is absent from infrastructure
- **WHEN** a statically identifiable handler method-and-path pattern has no matching SST route
- **THEN** validation fails and identifies the handler route and its source location

#### Scenario: Infrastructure route is absent from handler inventory
- **WHEN** an SST route targeting the monitor API handler has no matching handler route
- **THEN** validation fails and identifies the deployed route

#### Scenario: Authentication route helper is introduced
- **WHEN** route registration is about to be refactored through the authentication change's protected-route helper
- **THEN** handler/SST route validation already passes
- **AND** the refactor starts from the explicitly reconciled route inventory

#### Scenario: Dynamic behavior cannot be compared safely
- **WHEN** a handler branch cannot be converted to a deterministic method-and-path pattern
- **THEN** the validator does not infer a potentially incorrect route
- **AND** repository tests or diagnostics identify the unsupported pattern so coverage can be made explicit

### Requirement: Repository detects explicit merged OpenSpec route drift
The repository SHALL compare explicit HTTP method-and-path requirements in merged OpenSpec capabilities with SST routes without treating active, unimplemented change proposals as deployed behavior.

#### Scenario: Merged capability requires an unwired route
- **WHEN** a spec under `openspec/specs/` explicitly requires an HTTP method and API path that has no matching SST route
- **THEN** validation fails and identifies the capability file and normalized route

#### Scenario: Active change describes future behavior
- **WHEN** an unarchived change under `openspec/changes/` describes a route not yet deployed
- **THEN** the deployed-contract validator excludes that active change from merged-capability route coverage

### Requirement: Route authentication metadata remains consistent
When SST route declarations expose authentication metadata, the repository SHALL classify each API route as public or protected and SHALL validate equivalent metadata in applicable Bruno and OpenAPI operations.

#### Scenario: Public route metadata is available
- **WHEN** SST identifies a route as public
- **THEN** matching Bruno and OpenAPI operations declare that no bearer credential is required

#### Scenario: Protected route metadata is available
- **WHEN** SST identifies a route as protected
- **THEN** matching Bruno and OpenAPI operations declare the configured authentication requirement

#### Scenario: Authentication metadata disagrees
- **WHEN** applicable SST, Bruno, or OpenAPI metadata classifies the same route differently
- **THEN** validation fails with the normalized route and conflicting classifications

#### Scenario: Infrastructure has not introduced route authentication metadata
- **WHEN** SST routes do not yet expose public or protected classification
- **THEN** route coverage validation still runs
- **AND** the validator does not invent an authentication classification for non-health routes

### Requirement: Contract validators are tested and deterministic
The route and metadata validators SHALL run locally without network or cloud access and SHALL have fixture tests for successful and failing drift cases.

#### Scenario: Contributor runs contract validation locally
- **WHEN** the contributor invokes the documented Makefile target
- **THEN** validation reads only source-controlled repository inputs
- **AND** repeated runs against the same inputs produce the same result

#### Scenario: Validator regression tests run
- **WHEN** validator tests execute
- **THEN** fixtures cover missing, stale, parameter-normalized, handler-only, unhandled, and authentication-mismatch routes
- **AND** assertions verify actionable diagnostics

### Requirement: Pre-cutover contract gates cover all route representations
The repository SHALL require the deterministic SST, OpenAPI, Bruno, handler-route, health-contract, and documentation checks before the authentication security cutover.

#### Scenario: Pre-cutover release gates run
- **WHEN** the repository prepares to attach authentication to versioned API routes
- **THEN** SST, OpenAPI, Bruno, and handler route sets agree
- **AND** `make check-bruno` passes
- **AND** the public health operation and repository documentation describe the standard success envelope
