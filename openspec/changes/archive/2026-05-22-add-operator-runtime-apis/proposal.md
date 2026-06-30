## Why

Current API covers monitor CRUD, probe-location reads, latest status, and recent run history, but it still lacks operator-facing runtime surfaces needed to turn monitor configuration into an actual monitoring product. Operators cannot trigger a run on demand, inspect incidents, control recurring scheduling, or inspect audit history through explicit API contracts.

## What Changes

- Add a manual-run command API for triggering a monitor execution on demand.
- Add an incident API for reading incident state and acknowledging or resolving incidents through command endpoints.
- Add an admin scheduler configuration API for reading and changing recurring execution control state.
- Add an audit-event read API for inspecting monitor mutation history.

## Capabilities

### New Capabilities
- `manual-run-api`: Operator-triggered monitor execution command surface.
- `incident-management-api`: Incident read and operator action API.
- `scheduler-admin-api`: Admin control surface for recurring execution settings.
- `audit-event-read-api`: Read API for monitor audit history.

### Modified Capabilities

## Impact

- Expands public API shape in `infra/` and `services/monitor-api`.
- Introduces new operator-facing models that should be exposed directly through API contracts instead of only through internal business process state.
- Clarifies boundary between public operator/admin resources and internal execution pipeline artifacts.
