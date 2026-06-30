## MODIFIED Requirements

### Requirement: System exposes latest monitor status through HTTP API
System SHALL expose the latest persisted monitor status through HTTP API.

#### Scenario: Client requests monitor status
- **WHEN** client requests status for existing monitor
- **THEN** system returns the latest status snapshot for that monitor

### Requirement: System exposes recent run history through HTTP API
System SHALL expose recent raw execution history for a monitor through HTTP API.

#### Scenario: Client requests monitor run history
- **WHEN** client requests recent runs for existing monitor
- **THEN** system returns recent persisted `CheckRun` records for that monitor

### Requirement: System exposes dashboard-oriented monitor reads
System SHALL expose monitor read responses suitable for dashboard-style views.

#### Scenario: Client requests monitor collection for dashboard
- **WHEN** client requests monitor listing
- **THEN** system can return monitor resources with current status summary derived from persisted status data

### Requirement: Status response includes currentStatus field
System SHALL include the `currentStatus` field in monitor status responses, with values derived from the latest execution outcome.

#### Scenario: Status shows up
- **WHEN** latest execution outcome is "success"
- **THEN** `currentStatus` SHALL be "up"

#### Scenario: Status shows down
- **WHEN** latest execution outcome is "failure"
- **THEN** `currentStatus` SHALL be "down"

#### Scenario: Status shows unknown
- **WHEN** no execution has completed for the monitor
- **THEN** `currentStatus` SHALL be "unknown"

### Requirement: Status response includes lastDurationMs field
System SHALL include the `lastDurationMs` field in monitor status responses, representing the duration of the last execution in milliseconds.

#### Scenario: Duration is available
- **WHEN** a monitor has completed at least one execution
- **THEN** `lastDurationMs` SHALL be included in the status response
- **AND** SHALL contain the duration in milliseconds from the latest execution

#### Scenario: Duration is not available
- **WHEN** a monitor has never completed an execution
- **THEN** `lastDurationMs` SHALL be null or omitted from the response
