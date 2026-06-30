## Context

The dashboard root route currently renders a module landing placeholder inside the shared `AppShell`, while real operational workflows live under `/services`, `/incidents`, `/integrations`, and `/config`. Existing dashboard APIs already expose enough aggregate inputs for a useful root overview: services include lifecycle, rollup status, monitor counts, and nested monitors when fetched by service; incidents can be listed globally; scheduler configuration can be read; probe locations can be listed.

The dashboard should become an operator entry point without changing backend contracts. The first version should favor dense, actionable summaries over charts that require time-series aggregation the API does not yet provide.

## Goals / Non-Goals

**Goals:**

- Turn `/` into an operational overview that answers what is broken, what changed, and what needs setup attention.
- Use existing APIs and existing module routes for drill-down actions.
- Preserve the current monitoring-console visual language and shared shell.
- Keep the first implementation fast enough for server rendering by avoiding unnecessary per-monitor fan-out.

**Non-Goals:**

- Add new backend summary endpoints or persistence models.
- Add uptime percentages, latency charts, or historical trend graphs.
- Add real-time streaming updates.
- Move service, monitor, incident, scheduler, or probe-location management out of their current modules.
- Implement a multi-probe map while the default catalog is still effectively a single built-in location.

## Decisions

### Use existing aggregate APIs first

The root dashboard should fetch `listServices`, `listIncidents`, and `getSchedulerConfig` as the primary data sources. `listProbeLocations` can support a small setup/context card, but should not drive a map or region health view yet.

Alternative considered: add a backend `/dashboard-summary` endpoint. That would reduce frontend composition work, but it is unnecessary for the first version and would expand scope beyond the current need.

### Prioritize an attention queue over charts

The most valuable root-dashboard component is a ranked attention queue with items such as down services, open incidents, services without monitors, disabled monitor coverage, draft services, and disabled scheduler state.

Alternative considered: build dashboard charts first. Charts would look polished but risk being misleading because the available APIs do not expose aggregate historical time-series data.

### Keep service health compact and navigational

The dashboard should show a compact service health matrix with status, lifecycle, monitor coverage, updated time, and links to service details. The fuller card grid remains on `/services`.

Alternative considered: reuse the full service card layout on root. That duplicates `/services` and makes the dashboard less useful as a fast triage surface.

### Avoid global per-monitor detail fan-out initially

Global monitor-level status should come from data already present in `listServices` or each service summary. If disabled-monitor counts require nested monitor data, the implementation should either compute only from available service payloads or keep that item best-effort until an aggregate API exists.

Alternative considered: fetch every service detail and every nested monitor on the root page. That may be acceptable with tiny data but creates avoidable latency and failure coupling.

## Risks / Trade-offs

- Root page depends on multiple API calls -> render partial fallback sections where possible instead of failing the entire dashboard when non-critical context is unavailable.
- Existing service list may not include enough monitor-level detail for every setup-gap metric -> prefer accurate service-level metrics over guessed monitor-level counts.
- Dense dashboard can become noisy -> rank attention items and cap visible rows with links into the appropriate module.
- Scheduler disabled state may be more operationally important than service drafts -> attention ordering should treat scheduler disabled as high priority.

## Migration Plan

Deploy as a dashboard-only change. The root route can be replaced in one release because no persisted data or API contracts change. Rollback is a simple revert to the existing placeholder route if the new overview causes rendering problems.

## Open Questions

- Should acknowledged-but-unresolved incidents appear in the high-priority attention queue or only in the incidents panel?
- Should draft services be counted as setup gaps if they intentionally represent planned future coverage?
- Should the dashboard root use `NEXT_PUBLIC_MONITOR_API_BASE_URL` consistently with the current API client, or should the repo standardize on the documented `MONITOR_API_BASE_URL` separately?
