## Requirements

### Requirement: System exposes manual monitor run command through HTTP API
System SHALL allow an operator to request an on-demand run for an existing monitor through HTTP API.

#### Scenario: Operator triggers monitor run
- **WHEN** operator calls `POST /api/v1/services/{serviceId}/monitors/{monitorId}/run` for existing monitor
- **THEN** system executes one check attempt in the system execution environment
- **AND** the response includes a stable run identifier, trigger, timing, duration, outcome, status code when available, and error when available
- **AND** the response does not include probe-location or region identity

### Requirement: Manual run command only targets runnable monitors
System SHALL reject manual run commands for monitors that do not exist or are not runnable.

#### Scenario: Operator triggers run for missing monitor
- **WHEN** operator calls manual run command for unknown monitor ID
- **THEN** system returns not-found response

#### Scenario: Operator triggers run for disabled monitor
- **WHEN** operator calls manual run command for disabled monitor
- **THEN** system rejects request without scheduling execution

### Requirement: Manual run command uses manual trigger semantics
System SHALL mark on-demand monitor execution as a manual trigger distinct from recurring execution.

#### Scenario: Accepted manual run is processed downstream
- **WHEN** system materializes execution work for accepted manual run
- **THEN** resulting execution metadata identifies trigger type as manual

### Requirement: Manual run command remains separate from monitor CRUD
System SHALL expose on-demand execution as a command endpoint rather than as monitor mutation or generic run CRUD.

#### Scenario: Client inspects API shape
- **WHEN** client needs to trigger immediate execution
- **THEN** system provides a dedicated command route at `POST /api/v1/services/{serviceId}/monitors/{monitorId}/run`
- **AND** does not require client to create internal execution records directly
