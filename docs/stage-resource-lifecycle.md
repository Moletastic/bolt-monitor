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
| `AuthTable`, Cognito, durable parameters and secrets | `AuthTable` PITR/deletion protection/retain; Cognito protection/retain; AES parameter retained | no protection/PITR/retention; delete with stage | identifiers only, never values; auth details in `docs/auth-operations.md` |
| Bucket | removable; object expiry remains bounded | removable; object expiry remains bounded | ownership tags and SST state |
| Queues, schedules, API, functions, log groups, subscriptions | removable, not durable installation state | removable | ownership tags and SST state |
| Dashboard and generated SST support resources | removable | removable | SST state plus ownership tags where supported |

Provider default tags apply `service`, `stage`, `owner`, `lifecycle`, and, for
ephemeral targets, `expiresAt` to every taggable AWS resource. The bootstrap
stack has no stage-name conditionals: policy derives from validated target.

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

Credentialed smoke selects exactly one lifecycle: a unique, disposable
ephemeral target with verified cleanup, or declared persistent `staging` with
no teardown. Persistent target names beginning with `smoke` are rejected even
if configured, preventing unique retained smoke installations.

## Cost posture

Persistent retained tables and future identity material deliberately incur
storage and identity cost. Ephemeral orphaned tables, queues, buckets,
schedules, logs, APIs, functions, and generated resources can incur fixed or
usage cost; native TTL, message retention, log retention, and object expiry
reduce data lifetime but do not replace verified removal. No janitor, always-on
cleanup service, or new AWS service is added.
