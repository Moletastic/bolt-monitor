# Authentication Staging Cutover Evidence

Date: 2026-07-17

Target: `staging` in `us-east-1`, account `045104965990`.

## Completed Checks

- A bootstrapped `ACTIVE` `DEFAULT`/`ADMIN` operator completed invitation activation and custom dashboard sign-in.
- Dashboard logout invalidates the server-side session and expires the host cookie.
- Direct Cognito authentication with the no-secret operator client produced an access token; Bruno `GET /api/v1/services` succeeded with the inherited Bearer credential.
- The same request rejected an ID token with Gateway HTTP 401. Anonymous versioned requests also return Gateway HTTP 401, while `GET /api/health` returns the public success envelope.
- Optional TOTP enrollment and a subsequent dashboard TOTP challenge succeeded.
- A strongly consistent membership update to `DISABLED` denied the operator's next protected request. Restoring `ACTIVE` and advancing `AuthValidAfter` required a fresh full authentication ceremony before access resumed.
- Rotating the active SSM auth key from generation 1 to 2 invalidated the existing dashboard session. A new session encrypted under generation 2 was established after fresh sign-in; generation-1 session state was not usable.
- The AWS-admin membership disable/restore procedure was exercised as the all-admin-lockout break-glass dry run without removing JWT enforcement.
- AuthTable deletion protection and PITR are enabled. The retained inventory identifies AuthTable, Cognito user pool, and the non-secret SSM key reference.
- Auth refresh, storage, and key-loading alarms are `OK`; all use bounded `stage`, `component`, `operation`, and `outcome` dimensions. Auth dashboard logs retain for 14 days and recorded no credentials, tokens, cookies, session references, or request bodies.
- Production auth and protected-page responses provide nonce CSP, HSTS, `nosniff`, restrictive referrer and permissions policies, and frame/base/form/object restrictions.

## Cutover Decision

Staging evidence supports the v1 authentication cutover. Preserve JWT enforcement and persistent authentication resources in rollback; do not reopen anonymous `/api/v1/**` access.
