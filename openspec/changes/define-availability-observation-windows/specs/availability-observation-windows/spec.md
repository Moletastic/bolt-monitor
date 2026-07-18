## ADDED Requirements

### Requirement: Availability objectives are optional user-defined targets
The system SHALL allow a monitor to have no availability objective or one user-defined objective, and SHALL present the objective as an operational target rather than an SLA.

#### Scenario: Monitor has no objective
- **WHEN** an operator creates or updates a monitor without an availability objective
- **THEN** recurring execution, current status, and incidents continue to operate
- **AND** objective reads return an explicit no-objective state instead of a fabricated target or ratio

#### Scenario: Objective is displayed
- **WHEN** a client displays an availability objective or its evaluation
- **THEN** it labels the value as an objective based on observed recurring opportunities
- **AND** it does not describe the value as an SLA, contractual uptime, or service credit entitlement

### Requirement: System supports a minimal objective and rolling-window model
The system SHALL support exactly the `recurring-check-success` and `success-within-latency-threshold` objective types over the duration-based rolling-window family with allowed durations `24h`, `7d`, and `30d`, and SHALL represent objective precision as integer `targetBasisPoints` in the inclusive range `1..10000`.

#### Scenario: Recurring success objective is configured
- **WHEN** an operator configures `recurring-check-success` with valid integer target basis points and an allowed rolling duration
- **THEN** the system accepts the objective without a latency threshold

#### Scenario: Latency objective is configured
- **WHEN** an operator configures `success-within-latency-threshold`
- **THEN** the system requires a positive latency threshold in milliseconds, valid integer target basis points, and an allowed rolling duration

#### Scenario: Floating-point target is submitted
- **WHEN** an operator submits a floating-point target ratio or target basis points outside `1..10000`
- **THEN** the system rejects the configuration with a typed validation error identifying the objective target field

#### Scenario: Unsupported objective model is requested
- **WHEN** an operator submits another objective type, calendar window, percentile objective, arbitrary metric expression, or unsupported duration
- **THEN** the system rejects the configuration with a typed validation error identifying the objective field

### Requirement: Objective definitions are effective-dated and auditable
The system SHALL preserve immutable objective versions and select the version effective for each scheduled slot.

#### Scenario: Objective changes
- **WHEN** an operator changes a monitor objective
- **THEN** the system creates a new version effective at a recurring slot boundary
- **AND** previously classified slots retain their original objective version and classification

#### Scenario: Rolling window crosses an objective version
- **WHEN** a requested rolling window contains slots from more than one objective version
- **THEN** the report returns separate version segments
- **AND** it does not blend different targets into one compliant verdict
- **AND** the current version remains `INCOMPLETE` until it has a complete requested window

### Requirement: Immutable effective-dated definitions independently define expected opportunities
The system SHALL derive each expected recurring opportunity independently from immutable effective-dated schedule definitions, enabled/disabled intervals, and maintenance intervals, and SHALL identify it by tenant, service, monitor, schedule-definition version, and UTC scheduled time. Scheduler work, queue messages, and execution results SHALL satisfy or explain an expected opportunity but SHALL NOT create the expectation.

#### Scenario: Schedule definition becomes effective
- **WHEN** an immutable schedule definition covers a UTC range with an interval and alignment
- **THEN** the system can enumerate every recurring slot in that range deterministically without reading scheduler work records
- **AND** each slot selects the objective, enabled/disabled, and maintenance facts effective at its scheduled time

#### Scenario: Scheduler creates recurring work
- **WHEN** the scheduler determines that an enabled monitor is due
- **THEN** it uses the stable identity of the independently expected scheduled slot before queueing execution work
- **AND** every retry or redelivery for that opportunity retains the same identity

#### Scenario: Scheduler creates no work
- **WHEN** an expected enabled non-maintenance slot matures without any scheduler work record
- **THEN** the slot still exists as an expected opportunity
- **AND** absence of work does not remove or defer it from accounting

#### Scenario: Cadence changes
- **WHEN** a monitor cadence changes
- **THEN** future slots use a new schedule-definition version
- **AND** the system does not infer or rewrite past expected opportunities using the new cadence

#### Scenario: Schedule or state version is edited
- **WHEN** schedule cadence, alignment, enablement, disablement, or maintenance state changes
- **THEN** the system appends a non-overlapping effective-dated definition or interval rather than mutating the prior fact
- **AND** it rejects contradictory overlap or backdating across matured slots except through the authorized correction policy

### Requirement: Each slot has one canonical accounting result
The system SHALL select at most one canonical terminal recurring result for each scheduled slot and SHALL make duplicate processing idempotent.

#### Scenario: Duplicate results arrive
- **WHEN** retries or duplicate deliveries produce more than one result for the same scheduled-slot identity
- **THEN** one deterministic conditional write establishes the canonical result
- **AND** duplicates do not increment aggregate counts or replace the accepted result

#### Scenario: Manual check completes
- **WHEN** an operator-triggered manual check completes without a recurring scheduled-slot identity
- **THEN** the result is excluded from objective accounting by default
- **AND** it remains distinguishable as manual operational evidence

### Requirement: Mature slots have explicit exclusive classifications
The system SHALL classify every mature expected slot as exactly one of `good`, `bad`, `missing`, or `excluded` using objective and exclusion state effective at the slot's scheduled time.

#### Scenario: Successful recurring check meets objective
- **WHEN** a non-excluded mature slot has a canonical successful result and any configured latency threshold is met
- **THEN** the slot is classified `good`

#### Scenario: Recurring check fails or exceeds latency
- **WHEN** a non-excluded mature slot has a canonical failed result or a successful result above the configured latency threshold
- **THEN** the slot is classified `bad`

#### Scenario: Expected evidence is absent
- **WHEN** a non-excluded slot passes the bounded finalization grace without a canonical result
- **THEN** the slot is classified `missing`
- **AND** missing scheduler, queue, execution-provider, or result-persistence evidence is not silently converted to `excluded`

#### Scenario: Slot is still within finalization grace
- **WHEN** an expected slot could still complete within the bounded finalization grace
- **THEN** the reporting cutoff omits that immature slot from final counts
- **AND** the system does not prematurely classify it as `missing`

### Requirement: Independent finalizer materializes and classifies matured expectations
The system SHALL run finalization and compaction from an EventBridge event source independent of recurring scheduler success, SHALL enumerate matured expectations from immutable definitions, and SHALL make finalization idempotent and recoverable without requiring execution work to exist.

#### Scenario: Finalizer processes a mature range
- **WHEN** the independent finalizer runs with a maturity cutoff
- **THEN** it derives slots from immutable schedule and interval facts between its durable cursor and that cutoff
- **AND** it conditionally materializes and classifies each slot using deterministic identity and aggregate application markers

#### Scenario: Finalizer invocation reaches its budget
- **WHEN** finalization reaches a configured maximum slot, item, page, or safe-remaining-time budget
- **THEN** it stops at a durable slot boundary and persists versioned tenant/shard continuation state
- **AND** a later invocation resumes without omission, duplicate opportunity counts, or an unbounded table scan

#### Scenario: Finalizer fails before cursor advancement
- **WHEN** slot or aggregate writes commit but the finalizer fails before advancing its cursor
- **THEN** replay converges through deterministic slot keys and idempotent aggregate application markers
- **AND** the cursor advances only after every preceding slot is durably accounted

#### Scenario: Finalizer falls behind
- **WHEN** the durable finalized watermark exceeds its documented lag tolerance
- **THEN** bounded pipeline evidence reports finalizer delay or unknown state
- **AND** objective reads do not interpret the absent finalized range as complete

#### Scenario: Scheduler outage spans three intervals
- **WHEN** immutable definitions produce three enabled non-maintenance slots, all three mature, and no execution work or canonical result exists for any slot
- **THEN** independent finalization materializes exactly three expected slot facts
- **AND** classifies exactly three opportunities as `missing`, zero as `good`, zero as `bad`, and zero as `excluded`

### Requirement: Maintenance and disabled intervals are immutable prerequisite exclusions
The system SHALL consume immutable effective-dated maintenance and enabled/disabled interval facts as prerequisites, SHALL exclude a slot only when such a fact covers its scheduled time, and SHALL retain actor/source, reason, half-open effective bounds, recorded time, version, and immutable audit identity.

#### Scenario: Scheduled slot falls in maintenance
- **WHEN** an authorized maintenance interval was recorded no later than a slot's scheduled time and its half-open effective range covers that time
- **THEN** the slot is classified `excluded` with maintenance reason and interval identity

#### Scenario: Scheduled slot falls in disabled interval
- **WHEN** an immutable disabled interval was recorded no later than the slot's scheduled time and covers that time
- **THEN** the slot is classified `excluded` with disabled reason and configuration audit identity

#### Scenario: Exclusion is inspected
- **WHEN** a client inspects excluded counts
- **THEN** the system can trace them to actor or system source, reason, effective start, effective end, and immutable audit identity

#### Scenario: Interval is declared after a mature slot
- **WHEN** maintenance or disabled history is added or changed after a slot matured
- **THEN** the system does not retroactively hide that slot as excluded
- **AND** any authorized correction is represented as an auditable correction rather than silent history mutation

#### Scenario: Interval boundary equals scheduled time
- **WHEN** a slot is at an interval's inclusive start
- **THEN** that interval applies
- **WHEN** a slot is at the interval's exclusive end
- **THEN** that interval does not apply

### Requirement: Reports expose raw accounting and completeness
The system SHALL return raw `good`, `bad`, `missing`, and `excluded` integer counts together with eligible and observed counts and exact numerator/denominator representations for achieved ratio and data completeness. It SHALL NOT use floating-point arithmetic to store targets, calculate budgets, or decide evaluation boundaries.

#### Scenario: Report contains complete observations
- **WHEN** a report has `good=98`, `bad=2`, `missing=0`, and `excluded=5`
- **THEN** `eligible` is `100`, `observed` is `100`, achieved ratio is `98/100`, and data completeness is `100/100`
- **AND** excluded opportunities are disclosed but are not in the eligible denominator

#### Scenario: Report contains missing observations
- **WHEN** a report has `good=90`, `bad=5`, `missing=5`, and `excluded=0`
- **THEN** `eligible` is `100`, `observed` is `95`, achieved ratio is `90/95`, and data completeness is `95/100`
- **AND** the observed achieved ratio is not presented without its incomplete-data state

#### Scenario: Report has no eligible observations
- **WHEN** every expected slot is excluded or no mature expected slot exists
- **THEN** achieved ratio and data completeness are unavailable
- **AND** the evaluation state is `INCOMPLETE` with a machine-readable reason

### Requirement: Reports distinguish compliant, breached, and incomplete states
The system SHALL evaluate a report as `INCOMPLETE` when any eligible observation is missing, otherwise as `COMPLIANT` when `good * 10000 >= observed * targetBasisPoints` using overflow-safe integer/rational arithmetic, or `BREACHED` when that exact comparison fails.

#### Scenario: Complete report meets target
- **WHEN** a report has no missing slots and its achieved ratio is greater than or equal to the effective objective target
- **THEN** the evaluation state is `COMPLIANT`

#### Scenario: Achieved ratio equals target exactly
- **WHEN** a complete report satisfies `good * 10000 == observed * targetBasisPoints`
- **THEN** the evaluation state is `COMPLIANT`
- **AND** display rounding cannot change that verdict

#### Scenario: Complete report misses target
- **WHEN** a report has no missing slots and its achieved ratio is below the effective objective target
- **THEN** the evaluation state is `BREACHED`

#### Scenario: Partial observed ratio meets target
- **WHEN** a report's observed achieved ratio meets the target but at least one eligible slot is missing
- **THEN** the evaluation state is `INCOMPLETE`
- **AND** the system does not claim compliance or breach

### Requirement: Reports expose opportunity-based error budget and burn state
The system SHALL report integer `targetBasisPoints`, `allowedBad = floor(eligible * (10000 - targetBasisPoints) / 10000)`, `consumedBad`, `remainingBad`, and a basic burn state using overflow-safe integer arithmetic.

#### Scenario: Error budget is calculated
- **WHEN** a complete report has `eligible=1000`, `targetBasisPoints=9900`, and `bad=6`
- **THEN** `allowedBad` is `10`, `consumedBad` is `6`, and `remainingBad` is `4`
- **AND** the burn state is `BURNING` because more than half but not more than all allowed bad opportunities are consumed

#### Scenario: Error budget is exhausted
- **WHEN** consumed bad opportunities exceed allowed bad opportunities
- **THEN** remaining bad opportunities are negative
- **AND** the burn state is `EXHAUSTED`

#### Scenario: Error budget is exactly consumed
- **WHEN** `allowedBad` is positive and `consumedBad == allowedBad`
- **THEN** `remainingBad` is zero
- **AND** burn state is `BURNING`, not `EXHAUSTED`

#### Scenario: Error budget has zero allowance
- **WHEN** `allowedBad=0` and `consumedBad=0`
- **THEN** burn state is `HEALTHY`
- **WHEN** `allowedBad=0` and `consumedBad>0`
- **THEN** burn state is `EXHAUSTED` and `remainingBad` is negative

#### Scenario: Report is incomplete
- **WHEN** a report contains missing opportunities
- **THEN** raw error-budget counts remain visible
- **AND** burn state is `UNAVAILABLE` so missing evidence cannot produce a reassuring burn verdict

### Requirement: Durable aggregates survive raw run expiration with bounded cost
The system SHALL maintain idempotent versioned UTC hourly and daily count aggregates for objective reports, retain them for 400 days independently of raw `CheckRun` TTL, and bound steady-state accounting to one slot fact and one idempotent hourly aggregate application per mature expected opportunity plus bounded cursor and compaction work.

#### Scenario: Raw runs expire
- **WHEN** raw check runs age out under their configured TTL
- **THEN** hourly and daily aggregate records needed by a retained objective window remain readable
- **AND** objective reports do not require expired raw runs

#### Scenario: Rolling report is read
- **WHEN** the system evaluates a monitor's rolling window
- **THEN** it reads daily buckets for complete interior days and hourly buckets for boundary periods
- **AND** the read is bounded to at most 30 daily buckets and 48 hourly buckets per monitor and objective version

#### Scenario: Slot is processed more than once
- **WHEN** duplicate accounting events target the same slot and hour
- **THEN** idempotency prevents duplicate aggregate increments

### Requirement: Late results and closed-bucket corrections are bounded and versioned
The system SHALL automatically reconcile a result that arrives after initial `missing` classification only through 24 hours after that slot's maturity time. It SHALL close hourly buckets after every contained slot passes that horizon, compact a UTC day only after all its hourly buckets close, and SHALL NOT rewrite closed history except through the documented authorized correction policy.

#### Scenario: Late result arrives within correction horizon
- **WHEN** a canonical result arrives no later than 24 hours after a slot's maturity time
- **THEN** the slot and open hourly aggregate receive deterministic new revisions replacing `missing` with the result-derived classification
- **AND** idempotency applies the delta exactly once

#### Scenario: Ordinary result arrives after correction horizon
- **WHEN** a canonical result arrives more than 24 hours after maturity and the bucket is closed
- **THEN** the result remains visible as late operational evidence
- **AND** it does not change the closed slot classification, aggregate, or historical objective verdict

#### Scenario: Closed history receives authorized correction
- **WHEN** an authorized correction supplies actor, reason, authorization, prior classification, and new classification for a closed slot
- **THEN** the system appends the correction and writes new monotonic slot and affected bucket versions
- **AND** retains superseded versions and selects the latest authorized version for reads

#### Scenario: Closed history receives unapproved mutation
- **WHEN** an interval edit, repair, or late result attempts to mutate a closed bucket without an authorized correction
- **THEN** the system rejects or records the attempt without changing objective history

### Requirement: Monitor objective reads remain distinct from recent samples
The system SHALL expose objective reports through a monitor availability read surface that is distinct from latest status, raw run history, and recent-sample metrics.

#### Scenario: Client requests monitor objective report
- **WHEN** a client requests availability for a nested service monitor and allowed rolling duration
- **THEN** the response includes objective version, raw counts, ratios, completeness, evaluation, budget, burn, and exclusion summary
- **AND** it does not label the report as recent uptime or derive it from only the latest bounded raw samples

### Requirement: Service aggregation preserves monitor semantics
The system SHALL aggregate service objective reports by summing child-monitor opportunity counts and budget counts, recomputing ratios, and preserving each monitor's objective and evaluation.

#### Scenario: Service has objective-bearing monitors
- **WHEN** a client requests service availability
- **THEN** the system sums good, bad, missing, excluded, eligible, observed, allowed-bad, and consumed-bad counts across objective-bearing child monitors
- **AND** it recomputes aggregate ratios from summed counts instead of averaging monitor percentages
- **AND** it returns per-monitor objective type, target, window, version, and state

#### Scenario: Service has mixed monitor verdicts
- **WHEN** any included monitor is `INCOMPLETE`
- **THEN** service evaluation is `INCOMPLETE`
- **WHEN** no included monitor is incomplete and any included monitor is `BREACHED`
- **THEN** service evaluation is `BREACHED`
- **WHEN** all included monitors are `COMPLIANT`
- **THEN** service evaluation is `COMPLIANT`

#### Scenario: Service contains monitors without objectives
- **WHEN** service aggregation encounters child monitors without objectives
- **THEN** those monitors are excluded from aggregate arithmetic
- **AND** the response explicitly reports their count and identities rather than treating them as compliant

#### Scenario: Service has no objective-bearing monitors
- **WHEN** no child monitor has an objective
- **THEN** service evaluation is `INCOMPLETE` with a no-objectives reason
- **AND** the system does not invent a service objective

### Requirement: Golden timelines define accounting and arithmetic behavior
The system SHALL maintain deterministic golden fixtures covering slot generation, classification, aggregation, objective versioning, and report arithmetic.

#### Scenario: Golden timeline suite runs
- **WHEN** automated tests evaluate timelines containing successes, failures, latency breaches, duplicates, delayed and post-horizon results, missing pipeline stages, a three-interval scheduler outage with no work, maintenance, disablement, manual checks, objective/schedule changes, bucket closure, authorized corrections, and raw-run expiry
- **THEN** exact slot identities, classifications, bucket versions, counts, rational report ratios, integer budgets, burn states, and monitor/service evaluations match reviewed golden outputs

### Requirement: Availability implementation is gated by prerequisite contracts
The change SHALL be implemented and rolled out only after retry-safe execution, notification and escalation delivery assurance, monitoring-pipeline evidence, data recovery and capacity guardrails, and stage-resource lifecycle contracts have landed.

#### Scenario: Prerequisite gate is reviewed
- **WHEN** implementation begins or rollout is approved
- **THEN** evidence confirms the landed contracts for canonical recurring identity and deadlines, durable notification outcomes, bounded pipeline correlation, cursor/recovery/capacity behavior, and persistent-stage resource protection
- **AND** this change does not invent parallel execution, delivery, recovery, or stage-lifecycle semantics
