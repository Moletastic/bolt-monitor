## ADDED Requirements

### Requirement: Deployment postflight validates the public health envelope
After an explicit infrastructure deployment, the orchestrator SHALL validate the existing public health response as JSON with HTTP success, envelope `status: "success"`, and the documented healthy result. It SHALL use the deploy output API URL and shall not make a second health request solely for envelope validation.

#### Scenario: Public health contract is valid
- **WHEN** the deployed API health endpoint returns the documented success envelope
- **THEN** deployment postflight reports the public health endpoint as validated

#### Scenario: Public health contract is invalid
- **WHEN** the endpoint is unreachable, returns non-success HTTP status, malformed JSON, an error envelope, or an unexpected healthy result
- **THEN** deployment postflight fails with a non-secret health contract error
