## ADDED Requirements

### Requirement: AuthTable membership is the sole application authorization authority
The system SHALL authorize an authenticated request only after a strongly consistent read of the canonical AuthTable item `PK=MEMBER#<cognito-sub>`, `SK=MEMBERSHIP`. That item SHALL belong to tenant `DEFAULT`, have status `ACTIVE`, contain a supported fixed role, and satisfy the authentication boundary. Cognito groups, mutable identity attributes, AppTable records, and audit projections SHALL NOT grant application permissions.

#### Scenario: Active member presents an eligible token
- **WHEN** Cognito proves an access token whose subject matches an `ACTIVE` AuthTable membership and whose authentication ceremony is after the membership boundary
- **THEN** the system authorizes the request according to the role stored in that membership

#### Scenario: Cognito group claims disagree with membership
- **WHEN** a valid token contains Cognito groups or role-like attributes that differ from AuthTable membership
- **THEN** the system ignores those claims for authorization
- **AND** evaluates permissions only from the canonical AuthTable membership

#### Scenario: AppTable projection disagrees with AuthTable
- **WHEN** an AppTable audit projection is missing, delayed, or inconsistent with AuthTable
- **THEN** authorization uses AuthTable and does not infer membership or lifecycle state from AppTable

#### Scenario: Authenticated subject has no membership
- **WHEN** Cognito proves a token but its subject has no canonical AuthTable membership
- **THEN** the system denies the request without revealing membership existence details

### Requirement: Revocation uses access-token auth_time with an inclusive deny boundary
Each membership SHALL store a non-negative integer epoch-second `AuthValidAfter`. Authorization SHALL require the Cognito access-token `auth_time` claim to be a non-negative integer epoch second strictly greater than `AuthValidAfter`. Authentication ceremonies at or before the boundary SHALL be denied. The system SHALL NOT use access-token `iat` as revocation evidence.

#### Scenario: Authentication occurred at the revocation boundary
- **WHEN** an access token has `auth_time` equal to `AuthValidAfter`
- **THEN** the system denies the request

#### Scenario: Authentication occurred before the revocation boundary
- **WHEN** an access token has `auth_time` less than `AuthValidAfter`
- **THEN** the system denies the request

#### Scenario: Full authentication occurs after revocation
- **WHEN** an active member completes a full Cognito authentication ceremony whose access token has `auth_time` greater than `AuthValidAfter`
- **THEN** the token is eligible for role authorization

#### Scenario: Old refresh family issues a newer token
- **WHEN** a refresh-token family authenticated at or before `AuthValidAfter` produces an access token with a newer `iat` but unchanged old `auth_time`
- **THEN** the system denies the refreshed access token

#### Scenario: auth_time is missing or malformed
- **WHEN** an access token omits `auth_time` or represents it as a string, fraction, negative value, overflow, or other malformed shape
- **THEN** principal resolution fails closed before route work

### Requirement: Dashboard sessions obey the same revocation authority
Before using a stored Cognito token bundle, the dashboard SHALL strongly verify the subject's AuthTable membership and SHALL apply the same `Status`, role, tenant, and `auth_time > AuthValidAfter` rules as the API. Indexed session deletion and Cognito global sign-out SHALL be defense in depth and cleanup, not substitutes for the boundary.

#### Scenario: Dashboard session predates revocation
- **WHEN** a dashboard session contains a token family whose `auth_time` is at or before the membership boundary
- **THEN** the dashboard invalidates the session and requires a full authentication ceremony

#### Scenario: Session cleanup has not completed
- **WHEN** a session record remains discoverable after the membership boundary advances
- **THEN** the dashboard still denies it on its next validation

### Requirement: Membership subjects are immutable and tenant-bound
Each membership SHALL belong to the single tenant `DEFAULT`, SHALL use Cognito `sub` as its immutable identity binding, and SHALL require a new invite lifecycle rather than changing that binding in place. The system SHALL NOT create a duplicate membership identity item in AppTable or another AuthTable key family.

#### Scenario: Administrator attempts to replace a subject
- **WHEN** an administrator submits an update that would change a membership subject or tenant
- **THEN** the system rejects the update without modifying membership

#### Scenario: Membership is inspected across storage
- **WHEN** an operator traces authorization authority for a subject
- **THEN** exactly the canonical AuthTable membership supplies application authority
- **AND** any AppTable lifecycle event is treated only as audit evidence

### Requirement: System provides only fixed roles
The system SHALL support exactly `ADMIN`, `OPERATOR`, and `VIEWER` membership roles and SHALL reject unknown or custom roles.

#### Scenario: Administrator assigns a fixed role
- **WHEN** an administrator assigns `ADMIN`, `OPERATOR`, or `VIEWER`
- **THEN** the system stores that role in canonical AuthTable membership and uses it for subsequent authorization

#### Scenario: Client submits a custom role
- **WHEN** a client submits a role outside the fixed set
- **THEN** the system returns a validation error without changing permissions

### Requirement: Roles follow the explicit least-privilege permission matrix
The system SHALL enforce the following permissions consistently:

| Permission | ADMIN | OPERATOR | VIEWER |
| --- | --- | --- | --- |
| Read services, monitors, statuses, runs, incidents, audit history, scheduler state, channels, and policies | Allow | Allow | Allow |
| Create or update services and monitors | Allow | Allow | Deny |
| Delete services and monitors | Allow | Allow | Deny |
| Trigger manual monitor runs | Allow | Allow | Deny |
| Acknowledge or resolve incidents | Allow | Allow | Deny |
| Change global scheduler state | Allow | Deny | Deny |
| Create, update, delete, or test channels | Allow | Deny | Deny |
| Create, update, or delete escalation policies/routes | Allow | Deny | Deny |
| List or manage operator users and read tenant user-lifecycle audit events | Allow | Deny | Deny |

All secret-bearing read responses SHALL retain their existing redaction regardless of role.

#### Scenario: Viewer performs a read
- **WHEN** a `VIEWER` requests a permitted read resource
- **THEN** the system returns the same non-secret representation available to other roles

#### Scenario: Operator changes monitor configuration
- **WHEN** an `OPERATOR` creates, updates, or deletes a service or monitor
- **THEN** the system permits the operation subject to existing business validation

#### Scenario: Operator attempts an admin-only mutation
- **WHEN** an `OPERATOR` attempts scheduler, channel, policy, or user-management mutation
- **THEN** the system denies the request without applying side effects

#### Scenario: Viewer attempts a mutation
- **WHEN** a `VIEWER` attempts any configuration or operational mutation
- **THEN** the system denies the request without applying side effects

#### Scenario: Administrator performs an admin-only mutation
- **WHEN** an `ADMIN` performs a scheduler, channel, policy, or user-management operation
- **THEN** the system permits the operation subject to existing business validation

### Requirement: Direct APIs enforce the same permissions as the dashboard
Every protected HTTP route SHALL enforce authorization at the API boundary, independent of dashboard visibility or server-action checks.

#### Scenario: Hidden dashboard control is called directly
- **WHEN** a caller invokes a direct API operation that its role does not permit
- **THEN** the API returns a forbidden response
- **AND** performs no domain, Cognito, session, or audit mutation

#### Scenario: Dashboard renders role-scoped actions
- **WHEN** an authenticated operator opens a dashboard view
- **THEN** the dashboard omits or disables actions not granted to that role
- **AND** API enforcement remains authoritative

### Requirement: Authorization failures use generic safe responses
Authentication and authorization failures SHALL use the standard response envelope and typed error registry where generated by the application, SHALL distinguish unauthenticated from forbidden only at the protocol level, and SHALL NOT disclose membership existence, status, role, email, Cognito state, token claims, or credentials.

#### Scenario: Caller lacks permission
- **WHEN** an authenticated caller requests an operation not granted by its role
- **THEN** the system returns the registered generic forbidden error
- **AND** response details contain no sensitive identity or authorization data

#### Scenario: Authorization diagnostic is logged
- **WHEN** the system records an authorization denial
- **THEN** structured logs use opaque identifiers and bounded reason categories
- **AND** contain no email, token, credential, invitation material, or raw authorization header
