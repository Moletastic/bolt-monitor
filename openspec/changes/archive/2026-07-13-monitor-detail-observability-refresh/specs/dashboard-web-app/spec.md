## MODIFIED Requirements

### Requirement: System exposes monitor detail in dashboard
System SHALL provide a detailed monitor view for operational inspection.

#### Scenario: Operator views monitor detail
- **WHEN** operator opens an individual monitor
- **THEN** system shows monitor configuration, latest status, and recent run history using existing monitor read APIs
- **AND** keeps the monitor detail view inside the same module-oriented dashboard shell with `Services` treated as the active module

#### Scenario: Operator reviews current monitor status
- **WHEN** operator opens monitor detail
- **THEN** system shows a current-status summary with monitor name, protocol/type, target, enabled state, current status, last outcome, last check time, duration, and cadence
- **AND** system shows the latest error when status data includes an error
- **AND** raw service or monitor identifiers are not shown as primary status content

#### Scenario: Operator scans monitor identity and actions on desktop
- **WHEN** operator opens monitor detail on a wide viewport
- **THEN** system shows a dot-only monitor status indicator immediately before the monitor name in the left identity cluster
- **AND** system shows `Run now`, edit, and a vertical-more monitor actions menu in a right-aligned action cluster
- **AND** the edit action navigates to an edit monitor view preloaded with the monitor's existing configuration
- **AND** the vertical-more monitor actions menu contains maintenance, enable or disable, and delete monitor actions
- **AND** each visible action includes a leading icon or icon-only accessible label
- **AND** the `Run now` action is visually primary while using an outlined treatment
- **AND** system does not render a separate delete-monitor section below the evidence tabs

#### Scenario: Operator scans monitor request metadata
- **WHEN** operator opens monitor detail
- **THEN** system shows protocol, HTTP method, and endpoint below the monitor title as the executable check summary
- **AND** system shows frequency and timeout as compact badges below the request summary
- **AND** system does not render a separate `Check configuration` section

#### Scenario: Operator scans monitor indicators on desktop
- **WHEN** operator opens monitor detail on a wide viewport
- **THEN** system shows four indicator cards for current state, recent uptime, P99 latency, and error rate
- **AND** each indicator card includes a contextual icon at the top-left of the card
- **AND** recent uptime, P99 latency, and error rate are derived from the recent monitor run history already loaded for the page
- **AND** system shows an operator-readable no-data state for indicators that cannot be calculated from available runs

#### Scenario: Operator reviews monitor performance and configuration on desktop
- **WHEN** operator opens monitor detail on a wide viewport
- **THEN** system shows a full-width recent run timeline chart for the monitor
- **AND** the chart plots recent run duration datapoints and visually distinguishes failed outcomes from successful outcomes
- **AND** chart datapoint tooltips describe run time, duration, outcome, and available status or error context
- **AND** the chart includes operator-readable summary context for recent samples, uptime, P99 latency, and failures
- **AND** the chart includes a subtle coordinate plane so the latency trend does not float without scale context
- **AND** every visible datapoint has a hover and focus tooltip target aligned with the datapoint
- **AND** the chart uses a dedicated chart library rather than hand-written SVG coordinate rendering

#### Scenario: Operator uses monitor detail on mobile
- **WHEN** operator opens monitor detail on a narrow viewport
- **THEN** system shows monitor name with a dot-only status indicator before the action controls
- **AND** system keeps visible text for `Run now` while rendering edit and vertical-more controls with accessible names
- **AND** system shows a compact indicator picker for current state, recent uptime, P99 latency, and error rate, defaulting to current state
- **AND** system renders only the selected indicator card below the indicator picker
- **AND** system renders request metadata before the compact indicator picker and recent run timeline chart

#### Scenario: Operator deletes monitor from detail actions menu
- **WHEN** operator opens the vertical-more monitor actions menu and chooses delete monitor
- **THEN** system opens an in-app confirmation dialog
- **AND** the dialog requires the operator to type the monitor name before the destructive confirm action is enabled
- **AND** successful deletion uses the existing monitor delete action and returns the operator to the parent service detail view

#### Scenario: Operator reviews monitor evidence tabs
- **WHEN** operator opens monitor detail
- **THEN** system preserves the `Runs`, `Incidents`, and `Audit` tabs
- **AND** each tab label includes an icon matching the tab purpose
- **AND** each selected tab shows its corresponding table or empty state using the existing backing API data
