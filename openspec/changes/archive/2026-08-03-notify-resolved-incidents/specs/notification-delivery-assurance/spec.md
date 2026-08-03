## ADDED Requirements

### Requirement: Incident recovery notifies previously contacted policy channels
When an `incident.up` transition reaches the escalation runtime for an incident with escalation state, the system SHALL create and deliver one recovery notification to each unique channel selected by a step recorded in that state's `stepsFired` collection. It SHALL use the canonical recovery transition identity to create distinct durable delivery records before provider I/O, then suppress remaining escalation work.

#### Scenario: Incident recovers after multiple escalation steps fired
- **WHEN** an incident's escalation state records multiple fired steps and the system handles its canonical `incident.up` transition
- **THEN** it sends one recovery notification to every unique channel selected by those fired steps
- **AND** it marks the escalation state suppressed so no unfinished step can deliver an outage notification

#### Scenario: Incident recovers before later escalation step fires
- **WHEN** only the first escalation step fired before an incident recovers
- **THEN** the system sends recovery notifications only to unique channels from that fired step
- **AND** does not notify channels from later unfired steps

#### Scenario: Recovery transition is delivered more than once
- **WHEN** the notification queue delivers the same canonical recovery transition to concurrent or retried workers
- **THEN** all workers resolve the same transition-scoped recovery delivery identities
- **AND** at most one worker calls each provider for each recovery channel

#### Scenario: Recovery transition follows outage transition for same incident
- **WHEN** recovery deliveries are created for an incident that already has outage deliveries
- **THEN** recovery delivery identities use the canonical recovery transition identity
- **AND** they do not collide with delivery identities from the outage transition
