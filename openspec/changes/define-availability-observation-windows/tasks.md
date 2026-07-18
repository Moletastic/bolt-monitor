## 0. Prerequisite Landing Gates

- [ ] 0.1 Verify `make-check-execution-retry-safe` has landed canonical recurring `runId`/slot identity, terminal work, authoritative retry/lease deadlines, and bounded recovery semantics consumed by this change.
- [ ] 0.2 Verify `assure-notification-and-escalation-delivery` and `expose-monitoring-pipeline-health` have landed durable delivery outcomes and bounded scheduler/execution/notification evidence; document that notification state never changes slot classification.
- [ ] 0.3 Verify `establish-data-recovery-and-capacity-guardrails` and `standardize-stage-resource-lifecycle` have landed bounded cursor workflows, recovery inventory/validation, supported-envelope limits, and persistent/ephemeral stage protections.

## 1. Domain Contracts and Validation

- [ ] 1.1 Add shared objective types, rolling-duration enum, integer `targetBasisPoints` validation, latency validation, immutable effective-dated versions, evaluation states, burn states, and typed validation errors; reject floating-point targets.
- [ ] 1.2 Extend canonical monitor configuration and monitor API request/response mapping with an optional objective reference while preserving monitors without objectives.
- [ ] 1.3 Define immutable effective-dated schedule versions with interval, UTC alignment, half-open bounds, and non-overlap rules so recurring slots are enumerable without scheduler work; retain the retry-safe slot identity through work, queue, result, and persistence mappings.
- [ ] 1.4 Define immutable enabled/disabled and maintenance interval prerequisites plus authorized-correction facts that retain actor/source, reason, recorded time, half-open effective bounds, version, authorization, and audit identity.

## 2. Storage and Canonicalization

- [ ] 2.1 Add DynamoDB record types and tenant/service/monitor key patterns for schedule/objective versions, enabled/disabled and maintenance intervals, expected slots, canonical slot results, corrections, versioned finalizer cursors, and versioned hourly/daily aggregates.
- [ ] 2.2 Implement conditional canonical-result persistence so duplicate retries and redeliveries converge on one accepted result per scheduled slot.
- [ ] 2.3 Add deterministic expected-slot writes, idempotent aggregate application markers, monotonic slot/bucket versions, and `OPEN`/`CLOSED` metadata so each classification revision changes counts exactly once.
- [ ] 2.4 Configure 400-day TTL metadata for hourly and daily aggregates while leaving raw `CheckRun` retention unchanged.
- [ ] 2.5 Classify every new item family in the recovery inventory, preserve durable definitions/corrections without operational TTL, define cursor/aggregate rebuild and restore validation, and apply persistent-stage lifecycle protections.
- [ ] 2.6 Add repository tests for tenant isolation, key round trips, immutable-version overlap/backdate rejection, conditional duplicate handling, cursor replay, bucket versions/closure, corrections, TTL values, and bounded queries.

## 3. Slot Accounting and Aggregation

- [ ] 3.1 Derive expected slots independently from immutable schedule, objective, enabled/disabled, and maintenance facts; make scheduler work reference and satisfy the same identity without authoring the expectation.
- [ ] 3.2 Implement finalization-grace cutoff and exclusive `good`, `bad`, `missing`, and `excluded` classification precedence for both objective types.
- [ ] 3.3 Ensure manual checks are excluded by default and scheduler, queue, provider, or result gaps mature as `missing` with pipeline-health correlation rather than implicit exclusion.
- [ ] 3.4 Add an independent EventBridge finalizer/compactor event source and runtime mode that enumerates mature expectations without `RUN_REQUEST#` records and reads only bounded authoritative result/pipeline evidence.
- [ ] 3.5 Implement versioned tenant/shard cursors with finalized watermark, schedule position, opaque continuation, slot/item/page/time budgets, commit-before-advance semantics, idempotent crash replay, stale-cursor evidence, and bounded catch-up.
- [ ] 3.6 Implement hourly aggregate updates, a 24-hour post-maturity automatic late-result horizon, hourly closure, all-hours-closed daily compaction, and append-only authorized closed-history corrections that create new retained versions.
- [ ] 3.7 Implement rolling report math with integer counts and exact ratio numerators/denominators; calculate `allowedBad` by integer floor and compare compliance by cross multiplication without floating point.
- [ ] 3.8 Implement exact burn boundaries: incomplete/no-eligible is `UNAVAILABLE`; zero allowance is healthy only at zero consumption; equality with positive allowance is `BURNING`; only consumption above allowance is `EXHAUSTED`.
- [ ] 3.9 Enforce bounded rolling reads of at most 30 daily and 48 hourly buckets per monitor/objective version without scanning raw runs or scheduler work.

## 4. Status, Maintenance, and Incidents

- [ ] 4.1 Append immutable enabled/disabled intervals from monitor transitions and consume the landed immutable maintenance interval contract with half-open boundary and effective-date/version behavior; prohibit retroactive silent reclassification.
- [ ] 4.2 Transition maintenance exit to `UNKNOWN` with awaiting-recurring-observation reason and clear it only from a new recurring canonical result.
- [ ] 4.3 Prevent maintenance, disablement, exclusion, and manual success from resolving incidents; retain existing recovery thresholds for qualifying recurring results.
- [ ] 4.4 Add state-machine and incident tests for maintenance entry/exit, open incidents, manual checks, stale pre-maintenance status, and first post-maintenance recurring outcomes.

## 5. Monitor and Service Read Surfaces

- [ ] 5.1 Add nested monitor availability read handling with allowed-window validation and response-envelope fields for definition/bucket versions, integer target basis points, counts, exact ratio components, completeness, finalizer coverage, evaluation, budget, burn, and exclusions.
- [ ] 5.2 Add service availability aggregation that sums opportunity and budget counts, recomputes ratios, applies incomplete-first then breach semantics, and discloses monitors without objectives.
- [ ] 5.3 Keep objective response types and names distinct from latest status, raw run history, and `service-card-recent-metrics`; update OpenAPI and Bruno coverage for new routes.
- [ ] 5.4 Add API tests for no objective, no eligible slots, complete/compliant, exact target equality, complete/breached, incomplete/finalizer-lag, mixed service monitors, mixed objective versions, and bounded queries.

## 6. Dashboard Objective Experience

- [ ] 6.1 Add optional monitor objective form controls for integer basis-point targets, the two types, and three rolling durations with latency-threshold fields only when required.
- [ ] 6.2 Add monitor availability reporting that displays raw counts, exact-source formatted ratios, completeness/finalizer lag, evaluation, budget, burn, and auditable exclusion summaries using objective-not-SLA copy.
- [ ] 6.3 Add service availability reporting that preserves per-monitor objectives and explicitly identifies monitors without objectives without reusing recent-uptime/P99 presentation.
- [ ] 6.4 Add dashboard tests for objective configuration, awaiting-observation status, incomplete data, exclusions, no-objective state, service aggregation, and separation from recent metrics.

## 7. Golden Verification and Cost Guardrails

- [ ] 7.1 Add reviewed golden timeline fixtures covering cadence/objective/interval versions, successes, failures, latency breaches, duplicates, within-horizon and post-horizon results, missing pipeline stages, maintenance, disablement, manual checks, bucket closure, and authorized corrections.
- [ ] 7.2 Add the scheduler-outage golden fixture where three enabled non-maintenance intervals produce no work records and independent finalization yields exactly three mature `missing` opportunities.
- [ ] 7.3 Add golden integer/rational arithmetic fixtures for monitor and service counts, ratios, completeness, exact target equality, integer-floor budgets, zero allowance, consumed-equals-allowed, `EXHAUSTED`, burn thresholds, and no-data cases.
- [ ] 7.4 Verify reports remain correct after simulated raw-run/work TTL expiry, cursor crash/replay, 24-hour closure, hourly-to-daily compaction, and retained-version correction.
- [ ] 7.5 Document and test the bound of one slot fact plus one idempotent hourly application per expected opportunity, fixed-size cursor checkpoints, bounded backlog recovery, asynchronous daily compaction, 400-day aggregate retention, and bounded rolling reads.
- [ ] 7.6 Load-test finalizer catch-up at the supported installation envelope and verify slot/item/page/time budgets, no starvation, no scans, no duplicate opportunities, measurable lag recovery, and no Lambda timeout or DynamoDB throttle.
- [ ] 7.7 Run Go, dashboard, infra, OpenAPI/Bruno, strict OpenSpec, and production-build verification targets required by the touched components.
