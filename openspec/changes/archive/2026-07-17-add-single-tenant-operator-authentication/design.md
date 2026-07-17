## Context

Today API Gateway invokes `services/monitor-api` without an authorizer, and `monitorHandler` receives `DEFAULT` at process construction. The dashboard performs server-side fetches to that public API but has no operator identity or session. `GET /api/health` is intentionally public. This change is the v1 security boundary before operators place real monitoring configuration and incident data in the deployment.

The product model is deliberately small: one deployment owns exactly tenant `DEFAULT`, while the deployment can have N human operators. Cognito identifies a human; a separate application membership in the dedicated `AuthTable` decides whether that immutable Cognito subject is currently active, whether its authentication ceremony is new enough, and what application authority it has. Those states cannot be collapsed because disabling a Cognito user or waiting for JWT expiry does not provide the required next-request authorization semantics. `AppTable` remains the authority for monitoring data and must never become a membership or user-lifecycle authority.

The implementation spans SST infrastructure, the Next.js server runtime, the Go API, AWS-credentialed operational tooling, OpenAPI, and Bruno. It must follow existing repository boundaries: Go domain/service code behind internal facades, TypeScript I/O exceptions contained under `lib/io`, uniform Lambda response envelopes, service-owned tenant scope, and no secret material in UI state.

## Goals / Non-Goals

**Goals:**

- Provide invite-only Cognito Essentials authentication for multiple operators in one `DEFAULT` tenant.
- Keep all auth UI custom and first-party, including temporary-password, recovery, and optional TOTP paths.
- Give browsers a revocable opaque dashboard session while keeping Cognito tokens server-side.
- Authenticate every `/api/v1/**` request at API Gateway with an access-token scope and authorize current membership and authentication epoch in Go on every request.
- Make bootstrap, disable/recovery, cutover, retention, telemetry, and direct-client operation explicit and testable.
- Keep provider-specific code behind identity, token, principal, and storage boundaries without implementing a second provider.

**Non-Goals:**

- User-management UI or self-service registration.
- Roles beyond the named `ADMIN` value, resource-level permissions, or general RBAC.
- Machine-to-machine credentials, service accounts, API keys, or workload identity.
- Social, SAML, OIDC, custom, or non-Cognito identity-provider implementations.
- Migrating static users or existing anonymous clients automatically.
- Authentication-page responsive-layout work beyond preserving functional existing layout behavior.
- Multi-tenant selection, tenant claims, tenant switching, or accepting tenant input from clients.

## Decisions

### 1. Cognito Essentials is the only v1 identity implementation

Provision one user pool with the Essentials feature plan, administrator-only user creation, email sign-in/alias behavior, default Cognito email delivery, account recovery by verified email, MFA mode `OPTIONAL`, and software-token MFA enabled. Do not create a Cognito domain. Password policy and token validity are explicit infrastructure settings rather than console defaults.

Create two user-pool app clients:

- A dashboard server client for the custom Next.js handlers. It supports the password/challenge and refresh flows needed by those handlers and keeps any client secret server-only.
- A no-secret direct-operator client for Bruno and documented Cognito API authentication. Its client ID is public configuration, but passwords and returned tokens remain local secrets.

API Gateway accepts access JWTs from both approved clients and each protected route requires Cognito's `aws.cognito.signin.user.admin` access-token scope, which the approved user-pool API authentication flows support. This gives the Gateway an access-token discriminator and causes ID tokens, which do not carry the required access scope, to fail before Lambda invocation. The Go resolver still enforces `token_use=access` and an allowlisted `client_id` as defense in depth rather than trusting scope alone.

Alternatives considered:

- Cognito managed login was rejected because the required product experience uses custom Next pages.
- Amplify was rejected because it adds a browser-oriented abstraction and token custody that conflict with server-only sessions.
- A custom user/password store was rejected due to credential-handling risk and unnecessary operational burden.
- SES and custom DNS were deferred; Cognito default email is enough for the initial low-volume operator set, subject to its delivery quotas.

### 2. Identity flows are server-owned state machines

Define a provider-neutral TypeScript identity interface with operations equivalent to begin sign-in, answer new-password challenge, associate/verify software token, answer MFA, begin/confirm recovery, and refresh/revoke. Cognito request/response names remain inside `apps/dashboard/lib/io/auth/cognito-*`; route handlers and server actions consume discriminated application results.

Custom pages live in an `(auth)` route group outside `(monitoring)`. Challenge continuation is stored as a short-lived encrypted auth-transaction item. The browser receives a separate opaque `__Host-` transaction cookie or one-time reference, never Cognito's challenge `Session`. Transactions are hashed for lookup, bound to one flow and intended next challenge, expire after 10 minutes, enforce a bounded attempt count, and are conditionally consumed. Passwords, codes, and TOTP secrets are accepted only by server actions and are never reflected in URLs or returned state.

Alternatives considered:

- Putting Cognito challenge sessions in hidden fields or encrypted client tokens was rejected because it broadens browser and RSC secret exposure.
- Using middleware as the complete session authority was rejected because edge-runtime constraints and middleware bypass mistakes make protected server layouts/actions the stronger enforcement point.

### 3. A dedicated DynamoDB auth table is the permanent auth-state authority

Add an on-demand `AuthTable`, separate from `AppTable`, with `PK` and `SK`, native TTL on `TTL`, and server-side encryption. `AuthTable` is the permanent single authority for membership, dashboard sessions, authentication transactions, the per-member revocation epoch, and future user lifecycle state. `AppTable` stores monitoring-domain records only and must not duplicate, index as authority, or fall back to membership state. Separating auth state keeps access policy narrow and avoids adding volatile session traffic and security recovery concerns to the monitoring entity table.

Initial item families are:

| Item | PK | SK | Important attributes |
|---|---|---|---|
| Membership | `MEMBER#<cognito-sub>` | `MEMBERSHIP` | immutable `MembershipID`, immutable `Subject`, `TenantID=DEFAULT`, `Status=ACTIVE`, `Role=ADMIN`, `AuthValidAfter`, `Version`, `CreatedAt`, `UpdatedAt` |
| Dashboard session | `SESSION#<sha256(raw-id)>` | `META` | encrypted token bundle, `ExpiresAt`, `TTL`, version, refresh lease fields, subject, safe timestamps |
| Auth transaction | `AUTH_TX#<sha256(raw-id)>` | `META` | encrypted challenge state, flow/challenge discriminator, attempts, `ExpiresAt`, `TTL`, version |

`ExpiresAt` is authoritative and checked in application code on every use. `TTL` is merely eventual cleanup and can use the same epoch plus a short cleanup buffer. Sessions have a 12-hour absolute application lifetime, bounded by Cognito refresh-token validity; authentication transactions have a 10-minute lifetime. These values are named configuration constants and can be revised only with corresponding tests and operational documentation.

Membership reads in the API use `GetItem` with `ConsistentRead=true`. Membership has no TTL. `MembershipID` and `Subject` are immutable after creation; email is not an authorization key. `AuthValidAfter` is a non-negative integer Unix epoch second and is the sole application revocation boundary for Cognito authentication ceremonies. The initial write sets it to no earlier than membership creation, before invitation delivery. V1 authorizes only `Status=ACTIVE`, `TenantID=DEFAULT`, and `Role=ADMIN`; any other status or unsupported value fails closed. Future RBAC and user-lifecycle work extends this versioned record in place, without migrating membership authority or creating a second bootstrap path.

The Cognito adapter normalizes `auth_time` as an integer NumericDate. Authorization requires `auth_time > AuthValidAfter`; authentication ceremonies at or before the boundary are denied. Missing, fractional, string, negative, overflowed, or otherwise malformed `auth_time` fails authentication closed. `iat` can be captured as bounded diagnostic metadata but cannot satisfy, replace, or weaken the revocation check. A disable operation sets a non-active status for immediate next-request denial. Re-enabling or explicitly revoking existing ceremonies advances `AuthValidAfter` to the current documented whole-second boundary; only a later full Cognito authentication ceremony can restore eligibility while membership remains active.

Alternatives considered:

- JWT-only authorization was rejected because disabled membership must take effect on the next request.
- Storing sessions in signed cookies was rejected because it exposes token-bearing state to the browser and weakens immediate revocation.
- Reusing `AppTable` was rejected to preserve least privilege, independent retention/recovery, and clearer cost/traffic ownership.

### 4. Browser sessions use opaque identifiers and encrypted token bundles

Generate 32 random bytes with the platform cryptographic RNG and encode them without reducing entropy. Set the raw value only in `__Host-bolt-session` with `Secure`, `HttpOnly`, `SameSite=Lax`, `Path=/`, and no `Domain`. DynamoDB stores only SHA-256 of the raw value. A successful authentication always creates a fresh identifier and deletes any previous authenticated session or auth transaction reference to prevent fixation.

Encrypt the complete Cognito token bundle, not only the refresh token, with AES-256-GCM using one active installation-specific 256-bit key generation supplied through an SST Secret or bootstrap-safe secret reference. Bind ciphertext with authenticated additional data including application, stage, record kind, active generation identifier, and hashed record ID. Only the Next server role can retrieve the secret and read session records. API and browser roles cannot read either resource. Plaintext tokens and key material exist only during the server I/O operation and are never logged.

The secret value is generated at runtime by a credentialed non-printing helper and installed through SST Secret management or a bootstrap-safe equivalent without writing the value to source, scripts, shell history, files, logs, stack outputs, generated templates, or state-visible configuration. The value may exist transiently in process memory and the pinned SST CLI invocation boundary, which implementation tests and runbooks must treat as sensitive. At-rest protection uses the secret facility's AWS-managed encryption and does not introduce a customer-managed KMS key. There is no current/previous key ring and no online re-encryption path. Rotation replaces the sole active generation and intentionally invalidates every extant dashboard session and authentication transaction; operators expire cookies and require fresh sign-in after rotation.

Logout conditionally removes the session and expires the cookie; absence is success. Session records remain explicitly invalid after `ExpiresAt` even if DynamoDB has not deleted them.

Alternatives considered:

- DynamoDB encryption at rest alone was rejected because a table read would reveal reusable refresh tokens.
- Hashing refresh tokens was rejected because the server must present them to Cognito.

### 5. Refresh rotation uses a conditional single-writer protocol

Configure Cognito refresh-token rotation with a short grace period. Each session record has a monotonically increasing `Version`, `RefreshOwner`, and `RefreshLeaseUntil`.

When the access token is near expiry:

1. A request conditionally acquires a short refresh lease only for the version it read and only if no live lease exists.
2. The winner decrypts the current token bundle, calls Cognito, and conditionally writes the rotated bundle while incrementing `Version` and clearing the lease.
3. Losers wait with bounded jitter, strongly reread the item, and use the winner's newer bundle. They do not call Cognito with the old refresh token.
4. A stale winner cannot overwrite a newer version. An expired lease can be acquired by a later request.
5. Terminal Cognito refresh errors conditionally invalidate the session; transient provider/storage errors fail the current request without pretending the session is valid.

The API access token is refreshed before forwarding only when its remaining lifetime is below a small named skew. This avoids refreshing on every request while retaining enough time for API Gateway validation.

Alternatives considered:

- Uncoordinated refresh plus Cognito grace alone was rejected because concurrent server renders can race repeatedly and stale writes can strand the active token family.
- A distributed lock service was rejected because a conditional DynamoDB protocol is sufficient at this scale and adds no idle resource.

### 6. Protected dashboard boundaries validate server-side and forward access tokens

The `(monitoring)` server layout validates the session before protected children execute and redirects invalid sessions to `/login` with a sanitized root-relative return target. Every server action and route handler also invokes the same guard; layout validation is not treated as authorization for a later mutation. Authentication routes redirect an already valid session to `/`.

Refactor API fetches through one server-only authenticated I/O adapter that obtains a usable access token and sets `Authorization: Bearer ...`. Public health access, if used by the dashboard, remains a distinct unauthenticated adapter. No client component receives a token. An application-envelope `AUTHORIZATION_DENIED` invalidates the local session before returning an authentication-required result.

Fallible Cognito, secret/key-loading, DynamoDB, cookie, cryptographic, and fetch operations remain in `lib/io/**` and return `Result<T, ApiError>` or a specific auth error to the rest of `lib/**`, preserving the repository TypeScript error boundary rule.

Alternatives considered:

- Browser-to-API Bearer calls were rejected because they require browser token storage and expose tokens to client code.
- Relying only on Next route middleware was rejected because server actions and direct route invocation still require local checks.

### 7. CSRF, redirect, and response hardening are centralized

For every state-changing server action or route, compare normalized `Origin` and effective host/protocol with one required `DASHBOARD_ORIGIN`. Trust forwarded host/protocol only in the known SST/CloudFront runtime path. Requests with mismatches fail before side effects. For browser cases where `Origin` is legitimately absent, require same-site Fetch Metadata plus matching effective host; do not use absence as general permission. SameSite cookies are defense in depth, not the sole CSRF control.

One redirect sanitizer accepts only normalized root-relative application paths, rejects `//`, backslashes, control/encoded authority tricks, and auth-loop paths, and defaults to `/`. It is used for login, challenge completion, and expiry.

Set security headers centrally for auth and protected pages. Production CSP uses per-response nonces for required Next scripts and avoids broad `unsafe-inline` script permission; `frame-ancestors 'none'`, `base-uri 'self'`, `form-action 'self'`, and `object-src 'none'` are mandatory. Add `nosniff`, strict referrer and permissions policies, and HSTS only on production HTTPS responses. Tests inspect concrete headers and cookie attributes.

### 8. API Gateway authenticates; Go authorizes

Create one API Gateway HTTP API JWT authorizer using the Cognito issuer and both approved client IDs. Build a shared protected route definition/helper in `infra/stacks/bootstrap.ts` so every existing and future `/api/v1/**` registration receives that authorizer and `aws.cognito.signin.user.admin`. Keep `GET /api/health` registered separately without either. Add a guard test that parses the route declarations and fails if a versioned route lacks the authorizer or required scope, or if health gains auth. Deployment tests prove a valid ID token is rejected by API Gateway before the Lambda is invoked.

API Gateway performs signature, issuer, expiry/not-before, approved-client, and required-scope validation. The Lambda's provider-neutral Go `PrincipalResolver` then validates the integration claim shape, `token_use=access`, approved `client_id`, non-empty `sub`, and normalized integer `auth_time`. It performs a strongly consistent membership lookup and returns a principal only when `auth_time > AuthValidAfter` and the membership authority is supported:

```text
Principal {
  Subject: <immutable Cognito sub>
  TenantID: DEFAULT
  Role: ADMIN
}
```

`monitorHandler` receives the resolver and resolves once before route dispatch. Domain methods receive tenant from the resulting principal, removing the unauthenticated constructor-injected tenant from request handling. Cognito claim names remain inside the resolver adapter. Unknown role, wrong tenant, non-active or missing membership, and an authentication ceremony older than the membership boundary all fail closed. Invalid token claim shape is an authentication failure; valid-token membership and epoch denials share the same non-disclosing authorization result.

Alternatives considered:

- A Lambda authorizer was rejected because API Gateway's native JWT authorizer covers token cryptography more simply; current membership still belongs in the integration because it must be strongly read every request.
- Putting tenant/role in Cognito custom claims or groups was rejected because claims are stale until token renewal and make Cognito the application authorization store.

### 9. Auth errors preserve the envelope boundary honestly

Add `AUTHENTICATION_REQUIRED` (`401`) and `AUTHORIZATION_DENIED` (`403`) to `shared/errors`, mirror them in dashboard `ApiErrorCode`, and include registry drift tests. Lambda-originated failures use `errors.Respond` and the standard envelope with empty/non-sensitive details. Membership denial intentionally does not distinguish absent, non-active, old-epoch, wrong-tenant, or unknown-role records.

API Gateway rejects bad or absent JWTs before Lambda and therefore does not use the Go envelope helper. Documentation and clients treat a bare Gateway `401` as authentication failure rather than attempting to require `reason.code`. No custom Gateway mapping Lambda is added solely to reshape that edge.

### 10. Bootstrap and direct-client flows are operational, not public APIs

Add an idempotent Go administration command invoked through a Make target. It uses the existing AWS profile/credential chain, Cognito administrator APIs, and the shared AWS DynamoDB facade. For a normalized email it:

1. Finds an exact unique Cognito user or creates one with invitation delivery suppressed.
2. Reads the immutable `sub`.
3. Conditionally creates or reconciles the versioned `ACTIVE` `DEFAULT`/`ADMIN` membership, preserving immutable `MembershipID` and `Subject` and establishing `AuthValidAfter` before invitation delivery.
4. Sends or resends the default Cognito invitation only after membership is ready and only when activation is still required.
5. Leaves an established Cognito password and healthy existing identity untouched.
6. Fails on ambiguous identity or conflicting membership rather than overwriting it.

The runbook documents AWS CLI/console-authorized procedures for later invitations, membership disable/enable, credential reset, and break glass. No application bootstrap or recovery endpoint exists.

Bruno uses the no-secret direct client and Cognito API calls/helper flow to handle password auth and `NEW_PASSWORD_REQUIRED`, then stores only the resulting access token in a gitignored/local secret environment variable. Every `/api/v1/**` request adds Bearer auth; health does not. Direct clients never exchange or use the dashboard cookie.

### 11. Audit and telemetry are structured and secret-safe

Emit JSON security events from dashboard handlers, the Go resolver, and administration tooling. Use CloudWatch Embedded Metric Format or bounded metric filters for sign-in failure, recovery request, refresh failure, authorization denial, bootstrap failure, and auth infrastructure failure. Dimensions are stage, component, operation, and outcome only. Subject may appear in restricted audit logs when known but never as a metric dimension; email is omitted or irreversibly redacted.

Use correlation/request IDs across the dashboard-to-API hop without propagating secrets. CloudTrail and the selected secret facility's audit trail remain the AWS administrative evidence for Cognito, DynamoDB, secret administration, and break-glass operations. Configure finite log retention and alarms for sustained refresh failures and key-loading/DynamoDB auth errors. Do not emit submitted fields, cookie/header values, session hashes, token claims wholesale, or provider exception payloads that may contain challenge/token data.

### 12. Stage lifecycle, recovery, and cost favor a small serverless footprint

This change consumes the persistent/ephemeral classification and lifecycle contract from `standardize-stage-resource-lifecycle`; it defines no second taxonomy. Persistent stages retain and protect the user pool, authoritative `AuthTable`, and active AES secret and enable `AuthTable` point-in-time recovery. Ephemeral stages omit deletion protection/PITR where necessary and cleanly delete all three auth resources with the stage. Before protected-route cutover, `make check-auth-cutover-prerequisites` must prove validated classification, persistent `AppTable` protection, lifecycle-guarded ephemeral cleanup, and retained inventory. Authentication infrastructure tests assert the mapped lifecycle mode; auth runbooks document only auth-specific recovery and deliberate disposal steps.

Break glass requires an authorized AWS identity, direct Cognito administrator actions, and a conditional membership repair. It does not disable the JWT authorizer. Recovery ends with password/temporary credential rotation, session invalidation where possible, and audit review.

Material cost sources are Cognito Essentials monthly active users and default-email messaging limits/charges, on-demand DynamoDB reads/writes/storage and persistent-stage PITR, secret retrieval and AWS-managed at-rest encryption operations, and CloudWatch ingestion, retention, metrics, and alarms. Strong membership reads add one DynamoDB read per API request; this is an intentional immediate-disable and authentication-revocation cost. The design adds no customer-managed KMS key, NAT gateway, VPC endpoint, custom email/DNS service, always-on compute, or cross-region traffic.

## Risks / Trade-offs

- [Default Cognito email has low-volume quotas and less branding/control] -> State the quota in deployment docs, alarm on invitation/recovery failures, and require a future OpenSpec change before introducing SES.
- [One strong membership read per API request adds latency and read cost] -> Keep the item small and directly keyed; do not cache because immediate disable is a hard requirement.
- [Loading application key material adds cold-start latency and expands the Next server trust boundary] -> Retrieve the linked secret only in the server I/O adapter, cache it per warm runtime, never emit it, and fail closed when unavailable.
- [Concurrent React server requests can create refresh storms] -> Use versioned conditional leases, short Cognito rotation grace, bounded loser rereads, and race-focused tests.
- [Persistent auth resources can outlive a stack deployment and incur cost] -> Follow `standardize-stage-resource-lifecycle`, tag them, output only non-secret identifiers required for recovery, and document inventory/deletion; ephemeral stages delete them cleanly.
- [API Gateway `401` responses do not match the application envelope] -> Document and test the edge explicitly instead of adding a custom authorizer or response-shaping Lambda.
- [CSP nonces can conflict with Next rendering/caching behavior] -> Implement centrally against the deployed runtime, test auth and protected HTML, and fail CI on unsafe production fallback.
- [A compromised Next server role can decrypt refresh tokens] -> Isolate `AuthTable` and the AES secret reference with least-privilege grants, bind ciphertext to authenticated context, avoid token/key logs, and keep API/browser roles unable to retrieve either resource.
- [Bootstrap can partially create Cognito identity before membership] -> Make each step discoverable and idempotent, and test resume after each failure boundary.
- [Direct password auth expands the public app client's allowed flow] -> Limit it to the dedicated human-operator client, keep self-registration off, rely on Cognito password/MFA controls, and exclude M2M use.
- [Single-generation key rotation terminates all browser authentication state] -> Make rotation an explicit maintenance action, delete session/transaction records or let explicit expiry/TTL clean them up, expire cookies, and require fresh authentication; do not retain old keys or attempt online re-encryption.

## Migration Plan

1. Build auth resources, stage-lifecycle mapping, IAM grants, non-secret outputs, and required configuration as reviewable internal milestones while leaving the currently deployed API route posture unchanged.
2. Build and review the custom dashboard auth/session implementation, Go principal/membership implementation, administration command, documentation, telemetry, and protected-route configuration. Internal tasks and commits may land as milestones, but no environment may expose a mixed state in which only some `/api/v1/**` routes or application boundaries enforce authentication; do not create an optional-auth handler mode.
3. Run the AWS-credentialed bootstrap command in staging and production as applicable. Confirm exactly one `ACTIVE` `DEFAULT`/`ADMIN` membership with immutable identifiers and `AuthValidAfter`, then complete the invited user's temporary-password flow at or after that boundary.
4. Validate login, `NEW_PASSWORD_REQUIRED`, recovery, optional TOTP, logout, session expiry/refresh race, disabled membership, direct Cognito/Bruno access, envelope errors, public health, security headers, and secret-safe logs in staging.
5. Announce that existing anonymous `/api/v1/**` clients will break. Distribute the direct-client procedure to intended human operators; there is no static credential migration.
6. Apply the production security cutover atomically in one controlled deployment: every `/api/v1/**` route gains the JWT authorizer and access scope while the API begins requiring resolved membership and `auth_time`, and every dashboard protected boundary begins requiring its opaque session. Verify anonymous and ID-token v1 denial before Lambda, authenticated dashboard and Bruno success, and public health immediately.
7. If application behavior regresses, roll back the affected dashboard/API code to the last authentication-capable build or roll forward a fix. Preserve JWT enforcement and persistent-stage auth resources according to `standardize-stage-resource-lifecycle`; do not reopen anonymous v1 as routine rollback.
8. If identity infrastructure itself prevents all access, execute the AWS-authenticated break-glass runbook, then rotate recovered credentials, invalidate sessions, and review audit evidence.

## Open Questions

None for v1. Session lifetime, transaction lifetime, refresh skew/lease, log retention, and alert thresholds are implementation constants that tasks must document and test; changing the product/security behavior later requires a follow-on OpenSpec change.
