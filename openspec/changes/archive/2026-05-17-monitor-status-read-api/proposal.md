## Why

Once results and latest status exist, system needs read endpoints that expose them cleanly. Status read API should land before full frontend app so backend can prove useful monitor detail, recent history, and dashboard summary reads.

## What Changes

- Add read-focused monitor endpoints for latest status and recent run history.
- Define dashboard-friendly list shape that includes current status summary.
- Reuse persisted `MonitorStatus` and `CheckRun` models instead of recomputing state on demand.

## Capabilities

### New Capabilities
- `monitor-status-read-api`: HTTP API for latest monitor status, recent run history, and dashboard-oriented monitor reads.

### Modified Capabilities
- `monitor-crud-api`: Extend monitor read surface with operational status data.

## Impact

- Affects API routing, repository reads, and response modeling.
- Depends on result/status persistence being available.
- Gives frontend and operator workflows useful monitoring data before incidents/auth ship.
