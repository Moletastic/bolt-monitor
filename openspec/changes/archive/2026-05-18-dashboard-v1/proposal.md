## Why

Project now has enough backend surface to support a first real operator UI: monitor list reads include current status summary, monitor detail can read latest status and recent runs, and monitor configuration can be created, edited, enabled, and disabled through HTTP API. A dashboard v1 should land now to turn those APIs into a usable product surface and validate the frontend architecture before broader alerting, incident, and fleet-aggregation features exist.

## What Changes

- Add a new `apps/dashboard` Next.js + TypeScript application for operator workflows.
- Implement dashboard v1 around the existing monitor API surface: list monitors, view monitor detail, create monitor, edit monitor, and enable or disable monitors.
- Translate `DESIGN.md` tokens and the visual direction in `code.html` into reusable dashboard UI primitives using shadcn/ui.
- Keep v1 scoped to monitor operations and recent run history rather than inventing placeholder backend domains for alerts, incidents, logs, or scheduled maintenance.
- Prefer server-side or same-origin integration patterns so the frontend is not blocked on cross-origin browser concerns during bootstrap.

## Capabilities

### New Capabilities
- `dashboard-web-app`: Operator-facing dashboard for monitor management and operational status inspection.

### Modified Capabilities
- `monitor-crud-api`: Frontend consumption will pressure API response stability and error handling for real user flows.
- `monitor-status-read-api`: Frontend consumption will establish the first concrete dashboard read patterns for status summaries and recent runs.

## Non-goals

- Do not implement fleet-wide aggregate metrics such as global uptime, active alerts count, or load average unless backed by real APIs.
- Do not add alerts, incident management, logs, authentication, or multi-tenant workspace UX in this change.
- Do not reproduce every panel in `code.html` if the corresponding backend capability does not exist yet.

## Impact

- Adds the repository's first frontend app under `apps/`.
- Establishes UI conventions, token mapping, and data-fetching patterns for future product surfaces.
- Depends on existing monitor CRUD and monitor status read endpoints being the source of truth for dashboard data.
- May reveal small follow-on API gaps such as probe-location discovery or frontend-friendly error conventions, but those should be handled as explicit scope decisions rather than silently expanded into v1.
