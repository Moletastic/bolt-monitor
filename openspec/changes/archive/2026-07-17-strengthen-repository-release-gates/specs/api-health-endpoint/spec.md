## MODIFIED Requirements

### Requirement: Health endpoint returns deterministic JSON body
System SHALL return `200 OK` with the standard API success envelope containing a stable machine-readable healthy result.

#### Scenario: Health check succeeds
- **WHEN** Go Lambda handles a valid request for `/api/health`
- **THEN** response status is `200`
- **AND** response body has `status: "success"`
- **AND** response `data` contains the stable healthy result
- **AND** error-only envelope fields are omitted

### Requirement: Developer workflow covers health endpoint validation
System SHALL document commands and steps needed to run, deploy, and verify the public health endpoint and its standard response envelope.

#### Scenario: Developer validates endpoint after setup
- **WHEN** developer follows documented workflow
- **THEN** they can run the stack and confirm the health endpoint responds without authentication
- **AND** they can verify HTTP 200 and the documented success envelope

#### Scenario: Authentication security cutover begins
- **WHEN** versioned API routes are about to receive authentication
- **THEN** local handler tests, OpenAPI, Bruno, repository documentation, and deterministic contract gates already agree that health is public and uses the standard success envelope
