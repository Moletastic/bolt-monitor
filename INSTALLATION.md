# Installation

## Prerequisites

- AWS CLI authenticated to the intended account.
- Node.js 22, pnpm 10, and Go 1.26 or later.
- An HTTPS dashboard origin for Cognito callback configuration.

## First Deployment

1. Clone the repository and install locked dependencies:

   ```sh
   make setup
   ```

2. Create a target file from the example:

   ```sh
   cp infra/targets/example.target.json infra/targets/staging.target.json
   ```

3. Edit the ignored target file. Set its AWS profile, account, region, owner,
   lifecycle, and `dashboardOrigin`. Persistent targets also require budget
   settings or an explicit documented opt-out. See `docs/stage-resource-lifecycle.md`.

4. Verify selected target identity before mutation:

   ```sh
   make infra-status
   ```

5. Deploy:

   ```sh
   make deploy-infra
   ```

   Deployment verifies public `GET /api/health`. This is liveness only: it
   confirms the health Lambda runs and does not verify Cognito, authorization,
   DynamoDB, or protected monitor API reads.

6. Invite the initial administrator:

   ```sh
   make invite-admin EMAIL=operator@example.com
   ```

7. Complete invitation activation, then verify dashboard sign-in and one
   protected API read.

## Health Operations

- Public liveness: `GET /api/health` returns no dependency status.
- Protected readiness: configure a synthetic administrator after deployment and
  verify an authenticated v1 read separately. Use one explicit identity for both
  commands:

  ```sh
  TARGET=staging make setup-readiness EMAIL=readiness@example.com
  TARGET=staging make readiness-api EMAIL=readiness@example.com
  ```

  Setup stores a generated password in the target-scoped SSM SecureString
  `/<service>/<stage>/readiness/password`. Re-running setup is idempotent. Rotate
  the password only when required:

  ```sh
  TARGET=staging make setup-readiness EMAIL=readiness@example.com ROTATE=yes
  ```

- The readiness probe reads its password from SSM, authenticates with the
  dedicated Cognito client, and calls `GET /api/v1/services?limit=1`. Passwords
  and access tokens are never command arguments or output. Do not place either
  in target files, shell history, logs, or repository files.
- If readiness fails, first confirm the target identity with `make infra-status`.
  Then confirm the selected `EMAIL` matches setup, the stack has been deployed
  with the readiness client output, and the invoking AWS identity can decrypt the
  target SSM parameter. Rotate with `ROTATE=yes` if the user password is known to
  be compromised.

## Teardown

1. Disable external traffic, CI deploys, and any readiness probes for the target.
2. For an ephemeral target, run:

   ```sh
   make remove-infra
   ```

3. For a persistent target, confirm backup/export requirements, then provide
   explicit destructive intent:

   ```sh
   DESTROY=yes make remove-infra
   ```

4. Confirm CloudFormation/SST resources are gone and remove the local ignored
   target file only after no further operations need it.

Persistent removal is irreversible. Retained data or manually managed AWS
resources require their own documented cleanup before deletion.
