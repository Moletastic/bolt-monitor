## Why

The operator dashboard and versioned API currently trust network reachability and a hard-coded tenant instead of an authenticated operator identity. Before the service is used for real operational data, it needs a small, recoverable authentication boundary that supports multiple operators in the single deployment tenant without creating a premature multi-tenant or broad RBAC system.

## What Changes

- Add invite-only operator identities backed by Amazon Cognito Essentials, including custom dashboard pages for sign-in, invite activation, required password replacement, password recovery, and optional TOTP enrollment/challenge.
- Disable self-registration and use Cognito's default email delivery so the initial deployment does not require custom DNS or SES configuration.
- Add an idempotent, AWS-credentialed bootstrap command that creates or reconciles the initial `ACTIVE` `ADMIN` operator membership, including its immutable identity and authentication-revocation boundary, without exposing an unauthenticated bootstrap endpoint.
- Add opaque dashboard sessions stored in DynamoDB with 256-bit identifiers, hashed lookup keys, explicit application-enforced expiry, protected refresh-token storage, safe rotation, and concurrent refresh handling.
- Protect dashboard routes and state-changing actions with server-side session checks, CSRF Origin/Host validation, security headers, session-fixation prevention, safe redirects, and secret-safe rendering and logging.
- **BREAKING** Require a Cognito access token with `aws.cognito.signin.user.admin` at API Gateway for every `/api/v1/**` route while preserving public `GET /api/health`; existing unauthenticated v1 clients stop working at cutover, and ID tokens are rejected before Lambda invocation.
- Make a dedicated `AuthTable`, never `AppTable`, the permanent single authority for memberships, dashboard sessions, authentication transactions, per-member authentication revocation, and future user lifecycle state.
- Resolve authenticated API requests to a normalized Go principal, perform a strongly consistent `ACTIVE` membership read on every request, require normalized Cognito `auth_time > AuthValidAfter`, and derive the fixed `DEFAULT` tenant and initial `ADMIN` role exclusively from trusted server state; ceremonies at or before the boundary are denied and `iat` remains diagnostic only.
- Store the initial membership in a forward-compatible record with immutable membership ID and subject, `TenantID=DEFAULT`, `Status=ACTIVE`, `Role=ADMIN`, `AuthValidAfter`, version, and timestamps so later RBAC extends the same record without migration or a second bootstrap path.
- Document direct API and Bruno authentication, stable auth failure behavior, minimum audit/telemetry, stage-aware auth-resource lifecycle, break-glass recovery, and v1 cutover operations.
- Keep identity and session boundaries provider-neutral while shipping only a Cognito implementation in this change.
- Exclude user-management UI, broad RBAC, machine-to-machine credentials, custom identity providers, static-data migration, and responsive-layout work.

## Capabilities

### New Capabilities

- `operator-identity-lifecycle`: Invite-only Cognito operator identities, custom authentication and recovery flows, optional TOTP, and credentialed initial-admin bootstrap.
- `dashboard-auth-session`: Opaque server-side dashboard sessions, token custody and refresh, route protection, CSRF defenses, redirect safety, security headers, and secret handling.
- `api-operator-authorization`: API Gateway access-token authentication, normalized Go principals, per-request membership authorization, fixed-tenant enforcement, and auth error semantics.
- `auth-operations`: Authentication audit events, telemetry, retention, cost expectations, cutover, and break-glass recovery.

### Modified Capabilities

- `dashboard-web-app`: Require an authenticated operator with `ACTIVE` membership before rendering existing operator surfaces and add custom authentication pages outside the protected shell.
- `api-documentation`: Document the Cognito access-token flow and keep Bruno coverage usable after authenticated v1 cutover.

## Impact

- Infrastructure gains a Cognito Essentials user pool and app clients, a dedicated authoritative `AuthTable`, one active-generation AES-256 secret supplied through credentialed non-printing SST secret setup (or a bootstrap-safe equivalent), API Gateway JWT authorization with access scope, least-privilege IAM grants, stage-aware lifecycle behavior, and auth observability.
- The dashboard gains custom auth routes, server-only Cognito/session adapters, protected layouts/actions, cookie and security-header handling, and authenticated API token forwarding.
- The Go API gains a provider-neutral principal resolver, membership repository, authorization errors, and trusted request context instead of direct tenant injection.
- Bootstrap and operational documentation gain initial-admin, direct-client, cutover, recovery, cost, and validation procedures.
- Existing `/api/v1/**` consumers must obtain and send Cognito access JWTs at the security cutover; `/api/health` remains public.
