## ADDED Requirements

### Requirement: Accepted manual run commands produce downstream execution work
System SHALL turn every accepted manual run command into internal execution work for the shared runtime pipeline.

#### Scenario: Runnable monitor is requested manually
- **WHEN** operator calls `POST /api/v1/monitors/{id}/run` for an enabled existing monitor
- **THEN** system accepts the command
- **AND** creates internal execution work correlated to the returned `runId`

### Requirement: Accepted manual runs become observable through monitor run history
System SHALL make accepted manual runs visible in monitor run history after downstream execution completes.

#### Scenario: Operator checks run history after accepted manual run
- **WHEN** accepted manual execution later finishes successfully or unsuccessfully
- **THEN** operator can observe the resulting persisted run in `GET /api/v1/monitors/{id}/runs`
- **AND** that stored run identifies trigger type as manual
