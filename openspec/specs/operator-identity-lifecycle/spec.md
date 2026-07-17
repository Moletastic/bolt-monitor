## Purpose

Define invite-only operator identities, membership authority, and authentication lifecycle flows.

## Requirements

### Requirement: Deployment uses one invite-only Cognito operator directory
The system SHALL provision one Amazon Cognito user pool on the Cognito Essentials feature plan for each deployment, SHALL support multiple human operators in that directory, and SHALL associate every application member with the server-owned tenant `DEFAULT`. Self-registration SHALL be disabled. The implementation SHALL NOT use AWS Amplify or Cognito managed-login/hosted UI surfaces.

#### Scenario: Deployment provisions operator identity
- **WHEN** the infrastructure is deployed
- **THEN** it provisions a Cognito Essentials user pool configured for administrator-created users
- **AND** it does not provision a public sign-up flow, Amplify authentication resources, or a Cognito managed-login domain

#### Scenario: More than one operator is invited
- **WHEN** administrators create multiple Cognito users and corresponding `ACTIVE` memberships
- **THEN** each operator can authenticate independently to the same deployment
- **AND** each membership remains scoped to tenant `DEFAULT`

### Requirement: AuthTable is the permanent application-auth authority
The dedicated `AuthTable` SHALL be the single authoritative store for application memberships, dashboard sessions, authentication transactions, each member's authentication-revocation epoch, and future user lifecycle state. `AppTable` SHALL remain authoritative only for monitoring-domain data and SHALL NOT store, duplicate as authority, or provide fallback membership or user-lifecycle decisions.

#### Scenario: Application resolves operator authority
- **WHEN** the API or dashboard needs current application membership or revocation state
- **THEN** it reads the authoritative record from `AuthTable`
- **AND** it does not infer authority from `AppTable`, Cognito groups, email, or stale token role or tenant claims

#### Scenario: Future RBAC extends operator authority
- **WHEN** a later change adds roles or user lifecycle states
- **THEN** it extends the existing versioned `AuthTable` membership record
- **AND** it does not migrate membership authority or require a second initial-user bootstrap

### Requirement: Initial membership is forward-compatible and versioned
Bootstrap SHALL create one membership record with immutable `MembershipID`, immutable Cognito `Subject`, `TenantID=DEFAULT`, `Status=ACTIVE`, `Role=ADMIN`, non-negative integer Unix-second `AuthValidAfter`, positive `Version`, `CreatedAt`, and `UpdatedAt`. `MembershipID`, `Subject`, `TenantID`, and `CreatedAt` SHALL NOT change after creation. Membership has no TTL. Updates SHALL use optimistic version conditions and advance `UpdatedAt`; future RBAC and user lifecycle fields SHALL extend this record in place.

#### Scenario: Initial membership is created
- **WHEN** bootstrap resolves the initial operator's immutable Cognito subject
- **THEN** it writes the complete versioned membership shape to `AuthTable` before sending an invitation
- **AND** `AuthValidAfter` is no earlier than membership creation so prior authentication ceremonies do not acquire authority

#### Scenario: Membership is updated
- **WHEN** an administrator changes status, role, or authentication validity
- **THEN** the update preserves immutable identity fields and conditionally advances `Version` and `UpdatedAt`

#### Scenario: Membership shape is incomplete
- **WHEN** an authority check reads a record missing or malforming a required identity, authority, epoch, version, or timestamp field
- **THEN** authorization fails closed

### Requirement: Cognito provides initial email delivery without custom mail infrastructure
The user pool SHALL use Cognito's default email delivery for invitations and password recovery in v1. Deployment SHALL NOT require a custom domain, DNS records, Amazon SES identity, or external email provider.

#### Scenario: Fresh deployment uses default email
- **WHEN** an administrator invites the initial operator or an operator requests a password reset
- **THEN** Cognito sends the applicable message through its default email configuration
- **AND** deployment succeeds without custom DNS or SES configuration

### Requirement: Dashboard owns custom operator authentication pages
The Next.js dashboard SHALL provide custom pages and server-side handlers for sign-in, invitation activation and `NEW_PASSWORD_REQUIRED`, forgot-password initiation, reset-password confirmation, and optional TOTP enrollment and challenge. These pages SHALL use provider-neutral application interfaces whose only v1 adapter is Cognito.

#### Scenario: Operator signs in with an established password
- **WHEN** an operator with `ACTIVE` membership submits valid credentials on the custom sign-in page and Cognito returns tokens without another challenge
- **THEN** the dashboard establishes a server-side dashboard session
- **AND** no Cognito managed-login page or Amplify client code is used

#### Scenario: Invited operator must replace temporary password
- **WHEN** Cognito returns `NEW_PASSWORD_REQUIRED` after an invited operator submits a temporary password
- **THEN** the custom activation page collects and submits a compliant new password using the opaque server-held challenge continuation
- **AND** successful completion establishes a new dashboard session

#### Scenario: Operator begins password recovery
- **WHEN** an operator submits an email address on the forgot-password page
- **THEN** the dashboard requests Cognito password recovery and renders the same non-enumerating acknowledgement whether or not the account is usable

#### Scenario: Operator confirms password recovery
- **WHEN** an operator submits a valid recovery code and compliant new password on the custom reset page
- **THEN** Cognito confirms the password change
- **AND** the operator is returned to sign-in without exposing recovery codes or passwords beyond the server-side request boundary

#### Scenario: TOTP is required during sign-in
- **WHEN** Cognito returns a software-token MFA challenge for an operator
- **THEN** the custom TOTP challenge page accepts the one-time code and continues the same server-held authentication transaction

#### Scenario: Operator enrolls optional TOTP
- **WHEN** Cognito requires or offers software-token enrollment for an operator during authentication
- **THEN** the custom enrollment path presents the secret only for immediate authenticator setup and verifies a TOTP code before continuing
- **AND** the secret is not persisted in browser storage, an RSC payload, logs, or application records

### Requirement: Authentication transactions remain server-confidential and bounded
The dashboard SHALL keep Cognito challenge session values, passwords, recovery codes, TOTP secrets, and issued tokens out of URLs, browser storage, React Server Component payloads, client component props, and logs. Any pre-session authentication transaction stored across requests SHALL be opaque to the browser, single-purpose, short-lived, and invalid after successful use or expiry.

#### Scenario: Authentication requires multiple HTTP requests
- **WHEN** a Cognito challenge must continue on a later request
- **THEN** the browser receives only an opaque, short-lived transaction reference
- **AND** the server rejects a transaction used for another flow, after expiry, or after successful completion

#### Scenario: Authentication fails
- **WHEN** Cognito rejects credentials, a challenge, or a recovery code
- **THEN** operator-visible feedback does not disclose secrets or unnecessary account-existence information
- **AND** server logs contain no submitted password, code, challenge session, TOTP secret, or token

### Requirement: Initial administrator bootstrap is credentialed and idempotent
The system SHALL provide an operator-run bootstrap command that uses the caller's AWS credentials and AWS APIs directly, not a public application endpoint. Given an email address, repeated execution SHALL converge one Cognito identity and one `ACTIVE` versioned `DEFAULT` membership with role `ADMIN` in `AuthTable` without duplicating users, resetting an established password, replacing immutable membership identity, lowering `AuthValidAfter`, or weakening an existing account.

#### Scenario: Bootstrap creates the first administrator
- **WHEN** an authorized operator runs bootstrap for an email not yet represented in Cognito or application membership
- **THEN** the command creates the Cognito user through the administrator API with invitation delivery suppressed
- **AND** it creates the complete `ACTIVE` `DEFAULT` `ADMIN` membership before requesting default Cognito invitation delivery
- **AND** it stores immutable membership ID and Cognito subject, `AuthValidAfter`, version, and timestamps in the authoritative `AuthTable` record

#### Scenario: Bootstrap is retried after success
- **WHEN** the same bootstrap input is run after both identity and membership exist in the desired state
- **THEN** the command reports the reconciled state without creating another user, sending an unnecessary replacement invitation, or changing credentials

#### Scenario: Bootstrap resumes after partial failure
- **WHEN** the Cognito user exists but the membership write did not complete
- **THEN** a retry resolves the immutable Cognito subject and creates or reconciles the missing membership

#### Scenario: Bootstrap detects unsafe conflict
- **WHEN** the supplied email resolves ambiguously or an existing membership conflicts with the fixed tenant or supported role model
- **THEN** the command fails loudly without overwriting the conflicting identity or membership

### Requirement: V1 defines only the initial ADMIN application role
Application membership SHALL store status and role as named values. V1 SHALL authorize only `Status=ACTIVE` and `Role=ADMIN`; all other values fail closed. Cognito groups or mutable email claims SHALL NOT be the source of status, tenant, role, or authentication-revocation authorization.

#### Scenario: Operator identity has an active membership
- **WHEN** the application resolves a Cognito subject whose membership has tenant `DEFAULT`, role `ADMIN`, and status `ACTIVE`
- **THEN** the normalized application principal receives tenant `DEFAULT` and role `ADMIN`

#### Scenario: Unsupported role is encountered
- **WHEN** membership contains a role other than `ADMIN`
- **THEN** authorization fails closed rather than treating the unknown value as an administrator

### Requirement: Membership authentication epoch revokes prior ceremonies
The API SHALL authorize an `ACTIVE` membership only when the normalized Cognito `auth_time` integer NumericDate is strictly greater than the membership's `AuthValidAfter` integer Unix second. Authentication ceremonies at or before the boundary SHALL be denied. `iat` SHALL NOT be revocation authority and MAY be used only for diagnostics. Missing, string, fractional, negative, overflowed, or otherwise malformed `auth_time` SHALL fail authentication closed.

#### Scenario: Authentication occurred exactly at the boundary
- **WHEN** token `auth_time` equals membership `AuthValidAfter`
- **THEN** authorization is denied and a later full authentication ceremony is required

#### Scenario: Authentication occurred before the boundary
- **WHEN** token `auth_time` is less than membership `AuthValidAfter`
- **THEN** authorization is denied even if `iat` and token expiry are otherwise acceptable

#### Scenario: Administrator re-enables a member
- **WHEN** an administrator changes a non-active membership to `ACTIVE`
- **THEN** the operation advances `AuthValidAfter` to a documented whole-second boundary that excludes ceremonies completed through the update instant
- **AND** the operator must complete a fresh authentication ceremony strictly after that boundary
