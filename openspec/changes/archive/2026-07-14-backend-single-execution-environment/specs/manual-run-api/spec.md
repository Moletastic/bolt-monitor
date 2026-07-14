## MODIFIED Requirements

### Requirement: System can trigger a manual monitor run
System SHALL allow operators to trigger an immediate check for an enabled monitor.

#### Scenario: Manual run succeeds
- **WHEN** operator triggers a manual run for an enabled monitor
- **THEN** system executes one check attempt in the system execution environment
- **AND** the response includes run identity, trigger, timing, duration, outcome, status code when available, and error when available
- **AND** the response does not include probe-location or region identity
