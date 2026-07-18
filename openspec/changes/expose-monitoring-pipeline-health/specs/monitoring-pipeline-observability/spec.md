## ADDED Requirements

### Requirement: Runtime emits structured secret-free correlation logs
The system SHALL emit machine-parseable structured lifecycle logs for scheduler, execution, incident transition, notification dispatch, notification attempt, terminal delivery, suppression, and recovery using stable correlation identifiers available at each boundary.

#### Scenario: Execution crosses runtime boundaries
- **WHEN** the scheduler creates/publishes work and the worker processes its result
- **THEN** each applicable lifecycle log includes event name, component, stage, outcome, timestamp, and the same `runId`
- **AND** logs include `incidentId` when an incident transition exists

#### Scenario: Notification crosses dispatch, queue, and provider boundaries
- **WHEN** incident work creates, dispatches, consumes, or retries a notification delivery
- **THEN** applicable logs include stable `transitionId`, `deliveryId`, and `incidentId`
- **AND** queue transport identifiers are separately named `sqsMessageId`

#### Scenario: Runtime handles secret-bearing input
- **WHEN** runtime logs a monitor, queue, notification, provider, auth, or error event
- **THEN** it excludes targets, request headers/bodies, expected content, channel destinations/configuration, provider payloads/responses, credentials, token/cookie values, and raw queue bodies
- **AND** it records bounded reason classifications instead of unbounded exceptions where exposure is possible

### Requirement: Infrastructure provisions a repository-wide bounded default signal pack
The system SHALL provision the source-controlled repository-wide signal inventory for scheduler, execution, notification, protected API, evaluator, and deployed auth/RBAC operations. Alarm and metric resource count SHALL be fixed by runtime roles, SHALL use native AWS metrics first, and SHALL NOT scale with services, monitors, runs, incidents, deliveries, channels, users, or tenants.

#### Scenario: Monitoring runtime is deployed
- **WHEN** infrastructure creates repository runtime resources
- **THEN** it creates optional-SNS alarms for scheduler errors and missing heartbeat, execution and notification queue oldest age, execution and notification DLQ visible depth, fixed actionable runtime-role errors/throttles, and protected API `5xx`
- **AND** it creates fixed auth key/storage and sustained-refresh alarms only when auth/RBAC resources are deployed

#### Scenario: Alarm metric is selected
- **WHEN** an AWS service publishes the required signal
- **THEN** the alarm uses that native metric
- **AND** a custom metric is permitted only for fixed scheduler heartbeat or auth-domain state unavailable as a native metric

#### Scenario: Application fixture cardinality grows
- **WHEN** monitor, service, incident, delivery, channel, user, or tenant fixtures increase
- **THEN** alarm, metric, log-group, and dashboard-control counts remain unchanged
- **AND** no resource or dimension is created from an application row or correlation identity

#### Scenario: Custom metric is emitted
- **WHEN** scheduler heartbeat or an allowed auth-domain event emits a custom metric
- **THEN** dimensions are restricted to source-controlled fixed values such as service, stage, component, operation, and outcome
- **AND** no dimension contains tenant, monitor, service-domain, run, incident, transition, delivery, queue-message, channel, user, target, URL, or error text

### Requirement: Every default signal has an explicit action classification
The default inventory SHALL classify scheduler errors/heartbeat, queue ages, both DLQs, fixed runtime errors/throttles, protected API `5xx`, auth key/storage failure, and sustained auth refresh failure as optional-SNS alarms. It SHALL classify authorization denials, sign-in/recovery events, evaluator overdue/stuck/terminal-delivery aggregates, and target `DOWN` as dashboard/runbook evidence without default alarm actions.

#### Scenario: SNS topic ARN is configured
- **WHEN** deployment supplies the optional alarm-action topic ARN
- **THEN** every inventory entry classified `Optional SNS alarm` attaches that action according to its configuration
- **AND** dashboard-only evidence does not create an SNS action

#### Scenario: No SNS topic ARN is configured
- **WHEN** the optional alarm-action destination is absent
- **THEN** CloudWatch alarms remain deployed and visible without an action
- **AND** deployment output and runbooks state that no external notification destination is configured

#### Scenario: Authorization denials are observed
- **WHEN** authentication or membership authorization rejects requests
- **THEN** structured security evidence remains available to the dashboard/runbook path
- **AND** no default denial alarm or per-user metric is created

### Requirement: Alarm thresholds, missing data, and recovery are deterministic
Every alarm SHALL have source-controlled threshold, period, datapoints-to-alarm, missing-data behavior, and recovery behavior. Scheduler errors SHALL default to two of five one-minute periods, scheduler heartbeat to three missing one-minute periods, queue age to greater than five minutes for three of five periods, DLQ depth to greater than zero for one period, and remaining fixed alarms to their documented repository defaults.

#### Scenario: Isolated datapoint remains within tolerance
- **WHEN** one datapoint does not satisfy an alarm's documented sustained or terminal threshold
- **THEN** the alarm remains non-breaching according to its evaluation configuration

#### Scenario: Required liveness data is missing
- **WHEN** scheduler heartbeat or evaluator readiness data is absent for its configured periods
- **THEN** missing data is treated as breaching or unknown according to the paired liveness contract
- **AND** sparse event counters do not independently treat normal silence as failure

#### Scenario: Alarmed condition clears
- **WHEN** healthy datapoints satisfy recovery periods and relevant backlog or DLQ state is cleared
- **THEN** the alarm returns to `OK` without recreation

### Requirement: Alarms provide direct remediation context
Every alarm SHALL identify stage, owner, affected runtime scope, action/no-action status, and a stable runbook link with diagnosis, mitigation, recovery verification, and escalation guidance.

#### Scenario: Operator inspects an alarm
- **WHEN** an alarm changes state or appears in CloudWatch
- **THEN** its name/description identifies scheduler, execution, notification, API, evaluator, or auth scope
- **AND** its runbook uses safe correlation fields and bounded persisted/native evidence without requiring secret inspection

### Requirement: CloudWatch retention and FinOps acceptance are explicit
The system SHALL configure bounded retention for application Lambda log groups it owns and SHALL document low-use owner, expected validation, and 1,000-monitor stress-profile monthly costs using a stated pricing date and deployment region. Resources enabled by default for the low-use owner profile SHALL project at or below USD 1 incremental cost per persistent stage per month, excluding but separately itemizing optional SNS deliveries and external dead-man provider fees. Expected and stress profiles SHALL remain visible but SHALL NOT be represented as free-tier defaults.

#### Scenario: Runtime log group is managed by the stack
- **WHEN** the stack provisions or adopts an application Lambda log group
- **THEN** it applies the named 14-day low-use retention default unless the installation explicitly selects a longer documented cost profile

#### Scenario: Operator reviews observability cost
- **WHEN** the FinOps worksheet is generated
- **THEN** it includes structured log volume/retention, fixed alarms/custom metrics, evaluator Lambda, DynamoDB projection reads/writes/storage, SQS/API reads, and optional integrations for low, expected, and upper cases
- **AND** it records stage attribution, pricing date, region, cadence, item/log-size, and request assumptions

#### Scenario: Cost acceptance exceeds the cap
- **WHEN** projected or month-normalized cost introduced by the default low-use health pack exceeds USD 1 per persistent stage
- **THEN** default production enablement is blocked until the pack is reduced or an installation explicitly opts into and documents the higher-cost profile
- **AND** the system does not automatically disable monitoring as a budget response

### Requirement: Failure and recovery drills verify signals and health
The system SHALL provide repeatable staging-first drills for scheduler error/missing heartbeat, execution queue age/DLQ, notification dispatch/queue age/DLQ, projection incompleteness, and applicable auth alarms, including recovery verification and secret-safe evidence.

#### Scenario: Operator runs a supported failure drill
- **WHEN** the operator injects synthetic non-customer failure in staging
- **THEN** the drill identifies expected structured logs, alarm and action/no-action state, and pipeline API/dashboard state where enabled
- **AND** it does not send real customer notifications or expose secrets

#### Scenario: Operator completes drill recovery
- **WHEN** the injected condition is removed and affected work is safely replayed, redriven, quarantined, or repaired according to its runbook
- **THEN** the drill verifies the alarm returns from `ALARM` to `OK`
- **AND** complete evidence returns the API/dashboard to a non-failing state while incomplete evidence remains `UNKNOWN` rather than healthy

### Requirement: External dead-man assurance is represented accurately
The system SHALL document separately operated external dead-man monitoring as optional and SHALL distinguish it from internal CloudWatch alarms and optional SNS delivery.

#### Scenario: No external dead-man integration is configured
- **WHEN** operators inspect documentation or health UI
- **THEN** the system does not claim independent external assurance
- **AND** internal heartbeat alarms and optional SNS actions are described as same-installation detection/delivery

#### Scenario: Operator configures an external dead-man service
- **WHEN** a separately operated service expects a bounded heartbeat
- **THEN** documentation describes minimal secret-safe payload, timeout, ownership, test procedure, and separately itemized cost
- **AND** the integration remains optional and does not change pipeline health semantics
