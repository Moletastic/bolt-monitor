## Purpose

Define the source-controlled API contract and local documentation workflows for the HTTP API.

## Requirements

### Requirement: System maintains a source-controlled OpenAPI contract for the HTTP API
The system SHALL store a source-controlled OpenAPI document that describes the currently supported HTTP API routes, operations, and JSON payloads.

#### Scenario: Developer reviews API contract
- **WHEN** a developer opens the repository documentation assets
- **THEN** they can find one checked-in OpenAPI document that defines the current API surface

#### Scenario: OpenAPI contract covers current API routes
- **WHEN** a developer reads the OpenAPI document
- **THEN** it describes the current health, monitor CRUD, monitor status, and monitor run-history endpoints

### Requirement: System documents monitor API without probe-location contracts
System SHALL document the monitor API according to the single-execution-environment product contract.

#### Scenario: API documentation shows monitor payloads
- **WHEN** OpenAPI examples or schemas describe monitor create, update, read, status, runs, or manual-run responses
- **THEN** they do not include `probeLocations`, `probeLocationId`, `lastProbeLocationId`, or hard-coded location examples such as `iad`

#### Scenario: API documentation lists monitor API paths
- **WHEN** OpenAPI paths are rendered
- **THEN** probe-location catalog endpoints are not documented as supported product APIs

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

### Requirement: API documentation describes authenticated v1 access
The source-controlled API documentation SHALL define the Cognito Bearer access-token security scheme and required `aws.cognito.signin.user.admin` access scope for every `/api/v1/**` operation and SHALL leave `GET /api/health` explicitly unauthenticated. It SHALL state that ID tokens and dashboard session cookies are not API credentials and that API Gateway pre-integration authentication or scope failures can produce non-envelope `401` or `403` responses.

#### Scenario: Developer reviews a v1 operation
- **WHEN** a developer reads any `/api/v1/**` operation in the OpenAPI contract
- **THEN** the operation requires the Cognito Bearer access-token security scheme and documented access scope

#### Scenario: Developer reviews health operation
- **WHEN** a developer reads `GET /api/health` in the OpenAPI contract
- **THEN** the operation declares no authentication requirement

#### Scenario: Developer reviews auth failures
- **WHEN** a developer reads authentication and authorization response documentation
- **THEN** it distinguishes API Gateway pre-integration non-envelope `401` or `403` responses from application envelope `401` and `403` responses

### Requirement: Bruno documents and exercises direct Cognito authentication
The Bruno collection SHALL document a direct, human-operator Cognito authentication setup that obtains an access token without Cognito managed login, Amplify, dashboard cookies, or committed credentials. Protected requests SHALL source the Bearer token from a secret-capable local environment variable, and collection documentation SHALL explain invitation activation, token refresh or reacquisition, and cleanup without printing or persisting tokens in source control.

#### Scenario: Operator prepares Bruno after invitation
- **WHEN** an invited operator follows the Bruno setup documentation
- **THEN** they can complete the supported Cognito challenge flow and provide an access token through local secret configuration

#### Scenario: Bruno sends protected request
- **WHEN** any Bruno request targets `/api/v1/**`
- **THEN** it sends the configured access token as a Bearer credential
- **AND** it retains the route's existing domain and operation tags and documentation conventions

#### Scenario: Bruno sends health request
- **WHEN** the Bruno collection sends `GET /api/health`
- **THEN** the request succeeds without an Authorization header

#### Scenario: Repository is inspected for direct-client secrets
- **WHEN** committed Bruno files and examples are reviewed
- **THEN** no password, temporary password, challenge session, recovery code, access token, ID token, refresh token, or dashboard session value is present
