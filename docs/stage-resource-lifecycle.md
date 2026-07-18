# Stage Resource Lifecycle

## Target contract

Every SST mutation requires `SST_STAGE`, `SST_TARGET_CONFIG`, and a target
entry. Copy `infra/deployment-target.example.json` to an ignored local file and
replace all example identity values. The file contains no credentials; its
`credentialSource` is a display label only. AWS credentials come from the
explicit `AWS_PROFILE`, workload identity, or other AWS SDK/CLI source.

Persistent targets require `approved: true` and an explicitly configured stage
name. `prod`, `production`, and every approved persistent name, after removing
case and separators, are blocked from ephemeral use. Ephemeral targets require
`disposable: true` and a future `expiresAt`; expiry detects stale targets and
does not delete AWS resources automatically.

`staging` is persistent only when its installation configuration approves it
for deliberate shared validation. Prefer a developer-owned ephemeral name for
local work. Never omit a target and never use a unique persistent smoke stage.

## Resource policy matrix (v1)

| Resource | Persistent | Ephemeral | Inventory / cleanup |
| --- | --- | --- | --- |
| `AppTable` | PITR, deletion protection, retain on delete | no protection, no retain | retained table name and ARN / remove and verify tag ownership |
| `AuthTable`, Cognito, durable parameters and secrets | `AuthTable` PITR/deletion protection/retain; Cognito protection/retain; AES parameter retained | no protection/PITR/retention; exact-stage cleanup deletes the AES parameter with the stage | identifiers only, never values; auth details in `docs/auth-operations.md` |
| Bucket | removable; object expiry remains bounded | removable; object expiry remains bounded | ownership tags and SST state |
| Queues, schedules, API, functions, log groups, subscriptions | removable, not durable installation state | removable | ownership tags and SST state |
| Dashboard and generated SST support resources | removable | removable | SST state plus ownership tags where supported |

Provider default tags apply `service`, `stage`, `owner`, `lifecycle`, and, for
ephemeral targets, `expiresAt` to every taggable AWS resource. The bootstrap
stack has no stage-name conditionals: policy derives from validated target. The
non-printing AES-key helper applies the same policy tags to its SSM parameter;
persistent inventory lists its name but never its value.

SST is pinned to `4.14.1`. Ephemeral `make remove-infra` invokes the pinned SST
removal path and bounded Resource Groups Tagging API verification for exact
`service` and `stage` tags; it reports non-secret orphan ARNs. Resource kinds
covered are Cognito, DynamoDB, SSM/SST secrets, EventBridge, SQS, S3, functions,
APIs, dashboard resources, logs, subscriptions, and SST support resources. Do
not delete resources only because their names share a prefix. Cleanup also
requires SST state to report the target as not deployed, covering generated
resources that cannot be listed by ownership tags.

`make preview-infra` fails closed because SST `4.14.1` has no safe preview
command. It must never map to `sst diff`; that command enters SST's deploy path.

## Verification Evidence

On 2026-07-15, `smoke-20260715` deployed as an explicit ephemeral target in
AWS account `045104965990`, region `us-east-1`. The deployment included the
application table, queues, schedule, bucket, API, dashboard, and SST-generated
resources. AppTable deletion protection was disabled as required. Removal
initially exceeded the command time limit while CloudFront tore down; an exact
stage retry after SST unlock completed successfully. Final ownership-tag
inventory was zero and SST reported no resources left to remove.

Persistent staging was deployed in the same account and region. The existing
`bolt-monitor-staging-AppTableTable-coumsncm` physical table name remained
unchanged, with PITR, deletion protection, and `service`, `stage`, and `owner`
tags verified after deploy. SST `4.14.1` has no safe preview command, so the
workflow fails closed rather than claiming a no-replacement preview.

## Credentialed Smoke

On 2026-07-17, the disposable `smoke-auth-20260717` stage deployed the
current application together with an `AuthTable`, Cognito user pool, and
stage-scoped AES key parameter. Removal deleted the SST stack and the
stage-scoped key. The verifier initially detected a stale Cognito tagging
record, then confirmed through Cognito that the pool no longer existed and
completed with zero residual resources. The cleanup retry also confirmed that
an absent key parameter is idempotent.

The persistent `staging` inventory was checked without mutation: AppTable and
AuthTable retained their physical identifiers and deletion protection, AppTable
PITR was enabled, and AuthTable and Cognito ownership tags matched the target.
The key was recorded only as a SecureString parameter name and version. The
adoption entrypoint failed closed because SST `4.14.1` has no safe automated
adoption preview; the re-adoption runbook remains required before any mutation.

Credentialed smoke selects exactly one lifecycle: a unique, disposable
ephemeral target with verified cleanup, or declared persistent `staging` with
no teardown. Persistent target names beginning with `smoke` are rejected even
if configured, preventing unique retained smoke installations.

### Local staging smoke

Repository CI never deploys or receives AWS credentials. After deliberately
deploying the declared persistent staging target from a workstation, run
`make smoke-staging` with `SST_STAGE`, `SST_OUTPUT_PATH`, and a locally
acquired `STAGING_SMOKE_ACCESS_TOKEN` in the local environment. The helper reads the deployed API and direct Cognito client from
structured SST output, rejects production stage names, and checks public
health, anonymous Gateway `401`, and authenticated read-only access. It does
not print credentials, tokens, or authorization headers.

```sh
SST_STAGE=staging \
SST_OUTPUT_PATH=infra/.sst/outputs.json \
STAGING_SMOKE_ACCESS_TOKEN=<access-token-after-mfa> \
make smoke-staging
```

To set the token without copying it from Bruno, export local direct-client
credentials and the current TOTP code, then evaluate the helper output. The
token is emitted only to the calling shell's `eval` and is not written to a
file:

```sh
eval "$(node scripts/cognito-access-token.mjs)"
```

The helper requires local `COGNITO_REGION`, `COGNITO_CLIENT_ID`,
`COGNITO_USERNAME`, `COGNITO_PASSWORD`, and, when Cognito requests it,
`COGNITO_MFA_CODE` environment variables.

SST `4.14.1` writes `infra/.sst/outputs.json` after a persistent deploy; the
lifecycle verifier reads it directly. Point the smoke helper at that file:

```sh
AWS_PROFILE=bolt-monitor SST_STAGE=staging \
  SST_TARGET_CONFIG="$HOME/.config/bolt-monitor/deployment-target.json" \
  SST_LIFECYCLE_ACTION=deploy node scripts/sst-lifecycle.mjs
```

The deploy prints the path; pass it through unchanged:

```sh
SST_STAGE=staging \
SST_OUTPUT_PATH=infra/.sst/outputs.json \
STAGING_SMOKE_ACCESS_TOKEN=<token> \
make smoke-staging
```

Then point the helper at that file:

```sh
SST_STAGE=staging \
SST_OUTPUT_PATH=infra/.sst/outputs.json \
STAGING_SMOKE_ACCESS_TOKEN=<token> \
make smoke-staging
```

## Authentication Cutover Gate

Before protected-route cutover, run
`make check-auth-cutover-prerequisites`. This deterministic release gate proves
the validated stage classification, persistent `AppTable` deletion and
retain-on-delete protection, lifecycle-guarded ephemeral cleanup, and the
retained inventory including the auth key reference. It does not deploy AWS
resources or replace the later auth-resource ephemeral cleanup evidence.

## Cost posture

Persistent retained tables and future identity material deliberately incur
storage and identity cost. Ephemeral orphaned tables, queues, buckets,
schedules, logs, APIs, functions, and generated resources can incur fixed or
usage cost; native TTL, message retention, log retention, and object expiry
reduce data lifetime but do not replace verified removal. No janitor, always-on
cleanup service, or new AWS service is added.
