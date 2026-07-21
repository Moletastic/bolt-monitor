## MODIFIED Requirements

### Requirement: Initial administrator bootstrap is credentialed and idempotent
The system SHALL provide an operator-run `make invite-admin EMAIL=<email>` command that uses the selected deployment target, its AWS credentials, and deployed SST output to resolve the Cognito user pool and `AuthTable` directly, not a public application endpoint. Repeated execution for the same email SHALL converge one Cognito identity and one `ACTIVE` versioned `DEFAULT` membership with role `ADMIN` without duplicating users, resetting an established password, replacing immutable membership identity, lowering `AuthValidAfter`, or weakening an existing account.

#### Scenario: Bootstrap creates the first administrator
- **WHEN** an authorized operator invokes `make invite-admin EMAIL=<email>` after a target has deployed
- **THEN** the command resolves the target-selected user pool and `AuthTable` from SST output without requiring copied resource identifiers
- **AND** it creates the Cognito user through the administrator API with invitation delivery suppressed
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
