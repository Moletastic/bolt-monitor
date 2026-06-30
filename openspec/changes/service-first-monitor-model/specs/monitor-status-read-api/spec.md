## MODIFIED Requirements

### Requirement: System exposes latest monitor status through HTTP API
System SHALL expose the latest persisted monitor status through nested service-monitor HTTP API.

#### Scenario: Client requests monitor status
- **WHEN** client requests status for existing monitor through existing service path
- **THEN** system returns the latest status snapshot for that nested monitor

### Requirement: System exposes recent run history through HTTP API
System SHALL expose recent raw execution history for a monitor through nested service-monitor HTTP API.

#### Scenario: Client requests monitor run history
- **WHEN** client requests recent runs for existing monitor through existing service path
- **THEN** system returns recent persisted `CheckRun` records for that nested monitor

### Requirement: System exposes dashboard-oriented monitor reads
System SHALL expose monitor read responses suitable for dashboard-style views under service-first read surfaces.

#### Scenario: Client requests monitor collection for dashboard
- **WHEN** client requests nested monitor listing for one service
- **THEN** system can return monitor resources with current status summary derived from persisted status data
