# Stage Resource Lifecycle

## Target Contract

Every SST mutation reads one ignored target file at `infra/targets/<name>.target.json`. The file declares the stage name, AWS profile, expected AWS account, expected AWS region, lifecycle class (`persistent` or `ephemeral`), owner, service, dashboard origin, and required class-specific configuration. Copy `infra/targets/example.target.json` to a target file and fill in the local AWS identity.

AWS credentials come from standard AWS profile or credential-provider configuration, never from the target file. The orchestrator binds `AWS_PROFILE` and `AWS_REGION` from the selected target before invoking AWS APIs.

Persistent targets require `approved: true` and an explicitly configured stage name. `prod` and `production` are reserved as protected aliases even when an installation chooses another production name. Ephemeral targets require `disposable: true` and a future `expiresAt`; expiry detects stale targets and does not delete AWS resources automatically.

`staging` is persistent only when its target file explicitly approves it for deliberate shared validation. Prefer a developer-owned ephemeral target for local work. Never omit a target and never use a unique persistent smoke stage.

## Resource Policy Matrix (v1)

| Resource | Persistent | Ephemeral | Inventory / Cleanup |
| --- | --- | --- | --- |
| `AppTable` | PITR, deletion protection, retain on delete | no protection, no retain | retained table name and ARN / remove and verify tag ownership |
| `AuthTable`, Cognito, durable parameters and secrets | `AuthTable` PITR/deletion protection/retain; Cognito protection/retain; AES parameter retained | no protection/PITR/retention; exact-stage cleanup deletes the AES parameter with the stage | identifiers only, never values; auth details in `docs/auth-operations.md` |
| Bucket | removable; object expiry remains bounded | removable; object expiry remains bounded | ownership tags and SST state |
| Queues, schedules, API, functions, log groups, subscriptions | removable, not durable installation state | removable | ownership tags and SST state |
| Dashboard and generated SST support resources | removable | removable | SST state plus ownership tags where supported |

Provider default tags apply `service`, `stage`, `owner`, `lifecycle`, and, for ephemeral targets, `expiresAt` to every taggable AWS resource. The bootstrap stack has no stage-name conditionals: policy derives from validated target. The non-printing AES-key helper applies the same policy tags to its SSM parameter; persistent inventory lists its name but never its value.

SST is pinned to `4.14.1`. Ephemeral `make remove-infra` invokes the pinned SST removal path and bounded Resource Groups Tagging API verification for exact `service` and `stage` tags; it reports non-secret orphan ARNs. Resource kinds covered are Cognito, DynamoDB, SSM/SST secrets, EventBridge, SQS, S3, functions, APIs, dashboard resources, logs, subscriptions, and SST support resources. Cleanup also requires SST state to report the target as not deployed, covering generated resources that cannot be listed by ownership tags.

## Verification Evidence

On 2026-07-15, `smoke-20260715` deployed as an explicit ephemeral target in AWS account `045104965990`, region `us-east-1`. The deployment included the application table, queues, schedule, bucket, API, dashboard, and SST-generated resources. AppTable deletion protection was disabled as required. Removal initially exceeded the command time limit while CloudFront tore down; an exact stage retry after SST unlock completed successfully. Final ownership-tag inventory was zero and SST reported no resources left to remove.

Persistent staging was deployed in the same account and region. The existing `bolt-monitor-staging-AppTableTable-coumsncm` physical table name remained unchanged, with PITR, deletion protection, and `service`, `stage`, and `owner` tags verified after deploy.

The credentialed `smoke-auth-20260717` ephemeral stage deployed the current application together with an `AuthTable`, Cognito user pool, and stage-scoped AES key parameter. Removal deleted the SST stack and the stage-scoped key. The verifier initially detected a stale Cognito tagging record, then confirmed through Cognito that the pool no longer existed and completed with zero residual resources. The cleanup retry also confirmed that an absent key parameter is idempotent.

The persistent `staging` inventory was checked without mutation: AppTable and AuthTable retained their physical identifiers and deletion protection, AppTable PITR was enabled, and AuthTable and Cognito ownership tags matched the target. The key was recorded only as a SecureString parameter name and version.

### Local staging verification

Repository CI never deploys or receives AWS credentials. After deliberately deploying the configured persistent staging target from a workstation, run `make deploy-infra`; the orchestrator verifies SST outputs, persistent protections, and public health. Authentication flow validation is performed manually through the dashboard sign-in, invitation activation, optional TOTP enrollment, and protected API access paths.

## Authentication Cutover Gate

Before protected-route cutover, run `make check-auth-cutover-prerequisites`. This deterministic release gate proves the validated stage classification, persistent `AppTable` deletion and retain-on-delete protection, lifecycle-guarded ephemeral cleanup, the retained inventory including the auth key reference, and the destructive-intent gate for persistent removal. It does not deploy AWS resources or run credentialed smoke checks.

## Cost Posture

Persistent retained tables and future identity material deliberately incur storage and identity cost. Ephemeral orphaned tables, queues, buckets, schedules, logs, APIs, functions, and generated resources can incur fixed or usage cost; native TTL, message retention, log retention, and object expiry reduce data lifetime but do not replace verified removal. No janitor, always-on cleanup service, or new AWS service is added.