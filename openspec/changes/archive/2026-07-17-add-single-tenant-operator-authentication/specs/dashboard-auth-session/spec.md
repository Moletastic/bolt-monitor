## ADDED Requirements

### Requirement: Dashboard sessions use opaque high-entropy host cookies
After successful authentication, the dashboard SHALL issue a cryptographically random 256-bit session identifier and store only that opaque identifier in a cookie named with the `__Host-` prefix. The cookie SHALL be `Secure`, `HttpOnly`, `SameSite=Lax`, `Path=/`, SHALL omit `Domain`, and SHALL NOT contain Cognito tokens, identity claims, tenant, role, email, or serialized session state.

#### Scenario: Authentication establishes a session
- **WHEN** an operator completes authentication successfully
- **THEN** the dashboard creates a new 256-bit random session identifier using a cryptographically secure source
- **AND** the response sets the identifier in a compliant `__Host-` cookie

#### Scenario: Browser script inspects storage
- **WHEN** client-side code runs in the dashboard
- **THEN** it cannot read the HttpOnly session cookie
- **AND** it finds no Cognito token or dashboard session material in local storage, session storage, IndexedDB, or client-readable cookies

### Requirement: Session records use hashed lookup identifiers and explicit expiry
The dashboard SHALL store session records only in the dedicated authoritative `AuthTable`, under a deterministic cryptographic hash of the opaque cookie identifier, and SHALL NOT persist the raw identifier. `AppTable` SHALL NOT store or provide fallback session state. Every read SHALL reject a record whose application expiry is reached, regardless of whether DynamoDB TTL has deleted it; TTL SHALL be cleanup only.

#### Scenario: Valid cookie is presented
- **WHEN** the dashboard receives a session cookie before the record's explicit expiry
- **THEN** it hashes the cookie value, reads the corresponding record, and evaluates the stored expiry before accepting the session

#### Scenario: Expired record remains in DynamoDB
- **WHEN** a session's explicit expiry has passed but asynchronous TTL deletion has not occurred
- **THEN** the dashboard rejects the session and expires the browser cookie

#### Scenario: Session storage is inspected
- **WHEN** an authorized operator reads a session item
- **THEN** no raw browser session identifier is present

### Requirement: Cognito tokens remain protected on the server
Access, ID, and refresh tokens SHALL remain server-side. Token bundles stored for a dashboard session SHALL use authenticated encryption with one active generation of an installation-specific 256-bit AES key supplied through an SST Secret or bootstrap-safe secret reference, and ciphertext SHALL be bound to the application, stage, record kind, active generation identifier, and session context. Plaintext tokens and key material SHALL exist only transiently in the server I/O boundary. Tokens SHALL NOT be returned through RSC payloads, browser responses, client props, logs, errors, or telemetry.

#### Scenario: Dashboard calls the monitor API
- **WHEN** an authenticated server render or server action needs `/api/v1/**`
- **THEN** the server retrieves a usable Cognito access token from the protected session and sends it as the Bearer credential
- **AND** no token is exposed to browser JavaScript or rendered output

#### Scenario: Session record is copied without key access
- **WHEN** an actor obtains the DynamoDB session item but cannot retrieve the application key or satisfy the authenticated ciphertext context
- **THEN** the stored refresh-token ciphertext cannot be used as a Cognito refresh token

### Requirement: AES secret setup and rotation expose no value and retain no old generation
The active AES key SHALL be generated at runtime and installed through a credentialed non-printing helper using SST Secret management or a bootstrap-safe equivalent. Its value SHALL be absent from source, scripts, shell history, files, logs, stack outputs, generated templates, and state-visible configuration; any transient process-memory or pinned CLI boundary SHALL be treated as sensitive and documented. The system SHALL retain no previous decrypting generation and SHALL NOT perform online re-encryption. Replacing the active generation SHALL intentionally invalidate every existing dashboard session and authentication transaction, after which the dashboard SHALL expire presented cookies and require a fresh authentication ceremony.

#### Scenario: Infrastructure and state are inspected
- **WHEN** source, deployment output, generated configuration, and state-visible configuration are reviewed
- **THEN** they contain only the secret name or reference and non-sensitive generation metadata
- **AND** they contain no AES key value

#### Scenario: Active key generation rotates
- **WHEN** an authorized operator replaces the one active AES key generation
- **THEN** records encrypted by the prior generation are rejected rather than decrypted through a previous-key fallback
- **AND** all affected sessions and authentication transactions require fresh authentication

#### Scenario: Rotation implementation is inspected
- **WHEN** reviewers inspect session and transaction storage behavior
- **THEN** there is no current/previous key ring and no online record re-encryption path

### Requirement: Session and token rotation handles concurrent refresh safely
The dashboard SHALL rotate away any pre-authentication or prior authenticated session identifier when authentication succeeds. Cognito refresh-token rotation SHALL be enabled. Refresh shall use conditional ownership/version updates so one request refreshes an expiring access token while concurrent requests wait briefly and reread the winning result; stale refresh results SHALL NOT overwrite newer token state or revoke the active session family accidentally.

#### Scenario: Authentication succeeds while an old cookie exists
- **WHEN** a browser completes authentication with an existing anonymous, authentication-transaction, or authenticated cookie
- **THEN** the dashboard invalidates the prior reference and issues a newly generated session identifier

#### Scenario: Two requests detect an expired access token
- **WHEN** concurrent dashboard requests attempt to refresh the same session
- **THEN** only one request acquires the conditional refresh lease and calls Cognito
- **AND** the other request rereads the updated session and uses the winning token result

#### Scenario: Refresh owner fails
- **WHEN** a refresh lease expires without a successful version update
- **THEN** a later request can acquire a new lease without accepting stale token state

#### Scenario: Cognito rejects refresh
- **WHEN** Cognito determines the refresh token is invalid, revoked, or expired
- **THEN** the dashboard invalidates the server session, expires the cookie, and requires fresh sign-in

### Requirement: Dashboard validates sessions on protected server boundaries
Existing operator routes, server-rendered data loads, and server actions SHALL require a valid dashboard session before rendering protected data or invoking `/api/v1/**`. Authentication pages SHALL remain reachable without a session. The dashboard SHALL fail closed when session configuration or storage is unavailable.

#### Scenario: Unauthenticated operator requests a protected page
- **WHEN** a request without a valid session targets an operator route
- **THEN** the server redirects to the custom sign-in page before protected content or API data is rendered

#### Scenario: Unauthenticated request invokes a protected action
- **WHEN** a request without a valid session invokes a state-changing server action
- **THEN** the action performs no monitor API call and returns or navigates to an authentication-required outcome

#### Scenario: Authenticated operator opens an auth page
- **WHEN** an operator with a valid session opens sign-in or password-recovery entry pages
- **THEN** the server redirects to a safe default dashboard destination

### Requirement: State-changing dashboard requests enforce same-origin CSRF checks
Every dashboard endpoint or server action that changes authentication, session, or application state SHALL validate the request `Origin` against the configured canonical dashboard origin and validate the effective host against that same origin. Missing or mismatched origin/host evidence SHALL fail closed except for explicitly documented same-origin browser cases whose Fetch Metadata and host evidence are equivalent. Proxy-derived hosts SHALL be trusted only from the deployed platform's known forwarding boundary.

#### Scenario: Cross-origin form targets a dashboard action
- **WHEN** a state-changing request carries an `Origin` that does not exactly match the configured dashboard origin
- **THEN** the dashboard rejects the request before Cognito, session storage, or monitor API side effects occur

#### Scenario: Host header is manipulated
- **WHEN** a state-changing request's effective host does not match the configured dashboard origin
- **THEN** the dashboard rejects the request even if an ambient session cookie is present

### Requirement: Redirect targets cannot escape the dashboard origin
Post-authentication and session-expiry redirects SHALL accept only normalized root-relative paths from an explicit safe policy. Absolute URLs, protocol-relative values, encoded authority tricks, control characters, backslashes, and authentication-loop destinations SHALL be rejected in favor of a fixed dashboard default.

#### Scenario: Sign-in receives a safe return path
- **WHEN** the requested return target is a normalized allowed dashboard path
- **THEN** successful sign-in redirects to that path

#### Scenario: Sign-in receives an external return target
- **WHEN** the return target could resolve outside the dashboard origin or create an authentication loop
- **THEN** successful sign-in redirects to the fixed default dashboard path instead

### Requirement: Logout and invalidation remove server authority
Logout SHALL conditionally delete or revoke the server session before expiring the cookie. A missing session SHALL make logout idempotently successful. Membership-denied API responses SHALL invalidate the dashboard session so non-active or old-epoch operators cannot continue using cached server credentials.

#### Scenario: Operator logs out
- **WHEN** an authenticated operator submits the same-origin logout action
- **THEN** the server invalidates the session record and expires the `__Host-` cookie

#### Scenario: Logout is retried
- **WHEN** logout is submitted after the session has already been removed
- **THEN** the response still expires the cookie and reveals no session existence detail

#### Scenario: API rejects membership authority
- **WHEN** a dashboard API request receives the typed membership authorization response for non-active status, unsupported authority, or an expired authentication epoch
- **THEN** the dashboard invalidates its session and requires fresh authentication

### Requirement: Dashboard responses set a security-header baseline
Dashboard responses SHALL set an environment-appropriate Content Security Policy and SHALL set `frame-ancestors 'none'`, `base-uri 'self'`, `form-action 'self'`, `object-src 'none'`, `X-Content-Type-Options: nosniff`, a restrictive `Referrer-Policy`, and a restrictive `Permissions-Policy`. Production HTTPS responses SHALL enable HSTS. CSP configuration SHALL support required Next.js assets without allowing broadly unsafe script execution.

#### Scenario: Browser receives dashboard HTML
- **WHEN** the dashboard returns an authentication or protected HTML page
- **THEN** the response includes the configured security headers
- **AND** the page does not permit framing by another origin
