## 1. Stage Target Contract

- [x] 1.1 Add one typed deployment-target model for stage name, `persistent` or `ephemeral` class, owner, service, expected AWS account, expected region, and profile or credential-source label, including class-specific approval, disposal, and expiration fields.
- [x] 1.2 Add non-secret example configuration and loading rules that allow downstream installations to declare approved persistent stages without committing account-specific credentials or personal owner/profile values.
- [x] 1.3 Implement fail-closed validation for missing, unknown, inferred, incomplete, or contradictory stage configuration and for unapproved persistent names.
- [x] 1.4 Implement normalized protected-name validation so `prod`, `production`, and every approved persistent name or confusable normalized form cannot be selected as ephemeral.
- [x] 1.5 Add focused tests for valid persistent and ephemeral targets, omitted class, missing owner/account/region, disposal conflicts, expiration omission, unapproved persistent stages, and protected ephemeral names.

## 2. AWS Target Preflight

- [x] 2.1 Replace the hard-coded profile assumption with an explicit credential-source input that preserves the documented `bolt-monitor` local recipe without making it a silent fallback for every installation.
- [x] 2.2 Add a preflight that resolves the effective AWS caller account and region, compares them with expected configuration, and renders a secret-safe summary of application, stage, class, disposal posture, owner, account, region, and credential-source label.
- [x] 2.3 Require non-interactive confirmation bound to the rendered target for deploy, remove, import/adoption, and protection mutations, with separate destructive intent for persistent removal or protection changes.
- [x] 2.4 Add tests proving account/region mismatch, stale confirmation, and missing destructive intent fail before SST mutation and proving logs contain no raw credentials or secret values.

## 3. Central Resource Lifecycle Policy

- [x] 3.1 Inventory all directly declared and SST-generated resource kinds in the bootstrap stack and record the required persistent, ephemeral, tagging, expiration, output, and cleanup behavior for each.
- [x] 3.2 Add a central lifecycle policy consumed by bootstrap resource construction so covered resources receive class-derived options rather than independent stage-name conditionals.
- [x] 3.3 Apply `service`, `stage`, and `owner` tags to every taggable current resource and lifecycle/expiration metadata to ephemeral resources where AWS supports it.
- [x] 3.4 Configure persistent `AppTable` with PITR, DynamoDB deletion protection, and retain-on-delete behavior while keeping on-demand capacity and existing TTL support unchanged.
- [x] 3.5 Ensure persistent bucket, queue, schedule, API, dashboard, function, log, and generated-supporting-resource behavior is explicit without retaining resources that are not classified as durable installation state.
- [x] 3.6 Add stack outputs for lifecycle class, non-secret deployment identity, and the complete retained-resource inventory without outputting parameter, SST Secret, or credential values.
- [x] 3.7 Add infrastructure assertions that fail when a current or newly covered resource bypasses required class policy, tags, protection, disposal, or retained-inventory output.

## 4. Ephemeral Disposal And Verification

- [x] 4.1 Configure every ephemeral resource with no retain-on-delete and no deletion protection while preserving bounded native item TTL, message retention, log retention, and expiration controls where applicable.
- [x] 4.2 Implement an idempotent cleanup entrypoint around the pinned SST removal path that handles success, failure, and cancellation and does not invoke persistent retirement behavior.
- [x] 4.3 Verify pinned SST `4.14.1` state/output/removal behavior and implement bounded residual discovery using stack state, deterministic identifiers, and ownership tags rather than broad prefix-only deletion.
- [x] 4.4 Make residual verification cover Cognito user pools, DynamoDB tables, SSM parameters and SST Secrets, EventBridge schedules, SQS queues, S3 buckets, functions, APIs, dashboard resources, log groups, subscriptions, and SST-managed supporting resources.
- [x] 4.5 Report cleanup failures and non-secret orphan identifiers without hiding the original workflow result, and support safe idempotent retry by exact stage ownership.
- [x] 4.6 Add fixture tests for interrupted removal, non-empty or versioned buckets, partial provider failure, missing state, foreign similarly named resources, repeated cleanup, and zero-residual success.

## 5. Persistent Inventory And Operations

- [x] 5.1 Write a versioned retained-resource inventory and re-adoption/import runbook covering target confirmation, physical identity and tag checks, pinned SST/Pulumi support, no-replacement preview, apply verification, and fail-closed handling for unsupported resource kinds.
- [x] 5.2 Write a deliberate persistent-retirement runbook covering evidence or backup decisions, dependent-service shutdown, fresh inventory, separate destructive approval, targeted protection removal, deletion, and residual verification.
- [x] 5.3 Keep restore-to-new-table integrity validation, recovery cutover, rollback evidence, and measured restore drills referenced to `establish-data-recovery-and-capacity-guardrails` rather than implementing duplicate recovery procedures here.
- [x] 5.4 Add non-destructive runbook checks or fixtures that validate retained identifiers and replacement previews without deleting a real persistent installation.

## 6. Local, Staging, Smoke, And Auth Integration

- [x] 6.1 Update Make/package entrypoints and contributor documentation so local SST, preview, deploy, and remove require an explicit validated target and never infer `staging` or lifecycle class from omission.
- [x] 6.2 Document `staging` as approved persistent only for deliberate shared validation and provide a developer-owned ephemeral local recipe with explicit cleanup and stale-stage guidance.
- [x] 6.3 Reconcile credentialed release smoke so a unique current-revision stage is explicitly ephemeral with always-run zero-residual verification, while named persistent staging is the only non-ephemeral smoke alternative and is never torn down by smoke.
- [x] 6.4 Add lifecycle guards/tests that reject unique persistent smoke stages and prevent ephemeral cleanup from targeting approved persistent staging or protected production names.
- [x] 6.5 Make the lifecycle policy available to the authentication infrastructure so persistent `AuthTable`, Cognito user pool, and durable SSM/SST secret material receive required protection, inventory, and tags while ephemeral auth resources remain removable.
- [x] 6.6 Add an explicit release gate proving stage classification, basic persistent `AppTable` protection, ephemeral cleanup, and retained inventory are active before authentication route cutover proceeds.
- [x] 6.7 Reconcile overlapping auth, recovery/capacity, and release-gate implementation documentation to consume this capability without introducing a second stage taxonomy or a unique retained smoke stage.

## 7. Cost, Validation, And Rollout

- [x] 7.1 Document persistent retained-storage and identity cost posture, ephemeral fixed/usage orphan risks, expiration limitations, and the decision to add no always-on cleanup service or new AWS service.
- [x] 7.2 Run infrastructure formatting and type checks plus target-validation, lifecycle-policy, output, and cleanup test suites.
- [x] 7.3 Preview the approved persistent staging change, verify the effective account/region/profile summary, and record evidence that AppTable protection and tags do not replace the existing table.
- [x] 7.4 Deploy a representative explicitly ephemeral stage containing the current table, queues, schedules, bucket, API, dashboard, and generated resources; remove it through the supported path and record zero-residual inventory evidence.
- [x] 7.5 After auth resources are available, repeat ephemeral lifecycle validation with `AuthTable`, Cognito, and SSM/SST secret material and verify none are retained.
- [x] 7.6 Exercise persistent retained-inventory and non-destructive adoption preview against staging fixtures, leaving detailed restore drill execution to the recovery/capacity change.
- [x] 7.7 Run `openspec validate standardize-stage-resource-lifecycle --strict` and confirm the change remains apply-ready after implementation documentation is finalized.
