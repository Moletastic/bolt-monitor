## Why

The root dashboard route is currently a placeholder that points operators to the Services module. It should instead become the first screen operators use to understand whether monitoring is running, what is broken, and what needs attention.

## What Changes

- Replace the dashboard root placeholder with an operational overview backed by existing dashboard APIs.
- Add global health summary cards for service rollup health, open incidents, draft/setup state, and scheduler state.
- Add an attention queue that prioritizes down services, open incidents, services without monitor coverage, disabled monitors, and draft services.
- Add a compact service health matrix that links operators into existing service detail and monitor workflows.
- Add recent incident and setup-gap panels that expose operational work without requiring operators to visit every module.
- Keep monitor CRUD, service detail, incident detail, scheduler configuration, and probe-location management in their existing modules.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `dashboard-web-app`: Replace the permitted work-in-progress dashboard landing page with a required operational overview that summarizes service health, incidents, scheduler state, and setup gaps using existing APIs.

## Impact

- Affects `apps/dashboard/app/page.tsx` and may add small presentational helpers/components under `apps/dashboard/components/`.
- Uses existing dashboard API client functions including `listServices`, `listIncidents`, `getSchedulerConfig`, and optionally `listProbeLocations`.
- Does not require new backend APIs, new persistence, or changes to monitor/service/incident data contracts.
- Dashboard rendering must remain compatible with `MONITOR_API_BASE_URL`/`NEXT_PUBLIC_MONITOR_API_BASE_URL` configuration behavior already used by the app.
