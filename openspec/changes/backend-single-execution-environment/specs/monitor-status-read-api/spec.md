## MODIFIED Requirements

### Requirement: System exposes current monitor status
System SHALL expose the latest status snapshot for a monitor.

#### Scenario: Client reads monitor status
- **WHEN** client requests current monitor status
- **THEN** system returns current status, last check time, duration, last outcome, and latest error when available
- **AND** the response does not include last probe location or region identity
