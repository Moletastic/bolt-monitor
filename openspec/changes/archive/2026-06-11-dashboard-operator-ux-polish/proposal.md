## Why

The dashboard has enough core monitoring flows to be useful, but several operator-facing views still expose implementation details, placeholder pages, or low-signal summaries. This makes the product feel less like a monitoring console and more like an API wrapper.

Operators need clickable service surfaces, readable technology identity, monitor tables that expose protocol, structured service and monitor status summaries, notification channels that load on arrival, and non-empty incidents/settings modules.

## What Changes

- Make service cards in the Services view fully clickable while preserving nested actions and keyboard accessibility.
- Increase and normalize technology icon presentation so service technology is distinguishable at card and summary sizes.
- Replace raw service/monitor IDs in primary UI with human-readable names, status, protocol, target, lifecycle, and timing context.
- Improve Service detail summary into a monitoring-oriented operational snapshot with rollup status, lifecycle, coverage, enabled monitors, technology, and last update.
- Add protocol/type context to monitor overview tables and cards.
- Improve Monitor detail current status to show actionable status, last outcome, last check, duration, probe, target, interval, enabled state, and latest error when present.
- Preload notification channels when the Integrations page opens and present clear loading, empty, error, and configured states.
- Make Incidents and Settings modules useful landing pages instead of empty or placeholder experiences.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `dashboard-web-app`: Expand dashboard UI requirements for operational polish across services, monitor detail, integrations, incidents, and settings.

## Impact

- Affects dashboard app pages/components only unless incident views need richer human-readable incident references than current APIs provide.
- Likely touches `apps/dashboard/app/(monitoring)/services/page.tsx`, `apps/dashboard/app/(monitoring)/services/[serviceId]/page.tsx`, `apps/dashboard/components/monitor-table.tsx`, `apps/dashboard/app/(monitoring)/services/[serviceId]/monitors/[monitorId]/page.tsx`, `apps/dashboard/app/integrations/page.tsx`, `apps/dashboard/app/(monitoring)/incidents/page.tsx`, `apps/dashboard/app/config/page.tsx`, and shared presentational components.
- Uses existing APIs first: services, monitor status, monitor runs, incidents, scheduler config, probe locations, notification channels.
- Does not add synthetic incidents or fake settings. Empty states must explain real system state and next actions.
