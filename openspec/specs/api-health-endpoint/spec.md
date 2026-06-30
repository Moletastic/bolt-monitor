## Requirements

### Requirement: System exposes public health endpoint
System SHALL expose public HTTP `GET` endpoint for service health through API Gateway.

#### Scenario: Client calls health route
- **WHEN** client sends `GET` request to `/api/health`
- **THEN** API Gateway routes request to backend Lambda and returns successful HTTP response

### Requirement: Health endpoint is backed by Go Lambda
System SHALL implement health endpoint handler in Go code under `services/` and deploy it as Lambda.

#### Scenario: Infrastructure deploys health handler
- **WHEN** infrastructure stack is synthesized or deployed
- **THEN** health route targets Go Lambda built from repository service code

### Requirement: Health endpoint returns deterministic JSON body
System SHALL return `200 OK` with stable JSON response body that indicates service is healthy.

#### Scenario: Health check succeeds
- **WHEN** Go Lambda handles valid request for `/api/health`
- **THEN** response status is `200` and response body contains machine-readable healthy result

### Requirement: Developer workflow covers health endpoint validation
System SHALL document commands and steps needed to run, deploy, and verify health endpoint.

#### Scenario: Developer validates endpoint after setup
- **WHEN** developer follows documented workflow
- **THEN** they can run stack and confirm health endpoint responds successfully
