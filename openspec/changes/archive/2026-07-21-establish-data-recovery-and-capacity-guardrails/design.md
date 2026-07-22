## Context

Bolt Monitor has two DynamoDB recovery domains. `AppTable` stores monitoring configuration, operational history, incidents, and derived state with three sparse GSIs and native TTL on `TTL`. `AuthTable` is deliberately separate and stores application identity links, encrypted dashboard sessions/auth transactions, membership/RBAC authority, lifecycle operations, guards, and identity audit state. Cognito stores provider-managed identities, credentials, challenges, and token issuance; it is not a DynamoDB backup domain. `CheckRun` items expire after 30 days and `ExecutionWork` after 7 days, while auth sessions/transactions and terminal lifecycle operations follow the auth changes' explicit expiry rules.

Runtime reads are unevenly bounded. Monitor run, incident, and audit history and search use explicit query limits, but service, monitor, channel, policy, global incident, and activity lists use a single query without a limit or continuation. The scheduler first queries all service references and then queries monitors once per service, producing an unbounded N+1 traversal. Several service rollup paths also fetch every monitor status individually. DynamoDB's 1 MB query boundary can therefore look like completion while silently truncating work.

The deployment is single-region and serverless and uses the built-in `DEFAULT` tenant. Prerequisite `standardize-stage-resource-lifecycle` owns the basic persistent/ephemeral stage policy for PITR, deletion protection, and retain-on-delete behavior. This change consumes that baseline and adds detailed retention classes, recovery drills, bounded-read behavior, load evidence, and a measurable cost model.

## Goals / Non-Goals

**Goals:**
- Make durability, source of truth, retention, and recovery validation explicit for every item family.
- Reconcile `AppTable`, `AuthTable`, and Cognito recovery responsibilities without merging their authority, and provide tested restore-to-new-table procedures.
- Bound all growing runtime work and remove the scheduler's per-service N+1 enumeration.
- State and load-test an initial installation envelope instead of implying unlimited scale.
- Produce reproducible low-cost, expected-validation, and high-volume-stress cost scenarios plus optional budget setup guidance.

**Non-Goals:**
- No SLA, RPO, or RTO commitment; drill timings are observations only.
- No multi-region tables, active-active service, cross-region failover, or disaster-recovery control plane.
- No separate metrics database, S3 archive, data lake, or indefinite raw-run retention.
- No table per tenant/entity or infrastructure resources per monitor.
- No outbound monitored-response body limit owned by `harden-outbound-http-monitoring-boundaries`.
- No automatic shutdown when a budget threshold is reached.

## Decisions

### Keep monitoring, authorization, and credential authority separate

The inventory is table-specific. `AppTable` is authoritative for workspace/service/monitor and scheduler configuration, notification channels, escalation policies, incidents, incident activity, unresolved-incident escalation state, and monitoring audit history. These durable families receive no operational TTL. `AuthTable` is authoritative for immutable Cognito-subject links, `DEFAULT` membership status and role, session-valid-after boundaries, active-admin guards, durable lifecycle desired state, and identity lifecycle audit evidence. Cognito groups and mutable claims never become authorization authority.

Cognito remains authoritative for provider identity existence, credentials, password/MFA state, challenges, and token issuance. Neither DynamoDB table copies those credentials as recovery data. Encrypted dashboard session/token bundles and auth transactions remain short-lived `AuthTable` state governed by `ExpiresAt` and TTL; restoring them does not make them durable authority or override expiry. Cross-system recovery validates immutable subject links and desired-state convergence without pretending Cognito and DynamoDB form one transaction.

`MonitorStatus`, `ServiceStatus`, `AlertState`, search index entries, scheduler/pipeline-health projections, and aggregate rollups are reconstructable from canonical configuration and retained events/results. They remain compact overwrite-in-place or source-lifetime records rather than receiving arbitrary age TTLs. `CheckRun` remains 30-day raw evidence and `ExecutionWork` remains 7-day transient coordination state. Auth terminal operations/private invite intents retain their separately specified 30-day post-completion TTL. DynamoDB TTL is asynchronous, so APIs must not promise deletion at the exact second.

Alternative: retain every record indefinitely. Rejected because raw runs and work dominate storage without becoming a useful backup. Alternative: TTL incidents/audits now. Rejected because no approved compliance/product retention period exists; silent deletion is less safe than explicit durable retention.

### Inherit basic stage lifecycle protection

`standardize-stage-resource-lifecycle` is the prerequisite source of truth for persistent/ephemeral classification, PITR, deletion protection, retain-on-delete behavior, baseline ownership tags, and retained inventory for both tables. This change does not redefine or independently implement those defaults, nor does it introduce a recovery-specific smoke stage. Its runbook verifies and reapplies the prerequisite's current baseline after restore, while its evidence records the stage class used by each drill.

Both tables continue to use on-demand capacity because workloads are bursty and current measurements do not justify provisioned capacity management. The cost worksheet measures operations, storage, indexes, and PITR separately for `AppTable` and `AuthTable`. A controlled retirement procedure references the lifecycle prerequisite rather than duplicating its stage policy.

Alternative: restate lifecycle defaults here. Rejected because two active changes would own the same persistent/ephemeral policy. Alternative: add normal-operation backup tables or scheduled exports. Rejected because the prerequisite PITR baseline supplies same-region recovery with less operating surface.

### Restore each DynamoDB authority to a new table and coordinate cutover

The recovery runbook uses restore-to-point-in-time with immutable source tables and unique target names. It supports an `AppTable`-only incident, an `AuthTable`-only incident, or a coordinated incident requiring both. For each restored table it records source/target ARNs, recovery point, region, stage, operator, commands, and timings; verifies ACTIVE table/index state; and reapplies the prerequisite lifecycle settings, TTL configuration, tags, alarms, and least-privilege access that restore does not preserve automatically.

`AppTable` validation checks monitoring/configuration/history authority and its consumers: monitor API domain repositories, scheduler, check worker, escalation/notification workers, and health evaluator. `AuthTable` validation checks membership/RBAC/session-boundary authority and its consumers: monitor API principal/membership resolution, dashboard auth/session adapters, user-lifecycle repair, bootstrap/break-glass tooling, and administration workflows. Cognito validation uses provider APIs to verify referenced immutable subjects and required provider state without exporting credentials or treating provider state as table-owned.

Before `AppTable` cutover, recurring scheduling is paused and in-flight queues are accounted for. Before `AuthTable` cutover, operators preserve fail-closed authorization, verify at least one active administrator and the last-admin guard, and reconcile due lifecycle operations with Cognito. A coordinated incident uses explicit recovery points and a consistency decision log; it does not require identical timestamps or falsely claim cross-service atomicity. Each consumer set switches only to its corresponding validated target. Rollback points that set back to its unchanged source table and explicitly reconciles post-cutover writes.

Alternative: restore over the current table. DynamoDB does not support in-place PITR restore, and preserving the source is safer. Alternative: dual-write during recovery. Rejected because it creates conflict semantics beyond this single-region recovery scope.

### Validate durable integrity with a recovery-point manifest

Normal runtime paths remain query-only and scan-free. Recovery tooling runs against isolated tables and may use rate-limited paginated scans because it must inspect heterogeneous item families. Before drills, a manifest captures the selected fixture/recovery point, durable counts by entity type and tenant, known sentinel IDs, schema/tool version, and non-secret hashes of canonicalized durable records. For incident recovery where a pre-event manifest is unavailable, the validator uses PITR metadata, known sentinels, referential checks, and documented count tolerances rather than comparing with the later live table as if it represented the same point in time.

`AppTable` checks include key shape, required durable families, counts/hashes where available, tenant ownership, service-monitor and policy-channel references, incident/activity/audit linkage, retention classes, and representative monitoring reads. `AuthTable` checks include key/index shape, immutable subject and membership bindings, fixed roles/statuses, active-admin guard consistency, session-valid-after boundaries, lifecycle operation/audit references, due-work index readiness, and expiry rules for sessions, transactions, terminal operations, and invite intents. Cross-domain checks verify that auth-derived `DEFAULT` scope can read representative `AppTable` resources and that Cognito subjects referenced by memberships exist or are explicitly covered by a recovery exception. No check merges records or derives role from Cognito groups.

Alternative: compare only DynamoDB `ItemCount`. Rejected because it is approximate and cannot detect broken references or cross-tenant records. Alternative: require full cryptographic equality with the current source. Rejected because the source may have advanced beyond the restore point and TTL is asynchronous.

### Share one bounded AppTable scheduling projection with pipeline health

Monitor mutations will incrementally maintain compact scheduling projection items under a tenant-scoped, shardable key prefix in `AppTable`. The key shape and fields are designed with `expose-monitoring-pipeline-health` so scheduler due enumeration and bounded overdue/cadence evidence reuse one canonical projection/access pattern. The projection contains only scheduling-safe fields needed to identify canonical monitor configuration and authoritative due-time evidence; it contains no headers, expected bodies, channel secrets, or auth material.

The scheduler queries projection shards with explicit `Limit` and `ExclusiveStartKey`, enforces per-invocation page/item/enqueue/time budgets, and stores a compact checkpoint. Completion clears/rotates the checkpoint so new keys before a prior cursor cannot starve. Mutation transactions maintain canonical records and projections where transaction limits permit; a bounded reconciliation tool reports and repairs drift after partial legacy operations or recovery.

The existing table and indexes are preferred. The implementation first proves whether primary-key projection queries or an existing sparse index satisfy both scheduler and pipeline-health bounds. It SHALL NOT add separate scheduler-due and health-due indexes. A single new sparse due-time index is permitted only if before/after evidence shows the shared access path cannot meet the expected-validation or stress exercise criteria. `AuthTable`'s lifecycle-operation due-work GSI remains separate because it indexes security workflow repair, not monitor scheduling.

### Paginate APIs and replace repeated reads only where growth justifies it

All growing list endpoints will accept validated opaque cursors and enforce named default/maximum page sizes. Existing monitor history remains 20 records per page. Service, monitor, channel, policy, incident, and activity lists receive endpoint-appropriate bounded pages; no endpoint reads all pages to calculate totals. Internal delete/repair operations use bounded continuation and transaction chunks.

Service-card and rollup paths currently perform status/run reads per monitor. Implementation first measures these paths at the envelope. Where needed, it uses existing `ServiceStatus` incremental state, batch gets for known status keys, or a compact service metrics rollup updated with execution results. A rollup is justified only if it removes the measured blocker and includes rebuild/reconciliation behavior.

Alternative: precompute every dashboard view. Rejected as unnecessary write amplification. Alternative: keep unbounded list APIs because the initial tenant is small. Rejected because the high-volume stress profile is already large enough to cross DynamoDB and Lambda response boundaries.

### Separate owner, validation, and stress profiles

Three named profiles prevent stress evidence from becoming a default-cost claim:

- **Default low-cost owner profile:** one active tenant, up to 10 monitors at five-minute cadence, modest operator traffic, and 30-day raw-run retention. This is the owner-operated baseline used for low-cost estimates; it is not a guarantee of AWS Free Tier eligibility because account history, regions, logs, PITR, Cognito, and other services affect charges.
- **Expected validation profile:** one active tenant, up to 100 services and 100 monitors at one-minute cadence, representative histories, and up to 10 concurrent operator requests. This is the routine non-production validation target.
- **High-volume stress profile:** one active tenant, up to 100 services and 1,000 enabled monitors at one-minute cadence, representative 30-day history, and up to 25 concurrent operator requests. This is a deliberate stress exercise, explicitly not the default owner profile or a free-tier/cost promise.

The measured profile dimensions are warning and operational support boundaries. Crossing one emits an actionable signal and requires new evidence before support is claimed; it does not automatically reject otherwise safe configuration. Hard rejection is reserved for a separately named safety constraint such as a transaction limit, maximum payload/page size, minimum safe cadence, or resource invariant, and documentation must state that reason. The system does not claim unlimited scaling beyond measured profiles.

Exercise acceptance requires no sustained DynamoDB throttles, no Lambda timeout, bounded API response bytes/pages, correct continuation without omissions, scheduler checkpoint completion without starvation, and queue depth/age that returns toward baseline. The report records consumed capacity, request counts, p50/p95/p99 duration, errors, queue metrics, item sizes, and response bytes. These are repeatable evidence, not universal latency promises or service-level commitments.

Alternative: claim DynamoDB/Lambda automatic scaling is sufficient. Rejected because application query shape, Lambda duration, transaction limits, and queue production can fail before managed service capacity does.

### Model cost from measured profiles and recommend optional budgets

The cost worksheet will use current public pricing for the deployed region and record its pricing date. It calculates `AppTable` and `AuthTable` on-demand reads/writes, table/index and PITR storage, Cognito usage, SSM operations, TTL effects, Lambda requests/GB-seconds, SQS requests, API Gateway requests, logs/alarms, dashboard costs where attributable, and one non-production restore drill. The three named profiles share load-fixture assumptions so estimates are reproducible and identify the dominant driver. No normal-operation recovery table or always-on compute is added.

An AWS Budget is recommended but optional because budgets are account-level resources and a clean account may not have an approved notification endpoint or permission to create one. Documentation provides setup and verification for a stage-attributed configurable monthly budget, a forecast alert at 80 percent, and an actual alert at 100 percent. When enabled, recipients come from deployment configuration rather than source-controlled personal addresses. Missing budget configuration does not block default deployment, and budgets never stop monitors automatically.

Alternative: hard-code one dollar amount. Rejected because account credits, region, stage purpose, and ownership differ. Alternative: automatic disable at threshold. Rejected because it could create an unplanned monitoring outage.

## Risks / Trade-offs

- [Risk] PITR remains a same-region recovery mechanism and cannot cover a regional outage. → Mitigation: state the single-region boundary explicitly; multi-region recovery requires a separate approved change.
- [Risk] Restoring only one table can leave cross-domain references inconsistent with Cognito or the other table. → Mitigation: validate each authority independently, run explicit cross-domain subject/tenant checks, and record the coordinated recovery decision without merging data ownership.
- [Risk] Deletion protection plus retain-on-delete can block intended stack teardown and leave billable storage. → Mitigation: tag retained tables, document retirement, and alert on stage cost rather than silently weakening protection.
- [Risk] Restored-table settings or IAM links can differ from the source. → Mitigation: make post-restore settings and every consumer link explicit validation gates before cutover.
- [Risk] Scheduler projections can drift from canonical monitor configuration. → Mitigation: transactional maintenance, idempotent records, bounded reconciliation, and recovery validation.
- [Risk] Cursor traversal over a mutating projection can duplicate or defer work. → Mitigation: retry-safe execution identity, checkpoint rotation after complete passes, and starvation tests under concurrent mutation.
- [Risk] Adding pagination changes dashboard/API consumers. → Mitigation: preserve existing first-page shapes where possible, update server adapters and Bruno/OpenAPI together, and add multi-page contract tests.
- [Risk] Indefinite incidents/audits increase storage. → Mitigation: expose their measured growth in cost scenarios; do not invent a deletion period without product/compliance approval.
- [Risk] Offline validation scans consume capacity. → Mitigation: run against isolated non-production/restore tables with page/rate limits and report consumed capacity.

## Migration Plan

1. Confirm `standardize-stage-resource-lifecycle` supplies the stage-aware lifecycle baseline, then add the two-table data inventory, retention matrix, profile definitions, cost worksheet, runbook, and non-contractual drill evidence templates.
2. Add shared scheduler/pipeline-health projection records and mutation maintenance in `AppTable`, backfill/reconcile them with bounded tooling, and verify canonical/projection parity before switching scheduler reads.
3. Switch scheduler enumeration to bounded pages and checkpoints with retry, mutation, starvation, and high-volume stress tests.
4. Add bounded cursor contracts to remaining growing API and maintenance paths; replace measured N+1 rollup reads with batch or incremental state only where required.
5. Deploy to a non-production persistent stage, run the three named profiles, and record capacity and cost evidence.
6. Run restore-to-new-table drills for both domains, validate, cut over in non-production, roll back to each source, retain evidence, and explicitly clean up drill tables.
7. Optionally configure account/stage budget thresholds and recipients and verify alert configuration without inducing real spend; otherwise record why budget setup is unavailable and retain cost documentation.
8. Roll bounded paths and recovery procedures into the production stage only after non-production evidence passes and lifecycle prerequisites are active.

Rollback keeps both original tables and the scheduler path available until the corresponding cutover validation. API pagination additions are rolled back with their consumers. Projection writes may remain because they are reconstructable and ignored by old readers. Lifecycle protections supplied by the prerequisite remain enabled during application rollback.

## Open Questions

- Final scheduler shard count, page size, and safe remaining-time threshold will be selected from non-production measurements, bounded by the spec rather than guessed here.
- Final shared due-time key/index selection must be coordinated with pipeline-health implementation and justified by measurements; duplicate scheduler and health indexes are prohibited.
