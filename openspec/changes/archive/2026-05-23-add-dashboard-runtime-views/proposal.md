## Why

Runtime execution pipeline is built but the dashboard cannot yet show incident state, audit history, or manual run controls. UI must be ready to observe runtime behavior before or at the same time the runtime goes live — otherwise there is no way for operators to see results, trigger on-demand checks, or act on incidents.

## What Changes

- Add manual run trigger button to monitor detail so operators can fire a check on demand.
- Add incidents tab to monitor detail so operators can see open and resolved incidents per monitor.
- Add audit history tab to monitor detail so operators can see mutation history per monitor.
- Add incidents list view so operators can see all open and closed incidents across monitors.
- Add incident acknowledge and resolve action buttons so operators can close incidents from the UI.
- Add probe locations view so operators can see available probe regions.
- Add scheduler admin view so operators can read and toggle recurring execution state.

## Capabilities

### New Capabilities

- `runtime-observation-views`: Dashboard surfaces for observing and acting on runtime execution state — incidents, audit history, manual run trigger, and scheduler control.

### Modified Capabilities

- `dashboard-web-app`: Extend monitor detail view to include incident and audit tabs, and add manual run trigger action. Add full incidents overview page, scheduler admin page, and probe locations page.

## Impact

- Affects `apps/dashboard/` — new pages, components, and API client functions.
- No backend API changes required — all endpoints already exist.
- No changes to `services/` or `shared/` — pure UI work.
