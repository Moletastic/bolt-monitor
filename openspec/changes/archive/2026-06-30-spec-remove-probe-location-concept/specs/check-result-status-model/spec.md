## MODIFIED Requirements

### Requirement: System stores raw check execution results
System SHALL store raw execution results for completed healthchecks.

#### Scenario: Execution completes
- **WHEN** monitor execution finishes
- **THEN** system persists a `CheckRun` record containing monitor identity, timing, and outcome data
- **AND** the record does not require probe location or region identity
