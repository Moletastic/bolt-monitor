## Context

The system has three different operational views with different truth models: mutable current status, raw `CheckRun` history retained by DynamoDB TTL, and bounded recent-sample service-card metrics. None establishes how many recurring checks should have occurred, so none can distinguish a target failure from a missing scheduler, queue, provider, or result-persistence observation. This change has hard landing dependencies on `make-check-execution-retry-safe`, `assure-notification-and-escalation-delivery`, `expose-monitoring-pipeline-health`, `establish-data-recovery-and-capacity-guardrails`, and `standardize-stage-resource-lifecycle`. It consumes their canonical execution identities, authoritative retry/deadline and delivery evidence, bounded access/recovery contracts, and persistent-stage protections rather than redefining them. Notification delivery does not affect slot classification, but its assurance contract must be present before the product claims trustworthy end-to-end objective operations.

Availability objectives are optional operator targets, not contractual SLAs. They must remain useful after raw runs expire, must not rewrite history when configuration changes, and must avoid the cost and complexity of a general metrics platform.

## Goals / Non-Goals

**Goals:**

- Account deterministically for each mature recurring scheduled slot as good, bad, missing, or excluded.
- Derive expected slots without depending on scheduler execution or work records, and finalize them through an independent bounded process.
- Support optional `recurring-check-success` and `success-within-latency-threshold` objectives over one duration-based rolling-window family.
- Preserve objective and exclusion configuration history so every report is explainable.
- Make missing product-pipeline evidence visible through completeness and `INCOMPLETE` evaluation state.
- Keep rolling reads and aggregate storage bounded while surviving raw-run TTL.
- Aggregate service reports without averaging percentages or hiding monitors without objectives.
- Keep objective reports semantically separate from current status, incidents, and recent-sample metrics.

**Non-Goals:**

- Contractual SLA calculation, credits, compliance attestations, or legal reporting.
- PromQL, arbitrary metric ingestion, arbitrary query languages, or a general time-series database.
- A generalized alert-rule engine, status pages, multi-region analysis, or Grafana replacement.
- Objective types beyond recurring success and successful latency threshold.
- Calendar-month, request-volume, percentile-latency, or multiple simultaneous window families.

## Decisions

### Decision 1: Immutable schedule and interval facts create expectations

Each recurring opportunity is identified by tenant, service, monitor, immutable schedule-definition version, and UTC `scheduledFor`. A schedule version stores an effective half-open range, interval seconds, and UTC alignment/phase sufficient to enumerate every slot deterministically. Enabled and disabled state is represented by immutable effective-dated intervals, and maintenance uses the prerequisite immutable interval contract. Versions and intervals cannot overlap contradictorily, be mutated in place, or be backdated across a matured slot except through the authorized correction policy.

Expected slots are derived independently by enumerating immutable schedule definitions and joining the objective, enabled/disabled, and maintenance facts effective at `scheduledFor`. Enabled schedule slots are eligible expectations; slots covered by a disabled or maintenance interval remain expected but are classified `excluded` so the report discloses intentional non-observation. The scheduler's `RUN_REQUEST#`, queue envelope, and result may satisfy and explain an expectation, but their existence is never the authority that creates one. The retry-safe scheduler continues to derive the same slot identity before work side effects. Manual checks have no recurring slot identity and are excluded by default.

Alternative considered: infer expected checks from work records or divide elapsed time by the monitor's current interval. Both fail during scheduler outage and cannot handle cadence changes, disablement, maintenance, delayed execution, or retry duplicates without rewriting history.

### Decision 2: Mature slots have four mutually exclusive classifications

A slot becomes mature after a fixed finalization grace derived from the execution timeout and bounded delivery/retry allowance. Reports end at the maturity cutoff rather than labeling in-flight work missing.

Classification precedence is:

1. `excluded` when the slot was covered by an explicit maintenance interval or the monitor was disabled at `scheduledFor`.
2. `good` when the canonical recurring result succeeds and, for a latency objective, `durationMs` is at or below the configured threshold.
3. `bad` when the canonical result fails or a successful result exceeds the configured latency threshold.
4. `missing` when a non-excluded mature slot has no canonical result.

Maintenance and disabled intervals are explicit immutable audit facts with actor/source, reason, start, end, and effective timestamps. Retroactive interval creation cannot reclassify mature slots. Missing scheduler materialization, queue delivery, provider completion, or result persistence remains `missing`; pipeline-health evidence can explain the gap but cannot exclude it.

Alternative considered: omit unknown slots from the denominator. That produces optimistic results exactly when the monitoring system is unhealthy.

### Decision 3: An independent bounded finalizer materializes mature truth

A dedicated EventBridge schedule invokes a finalizer/compactor runtime mode independently of the recurring scheduler rule and scheduler success. For each tenant projection shard, it enumerates immutable schedule versions over the interval from a durable finalized watermark through `now - finalizationGrace`, joins effective objective/enabled/maintenance facts, and conditionally creates deterministic slot facts and classification revisions. It reads retry-safe canonical results and bounded pipeline-health evidence when available but never requires a `RUN_REQUEST#` to discover a slot.

Each invocation enforces named maximum slot, item, page, and safe-remaining-time budgets from the recovery/capacity contract. A versioned tenant/shard cursor contains the finalized watermark, schedule-definition position, and opaque continuation. The cursor advances only after all preceding derived slots and idempotent aggregate applications commit. Failure before cursor advancement replays deterministic slot keys and application markers; failure after advancement cannot omit committed slots. Cursor lag is emitted as pipeline evidence, and bounded recovery resumes from the cursor until caught up without an unbounded scan or a second opportunity count.

The steady-state write bound is one expected-slot fact and one idempotent hourly aggregate application per mature expected opportunity, plus fixed-size cursor checkpoints and asynchronous compaction. Scheduler outage does not lower this cost by hiding opportunities: three mature cadence slots with no work records produce exactly three `missing` facts. Backlog recovery remains bounded per invocation and scales with expected slots, not raw runs, retries, or API reads.

Alternative considered: finalize only from scheduler work/result events. Rejected because the scheduler cannot authoritatively report its own absence.

### Decision 4: Incomplete data blocks a compliance verdict and math is exact

For a report:

- `eligible = good + bad + missing`
- `observed = good + bad`
- `dataCompleteness = observed / eligible`, returned as exact numerator/denominator, or unavailable when `eligible = 0`
- `achievedRatio = good / observed`, returned as exact numerator/denominator, or unavailable when `observed = 0`
- `allowedBad = floor(eligible * (10000 - targetBasisPoints) / 10000)`
- `consumedBad = bad`
- `remainingBad = allowedBad - consumedBad`

Targets use integer basis points in `[1, 10000]`; floating-point objective input or storage is not accepted. Evaluation is `INCOMPLETE` whenever `missing > 0`, regardless of the observed ratio. Otherwise compliance compares `good * 10000 >= observed * targetBasisPoints` using overflow-safe integer/rational arithmetic; equality is `COMPLIANT`. It is `BREACHED` below that exact boundary and `INCOMPLETE` when no eligible or observed opportunities exist. Counts and exact ratio components are returned together so rounding cannot change a verdict.

Basic burn state uses integer budget opportunity counts, not extrapolated time-series math: `UNAVAILABLE` for incomplete/no-eligible reports; for positive allowance, `HEALTHY` when `2 * consumedBad <= allowedBad`, `BURNING` when more than half is consumed through `consumedBad == allowedBad`, and `EXHAUSTED` only when `consumedBad > allowedBad`. A zero allowance is `HEALTHY` at zero bad and `EXHAUSTED` after any bad slot. Thus consuming exactly the allowance is compliant at the objective boundary and `BURNING`, not `EXHAUSTED`.

Alternative considered: count missing as bad. That would blame the monitored service for monitoring-pipeline failure. The selected model reports uncertainty explicitly and refuses a compliance verdict.

### Decision 5: Objectives are optional, versioned monitor configuration

A monitor may have no objective. An objective has one type, integer `targetBasisPoints`, one allowed rolling duration (`24h`, `7d`, or `30d`), and a latency threshold only for `success-within-latency-threshold`. Objective updates create an immutable effective-dated version at a slot boundary. Reports do not combine objective versions into one verdict; a window crossing versions returns version segments and an overall `INCOMPLETE` state until a full window exists for the current version.

Alternative considered: mutate the objective in place and evaluate old slots against the latest target. That makes historical reports unauditable.

### Decision 6: Versioned aggregates close after a finite correction horizon

Canonical slot classification updates one idempotent hourly aggregate keyed by monitor, objective version, and UTC hour. A result arriving after initial `missing` classification may revise that slot only until 24 hours after the slot's maturity time. Hourly buckets remain `OPEN` until every contained slot has passed that horizon, then become `CLOSED`. A UTC day compacts only after all 24 hours are closed. Reports use closed daily records for whole interior days and hourly records for the unclosed recent period and start boundary; they do not scan raw runs.

Every slot classification and aggregate has a monotonic revision/version and an idempotent source marker. Ordinary late results after the 24-hour horizon remain visible as raw operational evidence but do not rewrite closed objective history. The only closed-history exception is a documented authorized correction: an append-only correction fact identifies actor, reason, prior/new classification, affected bucket versions, and authorization; processing writes a new bucket version while retaining superseded versions. Reads select the latest authorized version. No interval edit or unrecorded repair may silently mutate a closed bucket.

Aggregate retention is bounded to 400 days, longer than the current raw-run retention and the maximum 30-day objective window. The 24-hour correction horizon means a rolling read needs at most 24 unclosed recent hourly buckets plus 24 start-boundary hourly buckets, preserving the bound of 48 hourly and 30 daily buckets per monitor/objective version. Conditional updates and reconciliation make aggregate processing safe under duplicate delivery.

Alternative considered: compute every report from raw `CheckRun` records. Raw TTL would erase report truth and service reads would become expensive fan-out queries.

### Decision 7: Service aggregation sums opportunities and preserves monitor verdicts

Service reports include only child monitors with objectives in aggregate arithmetic and explicitly count/list monitors without objectives. Counts and budget opportunity units are summed; ratios are recomputed from summed counts rather than averaged. Service state is `INCOMPLETE` if any included monitor is incomplete, otherwise `BREACHED` if any included monitor is breached, otherwise `COMPLIANT`; with no objective-bearing monitors it is `INCOMPLETE` with a no-objectives reason. Different objective types, targets, windows, and versions remain visible per monitor and are not represented as one synthetic service objective.

Alternative considered: average monitor percentages. That overweights low-cadence monitors and invents a service target when monitor targets differ.

### Decision 8: Objective reports are distinct API and UI concepts

Nested monitor and service availability read surfaces return slot counts, completeness, objective version, evaluation, budget, burn state, and exclusion summaries. They do not reuse fields named `uptime`, P99, or trend from recent service-card metrics. UI copy uses “objective,” “observed,” “missing,” and “excluded,” and states that targets are not SLAs.

Maintenance entry can suppress scheduled execution and status transitions according to the maintenance capability. On exit, current monitor state becomes `UNKNOWN` with an awaiting-observation reason; only a new recurring canonical result may establish `UP` or drive incident recovery. Manual runs can remain operational evidence but do not enter objective accounting or clear awaiting-observation state by default.

## Risks / Trade-offs

- **Late canonical results can arrive after a slot was marked missing** -> Permit automatic revision only through the explicit 24-hour post-maturity horizon; after closure, preserve history unless an authorized append-only correction creates a new bucket version.
- **The independent finalizer can fall behind or fail with the scheduler** -> Use a separate event source/runtime mode, durable per-shard cursors, strict work budgets, stale-cursor pipeline evidence, and idempotent catch-up from immutable definitions.
- **Schedule/configuration clock skew can classify a boundary slot inconsistently** -> Use server-authored UTC effective timestamps and persist the selected schedule/objective version on the slot.
- **Aggregate write amplification increases DynamoDB cost** -> Use one expected-slot fact plus one idempotent hourly update, fixed-size cursor writes, asynchronous daily compaction, no raw scans, and the bounded read limits above; estimate cost from expected opportunities, not scheduler retries or request traffic.
- **Service reports can fan out across many monitors** -> Apply existing capacity guardrails, bounded monitor-page sizes, and precomputed per-monitor buckets; do not create per-service/per-window duplicate aggregates initially.
- **Incomplete reports may frustrate users** -> Return pipeline-health correlation and explicit missing counts, but never convert infrastructure gaps into exclusions or false compliance.
- **Objective edits near a window boundary reduce continuity** -> Segment by immutable versions and wait for one complete current-version window rather than blending targets.

## Migration Plan

1. Gate implementation on the five prerequisite changes and verify their landed execution identity/deadline, delivery, pipeline-evidence, bounded-cursor/recovery, and persistent-stage contracts.
2. Add immutable effective-dated schedule, objective, enabled/disabled, maintenance-reference, correction, slot, finalizer-cursor, and aggregate item families without enabling objective evaluation.
3. Propagate stable slot identities through scheduler, queue, provider, and canonical result persistence; verify duplicates converge while scheduler work remains non-authoritative for expectation.
4. Enable the independent finalizer in a persistent non-production stage, prove the three-slot scheduler-outage golden scenario, cursor crash recovery, bounded catch-up, and 24-hour closure/version behavior.
5. Backfill only periods covered by complete immutable schedule and interval definitions. Mark earlier periods unavailable; never infer expectations from current cadence or synthesize good observations.
6. Enable hourly accounting and closed-day compaction, compare golden timeline outputs and reconciliation counts, then expose monitor reports.
7. Enable service aggregation and dashboard presentation after report completeness and finalizer lag are operationally observable.
8. Roll back reads/UI and finalizer event source independently if needed; retain durable definitions, slot facts, cursors, corrections, and aggregates. Disabling objective collection creates an effective-dated objective end rather than deleting history.

## Open Questions

- None for apply readiness. Finalization grace is derived from landed authoritative retry/deadline bounds and encoded as a repository-owned configuration constant; the late-result correction horizon is fixed by this change at 24 hours after maturity.
