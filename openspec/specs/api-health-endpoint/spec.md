## MODIFIED Requirements

### Requirement: Health endpoint returns deterministic JSON body
System SHALL return `200 OK` with the standard API success envelope containing a stable machine-readable healthy result. This public endpoint SHALL be liveness-only and SHALL not depend on DynamoDB, authorization, or protected monitor API availability.

#### Scenario: Health check succeeds
- **WHEN** Go Lambda handles a valid request for `/api/health`
- **THEN** response status is `200`
- **AND** response body has `status: "success"`
- **AND** response `data` contains the stable healthy result
- **AND** error-only envelope fields are omitted

#### Scenario: Protected API dependency is unavailable
- **WHEN** a protected API dependency is unavailable
- **THEN** public health remains a liveness response
- **AND** authenticated readiness verification reports the protected-path failure separately

### Requirement: Developer workflow covers health endpoint validation
System SHALL document commands and steps needed to run, deploy, and verify the public health endpoint and its standard response envelope.

#### Scenario: Developer validates endpoint after setup
- **WHEN** developer follows documented workflow
- **THEN** they can run the stack and confirm the health endpoint responds without authentication
- **AND** they can verify HTTP 200 and the documented success envelope

#### Scenario: Authentication security cutover begins
- **WHEN** versioned API routes are about to receive authentication
- **THEN** local handler tests, OpenAPI, Bruno, repository documentation, and deterministic contract gates already agree that health is public and uses the standard success envelope

### Requirement: Deployment postflight validates the public health envelope
After an explicit infrastructure deployment, the orchestrator SHALL validate the existing public health response as JSON with HTTP success, envelope `status: "success"`, and the documented healthy result. It SHALL use the deploy output API URL and shall not make a second health request solely for envelope validation.

#### Scenario: Public health contract is valid
- **WHEN** the deployed API health endpoint returns the documented success envelope
- **THEN** deployment postflight reports the public health endpoint as validated

#### Scenario: Public health contract is invalid
- **WHEN** the endpoint is unreachable, returns non-success HTTP status, malformed JSON, an error envelope, or an unexpected healthy result
- **THEN** deployment postflight fails with a non-secret health contract error
