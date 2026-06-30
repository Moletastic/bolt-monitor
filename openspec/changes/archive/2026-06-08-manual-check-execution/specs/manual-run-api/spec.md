## MODIFIED Requirements

### Requirement: System exposes manual monitor run command through HTTP API
System SHALL allow an operator to request an on-demand run for an existing monitor through HTTP API. The command SHALL execute the check synchronously and return real-time results.

#### Scenario: Operator triggers monitor run
- **WHEN** operator calls `POST /api/v1/services/{serviceId}/monitors/{monitorId}/run` for existing monitor
- **THEN** system executes the HTTP check against the monitor's configured target
- **AND** returns execution result immediately in the response body

### Requirement: Manual run command only targets runnable monitors
System SHALL reject manual run commands for monitors that do not exist or are not runnable.

#### Scenario: Operator triggers run for missing monitor
- **WHEN** operator calls manual run command for unknown monitor ID
- **THEN** system returns not-found response

#### Scenario: Operator triggers run for disabled monitor
- **WHEN** operator calls manual run command for disabled monitor
- **THEN** system rejects request with conflict response without executing

### Requirement: Manual run command uses manual trigger semantics
System SHALL mark on-demand monitor execution as a manual trigger distinct from recurring execution.

#### Scenario: Manual run is recorded with manual trigger
- **WHEN** system records the execution result
- **THEN** resulting execution metadata identifies trigger type as manual

### Requirement: Manual run command remains separate from monitor CRUD
System SHALL expose on-demand execution as a command endpoint rather than as monitor mutation or generic run CRUD.

#### Scenario: Client inspects API shape
- **WHEN** client needs to trigger immediate execution
- **THEN** system provides a dedicated command route at `POST /api/v1/services/{serviceId}/monitors/{monitorId}/run`
- **AND** does not require client to create internal execution records directly

### Requirement: Manual run response includes execution outcome
System SHALL return execution outcome and details in the response body upon completion.

#### Scenario: Operator receives execution result
- **WHEN** operator calls manual run command
- **THEN** response includes `outcome` field with value "success", "failure", "timeout", or "error"
- **AND** response includes `durationMs` with execution time
- **AND** response includes `startedAt` and `finishedAt` timestamps