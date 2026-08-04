# Stage Resource Lifecycle

## Target Contract

Every SST mutation reads one ignored target file at `infra/targets/<name>.target.json`. The file declares the stage name, AWS profile, expected AWS account, expected AWS region, lifecycle class (`persistent` or `ephemeral`), owner, service, dashboard origin, and required class-specific configuration. Copy `infra/targets/example.target.json` to a target file and fill in the local AWS identity. Select it with `TARGET=<name>`; the orchestrator resolves its canonical path and passes that same path to SST as `SST_TARGET_FILE`.

AWS credentials come from standard AWS profile or credential-provider configuration, never from the target file. The orchestrator binds `AWS_PROFILE` and `AWS_REGION` from the selected target before invoking AWS APIs. `TARGET_FILE` is an explicit-path input for automation; if supplied with `TARGET`, it must resolve to that named target file or the command fails before preflight and SST invocation.

Persistent targets require `approved: true`, an explicitly configured stage name, and paired `budgetAmountUsd` plus `alertEmails` configuration. A persistent installation that deliberately cannot use AWS Budgets must instead set `budgetAlertsOptOut: true` with a non-empty `budgetAlertsOptOutReason`; omission, partial configuration, and malformed values fail before AWS mutation. `prod` and `production` are reserved as protected aliases even when an installation chooses another production name. Ephemeral targets require `disposable: true` and a valid `expiresAt`; they may omit budget configuration. Expired targets cannot deploy, develop, inspect, invite, or rotate keys; run `make remove-infra TARGET=<name>` to trigger exact-stage verified cleanup. Expiry never deletes AWS resources automatically.

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

SST is pinned to `4.14.1`. Ephemeral `make remove-infra` accepts an expired disposable target and captures its exact-stage SST inventory before invoking the pinned removal path. It reports this non-secret inventory after successful cleanup, then requires exact-stage SST state evidence that the target is not deployed and bounded Resource Groups Tagging API verification for exact `service` and `stage` tags. Missing state evidence, deployed-stage residue, or non-secret orphan ARNs fail cleanup with bounded diagnostics. Resource kinds covered are Cognito, DynamoDB, SSM/SST secrets, EventBridge, SQS, S3, functions, APIs, dashboard resources, logs, subscriptions, and SST support resources. SST state covers generated resources that cannot be listed by ownership tags.

## Verification Evidence

On 2026-07-15, `smoke-20260715` deployed as an explicit ephemeral target in AWS account `REDACTED_AWS_ACCOUNT_ID`, region `us-east-1`. The deployment included the application table, queues, schedule, bucket, API, dashboard, and SST-generated resources. AppTable deletion protection was disabled as required. Removal initially exceeded the command time limit while CloudFront tore down; an exact stage retry after SST unlock completed successfully. Final ownership-tag inventory was zero and SST reported no resources left to remove.

Persistent staging was deployed in the same account and region. The existing `bolt-monitor-staging-AppTableTable-coumsncm` physical table name remained unchanged, with PITR, deletion protection, and `service`, `stage`, and `owner` tags verified after deploy.

The credentialed `smoke-auth-20260717` ephemeral stage deployed the current application together with an `AuthTable`, Cognito user pool, and stage-scoped AES key parameter. Removal deleted the SST stack and the stage-scoped key. The verifier initially detected a stale Cognito tagging record, then confirmed through Cognito that the pool no longer existed and completed with zero residual resources. The cleanup retry also confirmed that an absent key parameter is idempotent.

The persistent `staging` inventory was checked without mutation: AppTable and AuthTable retained their physical identifiers and deletion protection, AppTable PITR was enabled, and AuthTable and Cognito ownership tags matched the target. The key was recorded only as a SecureString parameter name and version.

### Local staging verification

Repository CI never deploys or receives AWS credentials. After deliberately deploying a configured target from a workstation, run `make deploy-infra`; the orchestrator requires unambiguous flat or selected-stage SST outputs containing `apiUrl`, `dashboardUrl`, `appTableName`, and `authTableName`, then validates the existing public health JSON envelope (`status: "success"`, `data.status: "ok"`). Persistent targets additionally verify DynamoDB deletion protection and point-in-time recovery through exact reads for both output-named tables. Authentication flow validation is performed manually through the dashboard sign-in, invitation activation, optional TOTP enrollment, and protected API access paths.

## Authentication Cutover Gate

Before protected-route cutover, run `make test-infra`. Infrastructure tests prove validated stage classification, persistent `AppTable` deletion and retain-on-delete protection, lifecycle-guarded ephemeral cleanup, the retained inventory including the auth key reference, and the destructive-intent gate for persistent removal. They do not deploy AWS resources or run credentialed smoke checks.

## Cost Posture

Persistent retained tables and future identity material deliberately incur storage and identity cost. Ephemeral orphaned tables, queues, buckets, schedules, logs, APIs, functions, and generated resources can incur fixed or usage cost; native TTL, message retention, log retention, and object expiry reduce data lifetime but do not replace verified removal. No janitor, always-on cleanup service, or new AWS service is added.
