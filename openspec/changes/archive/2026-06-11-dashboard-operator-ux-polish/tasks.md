## 1. Services View Polish

- [x] 1.1 Make each service card in `/services` navigate to service detail from the full non-interactive card area.
- [x] 1.2 Preserve accessible link semantics and avoid nested interactive click conflicts.
- [x] 1.3 Remove raw service IDs from primary service card display.
- [x] 1.4 Normalize service technology icon container and icon size for overview tiles, service cards, and detail headers.

## 2. Service Detail Summary

- [x] 2.1 Replace service detail summary header ID label with service name, description, technology, lifecycle, and rollup status.
- [x] 2.2 Add monitoring-oriented summary metrics: total monitors, enabled monitor coverage, lifecycle, technology, and last update.
- [x] 2.3 Surface setup signals for no monitors, disabled monitor coverage, draft state, and down rollup status.
- [x] 2.4 Keep create-monitor action easy to find.

## 3. Monitor Overview And Detail

- [x] 3.1 Add a dedicated Protocol column to desktop monitor overview.
- [x] 3.2 Keep protocol visible on mobile monitor cards.
- [x] 3.3 Remove raw monitor IDs from primary monitor overview display.
- [x] 3.4 Improve monitor detail current status with status, last outcome, latest error, target, protocol, enabled state, last check, duration, probe, and cadence.
- [x] 3.5 Keep monitor configuration editing separate from current status triage content.

## 4. Integrations

- [x] 4.1 Load notification channels automatically when `/integrations` opens.
- [x] 4.2 Show explicit loading, empty, error, and configured-channel states.
- [x] 4.3 Preserve manual refresh after initial load.
- [x] 4.4 Avoid alert dialogs for success feedback if repository toast/status patterns can be reused.

## 5. Incidents

- [x] 5.1 Make `/incidents` useful when empty by explaining selected filter state and incident lifecycle.
- [x] 5.2 Avoid showing raw monitor IDs as primary incident labels.
- [x] 5.3 Keep incident rows actionable with links to incident detail or monitor detail where available.
- [x] 5.4 Handle incident API failures with an unavailable state instead of failing the whole page.

## 6. Settings

- [x] 6.1 Replace `/config` placeholder with a settings overview.
- [x] 6.2 Show scheduler recurring execution state and route operators to scheduler controls.
- [x] 6.3 Show probe location catalog summary from existing API data when available.
- [x] 6.4 Show safe environment/setup context without exposing secrets.
- [x] 6.5 Provide useful unavailable states for scheduler or probe API failures.

## 7. Verification

- [x] 7.1 Run `make lint-dashboard`.
- [x] 7.2 Run `make check-dashboard`.
- [x] 7.3 Review responsive layouts for `/services`, service detail, monitor detail, `/integrations`, `/incidents`, and `/config`.
- [x] 7.4 Verify no primary operator surface displays service IDs or monitor IDs except deliberate low-emphasis debug affordances.
