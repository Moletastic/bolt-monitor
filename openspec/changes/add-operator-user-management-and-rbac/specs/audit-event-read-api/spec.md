## MODIFIED Requirements

### Requirement: Audit read responses expose stable mutation history fields
System SHALL return audit-event metadata suitable for history views and operator investigation. Every mutation initiated by an authenticated operator SHALL include opaque immutable actor membership identifier and actor type; system-initiated events SHALL include explicit system origin. Audit output SHALL NOT contain email, Cognito subject, provider username/data, tokens, credentials, invitation material, session identifiers/hashes, authorization headers, idempotency keys, or private lifecycle operation input.

#### Scenario: Monitor has recorded mutations
- **WHEN** system returns audit history for a monitor
- **THEN** each event includes stable identity, type, timestamp, and actor or system-origin metadata

#### Scenario: Authenticated operator mutates a resource
- **WHEN** an authenticated operator successfully performs an audited mutation
- **THEN** event attributes it to opaque membership identifier
- **AND** does not use email, Cognito group, or subject as actor authority

#### Scenario: Audit response contains identity-provider material
- **WHEN** system serializes an audit response
- **THEN** it omits identity-provider and authentication material

## ADDED Requirements

### Requirement: AuthTable authority projects safe lifecycle audit events into AppTable
For each effective invite, invitation resend, activation, role assignment, enable, disable, or session-revocation outcome, the system SHALL keep authoritative operation/projection state in AuthTable and SHALL idempotently write one non-PII audit projection to AppTable at `PK=TENANT#DEFAULT`, `SK=AUDIT#USER_LIFECYCLE#<reverse-time>#<event-id>`. The projection SHALL NOT be used for authorization, membership state, last-admin checks, idempotency, or repair decisions.

#### Scenario: Administrator changes a user role
- **WHEN** AuthTable transaction effectively changes role
- **THEN** AppTable projection records event ID/type/time, opaque actor/target membership IDs, safe previous/new role, correlation ID, and outcome
- **AND** records no email, subject, provider data, token, session, or private operation input

#### Scenario: Lifecycle retry converges
- **WHEN** retry observes an already effective outcome
- **THEN** idempotent projection preserves exactly one event for that outcome

#### Scenario: AppTable projection write fails
- **WHEN** authority transition succeeds but projection fails
- **THEN** AuthTable operation retains `AuditProjectionState=PENDING`
- **AND** same-key retry or explicit repair writes projection without replaying authority transition

#### Scenario: AppTable projection is lost or stale
- **WHEN** projected audit differs from AuthTable operation state
- **THEN** application authorization and repair continue exclusively from AuthTable

### Requirement: Administrators can read tenant user-lifecycle audit history
The system SHALL expose bounded newest-first user-lifecycle audit history through `GET /api/v1/admin/audit-events`, restricted to `ADMIN`, by querying the AppTable `TENANT#DEFAULT` lifecycle-audit key range with cursor continuation.

#### Scenario: Administrator requests lifecycle audit history
- **WHEN** an administrator calls `GET /api/v1/admin/audit-events`
- **THEN** system returns a bounded newest-first page of safe projected lifecycle events and continuation metadata

#### Scenario: Non-admin requests lifecycle audit history
- **WHEN** an `OPERATOR` or `VIEWER` calls the endpoint
- **THEN** system returns generic forbidden response without audit data

#### Scenario: Projection is pending
- **WHEN** an AuthTable lifecycle outcome has not yet projected to AppTable
- **THEN** audit read may temporarily omit it
- **AND** administration operation status exposes safe pending projection state for explicit repair
