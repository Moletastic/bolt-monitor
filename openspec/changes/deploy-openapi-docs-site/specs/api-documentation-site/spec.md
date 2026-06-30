## ADDED Requirements

### Requirement: System deploys hosted API documentation through SST
The system SHALL deploy the repository OpenAPI documentation workspace as a hosted static site through SST-managed infrastructure.

#### Scenario: Developer deploys infrastructure
- **WHEN** the infrastructure stack is deployed
- **THEN** the system publishes a hosted API documentation site sourced from the repository docs workspace

### Requirement: Hosted API docs target the deployed API URL
The system SHALL publish a hosted OpenAPI contract that targets the deployed API URL for interactive documentation use.

#### Scenario: Developer uses hosted Swagger UI
- **WHEN** a developer opens the hosted Swagger UI and executes a documented operation
- **THEN** the request targets the deployed API host for that stack instead of a local placeholder host

### Requirement: Infrastructure outputs expose the hosted docs URL
The system SHALL expose the hosted docs site URL as part of infrastructure outputs.

#### Scenario: Developer checks deploy outputs
- **WHEN** a deployment finishes successfully
- **THEN** the stack outputs include the hosted API docs URL alongside other relevant application endpoints

### Requirement: Hosted docs reuse the existing OpenAPI docs workspace
The system SHALL build the hosted docs site from the existing checked-in OpenAPI docs workspace rather than maintaining a separate documentation source.

#### Scenario: Developer updates repository docs assets
- **WHEN** a developer changes the checked-in OpenAPI docs workspace and redeploys
- **THEN** the hosted docs site reflects those documentation updates without requiring a second docs source tree
