## 1. Shared Authentication Contracts

These sections are reviewable internal implementation milestones. Completing or deploying a subset MUST NOT create optional authentication or a mixed protected-route posture; task 10 performs the atomic fail-closed cutover.

- [x] 1.1 Add `AUTHENTICATION_REQUIRED`/HTTP 401 and `AUTHORIZATION_DENIED`/HTTP 403 to the canonical Go error registry, mirror them in the dashboard error enum/messages, and extend registry-drift tests.
- [x] 1.2 Define canonical Go authentication types for immutable membership ID and subject, fixed tenant, named status/role, `AuthValidAfter`, version/timestamps, and normalized principal, authorizing only `DEFAULT`, `ACTIVE`, and `ADMIN` in v1.
- [x] 1.3 Extend internal AWS facades and fakes needed for strongly consistent auth-table reads and credentialed Cognito administration so domain/service code does not depend on raw SDK clients.
- [x] 1.4 Add shared auth constants and validators for `AuthTable` item keys, required membership shape, integer Unix-second `auth_time`/`AuthValidAfter`, role/status/tenant values, explicit expiry, and secret-safe audit event names with focused unit tests.

## 2. Stage-Aware AWS Authentication Infrastructure

- [x] 2.1 Provision a Cognito user pool explicitly on the Essentials feature plan with self-registration disabled, default Cognito email, verified-email recovery, explicit password/token settings, optional software-token MFA, tags, and persistent/ephemeral protection behavior sourced from `standardize-stage-resource-lifecycle`.
- [x] 2.2 Provision separate dashboard-server and no-secret direct-operator app clients with only the required custom password/challenge and refresh flows, refresh-token rotation, approved validity windows, and no managed-login domain or Amplify resources.
- [x] 2.3 Provision the on-demand `AuthTable` with `PK`/`SK`, TTL on `TTL`, tags, and lifecycle-mapped PITR/protection, and encode that it is the sole authority for membership, sessions, authentication transactions, revocation epoch, and future user lifecycle while `AppTable` remains monitoring-only.
- [x] 2.4 Define one active generation of an installation-specific 256-bit AES key through an SST Secret or bootstrap-safe secret reference, with least-privilege retrieval limited to dashboard auth code; generate and install/rotate it through a credentialed non-printing helper without persisting the value in source, scripts, shell history, files, logs, outputs, templates, or state-visible configuration, and document the transient process/CLI boundary.
- [ ] 2.5 Link Cognito, `AuthTable`, the AES secret reference, canonical dashboard origin, approved client IDs, and stage configuration only to functions that require them; export only non-secret identifiers needed by bootstrap and recovery.
- [ ] 2.6 Add infrastructure tests asserting Essentials, no public sign-up/managed login, default email, TOTP/refresh rotation, authoritative `AuthTable`, persistent-stage protection/PITR, ephemeral-stage clean deletion, secret grants/value non-exposure, no customer-managed KMS key, and required resource tags.

## 3. API Gateway And Go Authorization

- [x] 3.1 Create the API Gateway JWT authorizer for the Cognito issuer and both approved app clients, require `aws.cognito.signin.user.admin` through a protected-v1 route helper, and leave `GET /api/health` public without authorizer or scope.
- [x] 3.2 Add a Bruno/SST route guard test that fails when any `/api/v1/**` route lacks the shared JWT authorizer or required scope, or when `/api/health` requires either.
- [ ] 3.3 Implement the Cognito API Gateway claim adapter behind a Go `PrincipalResolver`, including explicit `token_use=access`, approved `client_id`, non-empty `sub`, normalized non-negative integer `auth_time`, diagnostic-only `iat`, and fail-closed claim validation.
- [ ] 3.4 Implement authoritative `AuthTable` membership lookup with `ConsistentRead=true`; validate immutable IDs, `DEFAULT`/`ACTIVE`/`ADMIN`, version/timestamps, and `auth_time > AuthValidAfter`, denying ceremonies at or before the boundary before normalizing a principal.
- [ ] 3.5 Resolve and authorize the principal once before monitor route dispatch, replace constructor-injected request tenant use with `principal.TenantID`, and prevent request-controlled tenant overrides.
- [ ] 3.6 Return envelope `AUTHENTICATION_REQUIRED` for Lambda-side principal failures and indistinguishable envelope `AUTHORIZATION_DENIED` for all membership denials without leaking membership state.
- [ ] 3.7 Add Go tests for access-token defense in depth, malformed/missing/negative/fractional `auth_time`, diagnostic-only `iat`, wrong client, strongly consistent `AuthTable` reads, no `AppTable` fallback, incomplete/non-active/conflicting membership, unsupported role, equality and older-than-boundary cases, immediate disable, fixed tenant propagation, and public-detail redaction.
- [ ] 3.8 Update all affected monitor API handler/repository tests and run `make test-go-all`, `make lint-go`, and `make build-go`.

## 4. Credentialed Administrator Bootstrap

- [ ] 4.1 Implement an AWS-credentialed Go administration command and Make target that normalizes an email, uniquely discovers or creates the Cognito user with invitation delivery suppressed, resolves immutable `sub`, conditionally creates the complete `ACTIVE` `DEFAULT`/`ADMIN` `AuthTable` membership with immutable `MembershipID`, `AuthValidAfter`, version, and timestamps, then sends or resends the default Cognito invitation only when activation remains required.
- [ ] 4.2 Make bootstrap retries preserve immutable membership identity, established credentials, and non-decreasing `AuthValidAfter`; avoid duplicate users or replacement invitations while failing loudly on ambiguous identity, tenant, role, status, or record-shape conflicts.
- [ ] 4.3 Emit secret-safe structured bootstrap outcomes including stage, acting AWS principal when available, target subject, desired authority, and correlation data without credential values.
- [ ] 4.4 Add fake-backed tests for first creation, complete retry, Cognito-only partial recovery, complete membership reconciliation, invitation-after-membership ordering, immutable-field preservation, concurrent invocation, ambiguous users, conflicting membership, and denied AWS operations.

## 5. Dashboard Auth Storage And Provider Adapters

- [ ] 5.1 Add server-only, provider-neutral TypeScript contracts and discriminated results for sign-in challenges, password recovery, TOTP, token refresh/revocation, auth transactions, sessions, and memberships, with Cognito as the sole adapter.
- [ ] 5.2 Implement Cognito I/O adapters under `lib/io/**` for password sign-in, `NEW_PASSWORD_REQUIRED`, TOTP association/verification/challenge, forgot/confirm password, refresh rotation, and revocation without exposing raw provider errors.
- [ ] 5.3 Implement `AuthTable` auth-transaction storage using 256-bit opaque references, hashed DynamoDB keys, AES-256-GCM encrypted challenge state bound to application/stage/record/active-generation context, flow/attempt constraints, 10-minute explicit expiry, TTL cleanup, and conditional single use.
- [ ] 5.4 Implement `AuthTable` dashboard-session creation and lookup using 256-bit opaque IDs, `__Host-bolt-session` attributes, SHA-256 DynamoDB lookup keys, context-bound AES-256-GCM encrypted token bundles, 12-hour explicit expiry, TTL-only cleanup, and the sole active key generation.
- [ ] 5.5 Implement conditional version/lease refresh coordination, winner token rotation, bounded loser rereads, stale-writer rejection, expired-lease takeover, and terminal-refresh session invalidation.
- [ ] 5.6 Implement idempotent logout/session invalidation and cookie expiry, including replacement of any prior auth transaction or authenticated session after successful authentication.
- [ ] 5.7 Add tests proving cookie entropy/attributes, absence of raw IDs in storage, explicit expiry before TTL deletion, authenticated encryption context/generation enforcement, token and secret-value non-exposure, fixation prevention, logout idempotency, key-loading/storage fail-closed behavior, and rotation invalidation of all sessions/transactions with no previous-key fallback or online re-encryption.
- [ ] 5.8 Add deterministic concurrency tests for one-winner refresh, waiting readers, stale winner, expired lease takeover, transient failure, rotated-token persistence, and invalid refresh termination.

## 6. Custom Next.js Authentication Flows

- [ ] 6.1 Create the public `(auth)` route group and custom sign-in page/handler outside the polling and protected operator shell.
- [ ] 6.2 Implement invitation activation and `NEW_PASSWORD_REQUIRED` pages using only the opaque server-held transaction reference and rotate into a new dashboard session on success.
- [ ] 6.3 Implement non-enumerating forgot-password acknowledgement and reset-password confirmation pages without placing email, code, or password material in URLs or returned component state.
- [ ] 6.4 Implement optional TOTP enrollment and challenge pages that show enrollment material only for immediate setup, verify before continuation, and never persist the TOTP secret in browser/RSC/application storage.
- [ ] 6.5 Map Cognito and validation outcomes to typed, operator-safe auth feedback without raw exception text or account-state disclosure.
- [ ] 6.6 Add page/action tests for established sign-in, invalid credentials, invitation activation, reused/expired/wrong-flow transaction, recovery enumeration resistance, reset, TOTP enrollment/challenge, established-session auth-page redirect, and zero `/api/v1/**` reads from public auth pages.

## 7. Dashboard Protection And API Token Forwarding

- [ ] 7.1 Guard the `(monitoring)` server layout, protected route handlers, and every server action independently with server-side session validation before rendering data or causing side effects.
- [ ] 7.2 Refactor dashboard API I/O through one server-only authenticated adapter that obtains a usable access token, forwards only Bearer access tokens, and keeps public health access separate.
- [ ] 7.3 Invalidate the local dashboard session on `AUTHORIZATION_DENIED`, handle API Gateway non-envelope 401 separately from application envelopes, and return safe sign-in-required navigation/results.
- [ ] 7.4 Implement centralized CSRF validation for every state-changing auth, session, and application action using canonical Origin plus effective Host/protocol and tightly bounded missing-Origin Fetch Metadata handling.
- [ ] 7.5 Implement one return-target sanitizer that accepts only normalized safe root-relative paths, rejects external/authority/encoded/control/backslash/auth-loop targets, and defaults to `/`.
- [ ] 7.6 Configure nonce-capable production CSP, frame/base/form/object restrictions, `nosniff`, restrictive referrer and permissions policies, and production HTTPS HSTS for auth and protected responses.
- [ ] 7.7 Add tests for unauthenticated render/action denial, no protected-data leak, Bearer forwarding without browser token exposure, non-active/old-epoch membership invalidation, CSRF Origin/Host/proxy cases, redirect attacks, and exact security headers.
- [ ] 7.8 Add static/serialization guard tests that reject Cognito tokens, challenge sessions, passwords/codes, TOTP secrets, cookie values, or session records in client modules, RSC props, browser storage, URLs, logs, and telemetry fixtures.

## 8. Authentication Audit And Telemetry

- [ ] 8.1 Implement the structured security-event catalog across dashboard auth handlers, API authorization, and bootstrap for all required success/failure, membership status, and `AuthValidAfter` authority-change events.
- [ ] 8.2 Add correlation propagation and secret-redaction tests proving audit/log output excludes credentials, codes, TOTP secrets, transaction/session identifiers and hashes, JWTs, refresh tokens, cookies, encryption-key material, request bodies, and unsafe provider payloads.
- [ ] 8.3 Add bounded CloudWatch metrics or metric filters for sign-in failure, recovery request, refresh failure, authorization denial, bootstrap failure, and auth storage/key-loading errors without per-user dimensions.
- [ ] 8.4 Configure finite log retention and actionable alarms for sustained refresh failure and auth infrastructure errors, with infrastructure tests for thresholds, dimensions, tags, and cost-bearing resources.

## 9. Direct Client, API Contract, And Operations Documentation

- [ ] 9.1 Update OpenAPI security schemes so every `/api/v1/**` operation requires Cognito Bearer access auth and `aws.cognito.signin.user.admin`, `/api/health` explicitly requires neither, and documented 401/403 behavior distinguishes Gateway from envelope responses.
- [ ] 9.2 Add Bruno/direct-client setup for Cognito password and `NEW_PASSWORD_REQUIRED` flows using the no-secret client and gitignored/local secret variables; attach access Bearer auth to all versioned requests and none to health.
- [ ] 9.3 Extend Bruno validation to reject committed auth secrets and assert the correct authenticated/public request split while preserving domain/operation tags and Purpose/Setup/Expected result blocks.
- [ ] 9.4 Document prerequisites, default-email quotas, initial bootstrap, subsequent AWS-admin invitations, membership status changes, the inclusive-denial `AuthValidAfter` rule, direct token refresh/reacquisition, logout/session behavior, and why `iat`, ID tokens, and dashboard cookies are invalid revocation/API authority.
- [ ] 9.5 Reference `standardize-stage-resource-lifecycle` for stage classification; document auth-specific persistent resource inventory/PITR recovery, ephemeral clean deletion, all-admin-lockout break glass, required IAM permissions, credentialed non-printing single-generation key rotation with intentional session/transaction invalidation, audit review, and deliberate retained-resource deletion.
- [ ] 9.6 Document the explicit no-optional-auth v1 cutover checklist, anonymous-client break, prerequisite gate, staging evidence, post-cutover checks, and rollback that preserves JWT enforcement and persistent-stage auth data.
- [ ] 9.7 Document cost sources and limits for Cognito Essentials/default email, strong-read on-demand DynamoDB/persistent-stage PITR, secret retrieval/AWS-managed encryption operations, and CloudWatch while recording that v1 adds no customer-managed KMS key, SES, custom DNS, NAT, always-on compute, or cross-region auth traffic.

## 10. End-To-End Verification And Cutover Readiness

- [ ] 10.1 Add integration coverage for custom sign-in through dashboard session to authenticated API, invite activation, recovery, optional TOTP, logout, explicit session expiry, and concurrent refresh.
- [ ] 10.2 Add integration coverage proving anonymous/wrong-token/missing-scope denial, valid ID-token rejection before Lambda, equality and older-ceremony denial at `auth_time <= AuthValidAfter`, later full-auth access, immediate non-active denial with an unexpired JWT, fixed `DEFAULT` scope, and unauthenticated `/api/health` success.
- [ ] 10.3 Run `make check-bruno`, all Go/dashboard/infra format, lint, typecheck, test, and build targets, and resolve every auth-related failure without weakening test assertions.
- [ ] 10.4 Inspect generated/deployed configuration for no Amplify, no Cognito managed-login domain, no self-registration, no custom DNS/SES dependency, correct persistent/ephemeral lifecycle, no secret value in outputs/templates/state-visible configuration, least-privilege grants, and authorizer plus scope coverage on every v1 route.
- [ ] 10.5 Execute and record the atomic staging cutover checklist with a fully shaped bootstrapped administrator, custom dashboard and Bruno evidence, Gateway ID-token rejection, non-active and advanced-epoch denial, public health, security headers, alarms/log redaction, rotation invalidation, and break-glass dry run before approving production cutover.
