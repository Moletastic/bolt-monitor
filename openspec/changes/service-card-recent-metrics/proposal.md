## Why

The Services overview currently shows service identity, rollup status, and monitor coverage, but it does not communicate recent operational behavior like latency, P99, uptime, or trend shape. Operators need a compact service-card view that summarizes recent health from existing check-run data without waiting for a full historical analytics/SLO system.

## What Changes

- Add MVP service-card recent metrics derived from persisted `CheckRun` history.
- Expose dashboard-oriented per-service metric summaries that aggregate recent samples across each service's monitors.
- Update Services overview cards to show recent average latency, aggregate P99, recent uptime, monitor-up coverage, and a compact sparkline/empty-state treatment.
- Keep the metrics explicitly recent-sample based, not long-window or SLO-grade historical analytics.
- Preserve draft/no-monitor states so cards do not imply monitoring coverage where none exists.

## Capabilities

### New Capabilities

- `service-card-recent-metrics`: Defines the recent-sample service metrics contract used by dashboard service cards.

### Modified Capabilities

- `dashboard-web-app`: Services overview cards include recent metrics and trend context when available.

## Impact

- **API**: `services/monitor-api` should expose or embed recent service-card metrics without requiring dashboard N+1 fan-out across every service and monitor.
- **Storage**: No new persistence is required for MVP; metrics are computed from existing `CheckRun` records retained for raw run history.
- **Dashboard**: `apps/dashboard` service list card rendering and types need metric fields and unavailable/empty-state handling.
- **Tests**: Add backend unit coverage for metric aggregation and dashboard tests for healthy, degraded/down, draft, and no-monitor card states.
