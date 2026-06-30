## MODIFIED Requirements

### Requirement: System emits normalized execution result
System SHALL emit a normalized execution result shape for downstream result and status processing.

#### Scenario: Check finishes
- **WHEN** a healthcheck execution completes
- **THEN** system produces normalized result data describing monitor identity, location, timing, outcome, and protocol-specific details needed downstream

### Requirement: System enqueues notification event on state transition
System SHALL enqueue a notification event to the `notification-queue` SQS queue when a check run results in a state transition.

#### Scenario: Monitor transitions from UP to DOWN
- **WHEN** a check run transitions a monitor's status from UP to DOWN
- **THEN** system enqueues a notification event with `eventType: "incident.opened"` to `notification-queue`

#### Scenario: Monitor transitions from DOWN to UP
- **WHEN** a check run transitions a monitor's status from DOWN to UP
- **THEN** system enqueues a notification event with `eventType: "incident.resolved"` to `notification-queue`

#### Scenario: Monitor state unchanged
- **WHEN** a check run does not change the monitor's status (UP→UP or DOWN→DOWN)
- **THEN** system does NOT enqueue a notification event

#### Scenario: Monitor in maintenance
- **WHEN** a check run executes while monitor is in maintenance state
- **THEN** system does NOT enqueue a notification event even if status changes
