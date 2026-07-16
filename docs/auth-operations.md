# Authentication Operations (v1)

This runbook operates the invite-only operator authentication boundary. It is
for authorized AWS administrators; it does not create a public bootstrap or
recovery endpoint. The authoritative application record is `AuthTable`, not
`AppTable`, Cognito groups, email, or token claims.

## Prerequisites

- A completed deployment with an explicit validated target. Follow
  [stage-resource-lifecycle.md](./stage-resource-lifecycle.md); `staging` is
  persistent only when explicitly approved, and a unique smoke or developer
  stage must be explicitly ephemeral.
- AWS credentials for the declared account and region. Deployment wrappers
  require `SST_STAGE`, `SST_TARGET_CONFIG`, and a target-bound
  `SST_TARGET_CONFIRMATION`; persistent protection changes or retirement also
  require `SST_DESTRUCTIVE_CONFIRMATION`.
- The non-secret deployment outputs: `operatorUserPoolId`, `authTableName`,
  and, for key rotation, `authEncryptionKeyParameterName`. Do not put values
  from Cognito responses, dashboard cookies, or key material in shell history,
  tickets, source control, or deployment configuration.
- For dashboard use, the canonical `DASHBOARD_ORIGIN` and the deployed
  dashboard configuration. For direct API use, the direct-operator Cognito
  client ID and region in an ignored Bruno local environment.

The user pool uses Cognito Essentials and Cognito default email delivery. It
does not use self-registration, a managed-login domain, Amplify, custom DNS,
or SES. Default Cognito email is appropriate only for the small operator
population: account and Region delivery quotas, suppression, and rate limits
apply. Check the current Cognito service quota and delivery status before a
large invite/recovery event; investigate invitation or recovery failures in
CloudWatch and AWS service evidence. A requirement for branded, higher-volume,
or independently managed delivery needs a separate change, not an ad-hoc SES
configuration.

## Initial And Later Invitations

Create the first administrator only after the deployed target and outputs are
verified. The command is idempotent: it discovers one normalized email,
creates or reconciles exactly one Cognito identity and complete membership,
then sends an invitation only if activation is still needed. It preserves an
established password, immutable identity fields, and a non-decreasing
`AuthValidAfter` value. It stops on ambiguous email or conflicting membership
state rather than overwriting it.

```sh
SST_STAGE=<stage> \
EMAIL=operator@example.com \
USER_POOL_ID=<operatorUserPoolId> \
AUTH_TABLE_NAME=<authTableName> \
make bootstrap-admin
```

Run the same command with a later operator email for a subsequent
AWS-admin-controlled invitation. There is no self-service sign-up and no
dashboard user-management flow in v1. Bootstrap creates the membership before
invitation delivery, with immutable `MembershipID` and Cognito `Subject`,
`TenantID=DEFAULT`, `Status=ACTIVE`, `Role=ADMIN`, `AuthValidAfter`, version,
and timestamps. Confirm the invited person completes a full sign-in after the
record's `AuthValidAfter` boundary.

The command emits a secret-safe JSON audit outcome with stage, acting AWS
principal when available, target subject, authority, and correlation ID. Keep
that outcome and corresponding CloudTrail events as the invitation record.

## Membership Authority And Revocation

V1 recognizes only `ACTIVE` membership with `Role=ADMIN` in tenant `DEFAULT`.
Missing, malformed, non-active, wrong-tenant, or unsupported-role records deny
access without disclosing which condition occurred. `AuthTable` has no
membership TTL; session and transaction TTL is cleanup only, never authority.

`AuthValidAfter` is a non-negative whole Unix second and is the revocation
authority. Authorization requires an access token whose `auth_time` is
strictly greater than it:

```text
authorize only when auth_time > AuthValidAfter
```

The denial boundary is inclusive: a ceremony at exactly the stored second, or
earlier, is denied. Disabling a member makes the next request fail through the
strongly consistent membership read. Re-enabling a member or revoking existing
authentication must advance, never lower, `AuthValidAfter` to the documented
whole-second boundary; the operator must then perform a new full Cognito
authentication ceremony strictly after that moment.

`iat` is diagnostic metadata only. Refreshing a token family can yield a later
`iat`, but it does not create a later authentication ceremony and cannot pass
the `auth_time` test. ID tokens are not API credentials: API Gateway requires
the Cognito access-token scope and rejects an ID token before Lambda. A
dashboard `__Host-bolt-session` cookie is an opaque browser reference, not an
API credential; it is never accepted in place of `Authorization: Bearer
<access-token>`.

Membership status, role, or epoch changes are privileged AWS operations. Do
not edit a record casually in the DynamoDB console. Until the follow-on user
management capability lands, use the break-glass change controls below: read
the exact record strongly, preserve immutable `MembershipID`, `Subject`,
`TenantID`, and `CreatedAt`, conditionally update `Version` and `UpdatedAt`,
and never reduce `AuthValidAfter`. Record the change reason, subject,
correlation ID, and AWS principal without recording credentials or tokens.

## Dashboard And Direct Clients

The dashboard stores encrypted Cognito token bundles only server-side. Its
cookie is `__Host-bolt-session`, `Secure`, `HttpOnly`, `SameSite=Lax`,
`Path=/`, and has no `Domain`. Sessions expire in application code after 12
hours even if DynamoDB TTL has not yet removed the item. Successful sign-in
replaces any previous transaction or session. Logout conditionally invalidates
the server record, expires the cookie, and succeeds even when it was already
absent. A membership denial invalidates the dashboard session, requiring a
fresh sign-in.

For direct API use, follow
[the Bruno direct-client instructions](../.bruno/collections/bolt-monitor-api/README.md#direct-cognito-authentication).
Use the no-secret direct-operator app client and retain only the access token
in ignored local configuration. Every `/api/v1/**` request needs that Bearer
access token; `GET /api/health` intentionally needs neither token nor cookie.

Access tokens last 60 minutes. A direct client may use Cognito
`REFRESH_TOKEN_AUTH` with its locally held refresh token to acquire a new access
token, or repeat `USER_PASSWORD_AUTH` to reauthenticate. Reacquire through a
new full authentication after a membership epoch advance, re-enable, disable,
password reset, refresh failure, or any authorization denial; refresh alone
does not satisfy the `auth_time` boundary. Never commit or print the refresh
token, returned challenge `Session`, access token, ID token, password, or
dashboard cookie.

## Lifecycle, Recovery, And Deletion

The stage classification in
[stage-resource-lifecycle.md](./stage-resource-lifecycle.md) is the only
lifecycle policy. Auth does not create a second taxonomy.

| Resource | Persistent stage | Ephemeral stage |
| --- | --- | --- |
| Cognito operator user pool | deletion protection and retain-on-delete | no protection or retention; delete with stage |
| `AuthTable` | on-demand, PITR, deletion protection, retain-on-delete | no PITR/protection/retention; delete with stage |
| AES `SecureString` parameter | retain as durable installation material | delete with stage |

Persistent outputs inventory non-secret physical identifiers for `AuthTable`
and the user pool. Capture that inventory with its target summary before
re-adoption, recovery, or retirement. `AuthTable` PITR can restore memberships
and session records to a recovery table during its configured recovery window;
do not point production traffic at a restored table or replace the retained
table without the detailed integrity, cutover, and rollback procedure owned by
`establish-data-recovery-and-capacity-guardrails`.

For an ephemeral removal, run `make remove-infra` with the explicit target.
The lifecycle wrapper uses SST removal then verifies zero stage-owned Cognito,
DynamoDB, SSM/SST secret, and other covered resources by exact ownership. TTL,
log retention, and expiry metadata do not replace this verification. Retry the
same stage cleanup if it reports residual non-secret identifiers; never delete
by a broad name prefix.

For deliberate persistent retirement, follow
[persistent-resource-operations.md](./persistent-resource-operations.md):
capture a fresh inventory, make the evidence/backup decision, stop dependent
services, obtain separate destructive approval, remove protections only from
the exact inventoried resources, delete them deliberately, and record bounded
residual verification. Routine deploy failures never justify clearing auth
resource protection.

## Break Glass

Use break glass only when all administrators are locked out or identity
infrastructure prevents recovery. It requires an authorized AWS identity and
does not disable the API Gateway JWT authorizer, reopen anonymous v1 routes, or
depend on a dashboard session.

The responder role needs, at minimum, `sts:GetCallerIdentity`; scoped Cognito
administration on the target user pool (`ListUsers`, `AdminGetUser`,
`AdminCreateUser`, `AdminResetUserPassword`, `AdminEnableUser`,
`AdminDisableUser`, and `AdminUserGlobalSignOut` when used); and scoped
DynamoDB `GetItem`, `PutItem`, and conditional `UpdateItem` on the target
`AuthTable`. Key rotation additionally requires scoped `ssm:PutParameter` on
the target key parameter. Grant only the actions needed for the approved
incident and keep deployment, application, and break-glass roles separate.

1. Record the incident, exact target summary, AWS principal, and a correlation ID.
2. Use Cognito administration to locate or create one controlled operator and
   reset or enable credentials only when required.
3. Strongly read the subject's membership. Conditionally create or repair only
   a complete `DEFAULT`/`ACTIVE`/`ADMIN` record, preserving immutable fields
   when it already exists and advancing, never decreasing, `AuthValidAfter`.
4. Require the recovered operator to complete a new full sign-in after the
   boundary. If sessions may be compromised, perform Cognito global sign-out
   and invalidate affected dashboard sessions where possible.
5. Rotate the recovered credentials, review CloudTrail and application security
   events, confirm direct and dashboard access, and record the outcome.

AWS denial for a responder missing those permissions is expected; there is no
weaker application fallback.

## AES Key Rotation

Schedule rotation as an authentication-disrupting maintenance operation. It
creates a new 256-bit AES key generation only in process memory and the
non-printing AWS CLI boundary:

```sh
SST_STAGE=<stage> SST_TARGET_CONFIG=<local-target-config> make rotate-auth-key
```

The helper writes the AWS-managed `SecureString` at
`/<service>/<stage>/auth/aes-256-gcm` and does not print or persist the key
value. There is exactly one active generation: no previous-key fallback and no
online re-encryption. Therefore all dashboard sessions and authentication
transactions encrypted under the previous generation are intentionally
invalidated. Notify operators, expect presented cookies to expire, require
fresh sign-in, and review the rotation audit trail and key-loading/session
failure telemetry afterward. See [auth-key-rotation.md](./auth-key-rotation.md)
for the short operational reference.

## Cutover And Rollback

Authentication is an atomic security cutover, never optional per route or
handler. Before cutover, block the release unless all of the following are
true:

1. The target is explicitly classified and lifecycle protections, inventory,
   tags, and ephemeral cleanup verification are active.
2. Auth resources, least-privilege grants, canonical dashboard origin, alarms,
   finite log retention, and non-secret outputs are deployed.
3. At least one fully shaped `ACTIVE` `DEFAULT`/`ADMIN` membership exists,
   established by `make bootstrap-admin`, and the operator has completed a
   later full authentication ceremony.
4. Staging proves custom dashboard login, invitation activation,
   recovery, optional TOTP, logout, session expiry/refresh behavior, and direct
   Cognito/Bruno access.
5. Staging proves anonymous and ID-token `/api/v1/**` requests are rejected
   before Lambda; access-token requests are authorized only with current
   membership; `GET /api/health` remains public.
6. Staging records security headers, log redaction, alarm operation,
   key-rotation invalidation, an all-admin-lockout break-glass dry run, and
   ephemeral auth-resource zero-residual evidence where applicable.
7. Intended API consumers receive notice that anonymous v1 access stops; there
   is no static credential or cookie migration.

Deploy the authorizer and required access scope on every `/api/v1/**` route,
the Lambda membership/epoch enforcement, and dashboard session protection in
the same controlled release. Immediately verify dashboard success, Bruno
access-token success, anonymous and ID-token denial, an advanced-epoch or
non-active membership denial, public health, alarms, and audit events.

If application behavior regresses, roll forward a fix or revert the affected
dashboard/API build to the last authentication-capable version. Preserve JWT
enforcement and persistent-stage `AuthTable`, user pool, and key material. Do
not restore anonymous v1 as a routine rollback, and do not delete persistent
auth data. Use break glass when identity itself blocks all operator access.

## Cost Posture

Review current AWS regional pricing and service quotas for the target account.
Material v1 cost sources are:

- Cognito Essentials monthly active users and Cognito default-email delivery
  limits or applicable messaging charges.
- `AuthTable` on-demand reads, writes, storage, and persistent-stage PITR.
  Each protected API request intentionally performs a strongly consistent
  membership read to enforce immediate disable and epoch revocation.
- SSM `SecureString` retrieval and AWS-managed encryption operations for the
  active key reference.
- CloudWatch log ingestion and two-week retention, metrics or metric filters,
  and alarms. Refresh failures alarm at 5 events per five-minute period for
  three periods; storage and key-loading failures alarm at 3 on the same
  schedule.

V1 adds no customer-managed KMS key, SES, custom DNS, NAT gateway, VPC
endpoint, always-on compute, or cross-region authentication traffic. Persistent
retained identity resources and unremoved ephemeral resources can still incur
storage or usage costs; verified lifecycle cleanup is the control for the
latter.
