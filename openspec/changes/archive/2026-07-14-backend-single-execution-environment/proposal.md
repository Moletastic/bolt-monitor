## Why

The product contract no longer includes multi-regional or selectable probe-location health checks. Backend and runtime code should stop carrying the unused location model so monitor configuration, scheduling, execution, status, and API responses match the single-execution-environment product shape.

Because there is no production environment or data compatibility requirement, the implementation can remove fields and endpoint behavior directly rather than adding deprecation shims.

## What Changes

- Remove `probeLocations` from monitor create/update/read contracts and persistence models.
- Remove the probe-location catalog package and default `iad` catalog wiring.
- Remove `GET /api/v1/probe-locations` from the monitor API and infrastructure route wiring.
- Remove `probeLocationId` and `lastProbeLocationId` from execution results, run responses, manual run responses, and status responses.
- Simplify recurring execution so one enabled monitor creates one work item/check attempt.
- Simplify manual execution so it runs the monitor directly without selecting `monitor.ProbeLocations[0]`.

## Capabilities

### Modified Capabilities

- `monitor-crud-api`: Monitor payloads and responses no longer include probe-location fields.
- `check-execution-pipeline`: Execution requests/results no longer carry probe-location routing state.
- `manual-run-api`: Manual runs execute in the system environment and do not return probe location identifiers.
- `monitor-status-read-api`: Status responses no longer expose last probe location identifiers.

## Impact

- `shared/monitorconfig`, `shared/checkexecution`, `shared/resultstatus`, `shared/dynamodbrecord`
- `services/check-runtime`
- `services/monitor-api`
- `infra/stacks/bootstrap.ts`
- Go tests covering monitor validation, scheduling, manual runs, status, incidents, and repository records
