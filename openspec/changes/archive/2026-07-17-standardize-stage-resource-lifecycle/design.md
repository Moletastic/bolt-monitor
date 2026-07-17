## Context

The SST app currently accepts an optional stage, hard-codes AWS profile `bolt-monitor`, and creates `AppTable`, queues, subscriptions, a minute schedule, a bootstrap bucket, API/dashboard infrastructure, and supporting SST-managed resources without a lifecycle classification. The documented local and deploy commands both use `staging`. SST removal therefore has no product-level contract telling an operator which state must survive, which resources must disappear, or which AWS account and region are about to change.

Active changes make that ambiguity urgent. `add-single-tenant-operator-authentication` introduces an `AuthTable`, Cognito user pool, and SSM-backed encryption secret intended to retain identity state. `establish-data-recovery-and-capacity-guardrails` introduces AppTable PITR/deletion protection and a detailed restore drill. `strengthen-repository-release-gates` proposes unique smoke stages and cleanup. Without a shared stage policy, unique smoke stages could retain tables and user pools indefinitely, while a persistent stage could be removed as if it were disposable.

This design establishes the lifecycle prerequisite those changes consume. It covers resource ownership, basic protection, disposal, inventory, target confirmation, and retained-resource adoption or retirement. It does not duplicate the recovery change's restore validation, drill, capacity, or budget work.

## Goals / Non-Goals

**Goals:**

- Make every deployable SST stage explicitly `persistent` or `ephemeral` before resource evaluation.
- Protect durable storage and identity resources in approved persistent stages.
- Make ephemeral stages wholly disposable and verify that teardown leaves no stage-owned resources.
- Make AWS account, region, profile, stage, class, owner, and disposal posture visible before credentialed mutations.
- Preserve a straightforward open-source path through documented configuration and preflight commands rather than repository-specific hidden defaults.
- Reconcile named staging, local development, credentialed smoke, auth cutover, and recovery work under one lifecycle policy.

**Non-Goals:**

- Implement authentication, restore-to-new-table automation, recovery drills, load testing, budget alarms, or capacity changes.
- Add a janitor service, always-on cleanup process, control plane, or third-party deployment service.
- Define production account identifiers, owner values, personal profiles, or credentials in source control.
- Guarantee recovery solely from retention; detailed restore integrity and cutover evidence remain owned by `establish-data-recovery-and-capacity-guardrails`.
- Add application APIs or change monitoring-domain behavior.

## Decisions

### Use one explicit deployment-target contract

Every infrastructure mutation will resolve one validated target containing stage name, lifecycle class, owner, service, expected AWS account ID, expected AWS region, and intended AWS profile or credential source. Persistent targets additionally require membership in an explicit approved-stage configuration. Ephemeral targets additionally require `disposable=true` and an expiration/cleanup deadline used for reporting and enforcement by workflows.

The SST configuration and repository deployment wrappers consume the same target contract. Missing class, unknown class, unapproved persistent name, missing owner or AWS target, contradictory disposal flags, or unresolved credentials stop before resource creation. Stage-name prefixes can improve diagnostics but never infer class.

Alternative: treat `prod` as persistent and every other name as ephemeral. Rejected because a typo could silently weaken production protection or destroy a long-lived staging installation. Alternative: make every stage persistent by default. Rejected because unique CI and developer stages would strand retained resources and cost.

### Separate protected names from ephemeral naming

Persistent stage names are explicitly allowlisted in non-secret deployment configuration, with `production` and `prod` reserved as protected aliases even when an installation chooses another production name. Approved persistent names cannot be selected with the ephemeral class. Ephemeral names must satisfy a bounded, collision-resistant naming convention and cannot equal or impersonate protected names through case, separator, prefix, or normalization differences.

The repository can provide an example policy and documented `staging` configuration shape, but account ID, region, owner, and credential source are installation inputs. This keeps upstream open-source defaults safe without embedding one maintainer's AWS identity.

Alternative: commit one universal production/staging registry. Rejected because an open-source repository cannot know downstream AWS accounts or ownership. Alternative: rely only on an interactive prompt. Rejected because CI is non-interactive and prompts are not a durable policy.

### Apply resource behavior from class, not resource-by-resource convention

A central lifecycle policy is passed through the bootstrap wiring and any future infrastructure modules. It supplies required tags and protection/retention settings so new stateful resources do not independently guess behavior. Infrastructure assertions inventory resource kinds and fail when a covered resource omits the policy.

Persistent stages apply `service`, `stage`, and `owner` tags to taggable resources. `AppTable` and, when introduced, `AuthTable` use PITR, DynamoDB deletion protection, and infrastructure retain-on-delete behavior. The Cognito user pool uses deletion protection where supported and retain-on-delete behavior. SSM parameters and SST Secrets containing durable installation material retain on stack removal. Outputs expose names and ARNs or equivalent non-secret identifiers for every intentionally retained resource, never parameter or secret values.

Ephemeral stages apply the same ownership tags plus lifecycle/expiry metadata where supported, but use no retain-on-delete or deletion protection. Tables, user pools, parameters/secrets, schedules, queues, buckets, functions, APIs, dashboard resources, log groups, subscriptions, and SST-managed supporting resources must remain removable. Existing item/session TTLs, log retention, message retention, and other native expiration controls stay bounded where relevant, but resource-level cleanup does not rely on eventual TTL.

Alternative: configure only DynamoDB. Rejected because Cognito, SSM, buckets, queues, schedules, and generated hosting resources can also retain data or cost. Alternative: use AWS tag policies or a new janitor service. Rejected because they add account-level dependencies and do not replace correct SST ownership.

### Treat removal as a verified operation

Ephemeral cleanup runs for success, failure, and cancellation paths. It invokes the pinned SST removal path, handles known deletion prerequisites such as protected resources and non-empty/versioned buckets without weakening persistent policy, then performs a bounded residual inventory using the stage's SST state, CloudFormation/Pulumi/SST metadata as applicable, deterministic names, and ownership tags. Cleanup succeeds only when no stage-owned user pool, DynamoDB table, SSM parameter/SST Secret, EventBridge schedule, SQS queue, S3 bucket, or other stack resource remains. A cleanup failure preserves the original workflow result, reports non-secret identifiers, and provides an idempotent retry/manual procedure.

The expiry deadline is a guardrail for detecting and rejecting stale ephemeral stages in deploy/smoke workflows, not a claim that tags automatically delete resources. No cleanup daemon is added. Operators remain responsible for running the documented cleanup when a workstation cannot complete removal.

Alternative: trust a zero exit status from `sst remove`. Rejected because generated resources or interrupted providers can remain. Alternative: delete every resource matching a broad prefix. Rejected because prefix-only deletion can cross ownership boundaries.

### Require explicit target confirmation before AWS mutation

A preflight resolves the caller identity and effective region through AWS APIs and prints a non-secret deployment summary: application, stage, class, disposable posture, owner, account ID, region, and profile or credential-source label. It compares resolved account/region with expected values and fails on mismatch. Deploy and removal require a non-interactive confirmation value bound to that summary; local operator commands may provide an explicit confirmation after review, while CI supplies protected configuration. Secrets and raw credentials are never printed.

Persistent removal requires a separate destructive intent from ordinary deployment. It cannot be reached through the ephemeral cleanup path. This satisfies explicit confirmation without making automation depend on a TTY.

Alternative: continue trusting the profile hard-coded in `sst.config.ts`. Rejected because profile names are workstation-specific and do not prove account or region. Alternative: ask only `Are you sure?`. Rejected because a generic prompt neither identifies the target nor works safely in CI.

### Give retained resources an adoption and retirement lifecycle

Persistent stack outputs and a versioned runbook form the retained-resource inventory. Before changing a persistent logical name or removing a stack, operators capture the target summary and inventory. Re-adoption uses the pinned SST/Pulumi import or supported state-adoption mechanism, verifies physical identifiers and tags, previews no replacement, and only then applies. If the pinned tool cannot safely import a resource kind, the runbook blocks automated replacement and requires a separately reviewed migration.

Deliberate retirement requires evidence/backup decisions, dependent-service shutdown, a fresh inventory, explicit removal of deletion protection and retain policy for the named resources, deletion, and residual verification. The runbook never recommends clearing protection merely to make a routine deploy pass. AppTable restore-to-new-table, integrity validation, and recovery drills are referenced to the recovery/capacity change rather than redefined here.

Alternative: leave retained resources unmanaged after stack removal. Rejected because unowned tables, user pools, and parameters create recovery ambiguity and ongoing cost. Alternative: automatically delete retained resources after a timeout. Rejected because persistent data retirement requires an operator decision.

### Define staging, local, and smoke usage explicitly

The existing `staging` name becomes a deliberately configured long-lived persistent stage when it is used for shared deployment, auth cutover evidence, or recovery work. Local SST development must explicitly choose either that approved persistent staging target or a developer-owned ephemeral target; documentation must not silently map an omitted stage to either class. An ephemeral local target is removed and verified when no longer needed.

Credentialed release smoke has two valid modes: an isolated explicitly ephemeral stage with always-run verified cleanup, or the named long-lived persistent staging stage with no teardown. It may not create a unique persistent stage. The release-gate implementation must reconcile its current unique-stage plan with this rule before cloud smoke is enabled.

Alternative: preserve `staging` as an implicit catch-all for local, shared, and CI use. Rejected because concurrent ownership and cleanup intent remain ambiguous. Alternative: require every smoke to mutate persistent staging. Rejected because it weakens isolation; ephemeral smoke remains the preferred current-revision proof.

### Sequence basic protection before auth and detailed recovery

Implementation first lands target validation, tags, persistent AppTable PITR/deletion/retain protection, ephemeral behavior, inventory outputs, and runbooks. Authentication infrastructure then consumes the same policy for `AuthTable`, Cognito, and SSM/SST Secret resources before protected-route cutover. The recovery/capacity change remains responsible for detailed restore drills and can extend, but not redefine, the stage classes.

This sequencing prevents authentication from introducing blanket retained resources into disposable stages and prevents the recovery change from owning a second stage taxonomy.

## Risks / Trade-offs

- [Persistent protection can block legitimate replacement or teardown] -> Require preview, retained inventory, separate destructive intent, and the adoption/retirement runbook rather than automatic protection removal.
- [Ephemeral cleanup can miss provider-generated resources] -> Assert covered resource kinds, use SST state plus ownership tags and deterministic identifiers, and fail cleanup on residual inventory.
- [AWS services differ in tag, deletion-protection, and import support] -> Maintain a tested resource-policy matrix and block unsupported replacement instead of claiming uniform mechanics.
- [Explicit configuration adds setup steps for new contributors] -> Ship an example, one preflight path, actionable validation errors, and documented local ephemeral and named staging recipes.
- [Using shared persistent staging for local development can cause contention] -> Prefer developer-owned ephemeral stages for isolated work and reserve persistent staging for deliberate shared validation.
- [Expiration metadata does not delete resources by itself] -> Describe it as stale-stage detection only and retain always-run plus manual verified cleanup.
- [Cross-change requirements can drift] -> Make this change land first and update auth, recovery, and release-smoke implementation tasks to consume the shared policy without duplicating it.

## Migration Plan

1. Add the validated deployment-target model, protected-name rules, example configuration, preflight identity check, and infrastructure policy assertions without changing deployed resources.
2. Explicitly register the existing long-lived `staging` target for its intended account/region/owner outside committed personal configuration; require local users to choose persistent staging or an ephemeral target.
3. Apply tags and basic PITR, deletion protection, retain behavior, and retained-resource outputs to the existing `AppTable`; preview and verify that no table replacement occurs.
4. Add ephemeral resource options, always-run removal, residual inventory, stale-stage reporting, and idempotent cleanup documentation; prove zero residual resources with a representative ephemeral deployment.
5. Add the persistent inventory, re-adoption/import, and deliberate retirement runbook and test non-destructive preview/adoption checks against non-production fixtures.
6. Reconcile the release-smoke workflow so unique stages are ephemeral; retain named persistent staging as the alternative for credentialed smoke that must preserve state.
7. Require the authentication change to provision `AuthTable`, Cognito, and SSM/SST Secret resources through this policy before auth cutover.
8. Allow the recovery/capacity change to add detailed restore drills, capacity evidence, and budgets after the base lifecycle contract is active.

Rollback of application-independent validation can restore the previous wrapper only before protected resources are deployed. Once a persistent table or identity resource is protected, rollback must keep protection in place and use the adoption/retirement runbook; it must not delete or abandon the resource to recreate the old permissive behavior.

## Open Questions

None. Exact import command syntax and the generated-resource inventory adapters must be verified against pinned SST `4.14.1` during implementation, but unsupported mechanics fail closed and do not alter the lifecycle contract.
