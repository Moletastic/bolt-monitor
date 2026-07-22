## ADDED Requirements

### Requirement: Scheduler enumeration is bounded and resumable
The scheduler SHALL enumerate enabled monitors through bounded pages of the tenant-scoped `AppTable` scheduling projection. The projection and its authoritative due-time fields SHALL be coordinated with monitoring pipeline health so both consumers reuse one access pattern rather than create duplicate due-time indexes. Each invocation SHALL stop before its configured item, page, enqueue, or safe remaining-time budget is exhausted and SHALL persist continuation state that allows a later invocation to resume without starvation.

#### Scenario: All due monitors fit within one invocation
- **WHEN** scheduler enumeration reaches the end of the projection within all budgets
- **THEN** it completes the traversal and clears the completed continuation checkpoint

#### Scenario: Scheduler reaches an invocation budget
- **WHEN** enumeration or enqueueing reaches an invocation budget before the projection is exhausted
- **THEN** the scheduler persists the next opaque continuation position
- **AND** a later invocation resumes from that position

#### Scenario: Scheduler resumes after a failure
- **WHEN** an invocation fails after partially processing a page
- **THEN** retry-safe execution identity and existing enqueue idempotency prevent duplicate logical runs
- **AND** unprocessed monitors remain reachable by the current or prior safe checkpoint

#### Scenario: Projection changes during traversal
- **WHEN** monitors are added, updated, disabled, or deleted while a traversal is in progress
- **THEN** subsequent traversals converge on current canonical configuration
- **AND** no enabled monitor can be permanently starved by a stale cursor

#### Scenario: Pipeline health needs due-time evidence
- **WHEN** scheduler and pipeline-health designs require monitor due-time access
- **THEN** they reuse one AppTable projection and key/index access pattern
- **AND** a new shared sparse index is added only when measured evidence shows existing keys and indexes cannot meet bounded-read criteria
- **AND** AuthTable lifecycle due-work indexing remains a separate security workflow access pattern

### Requirement: Scheduler exposes capacity guardrail signals
The scheduler SHALL emit structured measurements for evaluated monitors, due monitors, pages, continuation, enqueue count, consumed capacity when available, duration, safe remaining time, and envelope violations without logging monitor secrets.

#### Scenario: Scheduler reaches a guardrail
- **WHEN** scheduler work stops because a configured budget is reached
- **THEN** telemetry identifies the exhausted budget and whether continuation was saved
- **AND** the invocation is not reported as a complete traversal

#### Scenario: Projection exceeds measured stress profile
- **WHEN** scheduler measurements observe more than 1,000 monitors for the supported tenant
- **THEN** the system emits an actionable operational support-boundary warning
- **AND** continues bounded processing rather than issuing unbounded queries or claiming unlimited scale
- **AND** rejects configuration only if a separately documented safety limit is reached
