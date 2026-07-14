## 1. Monitor Detail Data And Metrics

- [x] 1.1 Add dashboard helper logic to derive current state, recent uptime, P99 latency, and error rate from existing monitor status and recent run data.
- [x] 1.2 Add no-data handling for indicators when status or recent runs are unavailable.
- [x] 1.3 Add chart datapoint mapping from recent runs with run time, duration, outcome, status code, and error context.

## 2. Desktop Layout

- [x] 2.1 Refactor monitor detail header into left identity cluster with monitor name plus adjacent status badge and right action cluster.
- [x] 2.2 Add icon-leading desktop actions for run now, edit, enable or disable, and maintenance, with run now styled as primary outlined.
- [x] 2.3 Render four desktop indicator cards with top-left icons for current state, recent uptime, P99 latency, and error rate.
- [x] 2.4 Add recent performance chart with datapoints and tooltips.
- [x] 2.5 Replace the current configuration card with `Check configuration` showing endpoint, protocol, frequency, and timeout.
- [x] 2.6 Place recent performance chart and `Check configuration` in the same desktop row.

## 3. Mobile Layout

- [x] 3.1 Render mobile header with monitor name plus status badge above compact action controls.
- [x] 3.2 Keep `Run now` text visible on mobile and collapse edit, enable or disable, and maintenance controls to accessible icon-only buttons.
- [x] 3.3 Add mobile indicator picker for state, uptime, P99, and errors, defaulting to current state.
- [x] 3.4 Render only the selected indicator card on mobile.
- [x] 3.5 Order mobile content as actions, indicator picker/card, `Check configuration`, recent performance chart, then evidence tabs.

## 4. Evidence Tabs And Tables

- [x] 4.1 Add purpose-matching icons to `Runs`, `Incidents`, and `Audit` tabs while preserving link-based tab navigation.
- [x] 4.2 Preserve each tab's current table, empty state, and unavailable-state behavior.
- [x] 4.3 Ensure mobile table presentation remains readable and tappable for runs, incidents, and audit entries.

## 5. Verification

- [x] 5.1 Update or add dashboard tests for derived monitor indicators and no-data states.
- [x] 5.2 Update or add dashboard tests for monitor detail action accessibility and tab labels.
- [x] 5.3 Run `make lint-dashboard`.
- [x] 5.4 Run `make check-dashboard`.
- [x] 5.5 Run `make test-dashboard`.

## 6. Action Menu Refinement

- [x] 6.1 Align monitor header action buttons in a stable right-side cluster.
- [x] 6.2 Add a vertical-more monitor detail actions menu beside Run now and Edit.
- [x] 6.3 Move maintenance, enable or disable, and delete monitor actions into the vertical-more menu.
- [x] 6.4 Add a delete monitor confirmation dialog that requires typing the monitor name.
- [x] 6.5 Remove the separate delete monitor section from the monitor detail page.
- [x] 6.6 Update tests and rerun dashboard verification commands.

## 7. Header Metadata And Chart Refinement

- [x] 7.1 Move monitor status to a dot-only indicator before the monitor name.
- [x] 7.2 Move protocol badge to the request summary line before HTTP method and endpoint.
- [x] 7.3 Move frequency and timeout into compact metadata badges under the request summary.
- [x] 7.4 Remove the separate `Check configuration` section.
- [x] 7.5 Expand the recent performance chart into a full-width run timeline with summary context.
- [x] 7.6 Update tests and rerun dashboard verification commands.

## 8. Chart Polish

- [x] 8.1 Normalize monitor status before selecting dot color.
- [x] 8.2 Replace native SVG title tooltip with dashboard tooltip UI.
- [x] 8.3 Thin chart stroke and datapoints.
- [x] 8.4 Replace awkward chart endpoint labels and tag-like stat badges with quieter legend/summary text.
- [x] 8.5 Update tests and rerun dashboard verification commands.

## 9. Edit Route And Chart Interaction

- [x] 9.1 Add nested monitor edit route that loads service and monitor context.
- [x] 9.2 Enable detail header edit action and link it to the monitor edit route.
- [x] 9.3 Add a subtle coordinate plane to the run timeline chart.
- [x] 9.4 Align tooltip trigger hit targets with all visible datapoints.
- [x] 9.5 Update tests and rerun dashboard verification commands.

## 10. Chart Library Migration

- [x] 10.1 Add Recharts dependency for monitor timeline rendering.
- [x] 10.2 Replace custom SVG chart with Recharts line chart, grid, axes, and tooltip.
- [x] 10.3 Preserve success/failure datapoint styling and summary context.
- [x] 10.4 Update tests and rerun dashboard verification commands.
