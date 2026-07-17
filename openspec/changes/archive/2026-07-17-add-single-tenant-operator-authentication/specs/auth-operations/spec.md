## ADDED Requirements

### Requirement: Authentication emits minimum security audit events
The system SHALL emit structured security audit events for bootstrap reconciliation, sign-in success and failure, password-recovery request and completion, TOTP enrollment and challenge result, dashboard session creation and termination, token-refresh failure, membership authorization denial, membership status change, and `AuthValidAfter` advancement. Events SHALL include timestamp, event type, outcome, deployment/stage, correlation identifier where available, and immutable subject when known; they SHALL NOT include passwords, temporary passwords, recovery or TOTP codes, TOTP secrets, session identifiers or hashes, Cognito challenge sessions, JWTs, refresh tokens, cookie values, encryption-key material, or full request bodies.

#### Scenario: Sign-in fails
- **WHEN** an operator submits invalid credentials
- **THEN** the system records a structured failed-sign-in event with no credential or token material

#### Scenario: Membership denies a valid token
- **WHEN** the API rejects an authenticated subject during membership authorization
- **THEN** the system records a denial event that can be correlated operationally
- **AND** the client response and log event do not reveal the membership state distinction

#### Scenario: Bootstrap changes authority
- **WHEN** bootstrap creates or reconciles initial administrator membership
- **THEN** it records the acting AWS principal when available, target immutable subject, desired role, outcome, and stage without recording credentials

### Requirement: Authentication exposes bounded operational telemetry
The system SHALL publish or derive metrics for sign-in failures, password-recovery requests, refresh failures, authorization denials, bootstrap failures, and auth/session storage or encryption-key loading errors. It SHALL provide alarms or documented alert thresholds for sustained refresh failures and auth infrastructure errors while avoiding per-user high-cardinality dimensions.

#### Scenario: Refresh starts failing systemically
- **WHEN** refresh failures cross the documented threshold within the evaluation window
- **THEN** operations receives an actionable signal identifying the stage and failing auth component

#### Scenario: Metrics are inspected
- **WHEN** an operator reviews auth telemetry
- **THEN** dimensions are bounded and contain no email address, token, session value, or other authentication secret

### Requirement: Auth resources follow the shared stage lifecycle classification
Authentication infrastructure SHALL consume the persistent-versus-ephemeral stage classification defined by the forthcoming `standardize-stage-resource-lifecycle` change rather than defining a separate stage policy. Persistent stages SHALL retain and protect the Cognito user pool, authoritative `AuthTable`, and active AES secret and SHALL enable `AuthTable` point-in-time recovery. Ephemeral stages SHALL omit incompatible protection/PITR settings and SHALL cleanly delete the user pool, `AuthTable`, and secret when the stage is removed.

#### Scenario: Persistent stage is removed or replaced
- **WHEN** a stage classified as persistent undergoes stack removal or resource replacement
- **THEN** the user pool, `AuthTable`, and active AES secret retain compatible protection
- **AND** operators can identify the retained resources without exposing the secret value

#### Scenario: Ephemeral stage is removed
- **WHEN** a stage classified as ephemeral is deleted
- **THEN** the user pool, `AuthTable`, and active AES secret are deleted cleanly with the stage
- **AND** no retained authentication resource is orphaned

#### Scenario: Auth table is corrupted or data is removed accidentally
- **WHEN** an operator initiates recovery within the configured backup window
- **THEN** persistent-stage point-in-time recovery can restore memberships and session records to a recovery table

### Requirement: Break-glass recovery uses authenticated AWS administration
The system SHALL document a break-glass procedure using an authorized AWS identity and Cognito/DynamoDB administrative APIs to locate or create an operator, reset credentials when required, and create or re-enable the subject's `DEFAULT` `ADMIN` membership. The procedure SHALL NOT depend on a functioning dashboard session or expose an unauthenticated recovery endpoint, and SHALL require post-recovery audit review and credential rotation.

#### Scenario: All dashboard administrators are locked out
- **WHEN** an authorized responder follows break-glass recovery with valid AWS credentials
- **THEN** they can restore one controlled `ADMIN` path without disabling API authentication globally
- **AND** the procedure produces auditable AWS and application operations

#### Scenario: Responder lacks required AWS authority
- **WHEN** a caller without the documented IAM permissions attempts break-glass steps
- **THEN** AWS denies the operation and no weaker application fallback is available

### Requirement: V1 authentication cutover is explicit and fail-closed
The deployment procedure SHALL treat v1 authentication as a security cutover: provision and validate auth resources, bootstrap at least one administrator, validate dashboard and direct-client access, then atomically activate dashboard protection, membership/epoch enforcement, and JWT authorizer plus required access scope on all `/api/v1/**` routes in one controlled release. Implementation tasks and commits MAY use reviewable internal milestones, but no deployed environment SHALL operate protected routes in a mixed or optional-auth mode. Rollback SHALL preserve auth data according to stage lifecycle and SHALL prefer fixing or reverting application components without reopening anonymous v1 access.

#### Scenario: Cutover prerequisites are incomplete
- **WHEN** no complete `ACTIVE` administrator membership exists or protected-flow validation fails
- **THEN** the runbook blocks the authorization cutover

#### Scenario: Cutover completes
- **WHEN** the authenticated release is deployed
- **THEN** anonymous `/api/v1/**` requests stop working immediately
- **AND** `GET /api/health` remains publicly available

#### Scenario: Post-cutover application regression occurs
- **WHEN** operators roll back dashboard or API implementation
- **THEN** persistent-stage identity and membership data remain available
- **AND** the rollback does not intentionally remove the v1 JWT authorizer as a convenience fallback

### Requirement: Authentication additions carry an explicit cost posture
The implementation SHALL use Cognito Essentials with default email, an on-demand authoritative `AuthTable` with TTL cleanup and persistent-stage point-in-time recovery, one active generation of an installation-specific AES key supplied through an SST Secret or bootstrap-safe secret reference using AWS-managed at-rest protection, and bounded CloudWatch logs, metrics, and alarms. The design SHALL document Cognito monthly-active-user and messaging charges, DynamoDB request/storage/backup charges, secret retrieval and encryption-operation charges, and CloudWatch ingestion/retention/alarm charges. It SHALL avoid a fixed monthly customer-managed KMS key charge and SHALL add no NAT gateway, custom email service, custom domain, always-on compute, or cross-region auth traffic in v1.

#### Scenario: Infrastructure cost is reviewed
- **WHEN** reviewers evaluate the authentication change
- **THEN** they can identify each material recurring and usage-based AWS cost source
- **AND** low operator volume uses serverless/on-demand resources rather than idle provisioned capacity
