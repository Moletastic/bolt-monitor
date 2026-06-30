## MODIFIED Requirements

### Requirement: Scheduler creates execution requests for enabled monitors
System SHALL build one ExecutionRequest for each enabled monitor that is due to run.

#### Scenario: Execution requests built
- **WHEN** scheduler has a list of enabled monitors
- **THEN** one ExecutionRequest is created for each monitor
- **AND** trigger type is set to `recurring`
- **AND** no additional request fan-out occurs by probe location or region

## REMOVED Requirements

### Requirement: Scheduler creates execution requests for each monitor-probe combination
System SHALL build an ExecutionRequest for each enabled monitor and each probe location assigned to that monitor.

#### Scenario: Execution requests built
- **WHEN** scheduler has list of enabled monitors
- **THEN** for each monitor, for each probeLocation in monitor.ProbeLocations, an ExecutionRequest is created
- **AND** trigger type is set to "recurring"
