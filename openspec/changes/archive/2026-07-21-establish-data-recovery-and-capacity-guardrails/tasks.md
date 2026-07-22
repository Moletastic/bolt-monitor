## 1. Data Inventory and Operational Contracts

- [ ] 1.1 Add a canonical two-table item-family inventory covering `AppTable` monitoring/configuration/history records and `AuthTable` identity/session/membership/RBAC records, with durable/reconstructable/transient classification, authority, TTL, deletion trigger, recovery check, and sensitive-field handling; record Cognito credentials and token issuance as provider-managed rather than table recovery data.
- [ ] 1.2 Consolidate retention constants and documentation so raw `CheckRun` records use 30 days, `ExecutionWork` uses 7 days, auth sessions/transactions and lifecycle operations retain their auth-spec expiry, reconstructable records follow source lifetime/rebuild rules, and durable records in either table have no operational TTL.
- [ ] 1.3 Add tests that assert numeric TTL calculation for expiring families and absence of operational TTL on representative durable families.
- [ ] 1.4 Document the default low-cost owner, expected validation, and high-volume stress profiles; label 1,000 monitors per minute as stress rather than default/free-tier usage, and distinguish warning/support boundaries from separately justified hard safety limits.

## 2. Lifecycle Prerequisite and Table Contracts

- [ ] 2.1 Verify and document that `standardize-stage-resource-lifecycle` is the prerequisite owner of persistent/ephemeral classification, PITR, deletion protection, retain-on-delete behavior, and baseline ownership tags for `AppTable` and `AuthTable`; do not duplicate those controls in this change.
- [ ] 2.2 Verify both existing tables use on-demand capacity and native TTL configuration, and add assertions only for the detailed retention/capacity contracts owned here without restating stage lifecycle policy.
- [ ] 2.3 Preview restore adoption/cutover changes for both tables and document replacement, IAM, index, and deletion risks before a drill or deployment.

## 3. Scheduler Projection and Bounded Traversal

- [ ] 3.1 Define compact tenant-scoped scheduler projection and checkpoint records in existing `AppTable`, excluding secrets and documenting key/shard access patterns, authoritative due-time evidence, reconstructability, and reuse by pipeline health.
- [ ] 3.2 Maintain scheduler projections during monitor create, update, enable, disable, move, and delete operations with idempotent transactional writes where transaction limits permit.
- [ ] 3.3 Implement a bounded projection reconciliation/backfill command with dry-run output, explicit page/rate limits, continuation, consumed-capacity reporting, and repair mode.
- [ ] 3.4 Replace scheduler service-to-monitor N+1 enumeration with limited projection queries using opaque continuation and named page, item, enqueue, and safe remaining-time budgets.
- [ ] 3.5 Persist and rotate scheduler checkpoints so retries and concurrent projection mutation cannot permanently starve enabled monitors.
- [ ] 3.6 Emit secret-safe scheduler measurements for evaluated/due monitors, pages, enqueue count, continuation, duration, remaining time, consumed capacity when available, and envelope violations.
- [ ] 3.7 Add scheduler tests for multi-page completion, budget exhaustion/resume, retry after partial failure, projection mutation, checkpoint wraparound, idempotent enqueueing, drift repair, and stress above 1,000 monitors while continuing bounded processing and emitting a support-boundary warning.
- [ ] 3.8 Prove the scheduler and pipeline-health evaluator reuse one AppTable due-time projection/access pattern; prefer existing keys/indexes, prohibit duplicate due-time indexes, and add at most one shared sparse index only with before/after evidence. Keep the AuthTable lifecycle due-work GSI separate.

## 4. Bounded API and Internal Access Paths

- [ ] 4.1 Inventory every runtime DynamoDB query and scan, record its key/index, cardinality driver, explicit limit, cursor behavior, and N+1/full-history risk, and remove runtime scans or document measured exceptions permitted by the specs.
- [ ] 4.2 Add validated opaque cursor paging with named default and maximum sizes to service, monitor, notification-channel, escalation-policy, global-incident, and incident-activity collection APIs without calculating collection totals.
- [ ] 4.3 Update dashboard server adapters and views to consume bounded first pages and continuation for every changed collection without introducing eager all-page loading.
- [ ] 4.4 Update OpenAPI and Bruno requests for changed cursor contracts and add first-page, next-page, invalid-cursor, resource-mismatch, and maximum-limit coverage.
- [ ] 4.5 Bound internal delete, backfill, reconciliation, and filtered-query loops by page, item, transaction, and time budgets, preserving continuation instead of treating DynamoDB's 1 MB boundary as completion.
- [ ] 4.6 Measure service-card and service-rollup reads at the supported envelope and replace blocking per-monitor reads with batch gets or an incrementally maintained/rebuildable rollup only where before/after evidence justifies it.
- [ ] 4.7 Add repository and handler tests proving page bounds, stable continuation without duplicate/omitted records, no total-count traversal, filtered-page budget behavior, and no runtime table scans.

## 5. Recovery Tooling and Runbook

- [ ] 5.1 Add a versioned restore-to-new-table runbook for `AppTable`-only, `AuthTable`-only, and coordinated incidents, covering prerequisites, independent source/recovery-point selection, command capture, restore, lifecycle settings, authority-specific consumers, cutover approval, smoke checks, scheduler/queue and fail-closed auth handling, rollback, reconciliation, evidence, and cleanup.
- [ ] 5.2 Implement authority-specific recovery-point manifests that record schema/tool version, source metadata, durable counts by tenant/entity, sentinel IDs, and non-secret canonical hashes using bounded capacity-aware traversal.
- [ ] 5.3 Implement restored-table validators for AppTable monitoring/configuration/history integrity and AuthTable identity-link/membership/RBAC/guard/lifecycle integrity, plus cross-domain tenant and Cognito-subject checks that do not copy credentials or merge authority.
- [ ] 5.4 Ensure recovery tools redact secrets and sensitive configuration, emit machine-readable pass/fail evidence, report consumed capacity, and return non-zero when required integrity checks fail.
- [ ] 5.5 Add fixture-based validator tests for missing durable records, broken AppTable references, invalid AuthTable membership/guard/lifecycle state, missing or mismatched Cognito subjects, cross-tenant records, invalid key shapes, invalid/missing TTL, unexpected durable TTL, projection-only differences, and unavailable pre-event manifests.
- [ ] 5.6 Verify the runbook switches AppTable consumers and AuthTable consumers only to their corresponding validated target, validates Cognito separately, and preserves each unchanged source table for rollback.

## 6. Load, Recovery Drill, and Cost Evidence

- [ ] 6.1 Add deterministic default low-cost owner, expected validation, and high-volume stress fixtures with skewed service sizes, monitor cadence, item-size bands, retained run history, incidents, activities, audits, policies, channels, and representative AuthTable strong reads/lifecycle work.
- [ ] 6.2 Add a repeatable load runner that captures request counts, item sizes, consumed read/write capacity, throttles, queue depth/age, Lambda duration/errors/timeouts, page counts, response bytes, and p50/p95/p99 observations.
- [ ] 6.3 Run all three profiles in non-production and record whether scheduler traversal/checkpointing, API bounds, auth strong reads, queue recovery, no-sustained-throttle, and no-timeout exercise criteria pass; label the 1,000-per-minute case as stress evidence, not a default/free-tier claim.
- [ ] 6.4 Tune named scheduler/API budgets from measured evidence and repeat the high-volume stress test after any index, batching, projection, or rollup optimization, recording before/after evidence and avoiding duplicate scheduler/health indexes.
- [ ] 6.5 Execute a non-production restore-to-new-table drill for both recovery domains, validate each authority and cross-domain references, perform controlled consumer cutovers, representative read/write/auth checks, rollback, and explicit temporary-table cleanup.
- [ ] 6.6 Record drill dataset shape, region, tool version, restore/validation/cutover/rollback durations, results, and remediation as non-contractual observations without SLA, RPO, or RTO claims.
- [ ] 6.7 Build a reproducible pricing-date/region cost worksheet for the three named profiles covering both DynamoDB tables and indexes, Cognito, SSM, PITR, Lambda, SQS, API Gateway, logs/alarms, dashboard attribution, and one recovery drill; identify the dominant cost driver and disclaim Free Tier eligibility.

## 7. Optional Budget Guardrail and Final Verification

- [ ] 7.1 Document recommended optional account-level AWS Budget setup with stage attribution, forecast notification at 80 percent, actual notification at 100 percent, no automatic monitoring shutdown, and no personal address committed to source; default deployment must succeed without budget permission or a notification endpoint.
- [ ] 7.2 If optional budget provisioning is implemented, add infrastructure assertions for opt-in behavior, thresholds, alert-only behavior, and stage attribution; verify absent configuration creates no budget and does not fail deployment.
- [ ] 7.3 Update repository operations documentation with retained-table retirement, budget response, envelope breach handling, recovery cadence, evidence location, and ownership.
- [ ] 7.4 Run `make test-go-all`, `make lint-go`, `make build-go`, `make check-infra`, `make format-infra`, `make lint-dashboard`, `make check-dashboard`, `make test-dashboard`, and `make check-bruno`.
- [ ] 7.5 Confirm no multi-region resources, active-active flow, separate metrics database, new normal-operation table, merged AppTable/AuthTable authority, duplicate due-time index, per-monitor resource, or duplicate outbound HTTP response-body guardrail was introduced.
