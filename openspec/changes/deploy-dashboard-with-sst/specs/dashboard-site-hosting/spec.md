## ADDED Requirements

### Requirement: System deploys dashboard through SST
The system SHALL deploy `apps/dashboard` through the repository's SST application using a Next.js hosting component compatible with the dashboard's server-rendered runtime.

#### Scenario: Operator deploys infrastructure stack
- **WHEN** an operator runs the documented SST deployment workflow
- **THEN** the deployment provisions hosting resources for the dashboard application
- **AND** the hosted dashboard is reachable through an SST-generated URL without requiring custom DNS configuration

### Requirement: Deployed dashboard runtime is wired to deployed monitor API
The system SHALL configure the deployed dashboard runtime to use the monitor API URL from the same SST stack.

#### Scenario: Dashboard server fetches API data after deployment
- **WHEN** the deployed dashboard renders a page that needs monitor API data
- **THEN** the dashboard runtime uses `NEXT_PUBLIC_MONITOR_API_BASE_URL` configured from the stack's deployed API URL
- **AND** operators do not need to set that environment variable manually after deployment

### Requirement: Deployment publishes dashboard access URL
The system SHALL publish the dashboard hosting URL as a stack output.

#### Scenario: Operator needs deployed dashboard entrypoint
- **WHEN** SST deployment completes successfully
- **THEN** the stack outputs include a dashboard URL value
- **AND** that value points to the generated hosting URL for the deployed dashboard site

### Requirement: Deployment documentation covers dashboard hosting path
The system SHALL document how the dashboard is deployed and what runtime assumptions it has.

#### Scenario: Developer reviews deployment guidance
- **WHEN** a developer reads the repository documentation after dashboard hosting is added
- **THEN** they can identify that the dashboard is deployed through SST
- **AND** they can see that the first deployment uses a generated URL and stack-managed `NEXT_PUBLIC_MONITOR_API_BASE_URL`
