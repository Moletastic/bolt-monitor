## Why

The single-tenant operator authentication foundation proves identity and establishes `AuthTable` as the permanent authority for memberships, user lifecycle, revocation boundaries, and dashboard sessions. Bolt Monitor still needs administrator-managed later users and fixed least-privilege roles without creating a second identity authority or weakening immediate revocation.

## What Changes

- Extend the authentication-owned `AuthTable` membership contract to support lifecycle states `INVITED`, `ACTIVE`, and `DISABLED`; fixed roles `ADMIN`, `OPERATOR`, and `VIEWER`; an `AuthValidAfter` epoch; optimistic versioning; and current lifecycle-operation state.
- Add admin-only management of users after the initial authentication-owned bootstrap: invite, resend invitation, list, assign a fixed role, enable, disable, revoke sessions, and explicitly repair pending lifecycle operations.
- Keep `AuthTable` authoritative for membership, email uniqueness claims, active-admin guard state, lifecycle operations, and sessions. Do not copy membership or repair authority into `AppTable`.
- Add exact AuthTable access paths for strongly consistent subject authorization, bounded administration listing and membership-ID lookup, pending-operation repair, email uniqueness, and subject-session invalidation.
- Authorize access tokens from their Cognito `auth_time`: deny authentication ceremonies at or before `AuthValidAfter`, fail closed when `auth_time` is absent or malformed, and continue denying refreshed tokens from an old authentication family regardless of their newer `iat`.
- Make disable fail closed in this order: commit membership denial and the new authentication boundary, invalidate application sessions, globally sign out Cognito sessions, then disable the Cognito user. Make retries and explicit repair converge without restoring access.
- Keep revoke-only separate from disable: advance the authentication boundary and invalidate sessions while preserving active membership and role, then attempt Cognito global sign-out.
- Add fixed `ADMIN`, `OPERATOR`, and `VIEWER` roles with an explicit permission matrix covering reads, configuration, incidents, scheduler controls, channels/policies, and user administration.
- Prevent disabling or demoting the last active administrator with an AuthTable transaction.
- Project safe non-PII lifecycle audit events into `AppTable` so existing audit reads can serve them, while authority, workflow checkpoints, and repair state remain only in `AuthTable`.
- Prefer synchronous convergence, idempotent API retries, and an explicit administrator repair/reconciliation command over an always-running scheduled repair Lambda. Pending operations remain observable until repaired.
- Add user-management APIs/UI, audit attribution, tests, and operational documentation while following the shared stage lifecycle, cost-budget, and alarm-cap policies instead of redefining them.
- Exclude multi-tenancy, self-signup, custom roles, enterprise federation, machine-to-machine authentication, and email-change workflows.

## Capabilities

### New Capabilities

- `operator-user-management`: Administration of post-bootstrap operator lifecycle in AuthTable, fail-closed revocation/disable ordering, explicit repair, last-admin safety, and safe lifecycle audit projection.
- `operator-rbac`: AuthTable-owned fixed roles, least-privilege permissions, `auth_time` revocation enforcement, and uniform dashboard/direct-API authorization.

### Modified Capabilities

- `audit-event-read-api`: Return actor attribution and safe operator lifecycle events from the AppTable audit projection without making that projection an authorization or repair authority.

## Impact

- Depends on `add-single-tenant-operator-authentication` and extends its canonical AuthTable membership item rather than creating AppTable membership records or another bootstrap path.
- Affects AuthTable keys and sparse indexes, Cognito administration integration, monitor API authorization, dashboard session invalidation, audit projection, SST IAM/environment wiring, OpenAPI/Bruno coverage, and the dashboard Administration navigation and server actions.
- Adds usage-based Cognito, AuthTable, AppTable audit-projection, and CloudWatch activity but no recurring repair schedule or always-on compute. Resource lifecycle, stage attribution, budget notifications, and alarm-count limits defer to `standardize-stage-resource-lifecycle` and `establish-data-recovery-and-capacity-guardrails`.
