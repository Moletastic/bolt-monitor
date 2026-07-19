## Purpose

Define the on-demand manual run command as a retry-safe, idempotent service-scoped request that shares the canonical result commit without affecting recurring health projections.
## Requirements
### Requirement: System exposes manual monitor run command through HTTP API
System SHALL allow an operator to request an on-demand run for an existing monitor through the service-scoped HTTP API using the same durable result contract as recurring execution and a required `Idempotency-Key` header.

#### Scenario: Operator triggers monitor run
- **WHEN** operator calls `POST /api/v1/services/{serviceId}/monitors/{monitorId}/run` with a valid `Idempotency-Key` for an existing runnable monitor
- **THEN** system deterministically maps tenant, service, monitor, and idempotency key to one idempotency-record address, stores one stable `runId`, and sets `trigger=manual` before persistence or HTTP execution
- **AND** conditionally creates and claims manual work before the HTTP side effect
- **AND** commits one canonical `CheckRun` and terminal work through the shared result transaction
- **AND** the response includes the stable run identifier, trigger, timing, duration, outcome, status code when available, and error when available
- **AND** the response does not include probe-location or region identity

#### Scenario: Idempotency key is missing or invalid
- **WHEN** operator omits `Idempotency-Key` or supplies a key outside configured syntax/length bounds
- **THEN** system rejects the command through the standard validation error envelope
- **AND** performs no durable or HTTP side effect

### Requirement: Manual run command only targets runnable monitors
System SHALL reject manual run commands for monitors that do not exist or are not runnable.

#### Scenario: Operator triggers run for missing monitor
- **WHEN** operator calls manual run command for unknown monitor ID
- **THEN** system returns not-found response

#### Scenario: Operator triggers run for disabled monitor
- **WHEN** operator calls manual run command for disabled monitor
- **THEN** system rejects request without scheduling execution

### Requirement: Manual run command uses manual trigger semantics
System SHALL mark on-demand monitor execution as a manual trigger distinct from recurring execution and SHALL keep its effect on recurring health projections explicit.

#### Scenario: Accepted manual run is processed downstream
- **WHEN** system materializes and completes execution work for an accepted manual run
- **THEN** work, result, response, and `CheckRun` identify trigger type as manual and share one `runId`
- **AND** no `scheduleDefinitionVersion` or `scheduledFor` is assigned

#### Scenario: Manual diagnostic succeeds or fails
- **WHEN** a manual result is committed
- **THEN** it remains visible in raw run history
- **AND** it does not advance recurring counters or status cursor, rewrite monitor status, open or resolve an incident, create incident activity/outbox, or change service rollup

### Requirement: Manual run command remains separate from monitor CRUD
System SHALL expose on-demand execution as a command endpoint rather than as monitor mutation or generic run CRUD.

#### Scenario: Client inspects API shape
- **WHEN** client needs to trigger immediate execution
- **THEN** system provides a dedicated command route at `POST /api/v1/services/{serviceId}/monitors/{monitorId}/run`
- **AND** does not require client to create internal execution records directly

### Requirement: Manual request and result retries are idempotent
System SHALL canonicalize the service-scoped command into a deterministic request fingerprint and retain a bounded idempotency record so repeated use of the same key and request converges on one manual `runId`, terminal work record, canonical `CheckRun`, and response.

#### Scenario: Same request is replayed
- **WHEN** the same scoped `Idempotency-Key` and request fingerprint are received within `MANUAL_IDEMPOTENCY_RETENTION`
- **THEN** system resumes the same in-progress run or returns the same canonical completed response
- **AND** does not execute another completed run or create recurring projection effects

#### Scenario: Same key is reused for a different request
- **WHEN** the same scoped `Idempotency-Key` is received with a different canonical request fingerprint
- **THEN** system returns an existing public conflict error in the standard envelope
- **AND** does not mutate or execute either request

#### Scenario: Idempotency retention is configured
- **WHEN** a manual idempotency record is created
- **THEN** it stores fingerprint, `runId`, replay/result reference, and TTL only
- **AND** `MANUAL_IDEMPOTENCY_RETENTION` is bounded, documented, and tested

### Requirement: Manual execution failures are typed
System SHALL classify durable-work, lease, execution, and result-commit failures for manual runs using the shared typed runtime failure model while preserving existing public response-envelope conventions.

#### Scenario: Manual durable operation fails
- **WHEN** work creation, claim, or result commit fails
- **THEN** the API maps the typed internal failure to an existing appropriate public error code
- **AND** safe details include the operation and stable `runId` when available

