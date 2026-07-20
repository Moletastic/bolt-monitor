## ADDED Requirements

### Requirement: Delayed escalation steps use self-cleaning one-time schedules
The system SHALL schedule delayed escalation steps with EventBridge Scheduler one-time `at` schedules, or an equivalent mechanism that automatically removes completed schedules. Each schedule SHALL use a deterministic bounded-length name derived from tenant, incident transition, and policy step identity, SHALL disable flexible time windows, and SHALL specify action-after-completion deletion.

#### Scenario: Delayed step is scheduled
- **WHEN** an active escalation advances to a step with a positive delay
- **THEN** the system creates one one-time schedule for the exact UTC execution time
- **AND** configures the schedule to delete after completion

#### Scenario: Schedule creation is retried
- **WHEN** duplicate processing attempts to schedule the same transition and step
- **THEN** the deterministic name resolves to the same schedule
- **AND** the operation succeeds idempotently without creating another schedule

#### Scenario: Schedule completes
- **WHEN** the one-time schedule completes its target invocation
- **THEN** the scheduling service removes the schedule automatically
- **AND** no annual cron rule remains capable of replaying the step

### Requirement: Scheduled work is delivered through the existing notification queue
The one-time schedule SHALL target the existing notification SQS queue with a canonical message containing stable transition and step identity plus `sourceKind=scheduler_target`. Scheduler target delivery SHALL have named finite retry age and attempt count, the maximum retry age SHALL cover the configured maximum Scheduler target backoff, and target-delivery exhaustion SHALL use the existing notification DLQ without consuming a provider delivery attempt.

#### Scenario: Scheduler target invocation succeeds
- **WHEN** the schedule sends its message to the notification queue
- **THEN** the escalation runtime processes the scheduled step through the same idempotent queue path as immediate work

#### Scenario: Scheduler cannot send to target queue
- **WHEN** Scheduler exhausts its bounded target retry policy
- **THEN** Scheduler sends the failed target invocation to the existing notification DLQ
- **AND** the schedule remains eligible for action-after-completion deletion

#### Scenario: Scheduler DLQ work is considered for redrive
- **WHEN** an operator inspects a `scheduler_target` failure envelope
- **THEN** the system revalidates canonical step identity and current incident/escalation eligibility before enqueue
- **AND** stale, recovered, malformed, or unknown work is quarantined or suppressed rather than blindly redriven

### Requirement: Scheduler retries are separate from delivery attempts
Scheduler target retries SHALL only attempt to enqueue the scheduled-step message. Provider attempts SHALL begin only after SQS processing conditionally claims a delivery, and Scheduler exhaustion SHALL NOT increment delivery attempt count. Infrastructure SHALL assert finite Scheduler retry attempts/age and notification-DLQ configuration alongside notification Lambda timeout, delivery lease, SQS visibility, and receive-budget ordering.

#### Scenario: Scheduler retries queue delivery
- **WHEN** Scheduler retries a failed target send
- **THEN** no delivery is created or claimed solely because of that retry

#### Scenario: Scheduler enqueue succeeds ambiguously more than once
- **WHEN** duplicate canonical scheduled-step messages reach SQS
- **THEN** deterministic delivery identity and conditional claims prevent concurrent provider sends

### Requirement: Scheduler permissions are least privilege and do not leak per schedule
Infrastructure SHALL provide a Scheduler execution role restricted to sending messages to the existing notification queue and notification DLQ as required by the target and dead-letter configuration. Runtime scheduling permissions SHALL be scoped to the managed schedule group and SHALL NOT add per-schedule Lambda resource-policy statements.

#### Scenario: Scheduler invokes delayed work
- **WHEN** a one-time schedule executes
- **THEN** its execution role can send only to the configured notification queue and DLQ resources needed by the schedule

#### Scenario: Runtime creates a schedule
- **WHEN** the escalation runtime creates or reconciles a one-time schedule
- **THEN** its IAM permissions are limited to the managed schedule group and passing the dedicated Scheduler execution role

#### Scenario: Many incidents schedule steps
- **WHEN** many one-time schedules are created and completed
- **THEN** no per-incident Lambda invoke permission statements accumulate
- **AND** no persistent EventBridge rules or targets accumulate

### Requirement: Legacy annual escalation rules are eliminated safely
The deployment and operations procedure SHALL identify and remove legacy `esc-*-step-*` EventBridge rules, their targets, and corresponding `allow-events-*` Lambda permissions after the new scheduler path is active. The runtime SHALL stop creating those resources.

#### Scenario: New version is deployed
- **WHEN** escalation scheduling runs after cutover
- **THEN** it does not call EventBridge `PutRule` or `PutTargets`
- **AND** it does not call Lambda `AddPermission`

#### Scenario: Legacy resources exist at cutover
- **WHEN** operators run the documented migration cleanup
- **THEN** matching legacy escalation rules and targets are inventoried and removed
- **AND** matching stale Lambda permissions are removed without modifying unrelated rules or statements
