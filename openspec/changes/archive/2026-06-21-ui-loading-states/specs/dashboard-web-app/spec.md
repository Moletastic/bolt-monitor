## MODIFIED Requirements

### Requirement: System degrades gracefully when dashboard APIs are unavailable

The dashboard SHALL render an operator-readable unavailable state when a backing API request fails, and SHALL render a loading placeholder that matches the destination page's shape while a backing API request is in flight.

#### Scenario: Top-level unhandled error boundary

- **WHEN** an unhandled error occurs in any dashboard route
- **THEN** the dashboard renders an unavailable-state page inside the shared AppShell
- **AND** the page exposes a retry control that calls the Next.js error boundary reset

#### Scenario: Per-section API failure on a multi-fetch page

- **WHEN** a page issues several parallel API requests and one or more fail
- **THEN** the affected section renders an unavailable card inside the shared shell
- **AND** sections whose data loaded successfully continue to render normally

#### Scenario: Top-level awaits on single-fetch pages

- **WHEN** a page awaits a single API call that fails (for example `/admin/scheduler` or `/locations`)
- **THEN** the page renders an unavailable state inside the shared shell
- **AND** the dashboard does not surface the raw error stack trace to operators

#### Scenario: Page is loading server data

- **WHEN** a dashboard segment is in flight for any server data fetch
- **THEN** system renders a loading placeholder matching the destination page's card grid, table column count, and primary heading shape
- **AND** the loading placeholder lives inside the shared AppShell

#### Scenario: Table data is loading

- **WHEN** a table on a multi-fetch page is still resolving its data
- **THEN** the table renders skeleton rows inside the body matching the destination column count
- **AND** the table headers remain visible and stable
