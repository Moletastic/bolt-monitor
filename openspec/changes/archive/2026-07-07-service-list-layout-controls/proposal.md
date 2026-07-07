## Why

The Services landing page currently spends its highest-value area on a “Service operations” card that duplicates lower-fidelity service information already available in the service card list. Operators need a clearer service-list hierarchy with health indicators, search/filter controls, and creation access arranged for both desktop scanning and mobile operation.

## What Changes

- Replace the duplicated “Service operations” card with a visible desktop page header and concise service-list description.
- Present `Active`, `Drafts`, and `Down now` as equal summary indicators rather than mixed-size cards.
- Add a service-list controls row above the service cards with search and filter affordances.
- Keep the desktop `Create service` action in the controls row.
- Render the mobile layout as health indicators first, controls second, one service card per row, and a floating bottom-right create-service action.
- Update loading states to match the final service-list layout.

## Capabilities

### New Capabilities

- None.

### Modified Capabilities

- `dashboard-web-app`: Update Services module requirements for service-list header, health indicators, list controls, responsive layout, and create-service placement.

## Impact

- Affected app surface: `apps/dashboard/app/(monitoring)/services/page.tsx`.
- Affected loading surface: `apps/dashboard/app/(monitoring)/services/loading.tsx`.
- Potential dashboard component impact if shared stat-card, search/filter, or floating-action patterns are extracted.
- No backend API changes are expected.
