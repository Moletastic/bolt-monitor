## Purpose

Define operator incident visibility, acknowledgement, and resolution semantics, and lock in canonical execution-driven incident transitions tied to the retry-safe execution pipeline.
## Requirements
### Requirement: System exposes incident collection through HTTP API
System SHALL allow operators to read incident records through HTTP API.

#### Scenario: Operator requests incident collection
- **WHEN** operator calls `GET /api/v1/incidents`
- **THEN** system returns incident resources suitable for operational overview use

### Requirement: System exposes incident detail and monitor-scoped incident reads
System SHALL allow operators to inspect one incident directly and inspect incidents for one monitor.

#### Scenario: Operator requests incident by ID
- **WHEN** operator calls `GET /api/v1/incidents/{incidentId}` for existing incident
- **THEN** system returns incident resource for that incident

#### Scenario: Operator requests incidents for one monitor
- **WHEN** operator calls `GET /api/v1/services/{serviceId}/monitors/{monitorId}/incidents` for existing monitor
- **THEN** system returns incidents associated with that monitor

### Requirement: System exposes incident acknowledgement and resolution as commands
System SHALL allow operators to acknowledge or resolve incidents through dedicated action endpoints.

#### Scenario: Operator acknowledges incident
- **WHEN** operator calls `POST /api/v1/incidents/{incidentId}/ack` for actionable incident
- **THEN** system records acknowledgement state for that incident

#### Scenario: Operator resolves incident
- **WHEN** operator calls `POST /api/v1/incidents/{incidentId}/resolve` for actionable incident
- **THEN** system records resolution state for that incident

### Requirement: System keeps incident lifecycle ownership in business rules
System SHALL keep incident creation and default closure under ordered recurring execution business rules rather than exposing generic incident create/delete CRUD or allowing manual/stale results to change lifecycle.

#### Scenario: Client inspects API shape
- **WHEN** client needs incident visibility and operator actions
- **THEN** system provides read routes and acknowledgement/resolution action routes
- **AND** does not require or allow clients to create incidents directly through generic CRUD

#### Scenario: In-order recurring observation crosses a threshold
- **WHEN** a canonical recurring result newer than the status cursor crosses the configured failure or recovery threshold
- **THEN** system conditionally creates or updates one incident transition causally identified by `runId`, `scheduleDefinitionVersion`, and `scheduledFor`

#### Scenario: Manual or stale observation is committed
- **WHEN** a manual result or out-of-order recurring result is stored as a canonical `CheckRun`
- **THEN** it does not open, update, resolve, or notify an incident

### Requirement: Execution-driven incident transitions are idempotent
System SHALL persist incident state, one deterministic incident activity/audit identity set, and one canonical pending notification outbox item atomically with the causal recurring result.

#### Scenario: Result transaction is retried
- **WHEN** a recurring result that caused an incident transition is committed again
- **THEN** system retains one incident transition, one deterministic activity/audit identity set, and one canonical outbox item
- **AND** `transitionId`, persisted activity `activityId`, and outbox-envelope `eventId` are the same value
- **AND** it does not apply threshold counters again

#### Scenario: Older result arrives after transition
- **WHEN** an older recurring ordering key finishes after a newer key has advanced incident state
- **THEN** the older result cannot reopen, resolve, or rewrite that incident
- **AND** it creates no notification outbox item

### Requirement: Incident notification outbox carries causal identity
System SHALL include stable transition and execution identity in every execution-driven canonical notification outbox item.

#### Scenario: Incident transition creates event
- **WHEN** an in-order recurring result opens or resolves an incident
- **THEN** the durable outbox item contains equal `transitionId`/`eventId`, causal `runId`, `trigger=recurring`, `scheduleDefinitionVersion`, UTC `scheduledFor`, incident ID, transition type, and transition time
- **AND** it uses the item model/key consumed by `assure-notification-and-escalation-delivery`

### Requirement: Notification dispatch ownership is exclusive
System SHALL stop at atomic creation of the canonical pending notification outbox item. `assure-notification-and-escalation-delivery` SHALL be the sole owner of notification SQS send, dispatch acknowledgement, downstream claim/deduplication, route initiation, and per-channel outcomes.

#### Scenario: Result and incident transition commit
- **WHEN** the transaction creates a notification-relevant transition
- **THEN** execution runtime performs no direct notification queue send or dispatch acknowledgement
- **AND** the pending outbox item remains safe during temporary dispatcher unavailability

#### Scenario: Dispatcher is rolled out
- **WHEN** outbox-producing execution paths are enabled
- **THEN** the notification-assurance dispatcher is already provisioned and enabled, or both are enabled atomically
- **AND** there is no direct-send fallback or silent notification gap

#### Scenario: Delivery protocol proceeds
- **WHEN** notification assurance dispatches or consumes the transition
- **THEN** its specifications exclusively define acknowledgement, claims, retries, and channel outcomes

