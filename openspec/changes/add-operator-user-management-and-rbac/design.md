## Context

`add-single-tenant-operator-authentication` owns Cognito, the retained `AuthTable`, opaque dashboard sessions, per-request strongly consistent membership authorization, and the sole initial-administrator bootstrap. This change extends that foundation for later-user management and fixed RBAC. It does not move or duplicate membership into `AppTable`, add another bootstrap, or migrate an initial membership between authorities.

Cognito and DynamoDB cannot participate in one transaction. Lifecycle workflows therefore need durable, idempotent checkpoints. Revocation must use the Cognito access-token `auth_time` claim because `iat` changes when an old refresh-token family produces a new access token and does not prove a new authentication ceremony.

## Goals / Non-Goals

**Goals:**

- Keep AuthTable as the permanent user, membership, revocation, session, lifecycle-operation, and repair authority.
- Extend the authentication membership record forward-compatibly for three statuses, three fixed roles, revocation boundary, version, and current operation state.
- Enforce one permission matrix at every API route and reflect it in dashboard affordances.
- Give administrators post-bootstrap invite/resend, list, role, enable/disable, revoke, and explicit repair controls.
- Fail closed through partial Cognito/AuthTable failures and converge idempotently without recurring compute by default.
- Protect the last active administrator transactionally.
- Keep safe lifecycle audit reads in the existing AppTable audit surface without moving authority there.
- Keep operational cost and alarms within shared stage policies.

**Non-Goals:**

- A second bootstrap command, one-time membership migration, AppTable membership, duplicate identity records, or a separate user authority.
- Multiple tenants, tenant switching, self-signup, custom roles, delegated role design, or enterprise federation.
- Machine-to-machine identities, API keys, service accounts, Cognito groups as authority, or email changes.
- Exactly-once Cognito email delivery.
- A scheduled repair Lambda or EventBridge rule without measured evidence and a follow-on approved change.

## Decisions

### 1. AuthTable remains the sole authority

The authentication change's canonical membership item remains at `PK=MEMBER#<cognito-sub>`, `SK=MEMBERSHIP` and is read with `ConsistentRead=true` on every protected API request. RBAC extends that item with:

| Attribute | Contract |
|---|---|
| `Subject` | Immutable Cognito `sub`; duplicates the key value only as a typed attribute, not as another identity record |
| `MembershipID` | Immutable opaque administrator-facing identifier |
| `TenantID` | Always `DEFAULT` |
| `Status` | `INVITED`, `ACTIVE`, or `DISABLED` |
| `Role` | `ADMIN`, `OPERATOR`, or `VIEWER` |
| `AuthValidAfter` | Non-negative integer epoch seconds; authentication with `auth_time <= AuthValidAfter` is denied |
| `Version` | Monotonically increasing optimistic-concurrency version |
| `LifecycleOperationID` | Current operation ID or absent |
| `LifecycleOperationType` | Current lifecycle operation type or absent |
| `LifecycleOperationState` | `PENDING`, `RETRYABLE`, `SUCCEEDED`, `FAILED`, or `DELIVERY_UNKNOWN`, or absent when no operation has been attached |
| timestamps | Safe created/updated and lifecycle timestamps |

The authentication-owned bootstrap remains the only initial bootstrap. It creates the first canonical `ACTIVE`/`ADMIN` membership and initializes the AuthTable active-admin guard using this forward-compatible schema. This RBAC change manages later users only and adds no migration, cutover bootstrap, duplicate membership key, or AppTable backfill.

Email is mutable PII and never grants access. A conditional digest claim supports invite uniqueness, but it is neither an identity nor authorization record. Cognito remains credential and authentication-ceremony authority; AuthTable remains application access and lifecycle authority.

### 2. AuthTable access paths are explicit and bounded

AuthTable retains `PK`/`SK`, TTL on `TTL`, and adds three sparse indexes. Every runtime query sets an explicit limit and returns opaque continuation where applicable.

| Purpose | Item/index and exact key |
|---|---|
| Authorize by subject | Base-table `GetItem(PK=MEMBER#<sub>, SK=MEMBERSHIP, ConsistentRead=true)` |
| Claim canonical email | Base-table `Put/GetItem(PK=EMAIL#<sha256(canonical-email)>, SK=CLAIM)`; item points to one invite operation before subject resolution and then one membership ID |
| Active-admin invariant | Base-table `Get/UpdateItem(PK=TENANT#DEFAULT, SK=ADMIN_GUARD)` in the same transaction as membership transitions |
| Load operation/idempotency | Base-table `GetItem(PK=LIFECYCLE_OP#<sha256(operation-scope and idempotency-key)>, SK=META)` |
| List users | `GSI1PK=TENANT#DEFAULT#MEMBERS`, `GSI1SK=MEMBER#<membership-id>` on membership items; query exact partition with bounded limit/cursor |
| Resolve administrator membership ID | Same GSI1 with exact partition and exact sort-key equality; result is re-read from the base table before mutation |
| List repairable operations | `GSI2PK=LIFECYCLE#PENDING`, `GSI2SK=<zero-padded-next-attempt-epoch>#<operation-id>` on non-terminal operation items; bounded query through `<= now`; terminal operations remove these keys |
| Invalidate subject sessions | `GSI3PK=MEMBER#<sub>#SESSIONS`, `GSI3SK=<zero-padded-expiry-epoch>#<session-hash>` on dashboard-session items; bounded query/delete traversal |

Lifecycle operation items contain operation ID/type, target membership ID or private invite intent, normalized-input fingerprint, desired state, current step, lease owner/expiry, version, attempts, next-attempt time, initiating actor, projection state, and terminal outcome. Private canonical email is restricted to the email claim/invite operation needed for delivery. Terminal operation items use the shared retention classification; authority-bearing membership, guard, and email claim items have no operational TTL.

No AppTable item can satisfy authorization, last-admin checks, idempotency, or repair. GSI reads are used only for discovery; security-changing writes re-read canonical base-table items and use conditions/transactions.

### 3. Authorization uses Cognito `auth_time`, never `iat`

The claim adapter requires access-token `auth_time` to be a non-negative integer epoch second. Missing, string-shaped, fractional, overflowing, or otherwise malformed values fail closed before route work. After authentication, authorization strongly reads membership and requires `Status=ACTIVE`, `TenantID=DEFAULT`, a known fixed role, and `auth_time > AuthValidAfter`.

The boundary is inclusive on denial: authentication ceremonies at or before `AuthValidAfter` are denied. Only a full Cognito authentication ceremony with `auth_time` strictly greater than the boundary can restore eligibility, and only while membership is active. Refreshing an older token family does not help: even when the refreshed token has a later `iat`, it retains the family's old `auth_time` and remains denied. `iat` may still support token diagnostics but never revocation authorization.

Dashboard session validation applies the same membership and `auth_time` rule before using its token bundle. This makes the AuthTable boundary immediately authoritative even while indexed session cleanup or Cognito propagation is incomplete.

### 4. Central application authorization policy

Define typed permissions rather than scattering role checks: `read`, `configure-resources`, `operate-incidents`, `run-manual-check`, `manage-scheduler`, `manage-notification-config`, and `manage-users`. One role-to-permission table implements the matrix in `operator-rbac`; every protected monitor API route declares and checks one permission after authentication but before body parsing or side effects. The dashboard uses safe capabilities for affordances, but direct API enforcement remains authoritative.

Cognito groups, email, username, `iat`, and custom claims never select role or tenant. Unknown roles and statuses fail closed.

### 5. Durable lifecycle workflows converge synchronously or by explicit repair

Each cross-system command conditionally creates or loads an AuthTable operation using the hashed scoped idempotency key. Reuse with a different fingerprint is a conflict. The request acquires a short conditional lease, verifies observed state before each step, checkpoints completion, and continues while its execution budget allows. Retryable failures leave the operation `RETRYABLE`, visible with safe operation ID/type/step/age and no PII. Retrying the same API command or invoking the explicit administration repair command resumes it.

Provide an AWS-credentialed, auditable command that can repair one operation ID or query a bounded page of due GSI2 operations and reconcile them. It uses the same workflow service and conditional leases as API retries. It never scans AuthTable or the Cognito user pool. There is no EventBridge schedule or repair Lambda by default. Recurring repair requires evidence that explicit repair leaves unacceptable pending age/volume and a separate approved change that fits shared cost and alarm caps.

Provider usernames for invites are generated once and stored in the AuthTable operation before `AdminCreateUser`. A lost create response is repaired by `AdminGetUser` for that exact username, never by creating a second identity. Membership is created only after immutable `sub` recovery.

Cognito resend has no idempotency token or delivery receipt. The operation claims one delivery attempt before calling Cognito. Loss in the ambiguity window produces `DELIVERY_UNKNOWN`; neither API retry nor repair sends again. An administrator must start a new resend operation with a new idempotency key.

### 6. Lifecycle ordering is fail closed

- Invite: persist operation/private intent and email digest claim, create or recover the deterministic Cognito user, recover immutable `sub`, create one `INVITED` membership in AuthTable, then project one safe audit outcome.
- Activation: on the first completed full Cognito authentication, transactionally change `INVITED` to `ACTIVE`, update the AuthTable admin guard when applicable, clear/update current lifecycle state, and project the effective audit event.
- Role change: transactionally update AuthTable membership/version and admin guard, enforcing last-admin safety, then project audit.
- Disable: in one AuthTable transaction mark membership `DISABLED`, advance `AuthValidAfter`, update version/current operation, and enforce/update the admin guard; next invalidate dashboard sessions through the subject-session access path; next call Cognito disable; next call Cognito global sign-out; finally project audit and complete the operation. Any partial state after the first transaction remains denied.
- Enable: call Cognito enable first; then transactionally set membership `ACTIVE` without reducing `AuthValidAfter`, update the guard/version, and project audit. Access requires a later full authentication with `auth_time > AuthValidAfter`.
- Revoke sessions only: advance `AuthValidAfter` and operation/version while leaving `Status` and `Role` unchanged; invalidate dashboard sessions; call Cognito global sign-out; project audit. It never disables membership or the Cognito user.

Advancing a boundary means atomically setting it to at least the current epoch second and strictly greater than its prior value. Repeated operations never move it backward. Session invalidation is idempotent: authorization denies by the boundary first, then the workflow removes all currently discoverable subject sessions in bounded pages and records continuation until complete. Cognito not-found/already-disabled/already-enabled outcomes are normalized against desired state; transient failures remain repairable. No retry path re-enables a disabled membership.

### 7. Last-active-admin guard is transactional

An active administrator is exactly an AuthTable membership with `Status=ACTIVE` and `Role=ADMIN`. Invitations and disabled memberships do not count. The authentication-owned initial bootstrap initializes `PK=TENANT#DEFAULT`, `SK=ADMIN_GUARD` with count one in the same authority setup. Every later transition out of the set conditionally requires count greater than one and decrements it in the same AuthTable transaction as membership/version changes; every transition into the set increments it. Missing or inconsistent guard state fails closed and is reported for explicit reconciliation, never reconstructed from AppTable.

### 8. API and dashboard surfaces

Add these admin routes under existing envelope and typed-error conventions:

- `GET /api/v1/admin/users`
- `POST /api/v1/admin/users/invitations`
- `POST /api/v1/admin/users/{membershipId}/resend-invitation`
- `PATCH /api/v1/admin/users/{membershipId}/role`
- `POST /api/v1/admin/users/{membershipId}/enable`
- `POST /api/v1/admin/users/{membershipId}/disable`
- `POST /api/v1/admin/users/{membershipId}/revoke-sessions`
- `GET /api/v1/admin/audit-events`

Cross-system mutations require `Idempotency-Key`. Responses may expose a safe operation ID and pending/retryable status so incomplete work is observable. Explicit operational repair remains the AWS-credentialed command, not a public repair endpoint.

Add an Administration/Users dashboard route for administrators. Use server-side fetches/actions, `<Link>`, forms, shared `ConfirmDialog` for disable/revoke, and existing `Result` conventions; do not add imperative router calls. Errors distinguish authentication, forbidden, validation, conflict/last-admin/invalid state, and retriable dependency failure with generic safe details.

### 9. AppTable stores only the safe lifecycle audit projection

AuthTable operation records and membership transitions are authoritative. Once an effective lifecycle outcome is durable, write an idempotent non-PII projection to AppTable using `PK=TENANT#DEFAULT`, `SK=AUDIT#USER_LIFECYCLE#<reverse-time>#<event-id>`. The projection contains event ID/type/time, opaque actor membership ID or system origin, opaque target membership ID, safe prior/new status or role, correlation ID, outcome, and repair completion origin. It contains no subject, email, provider username/payload, token data, session key/hash, idempotency key, or private operation input.

`GET /api/v1/admin/audit-events` reads this bounded newest-first AppTable partition range so lifecycle events join the existing audit read/storage boundary. Projection failure leaves `AuditProjectionState=PENDING` in the AuthTable lifecycle operation and is repaired idempotently by API retry or the explicit repair command. AppTable loss or lag never changes authorization or workflow truth.

Existing application mutation audits gain opaque actor membership ID or explicit system origin. User-management list/form responses remain the only application responses allowed to contain managed email.

### 10. Stage lifecycle, telemetry, cost, and alarms reuse shared policy

Emit bounded non-PII signals for lifecycle outcomes, pending/retryable operation count and oldest age observed by requests or explicit repair, repair command outcomes, authorization denial categories, and audit-projection lag. Do not create per-user dimensions. Use existing shared log retention, persistent/ephemeral resource lifecycle, stage ownership/tagging, budget notifications, and alarm-count/cost caps from `standardize-stage-resource-lifecycle` and `establish-data-recovery-and-capacity-guardrails`; this change does not restate thresholds or create one alarm per workflow.

Incremental costs are Cognito admin calls/MAUs, AuthTable strongly consistent reads and transactions/index storage, AppTable audit projection writes/storage, command-invoked repair, and bounded CloudWatch telemetry. There is no scheduled invocation cost by default and no new always-on compute.

## Risks / Trade-offs

- [Cognito resend cannot prove delivery] -> Claim one attempt and use `DELIVERY_UNKNOWN`; require a new command for another send.
- [AuthTable and Cognito can drift] -> Deny through AuthTable first, persist desired state/checkpoints, and converge via same-key retry or explicit repair.
- [GSI session discovery is eventually consistent] -> Enforce `AuthValidAfter` on every API and dashboard session use before cleanup, then repeat bounded indexed cleanup until complete.
- [A refreshed token has a new `iat`] -> Ignore `iat` for revocation and compare the required `auth_time` to `AuthValidAfter`.
- [Admin guard can drift through defects/manual edits] -> Restrict writes to AuthTable transactions and provide report-first explicit reconciliation; never infer authority from AppTable.
- [Audit projection can lag] -> Keep projection state in AuthTable and repair idempotently; never block fail-closed access removal on AppTable availability.
- [Explicit repair can leave work pending longer than a schedule] -> Expose pending state/age and document response ownership; add recurring compute only with measured evidence and approved cost/alarm impact.

## Migration Plan

1. Implement the forward-compatible AuthTable membership schema, indexes, operation items, session subject index, and active-admin guard as part of the authentication dependency before its initial bootstrap is used. Do not create AppTable membership data or a separate migration.
2. Run only the authentication-owned initial bootstrap and verify its canonical `ACTIVE`/`ADMIN` AuthTable membership and initialized admin guard.
3. Deploy fixed-role authorization, lifecycle APIs, explicit repair command, audit projection, and dashboard administration in a non-production persistent stage.
4. Verify all three roles, `auth_time` parsing/boundaries, old-family refresh denial, later full-auth acceptance, disable/revoke ordering, session cleanup continuation, last-admin races, partial failures, and explicit repair.
5. Update OpenAPI/Bruno and enable route-wide permission enforcement only after the authentication dependency and initial membership/guard verification pass.
6. Follow the shared stage lifecycle and cost/alarm review gates. Roll back UI/API additions without deleting Cognito users, AuthTable authority/operation state, AppTable audit evidence, or weakening the last-admin guard; prefer forward repair for partial lifecycle operations.

## Open Questions

None. Recurring repair compute is explicitly deferred unless production evidence demonstrates that observable, operator-owned repair cannot meet lifecycle recovery needs.
