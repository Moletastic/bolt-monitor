## ADDED Requirements

### Requirement: System maintains a source-controlled OpenAPI contract for the HTTP API
The system SHALL store a source-controlled OpenAPI document that describes the currently supported HTTP API routes, operations, and JSON payloads.

#### Scenario: Developer reviews API contract
- **WHEN** a developer opens the repository documentation assets
- **THEN** they can find one checked-in OpenAPI document that defines the current API surface

#### Scenario: OpenAPI contract covers current API routes
- **WHEN** a developer reads the OpenAPI document
- **THEN** it describes the current health, probe-location, monitor CRUD, monitor status, and monitor run-history endpoints

### Requirement: System provides local Swagger UI for interactive API documentation
The system SHALL provide a local Swagger UI workflow that renders the checked-in OpenAPI document for interactive API exploration.

#### Scenario: Developer opens Swagger UI locally
- **WHEN** a developer runs the documented local docs workflow for Swagger UI
- **THEN** the system serves a local Swagger UI page backed by the repository OpenAPI document

### Requirement: System provides local Redoc for static API reference
The system SHALL provide a local Redoc workflow that renders the same checked-in OpenAPI document as a static API reference.

#### Scenario: Developer opens Redoc locally
- **WHEN** a developer runs the documented local docs workflow for Redoc
- **THEN** the system serves a local Redoc page backed by the same repository OpenAPI document used by Swagger UI

### Requirement: System documents a stable local workflow for viewing API docs
The system SHALL document a stable local command workflow for viewing API documentation.

#### Scenario: Developer wants to view API docs
- **WHEN** a developer follows the repository docs workflow
- **THEN** they can run one documented command surface to open or serve the local API docs without modifying application code or deployment configuration
