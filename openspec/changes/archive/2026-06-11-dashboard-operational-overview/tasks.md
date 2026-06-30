## 1. Data Composition

- [x] 1.1 Update the dashboard root page to fetch service, incident, and scheduler overview data through existing API client functions.
- [x] 1.2 Derive summary counts for down services, open incidents, draft services, services without monitors, and scheduler enabled/disabled state.
- [x] 1.3 Derive an ordered attention queue that prioritizes disabled scheduler state, down services, open incidents, services without monitors, disabled monitor coverage when available, and draft services.
- [x] 1.4 Handle empty service data separately from healthy zero-count summaries.

## 2. Dashboard UI

- [x] 2.1 Replace the placeholder dashboard root content with operational summary cards inside the existing `AppShell`.
- [x] 2.2 Add a prioritized attention area with links to the relevant service, incident, or scheduler routes.
- [x] 2.3 Add a compact service health matrix with service identity, rollup status, lifecycle state, monitor coverage, updated context, and service-detail links.
- [x] 2.4 Add recent incident and setup-gap panels using existing dashboard visual patterns.
- [x] 2.5 Preserve responsive behavior for desktop and mobile layouts.

## 3. Error And Empty States

- [x] 3.1 Show an actionable empty state with a create-service path when no services exist.
- [x] 3.2 Show useful unavailable or partial-state messaging when overview API data cannot be loaded.
- [x] 3.3 Avoid rendering misleading health summaries when required source data is unavailable.

## 4. Verification

- [x] 4.1 Run `make lint-dashboard`.
- [x] 4.2 Run `make check-dashboard`.
- [x] 4.3 Manually review the dashboard root with representative states: no services, healthy services, down services, open incidents, and disabled scheduler.
