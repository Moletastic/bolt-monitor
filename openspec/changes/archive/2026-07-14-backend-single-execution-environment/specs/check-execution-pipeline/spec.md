## MODIFIED Requirements

### Requirement: System builds execution work for due monitors
System SHALL create execution work for each enabled monitor that is due to run.

#### Scenario: Enabled monitor is due
- **WHEN** a monitor is enabled and due for execution
- **THEN** system creates one work item for that monitor and run
- **AND** the work item does not include probe-location routing state

### Requirement: System records execution result
System SHALL record each completed execution result with monitor identity, trigger, timing, and outcome.

#### Scenario: Execution completes
- **WHEN** a check attempt completes
- **THEN** system records the result without probe-location or region identity
