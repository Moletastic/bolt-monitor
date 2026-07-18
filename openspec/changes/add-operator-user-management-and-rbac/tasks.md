## 1. Authentication-Owned AuthTable Contract

- [ ] 1.1 Extend canonical authentication membership types with `INVITED|ACTIVE|DISABLED`, `ADMIN|OPERATOR|VIEWER`, non-negative epoch-second `AuthValidAfter`, monotonic `Version`, and current lifecycle operation ID/type/state; reject unknown or malformed values closed.
- [ ] 1.2 Keep canonical membership at `PK=MEMBER#<sub>`, `SK=MEMBERSHIP`; add no AppTable membership, alternate identity item, second user authority, or membership migration.
- [ ] 1.3 Coordinate the authentication-owned initial bootstrap to write the forward-compatible `ACTIVE`/`ADMIN` membership and initialize `PK=TENANT#DEFAULT`, `SK=ADMIN_GUARD`; add no RBAC bootstrap command or one-time cutover path.
- [ ] 1.4 Add AuthTable item contracts for conditional email-digest claims and lifecycle operations containing desired state, fingerprint, current step, lease/version, attempts, next-attempt time, initiating actor, terminal outcome, and audit-projection state.
- [ ] 1.5 Add sparse AuthTable GSI1 keys for bounded `TENANT#DEFAULT#MEMBERS` listing/exact membership-ID discovery, GSI2 keys for bounded `LIFECYCLE#PENDING` due-operation queries, and GSI3 keys for bounded subject-session traversal.
- [ ] 1.6 Add infrastructure and repository tests for exact key/index shapes, sparse-index membership, strong base-table subject reads, bounded queries/cursors, no runtime scans, immutable subject/tenant, and no AppTable authority records.

## 2. auth_time Revocation And Authorization

- [ ] 2.1 Extend the Cognito access-token claim adapter to require `auth_time` as a non-negative integer epoch second and fail closed for missing, string, fractional, negative, overflow, or malformed values.
- [ ] 2.2 Implement authorization from a strongly consistent canonical AuthTable membership read requiring `ACTIVE`, `DEFAULT`, supported role, and `auth_time > AuthValidAfter`; never use `iat`, Cognito groups, AppTable, email, or username as authority.
- [ ] 2.3 Apply the same membership and `auth_time` check to dashboard session validation before stored token use so indexed deletion and Cognito sign-out are cleanup/defense in depth.
- [ ] 2.4 Implement the central role-to-permission table exactly as specified and unit-test every role/permission pair, including unknown-role/status denial.
- [ ] 2.5 Inventory every protected monitor API method/path and assign a typed permission covering reads, service/monitor configuration, manual runs, incidents, scheduler, channels/tests, policies, audit, and user administration.
- [ ] 2.6 Apply authorization before body parsing and domain/external effects on every protected route; add tests proving denied calls perform no writes, Cognito calls, session mutation, or audit mutation.
- [ ] 2.7 Add boundary tests for `auth_time` before/equal/after `AuthValidAfter`, missing/malformed claim, refreshed old family with later `iat`, later full authentication, disabled/invited membership, unknown authority, and immediate dashboard/API denial.

## 3. Transactional Membership And Last-Admin Safety

- [ ] 3.1 Implement versioned AuthTable membership transitions that atomically update current operation state and never move `AuthValidAfter` backward.
- [ ] 3.2 Implement active-admin guard transactions for activation, role transition, enable, and disable, requiring another active administrator before a transition out of `ACTIVE`/`ADMIN`.
- [ ] 3.3 Make missing/inconsistent guard fail closed and add an explicit report-first reconciliation mode that never infers authority from AppTable.
- [ ] 3.4 Add repository tests for stale versions, idempotent transitions, initial guard compatibility, invited/disabled admin exclusion, concurrent demotion/disable races, and no boundary rollback.

## 4. Durable Cognito Lifecycle Workflows

- [ ] 4.1 Extend shared AWS facades only with Cognito admin create/get, resend, enable, disable, and global-sign-out operations plus deterministic failure-injection fakes.
- [ ] 4.2 Implement hashed scoped-idempotency operation creation/loading, fingerprint conflicts, conditional leases, verified step checkpoints, bounded work/backoff metadata, terminal outcomes, and safe diagnostics in AuthTable.
- [ ] 4.3 Implement invite using AuthTable private intent/email claim, one persisted deterministic provider username, exact provider-user recovery, immutable subject binding, and one canonical `INVITED` membership.
- [ ] 4.4 Implement resend with one conditionally claimed delivery attempt and terminal `DELIVERY_UNKNOWN`; never resend automatically from retry or repair.
- [ ] 4.5 Implement first-full-auth activation and role assignment with membership versioning, active-admin maintenance, and idempotent audit-projection state.
- [ ] 4.6 Implement disable strictly as AuthTable `DISABLED` plus advanced boundary/guard first, bounded dashboard-session invalidation second, Cognito disable third, Cognito global sign-out fourth, and audit projection/completion last.
- [ ] 4.7 Implement enable as Cognito enable first and AuthTable `ACTIVE` transition second without reducing boundary; require a later full authentication ceremony for access.
- [ ] 4.8 Implement revoke-only as advanced AuthTable boundary with unchanged status/role, session invalidation, Cognito global sign-out, and audit projection; never disable membership or Cognito user.
- [ ] 4.9 Normalize provider already-desired/not-found outcomes safely and ensure retries cannot re-enable disabled access or skip unfinished ordered steps.
- [ ] 4.10 Add workflow tests with process-loss/failure injection around every AuthTable, session, Cognito, and projection step; cover repeated keys, conflicting inputs, concurrent leases, bounded session continuation, old-family refresh, and final convergence.

## 5. Explicit Repair And Observability

- [ ] 5.1 Add an AWS-credentialed administration command that repairs one operation ID through the same workflow service and conditional lease protocol used by API retries.
- [ ] 5.2 Add a bounded reconciliation mode that queries only GSI2 due operations with strict batch/time limits and opaque continuation; prohibit AuthTable and Cognito scans.
- [ ] 5.3 Expose safe pending/retryable operation ID/type/step/age and audit-projection state in admin responses while retaining all authority and repair state in AuthTable.
- [ ] 5.4 Emit PII-free command/request metrics and logs for outcomes, observed pending count/oldest age, retry categories, and projection lag using bounded dimensions.
- [ ] 5.5 Add repair tests for exact-operation repair, due-query continuation, lease expiry/concurrency, no-work behavior, ambiguous resend non-replay, projection-only repair, and secret-safe output.
- [ ] 5.6 Assert infrastructure provisions no EventBridge repair schedule or recurring repair Lambda; document that recurring compute requires measured evidence and a follow-on approved change.

## 6. Safe AppTable Audit Projection

- [ ] 6.1 Implement idempotent lifecycle audit projection at `PK=TENANT#DEFAULT`, `SK=AUDIT#USER_LIFECYCLE#<reverse-time>#<event-id>` containing only approved non-PII actor/target/outcome fields.
- [ ] 6.2 Keep `AuditProjectionState` and repair checkpoints in AuthTable; make projection failure repairable without replaying completed authority or Cognito transitions.
- [ ] 6.3 Propagate opaque actor membership context through existing authenticated mutations and require operator attribution or explicit system origin in application audit tests.
- [ ] 6.4 Implement bounded newest-first AppTable lifecycle-audit reads with cursor continuation for `GET /api/v1/admin/audit-events`.
- [ ] 6.5 Add tests proving AppTable loss/lag cannot grant/deny access or drive repair, projection retries create one event, and projections exclude subject, email, provider data, token/session material, idempotency keys, and private operation input.

## 7. User Management API And Contracts

- [ ] 7.1 Add admin handlers and SST routes for bounded user list, invite, resend, role assignment, enable, disable, revoke sessions, and lifecycle audit list using standard envelopes and typed errors.
- [ ] 7.2 Require and validate `Idempotency-Key` on cross-system mutations, expose safe pending operation status, and use generic 401/403/409/503 outcomes without membership/provider disclosure.
- [ ] 7.3 Resolve route `membershipId` through exact bounded GSI1 discovery, then strongly re-read canonical subject-keyed membership before every mutation.
- [ ] 7.4 Add handler tests for success, pagination, validation, duplicate/conflicting requests, invalid lifecycle states, last-admin conflict, dependency failure, pending repair status, and response/log redaction.
- [ ] 7.5 Update OpenAPI and Bruno requests/tags/docs for every new route and `auth_time`/error semantics, then run route coverage validation.

## 8. Dashboard Administration

- [ ] 8.1 Add typed dashboard adapters and `Result`-based server actions for user lifecycle and lifecycle-audit endpoints without catching outside `lib/io`.
- [ ] 8.2 Add an admin-only Administration/Users navigation entry and bounded user list/invite views with fixed role, lifecycle/invitation status, timestamps, pagination, safe current operation state, and state-appropriate actions.
- [ ] 8.3 Add role, resend, enable, disable, and revoke forms using generic safe feedback and shared `ConfirmDialog` for access-removing actions.
- [ ] 8.4 Apply permission capabilities to existing dashboard affordances while preserving API authority and Link/form router convention.
- [ ] 8.5 Add dashboard tests for admin flows, non-admin data isolation, role-scoped visibility, confirmations, pending/repair messaging, session invalidation after boundary change, safe errors, and no imperative router calls.

## 9. Stage, Security, Cost, And End-To-End Verification

- [ ] 9.1 Reference shared persistent/ephemeral stage lifecycle, ownership/tagging, log retention, budget notifications, and alarm-count/cost caps from `standardize-stage-resource-lifecycle` and `establish-data-recovery-and-capacity-guardrails`; do not duplicate thresholds or budgets.
- [ ] 9.2 Document incremental Cognito, AuthTable strong-read/transaction/index, AppTable projection, command-invoked repair, and bounded telemetry costs, explicitly recording no scheduled repair compute by default.
- [ ] 9.3 Add end-to-end role matrix tests through direct APIs and dashboard for all roles/permissions, Cognito-group disagreement, immutable subject/tenant, and AppTable projection disagreement.
- [ ] 9.4 Add end-to-end disable/revoke tests proving inclusive `auth_time` boundary denial, missing/malformed fail-closed behavior, old-family refresh denial despite later `iat`, later full-auth eligibility by status, required ordering, and explicit repair convergence.
- [ ] 9.5 Add automated redaction tests over envelopes, UI, logs, metrics, audit projection, operation status, and repair output for email outside approved admin forms, subject, provider username/payload, tokens, temporary passwords, invitation material, credentials, session identifiers/hashes, authorization headers, idempotency keys, and private intent.
- [ ] 9.6 Verify only the authentication-owned initial bootstrap exists and produces the canonical membership/guard; verify there is no AppTable membership, one-time membership migration, duplicate identity item, repair schedule, or always-on repair Lambda.
- [ ] 9.7 Run `openspec validate add-operator-user-management-and-rbac --strict`, `make test-go-all`, `make lint-go`, `make build-go`, `make lint-dashboard`, `make check-dashboard`, `make test-dashboard`, `make check-infra`, `make check-bruno`, and the production dashboard build; resolve failures before rollout.
