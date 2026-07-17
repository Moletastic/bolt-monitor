## ADDED Requirements

### Requirement: Repository provides an opt-in local staging release smoke helper
After authentication is implemented, the repository SHALL provide an explicitly invoked local helper that an operator runs from a workstation after deliberately deploying a non-production SST stage governed by `standardize-stage-resource-lifecycle`. Repository CI SHALL not deploy infrastructure or receive AWS credentials.

#### Scenario: Authorized operator starts staging smoke
- **WHEN** authentication and stage-lifecycle dependencies are complete and an authorized operator explicitly invokes the local helper with local credentials available
- **THEN** the helper validates the deliberately deployed declared non-production stage
- **AND** it obtains the deployed API URL from SST output rather than a production endpoint

#### Scenario: Credentials are unavailable
- **WHEN** the local staging smoke helper is invoked without its required credentials
- **THEN** it stops with an actionable explanation
- **AND** ordinary pull-request CI remains unaffected and credential-free

#### Scenario: Production target is supplied
- **WHEN** local input or resolved configuration identifies a production stage
- **THEN** the helper rejects the target before requests are sent

#### Scenario: Authentication is not implemented
- **WHEN** protected routes and their staging token flow do not yet exist
- **THEN** the credentialed staging auth-smoke helper is not enabled or treated as a release gate
- **AND** local deterministic Phase 0 release gates remain required

### Requirement: Staging smoke uses a retention-safe lifecycle
The local staging smoke helper SHALL use the declared long-lived persistent staging environment. It SHALL NOT create, deploy, or remove any stage and SHALL NOT target a unique or per-run stage containing retained resources.

#### Scenario: Persistent staging lifecycle is selected
- **WHEN** smoke targets infrastructure with retained resources
- **THEN** an operator has deliberately deployed the declared long-lived persistent staging environment first
- **AND** the helper only reads its structured deployment output and calls its API
- **AND** the run does not create a unique retained stage

### Requirement: Staging smoke proves public health and protected authentication behavior
The local staging smoke helper SHALL verify the public health contract and, when protected routes are deployed, both rejected and accepted authentication outcomes against a non-destructive protected route.

#### Scenario: Public health is healthy
- **WHEN** the isolated staging deployment completes
- **THEN** unauthenticated `GET /api/health` returns HTTP 200
- **AND** the JSON body matches the standard success envelope and reports a healthy result

#### Scenario: Protected route receives no token
- **WHEN** smoke calls the selected protected route without an authentication token
- **THEN** API Gateway rejects the request with HTTP 401 before Lambda invocation
- **AND** smoke does not require an application response envelope or `reason.code` from this edge response

#### Scenario: Protected route receives a valid token
- **WHEN** smoke calls the same non-destructive protected route with a valid staging token
- **THEN** the request passes authentication and returns the route's documented non-authentication response

### Requirement: Staging smoke protects credentials and bounds cloud cost
The local staging smoke helper SHALL prevent credential disclosure. It SHALL not deploy, remove, or otherwise mutate cloud infrastructure.

#### Scenario: Smoke commands execute with credentials
- **WHEN** the helper acquires an API token from local credentials
- **THEN** secret values are not printed
- **AND** command output does not print authorization headers, token values, or credential-bearing request traces

#### Scenario: Smoke run finishes or fails
- **WHEN** validation completes or fails
- **THEN** the declared persistent staging environment remains under its lifecycle policy
- **AND** the helper reports its validation result without deleting or changing infrastructure

#### Scenario: Repository CI runs ordinary validation
- **WHEN** an untrusted pull-request event executes repository validation
- **THEN** the local smoke helper does not run
- **AND** AWS credentials and staging tokens are not exposed to any CI event
