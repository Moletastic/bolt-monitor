## Why

The monitor detail page has the right raw data, but the layout does not make the operator's first questions easy to answer: what is this monitor, is it healthy, what is it checking, and what happened recently? This change refreshes the monitor detail page into an observability-focused view using existing monitor, status, run, incident, and audit data.

## What Changes

- Refresh the monitor detail header so monitor name and status badge form the left identity cluster while monitor actions sit in a right-aligned action cluster.
- Present four monitor indicator cards for current state, recent uptime, P99 latency, and error rate, each with a contextual icon.
- Add a monitor performance chart for recent run latency and outcome datapoints with operator-readable tooltips.
- Add a `Check configuration` panel summarizing endpoint, protocol, frequency, and timeout beside the chart on desktop and before the chart on mobile.
- Preserve the existing `Runs`, `Incidents`, and `Audit` tabs, adding tab icons and keeping each tab focused on its corresponding table.
- Improve the mobile layout with compact action buttons, a single selected indicator card controlled by indicator tabs, and card-style table rows where needed.
- Do not add latest-response capture or response body persistence in this change.

## Capabilities

### New Capabilities

- None.

### Modified Capabilities

- `dashboard-web-app`: refine monitor detail layout, indicators, chart, configuration summary, responsive behavior, and tab presentation.

## Impact

- Affected code: `apps/dashboard` monitor detail route, monitor detail components, chart/presentation helpers, responsive styles, and related tests.
- APIs: no new API endpoints and no response body persistence; the change uses existing monitor read, status, run history, incident, audit, and manual-run APIs.
- Dependencies: no new runtime dependency expected; icons should come from the dashboard's existing icon set.
- Systems: no backend, infra, or DynamoDB schema changes expected.
