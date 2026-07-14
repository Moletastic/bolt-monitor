## Context

The dashboard already exposes monitor detail at `/services/{serviceId}/monitors/{monitorId}`. The page loads service, monitor, latest status, recent runs, monitor incidents, and monitor audit events through existing APIs, then renders a status summary, configuration card, action forms, and `Runs` / `Incidents` / `Audit` tabs.

The current layout is data-rich but not scan-first. It mixes identity, status, cards, configuration, and actions inside the same large status card, while recent run data is only visible after moving below the fold. The refresh should make the page answer operator questions in order: what monitor is this, what can I do, is it healthy, what is it checking, what changed recently, and what evidence backs that up.

## Goals / Non-Goals

**Goals:**

- Reframe monitor detail as an observability surface using existing APIs and dashboard primitives.
- Keep monitor identity and status together in the header, with actions grouped separately at the end of the row.
- Show four health indicators: current state, recent uptime, P99 latency, and error rate.
- Derive recent uptime, error rate, P99 latency, and chart datapoints from existing recent run history.
- Present `Check configuration` as an operator-readable reference panel with endpoint, protocol, frequency, and timeout.
- Preserve the existing `Runs`, `Incidents`, and `Audit` tabs and their current backing APIs.
- Make mobile layout compact by hiding non-primary action text and showing one selected indicator card at a time.

**Non-Goals:**

- Do not capture, persist, or render latest response bodies.
- Do not add new backend endpoints, DynamoDB fields, or response DTO fields.
- Do not change monitor execution, status derivation, incident creation, audit semantics, or manual-run behavior.
- Do not introduce monitor execution location controls.

## Decisions

1. **Use existing recent runs as the metric and chart source.**

   The page already calls `getMonitorRuns(serviceId, monitorId)`, and each run includes `startedAt`, `finishedAt`, `durationMs`, `outcome`, `statusCode`, `error`, and `trigger`. This is enough to compute recent uptime, error rate, P99 latency, and a latency/outcome chart without API expansion.

   Alternative considered: extend status API with precomputed metrics. That would simplify the dashboard but adds backend work and duplicate aggregation logic for a layout-only refresh.

2. **Keep status badge beside the monitor name.**

   The header should have a left identity cluster: monitor name immediately followed by status badge. Actions occupy the right side so the badge reads as part of identity, not as an action or page-level aside.

   Alternative considered: right-align status badge. The user rejected this because the right side belongs to actions.

3. **Use icon-leading action buttons on desktop and compact icon buttons on mobile.**

   Desktop buttons show icon and text for `Run now`, `Edit`, enable/disable, and maintenance. Mobile keeps `Run now` text visible and collapses the other actions to icon-only buttons with accessible labels.

   Alternative considered: stack all full-text buttons on mobile. This consumes too much vertical space before health content.

4. **Use a local indicator tab picker on mobile.**

   Desktop shows all four indicator cards. Mobile shows a tab/pill selector for `State`, `Uptime`, `P99`, and `Errors`, defaulting to `State`, then renders the selected card only. This avoids four stacked KPI cards crowding the viewport.

   The indicator picker should not reuse the route `tab` query parameter because that parameter already controls `Runs`, `Incidents`, and `Audit`. Use component state or another non-conflicting local control.

   Alternative considered: make all four cards horizontally scrollable. That hides important state and makes comparison harder.

5. **Keep the evidence tabs route-addressable.**

   The existing `Runs`, `Incidents`, and `Audit` tabs use link-based tab navigation via query string. Preserve that behavior and add icons to improve scanability.

   Alternative considered: convert all tabs to client state. That would reduce URL shareability and conflict with existing dashboard navigation patterns.

6. **Name the configuration panel `Check configuration`.**

   The section is read-only operational reference, not an edit form. `Check configuration` is more precise than `Configuration` or `Monitor config details` because it describes what the monitor actually executes.

## Risks / Trade-offs

- **Recent-run metrics can look authoritative despite limited samples** → Show sample context such as recent run count and use `No data` states when no runs exist.
- **P99 is noisy with small run counts** → Derive only from available recent successful durations and label it as recent tail latency, not SLO-grade latency.
- **Chart tooltip on mobile can be hard to use** → Support tap/focus interaction in addition to hover and keep tooltip content concise.
- **Icon-only mobile actions can be inaccessible** → Provide accessible names for every icon-only button and preserve visible text for the primary `Run now` action.
- **Two tab systems can confuse users** → Keep indicator picker visually scoped to the KPI card and keep evidence tabs below chart/config with `Runs`, `Incidents`, and `Audit` labels plus icons.
