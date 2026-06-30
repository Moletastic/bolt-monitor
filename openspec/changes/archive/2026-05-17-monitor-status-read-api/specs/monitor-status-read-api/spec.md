## ADDED Requirements

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
