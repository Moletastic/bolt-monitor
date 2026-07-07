## Context

The monitor API already stores raw `CheckRun` records with `startedAt`, `finishedAt`, `durationMs`, `outcome`, status code, error, monitor identity, and service identity. Raw runs are retained for 30 days, but the public monitor run endpoint currently returns only the latest 20 runs for a single monitor.

The Services overview currently renders one card per service using `listServices()` data: service identity, lifecycle, rollup status, technology, monitor coverage, update time, and current monitor traffic-light dots. It does not receive historical runs or aggregate metrics.

The MVP target is a visual service-card treatment like the example in `tmp/service-cards.png`: status chip, monitor-up coverage, health bars, recent average latency, aggregate P99, recent uptime percentage, sparkline, and honest no-data/no-monitor states. The example's icons are not a replacement source; dashboard cards must reuse the existing service technology icon system already used by the Services module.

## Goals / Non-Goals

**Goals:**

- Compute service-card metrics from recent run samples already stored as `CheckRun` records.
- Avoid dashboard fan-out where the service list calls every monitor's `/runs` endpoint.
- Keep the metric semantics honest: recent observed samples, not SLO-grade historical uptime.
- Support healthy, degraded, down, draft, and no-monitor cards with clear empty states.
- Reuse existing dashboard service technology icons instead of introducing a new icon set for the cards.
- Keep the API response suitable for dashboard rendering without leaking storage implementation details.

**Non-Goals:**

- No new long-term rollup table or time-series store for this MVP.
- No configurable 24h/7d/30d analytics windows.
- No persisted SLO calculations, error budgets, or time-weighted availability.
- No change to check execution, status derivation, incident behavior, or run retention.

## Decisions

### Decision 1: Backend Computes Service-Card Metrics

The monitor API will compute recent service-card metrics server-side and include them in service list/detail responses or expose a dashboard-oriented service-card endpoint.

Rationale: The service list page needs multiple service cards at once. If the dashboard fans out to every monitor run endpoint, page latency and request count grow quickly with service count. Server-side aggregation keeps the dashboard contract simple and lets the backend control limits.

Alternative considered: Dashboard calls `listMonitorRuns()` per monitor. This is acceptable for a spike but poor as the product path because it creates N+1 requests and duplicates metric semantics in the UI.

### Decision 2: Use Recent Sample Semantics

Metrics are derived from the latest bounded set of run samples per monitor. The initial bound should match the existing monitor history contract unless implementation discovers a safe backend-only limit is needed.

- Average latency: arithmetic mean of successful run `durationMs` values.
- Aggregate P99: nearest-rank percentile over successful run `durationMs` values.
- Recent uptime: successful run count divided by total returned run count.
- Sparkline: recent aggregate run points ordered by time, using latency for successful runs and a failure marker for unsuccessful runs.
- Monitor-up coverage: current UP monitor count divided by total configured monitors, using current persisted monitor status.

Rationale: This matches available data and keeps the card useful without pretending to be a complete availability report.

### Decision 3: Treat Empty and Partial Data as First-Class States

Cards must distinguish:

- Draft/no monitors: no configured monitoring; show setup-oriented empty state.
- Monitors exist but no runs: waiting for data; show placeholders.
- Down service: show failed state and recency context using current status timestamps where available.
- Partial monitor data: calculate from available samples and avoid fabricating values for missing monitors.

Rationale: Monitoring UI loses trust quickly if it shows `0%` or `N/A` without differentiating no coverage from real failure.

## Risks / Trade-offs

- **Small sample P99 can be noisy** -> Label metrics as recent and prefer bounded-sample language in UI copy.
- **Server-side aggregation can increase DynamoDB reads** -> Limit samples per monitor and avoid dashboard N+1 requests.
- **Service list could become slower with many services** -> Keep MVP bounded and consider follow-up rollup persistence only if needed.
- **Uptime may be mistaken for SLO uptime** -> Use `Recent uptime` or helper copy instead of implying 24h/30d availability.
- **Failed runs have duration values but may represent timeout/error behavior** -> Use success-only latency metrics for average/P99, while uptime uses all outcomes.

## Migration Plan

1. Add response types and aggregation logic behind existing monitor API patterns.
2. Add dashboard type parsing/rendering for metric-rich service cards.
3. Deploy without storage migration because metrics derive from existing raw run records.
4. If the API cannot compute metrics, dashboard should still render existing service cards with placeholders for unavailable metrics.
