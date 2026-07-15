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
| `AuthTable`, Cognito, durable parameters and secrets | future auth modules consume durable policy | future auth modules consume removable policy | identifiers only, never values |
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
not delete resources only because their names share a prefix.

`make preview-infra` fails closed because SST `4.14.1` has no safe preview
command. It must never map to `sst diff`; that command enters SST's deploy path.

## Cost posture

Persistent retained tables and future identity material deliberately incur
storage and identity cost. Ephemeral orphaned tables, queues, buckets,
schedules, logs, APIs, functions, and generated resources can incur fixed or
usage cost; native TTL, message retention, log retention, and object expiry
reduce data lifetime but do not replace verified removal. No janitor, always-on
cleanup service, or new AWS service is added.
