## MODIFIED Requirements

### Requirement: System exposes monitor overview under services module

System SHALL provide the current monitor-oriented overview inside the `Services` module rather than on the root dashboard route.

#### Scenario: Operator opens services module

- **WHEN** operator navigates to the services landing route
- **THEN** system shows the current monitor overview backed by real monitor API data

#### Scenario: Operator opens service list with services

- **WHEN** services exist in the Services module
- **THEN** system shows each service as an actionable card that navigates to service detail from the full non-interactive card area
- **AND** each card emphasizes service name centered next to a service category icon (server/database/cache/http/queue/container/function) wrapped in a square tile colored by service state
- **AND** each card includes a segmented health bar where each segment represents a monitor, colored by that monitor's status (UP/DOWN/DEGRADED), with segments dividing the bar width equally
- **AND** each card shows a metrics row with Avg latency (left-aligned), Agg. P99 (center-aligned), and Recent uptime (right-aligned)
- **AND** each card shows "Last checked at {local time}" when history exists, or "Never checked" in normal font when no history exists
- **AND** when the service has child monitors, each card shows the segmented health bar reflecting each monitor's current status
- **AND** when recent metric data is unavailable, each card shows an honest no-data, no-monitor, or draft state rather than fabricated metric values
- **AND** the "Pending config" label is hidden for draft services with no monitors
- **AND** raw service identifiers are not shown as primary card content
- **AND** the services list grid renders up to four service cards per row on wide viewports

#### Scenario: Service card visual state mirrors service health

- **WHEN** a service card is rendered with a rollup status of UP, DEGRADED, or DOWN
- **THEN** the technology icon tile, the monitor-up coverage counter, and the card outline use the same color family for that state
- **AND** counters show `X/Y monitors up` and use the warning color for partial coverage, the down color when zero monitors are up, the up color when all monitors are up, and the muted color when metric data is not available
- **AND** the technology icon is rendered smaller than its surrounding square tile so the colored tile is visually noticeable
- **AND** the monitor-up counter is rendered directly under the service-state status chip on the right side of the card with a small visual gap
- **AND** the service card does not render a secondary lifecycle-state label under the service name
- **AND** the service card does not render a "Based on … runs" or similar recent-sample label when metric data is available
- **AND** each monitor protocol badge text is colored by the corresponding monitor's current status using the same status color system

#### Scenario: Operator scans service technology

- **WHEN** service technology is available
- **THEN** system shows a consistently sized technology icon that is visually distinguishable in service overview, service cards, and service detail summary
- **AND** unknown or missing technology uses a consistent fallback icon with the same visual footprint
