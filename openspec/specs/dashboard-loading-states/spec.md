# dashboard-loading-states Specification

## Purpose
TBD - created by archiving change ui-loading-states. Update Purpose after archive.
## Requirements
### Requirement: System renders per-segment loading UI

The dashboard SHALL render a per-segment loading placeholder while server data is fetching on any route that depends on a server fetch.

#### Scenario: Operator navigates to a data-fetching segment

- **WHEN** operator opens any dashboard route that performs a server-side data fetch
- **THEN** system renders a loading placeholder matching the destination page's card grid, table column count, and primary heading shape
- **AND** placeholders use the dashboard's surface and skeleton styling tokens

#### Scenario: Operator navigates between segments

- **WHEN** operator navigates from one segment to another
- **THEN** the new segment's loading placeholder appears immediately
- **AND** the previous segment is removed

### Requirement: System renders skeleton rows inside tables while data resolves

The dashboard SHALL render skeleton rows inside the `<TableBody>` of any dashboard table whose data is still resolving.

#### Scenario: Operator scans a table on a multi-fetch page

- **WHEN** a table on a multi-fetch page is still resolving its data
- **THEN** the table renders skeleton rows matching the destination column count
- **AND** the table headers remain visible and stable
- **AND** skeleton cells use variable widths so rows do not appear mechanical

#### Scenario: Table data resolves

- **WHEN** the table data fetch resolves
- **THEN** the skeleton rows are replaced by the real rows in a single layout-stable update

### Requirement: System applies a subtle page transition on segment navigation

The dashboard SHALL apply a short fade-in animation when navigation between segments begins, while respecting `prefers-reduced-motion`.

#### Scenario: Operator navigates between segments

- **WHEN** operator navigates from one dashboard segment to another
- **THEN** the destination segment fades in over a short duration
- **AND** the animation respects `prefers-reduced-motion: reduce` and is suppressed when reduced motion is requested

#### Scenario: Operator prefers reduced motion

- **WHEN** operator has `prefers-reduced-motion: reduce` set
- **THEN** system renders the new segment without any fade animation

