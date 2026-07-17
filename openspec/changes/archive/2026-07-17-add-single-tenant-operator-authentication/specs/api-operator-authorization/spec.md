## ADDED Requirements

### Requirement: API Gateway authenticates every versioned API route
Every route matching `/api/v1/**` SHALL require the deployment's API Gateway JWT authorizer configured for the Cognito issuer and approved application client audience and SHALL require Cognito's `aws.cognito.signin.user.admin` access-token scope, which the approved user-pool API authentication flows support. `GET /api/health` SHALL remain public and SHALL NOT inherit the JWT authorizer or scope. The v1 cutover SHALL provide no anonymous compatibility mode.

#### Scenario: Client calls a versioned route without a JWT
- **WHEN** a client requests any `/api/v1/**` route without an accepted Bearer JWT
- **THEN** API Gateway rejects the request before invoking the monitor API Lambda

#### Scenario: Client calls public health
- **WHEN** a client requests `GET /api/health` without credentials
- **THEN** API Gateway invokes the health Lambda and preserves the existing public health response

#### Scenario: New v1 route is added
- **WHEN** infrastructure adds another route under `/api/v1/**`
- **THEN** route validation fails unless the shared v1 JWT authorizer and required access-token scope are attached

### Requirement: Only Cognito access tokens authorize API requests
API Gateway SHALL require the configured access-token scope on protected routes wherever the approved Cognito flow supplies that scope. The monitor API SHALL accept only a JWT whose validated claims identify it as a Cognito access token issued by the configured pool for an approved app client and SHALL retain explicit `token_use=access` validation as defense in depth. An ID token SHALL NOT authorize API access even if its signature, issuer, expiry, and client association are otherwise valid.

#### Scenario: Client sends an access token
- **WHEN** API Gateway validates the token and required scope and the integration claims contain the expected `token_use=access` and approved client identifier
- **THEN** principal resolution may continue

#### Scenario: Client sends an ID token
- **WHEN** a client sends a valid Cognito ID token to any `/api/v1/**` route
- **THEN** API Gateway rejects it for lacking the required access-token scope before invoking the monitor API Lambda

#### Scenario: Defense-in-depth token-use check fails
- **WHEN** an integration claim set lacks `token_use=access` despite passing the Gateway boundary
- **THEN** the monitor API fails authentication closed and performs no domain operation

### Requirement: Go resolves a normalized provider-neutral principal
The Go API SHALL resolve gateway identity claims through a `PrincipalResolver` boundary into a normalized principal containing immutable subject, tenant, and named role. The Cognito adapter SHALL normalize `auth_time` as a non-negative integer NumericDate for membership revocation checks. Domain handlers SHALL depend on the normalized principal rather than Cognito-specific claim maps or API Gateway event internals. V1 SHALL provide only the Cognito-backed resolver.

#### Scenario: Gateway claims are valid
- **WHEN** the Cognito resolver receives validated access-token claims with a non-empty immutable subject
- **THEN** it returns a normalized principal only after membership authorization succeeds

#### Scenario: Required claim is missing or malformed
- **WHEN** the resolver cannot obtain the expected access-token use, client identifier, immutable subject, or a non-negative integer `auth_time`
- **THEN** it returns a typed authentication failure and no handler operation runs

#### Scenario: Issued-at claim differs from authentication time
- **WHEN** a token has an acceptable `iat` but its `auth_time` is missing, malformed, or older than the membership boundary
- **THEN** `iat` does not authorize the request
- **AND** it may be retained only as diagnostic metadata

### Requirement: Every API request strongly verifies authoritative application membership
After token authentication and before dispatching a domain operation, the API SHALL perform a strongly consistent read from the dedicated `AuthTable`, keyed by the immutable Cognito subject. `AuthTable` SHALL be the single membership and authentication-revocation authority; `AppTable`, cached decisions, and JWT-carried tenant or role claims SHALL NOT substitute for this read. Authorization SHALL require an existing membership with immutable membership ID and subject, tenant `DEFAULT`, status `ACTIVE`, role `ADMIN`, a valid version and timestamps, and normalized `auth_time > AuthValidAfter` on every request. `AuthValidAfter` and `auth_time` SHALL use integer Unix epoch seconds, and authentication ceremonies at or before the boundary SHALL be denied.

#### Scenario: Active member calls the API at the boundary
- **WHEN** an authenticated subject has a strongly read `ACTIVE` `DEFAULT` membership with role `ADMIN` and `auth_time` equal to `AuthValidAfter`
- **THEN** the API denies authorization and requires a later full authentication ceremony

#### Scenario: Membership is disabled after token issuance
- **WHEN** an administrator changes membership to a non-active status and the operator sends a previously issued unexpired access token on the next request
- **THEN** the strongly consistent read observes the non-active state and the API denies the request immediately

#### Scenario: Authentication ceremony predates revocation boundary
- **WHEN** an `ACTIVE` member sends an otherwise valid access token whose normalized `auth_time` is less than `AuthValidAfter`
- **THEN** the API denies authorization and requires a new authentication ceremony

#### Scenario: Authentication time is malformed
- **WHEN** the token has missing, string, fractional, negative, overflowed, or otherwise malformed `auth_time`
- **THEN** principal resolution fails authentication closed before membership authority is granted

#### Scenario: Membership is absent
- **WHEN** a valid Cognito access-token subject has no application membership
- **THEN** the API denies authorization and performs no requested operation

#### Scenario: Membership contains unsupported authority
- **WHEN** membership names another tenant, a status other than `ACTIVE`, a role other than `ADMIN`, or an invalid authority record shape
- **THEN** the API fails closed and does not derive authority from token claims

### Requirement: Tenant scope comes only from trusted server state
The monitor API SHALL derive tenant `DEFAULT` from validated application membership and server configuration. It SHALL ignore or reject tenant authority supplied through request bodies, query parameters, path parameters, headers, or mutable identity attributes.

#### Scenario: Authenticated request attempts tenant override
- **WHEN** a client supplies a tenant identifier different from `DEFAULT` in any request-controlled location
- **THEN** the request cannot access or mutate another tenant
- **AND** domain repository calls receive only the principal's server-derived `DEFAULT` tenant

#### Scenario: Handler executes domain operation
- **WHEN** authorization succeeds
- **THEN** handler tenant scope comes from the normalized principal rather than a globally injected unauthenticated tenant value

### Requirement: Authentication and authorization failures have explicit edge semantics
Application-generated authentication failures SHALL use typed code `AUTHENTICATION_REQUIRED` with HTTP `401`; missing, non-active, unsupported, or authentication-epoch-denied membership SHALL use typed code `AUTHORIZATION_DENIED` with HTTP `403`. These Lambda responses SHALL use the shared error registry and response envelope without sensitive details. Rejections generated by API Gateway before Lambda invocation, including missing required scope and ID-token rejection, SHALL use Gateway-generated non-envelope `401` or `403` semantics and are explicitly outside the application envelope guarantee.

#### Scenario: Principal resolution fails inside Lambda
- **WHEN** integration identity context is present but required access-token claims cannot be resolved safely
- **THEN** Lambda returns HTTP `401` with the standard error envelope and `reason.code` equal to `AUTHENTICATION_REQUIRED`

#### Scenario: Membership authorization fails
- **WHEN** the strongly consistent membership check denies the subject
- **THEN** Lambda returns HTTP `403` with the standard error envelope and `reason.code` equal to `AUTHORIZATION_DENIED`
- **AND** details do not disclose whether membership was missing, non-active, assigned another tenant or role, malformed, or rejected by `AuthValidAfter`

#### Scenario: Gateway rejects token before integration
- **WHEN** API Gateway rejects a missing, malformed, expired, wrong-issuer, or wrong-audience JWT
- **THEN** the client receives an HTTP `401` generated by API Gateway
- **AND** clients do not assume that this pre-integration body conforms to the application envelope

### Requirement: Direct clients can authenticate without dashboard sessions
Bruno and other human-operated direct API clients SHALL authenticate directly against Cognito's supported non-managed-login API flow, obtain an access token, and send it as `Authorization: Bearer <access-token>`. Dashboard session cookies SHALL NOT be accepted as API credentials.

#### Scenario: Bruno calls protected API
- **WHEN** an invited operator completes the documented Cognito direct-client authentication flow and Bruno sends the resulting access token
- **THEN** API Gateway authenticates the token and the API applies the same membership authorization as dashboard traffic

#### Scenario: Direct client sends dashboard cookie only
- **WHEN** a direct client calls `/api/v1/**` with a dashboard session cookie but no accepted Bearer access token
- **THEN** API Gateway rejects the request before Lambda invocation
