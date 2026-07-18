## ADDED Requirements

### Requirement: Authentication owns the only initial administrator bootstrap
The authentication change's AWS-credentialed bootstrap SHALL remain the sole initial bootstrap. It SHALL create the canonical AuthTable `ACTIVE`/`ADMIN` membership and initialize the AuthTable active-administrator guard using the forward-compatible membership schema. RBAC SHALL manage later users and SHALL NOT add a second bootstrap, one-time membership migration, AppTable membership, or duplicate identity key.

#### Scenario: Initial installation is prepared
- **WHEN** an authorized operator initializes the first administrator
- **THEN** they use the authentication-owned bootstrap
- **AND** no RBAC bootstrap or membership migration is run

#### Scenario: RBAC is enabled after authentication
- **WHEN** the RBAC surfaces are deployed
- **THEN** they use the existing canonical AuthTable membership and guard
- **AND** they do not copy membership authority into AppTable

### Requirement: AuthTable membership records are forward-compatible lifecycle records
Canonical membership SHALL store immutable subject and membership ID, tenant `DEFAULT`, status `INVITED|ACTIVE|DISABLED`, role `ADMIN|OPERATOR|VIEWER`, non-negative integer epoch `AuthValidAfter`, monotonically increasing `Version`, and current lifecycle operation ID, type, and state when an operation is attached. Unknown states, roles, malformed boundaries, and invalid versions SHALL fail closed.

#### Scenario: Later user membership is created
- **WHEN** an invitation resolves an immutable Cognito subject
- **THEN** the system creates one canonical AuthTable membership with status `INVITED`, its assigned fixed role, boundary, version, and lifecycle operation state

#### Scenario: Membership transition is retried concurrently
- **WHEN** workers attempt to advance the same membership from an older version
- **THEN** conditional version checks permit only a valid current transition
- **AND** stale writes cannot overwrite newer lifecycle or revocation state

### Requirement: AuthTable provides exact bounded lifecycle access paths
The system SHALL implement these AuthTable access paths: strongly consistent base-table subject membership lookup; base-table email-digest claim; base-table active-admin guard; base-table hashed scoped idempotency operation lookup; sparse GSI1 tenant membership listing and exact membership-ID discovery; sparse GSI2 due pending-operation listing; and sparse GSI3 subject-session listing. GSI discovery SHALL be bounded, and every security mutation SHALL re-read canonical base-table state and use conditional writes or transactions.

#### Scenario: API authorizes a subject
- **WHEN** a protected request resolves Cognito `sub`
- **THEN** it calls `GetItem` for `PK=MEMBER#<sub>`, `SK=MEMBERSHIP` with `ConsistentRead=true`

#### Scenario: Administrator lists users
- **WHEN** an `ADMIN` requests a page of users
- **THEN** the system queries `GSI1PK=TENANT#DEFAULT#MEMBERS` with a bounded limit and opaque continuation

#### Scenario: Workflow lists due repair work
- **WHEN** an explicit repair command requests due operations
- **THEN** it queries `GSI2PK=LIFECYCLE#PENDING` through the current epoch with a strict batch limit
- **AND** it does not scan AuthTable or Cognito

#### Scenario: Workflow invalidates dashboard sessions
- **WHEN** disable or revoke-session cleanup runs for a subject
- **THEN** it traverses bounded pages under `GSI3PK=MEMBER#<sub>#SESSIONS`
- **AND** persists continuation until the current traversal completes

### Requirement: Administrators can list operator memberships
The system SHALL expose admin-only `GET /api/v1/admin/users` and a dashboard administration view listing `DEFAULT` AuthTable memberships with opaque membership ID, administrator-visible email, fixed role, lifecycle status, invitation status, current safe operation status, and timestamps. Responses SHALL NOT contain subject, tokens, credentials, invitation codes, session identifiers, Cognito groups, or raw provider responses.

#### Scenario: Administrator lists users
- **WHEN** an `ADMIN` requests the user list
- **THEN** the system returns a bounded page from the AuthTable membership index
- **AND** each item contains only defined safe administration fields

#### Scenario: Non-admin lists users directly
- **WHEN** an `OPERATOR` or `VIEWER` calls the user list API
- **THEN** the system returns the generic forbidden response

### Requirement: Administrators can invite a later operator with a fixed role
The system SHALL expose an admin-only invite accepting email, fixed role, and idempotency key. It SHALL persist the AuthTable operation/private invite intent and conditional email-digest claim, generate one deterministic provider username, create or recover the Cognito user, obtain immutable subject, and create one AuthTable `INVITED` membership. Membership authority SHALL bind only to Cognito subject.

#### Scenario: New invitation succeeds
- **WHEN** an administrator submits an unused email, fixed role, and new idempotency key
- **THEN** the workflow creates or recovers one Cognito user and one canonical `INVITED` membership
- **AND** Cognito attempts invitation delivery
- **AND** the API returns only safe membership and operation fields

#### Scenario: Same invitation request is retried
- **WHEN** the request is retried with the same idempotency key and fingerprint
- **THEN** the workflow converges on the original provider user, membership, and delivery claim
- **AND** does not create duplicate identity, membership, or invitation attempts

#### Scenario: Idempotency key is reused with different input
- **WHEN** an invite key is reused with a different normalized email or role
- **THEN** the system rejects it as a conflict without new side effects

#### Scenario: Provider create response is lost
- **WHEN** Cognito created the deterministic username but the process ended before recording subject
- **THEN** retry or explicit repair gets that exact provider user and recovers subject
- **AND** never creates another user to compensate

### Requirement: Administrators can resend a pending invitation
The system SHALL expose an admin-only resend for `INVITED` membership. Each logical resend SHALL require an idempotency key and claim at most one provider delivery attempt. An interrupted attempt with unknown Cognito outcome SHALL terminate as `DELIVERY_UNKNOWN`; retries and repair SHALL NOT automatically resend.

#### Scenario: Pending invitation is resent
- **WHEN** an administrator requests resend with a new idempotency key
- **THEN** Cognito attempts one replacement invitation
- **AND** the system records completion without returning invitation material

#### Scenario: Same resend is retried
- **WHEN** the same resend request repeats
- **THEN** the system returns its recorded outcome without another delivery attempt

#### Scenario: Resend outcome is ambiguous
- **WHEN** execution stops after claiming delivery but before durably recording Cognito result
- **THEN** the operation becomes `DELIVERY_UNKNOWN`
- **AND** another delivery requires an explicit new resend key

### Requirement: Invited membership activates after a full Cognito authentication
The system SHALL transactionally transition an `INVITED` membership to `ACTIVE` only when Cognito proves a completed full authentication for that subject and the access token has valid `auth_time`. If role is `ADMIN`, the transaction SHALL update the active-admin guard.

#### Scenario: Invited operator completes first authentication
- **WHEN** Cognito proves a completed full authentication for the invited subject
- **THEN** the system transactionally marks membership `ACTIVE`
- **AND** authorizes only if `auth_time > AuthValidAfter`

#### Scenario: Concurrent first requests arrive
- **WHEN** multiple first requests race for one invited subject
- **THEN** versioned transaction permits one effective activation and guard update
- **AND** retries do not duplicate lifecycle audit projection

### Requirement: Administrators can assign membership roles
The system SHALL expose an admin-only role assignment that changes only fixed role and version/current operation state, without changing subject, tenant, provider identity, or revocation boundary.

#### Scenario: Administrator changes a role
- **WHEN** an administrator assigns another fixed role to active or invited membership
- **THEN** AuthTable stores the role transactionally
- **AND** subsequent requests use the new permission set

#### Scenario: Role update is retried
- **WHEN** a completed role assignment repeats with the same desired role
- **THEN** it converges without duplicate state transitions or audit projections

### Requirement: Last active administrator is protected transactionally
Any operation reducing active `ADMIN` membership count SHALL conditionally prove another active administrator remains and update the AuthTable `ADMIN_GUARD` in the same transaction as membership. Invited or disabled administrators SHALL NOT count. Missing or inconsistent guard state SHALL fail closed for explicit reconciliation.

#### Scenario: Sole active administrator is demoted
- **WHEN** an administrator attempts to demote the sole active `ADMIN`
- **THEN** the transaction fails with generic last-administrator conflict
- **AND** membership and guard remain unchanged

#### Scenario: Sole active administrator is disabled
- **WHEN** an administrator attempts to disable the sole active `ADMIN`
- **THEN** the transaction fails before boundary, session, or Cognito effects

#### Scenario: Concurrent administrators are removed
- **WHEN** concurrent demotion or disable requests would collectively remove every active administrator
- **THEN** AuthTable transactions permit only operations leaving at least one active administrator

### Requirement: Disable follows boundary, session, provider-disable, then sign-out ordering
Disable SHALL execute in this order: transactionally set AuthTable membership `DISABLED`, monotonically advance `AuthValidAfter`, update version/current operation, and maintain the last-admin guard; invalidate subject dashboard sessions; invoke Cognito disable; invoke Cognito global sign-out; then project audit and complete. The operation SHALL persist every step and converge idempotently without restoring application access.

#### Scenario: Disable completes
- **WHEN** an administrator disables eligible membership
- **THEN** AuthTable denies membership and old authentication families before external calls
- **AND** dashboard sessions are invalidated before Cognito disable
- **AND** Cognito disable occurs before global sign-out

#### Scenario: Session cleanup is incomplete
- **WHEN** boundary/status commit succeeds but indexed session cleanup stops mid-page
- **THEN** every session remains logically denied by membership/boundary validation
- **AND** retry or explicit repair resumes cleanup before Cognito steps

#### Scenario: Cognito disable fails
- **WHEN** membership denial and session invalidation complete but provider disable fails
- **THEN** requests remain denied
- **AND** retry or explicit repair resumes provider disable before global sign-out

#### Scenario: Cognito global sign-out fails
- **WHEN** membership denial, session invalidation, and provider disable complete but global sign-out fails
- **THEN** requests remain denied
- **AND** retry or explicit repair resumes global sign-out

#### Scenario: Disable repeats after completion
- **WHEN** disable is repeated for already disabled membership
- **THEN** it converges without moving boundary backward or duplicating audit

### Requirement: Administrators can enable disabled membership
Enable SHALL call Cognito enable first and only then transactionally set AuthTable membership `ACTIVE`, update version/current operation and admin guard, and project audit. Enable SHALL NOT reduce `AuthValidAfter`; access requires a later full authentication with `auth_time` strictly greater than the retained boundary.

#### Scenario: Enable completes
- **WHEN** an administrator enables disabled membership
- **THEN** Cognito identity is enabled before membership activation
- **AND** old authentication families remain denied

#### Scenario: Membership activation fails after provider enable
- **WHEN** Cognito enable succeeds but AuthTable transaction fails
- **THEN** disabled membership continues denying access
- **AND** retry or explicit repair completes activation

### Requirement: Administrators can revoke sessions without disabling membership
Revoke-only SHALL transactionally and monotonically advance `AuthValidAfter` while preserving `Status` and `Role`, invalidate subject dashboard sessions, invoke Cognito global sign-out, project audit, and complete. It SHALL NOT disable AuthTable membership or Cognito user.

#### Scenario: Active sessions are revoked
- **WHEN** an administrator revokes sessions for active membership
- **THEN** authentication ceremonies at or before the new boundary are immediately denied
- **AND** dashboard sessions are invalidated before global sign-out
- **AND** membership remains active

#### Scenario: Old family refreshes after revocation
- **WHEN** an old refresh family returns an access token with later `iat` and old `auth_time`
- **THEN** application authorization continues to deny it

#### Scenario: User authenticates fully after revocation
- **WHEN** active membership completes a full authentication with `auth_time` strictly greater than boundary
- **THEN** the new session is eligible without an enable operation

### Requirement: Lifecycle operations are durable, observable, and explicitly repairable
Invite, resend, enable, disable, and revoke workflows SHALL store authority and repair state only in AuthTable. Retryable incomplete operations SHALL remain visibly `PENDING` or `RETRYABLE` with safe operation ID/type/step/age. Same-key API retry or an AWS-credentialed explicit repair command SHALL resume them using conditional leases, verified steps, and bounded work. No recurring repair Lambda or EventBridge schedule SHALL be provisioned by default.

#### Scenario: Process stops between systems
- **WHEN** execution stops after a Cognito or DynamoDB step
- **THEN** durable AuthTable checkpoints preserve desired state and last verified step
- **AND** retry or repair converges without compensation that restores access

#### Scenario: Administrator repairs one operation
- **WHEN** an authorized operator invokes repair with one operation ID
- **THEN** the command conditionally leases and resumes only that operation
- **AND** emits a safe auditable result

#### Scenario: Administrator reconciles due operations
- **WHEN** an authorized operator invokes bounded due reconciliation
- **THEN** the command queries GSI2 with a strict batch limit and continuation
- **AND** never scans AuthTable or Cognito

#### Scenario: No repair command is running
- **WHEN** work remains incomplete
- **THEN** it remains denied where access removal began and remains observable for operator action
- **AND** no scheduled compute runs automatically

### Requirement: User-management UI exposes safe administration actions
The dashboard SHALL provide an admin-only user list and invite flow plus role, resend, enable, disable, and revoke actions appropriate to state. Destructive/access-removing actions SHALL use shared confirmation dialog, and pending/retryable lifecycle state SHALL be visible without PII-bearing diagnostics.

#### Scenario: Administrator opens user administration
- **WHEN** an `ADMIN` visits administration
- **THEN** dashboard shows safe membership data, operation state, and state-appropriate actions

#### Scenario: Non-admin navigates to administration
- **WHEN** an `OPERATOR` or `VIEWER` requests administration
- **THEN** dashboard renders generic forbidden experience and no user data

#### Scenario: Administrator disables a user
- **WHEN** administrator selects disable
- **THEN** dashboard requires `ConfirmDialog`
- **AND** submits through server action without imperative router mutation

### Requirement: User lifecycle data and errors minimize sensitive information
Logs, metrics, safe operation status, audit projection, repair output, and errors SHALL NOT contain email, subject, provider username, tokens, temporary passwords, invitation links/codes, credentials, session identifiers/hashes, raw Cognito payloads, authorization headers, or idempotency keys. Administrator list/form responses MAY contain managed email only where required for identification/delivery.

#### Scenario: Provider returns detailed failure
- **WHEN** Cognito returns identity or delivery detail
- **THEN** public response maps to registered generic error
- **AND** diagnostics retain only opaque IDs, category, step, and correlation ID

#### Scenario: Pending operation is displayed
- **WHEN** administration UI reports pending repair
- **THEN** it exposes safe operation ID/type/state/age
- **AND** omits private operation input and provider identity data

### Requirement: User-management operations follow shared stage and cost policy
The implementation SHALL add no recurring repair compute by default and SHALL use bounded on-demand Cognito, AuthTable, AppTable projection, command, log, metric, and alarm resources. Stage persistence/disposal, ownership tags, budget alerts, retention, and alarm-count/cost caps SHALL follow the shared stage lifecycle and capacity/cost capabilities rather than defining conflicting thresholds here.

#### Scenario: No lifecycle work occurs
- **WHEN** no user lifecycle command or explicit repair runs
- **THEN** the change incurs no scheduled repair invocation

#### Scenario: Infrastructure is reviewed
- **WHEN** reviewers inspect user-management resources
- **THEN** stage lifecycle and cost/alarm decisions reference the shared policies
- **AND** no per-user alarm or duplicate stage budget is introduced
