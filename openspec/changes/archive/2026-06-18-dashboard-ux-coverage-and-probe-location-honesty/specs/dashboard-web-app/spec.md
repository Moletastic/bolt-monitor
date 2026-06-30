## ADDED Requirements

### Requirement: System presents probe-location selection honestly

The dashboard SHALL derive any monitor probe-location selection from the enabled subset of the canonical probe-location catalog read from the monitor API at request time. The dashboard SHALL NOT hard-code probe-location identifiers in client components or server actions as the source of selection.

#### Scenario: Catalog contains a single enabled location

- **WHEN** the enabled subset of the probe-location catalog contains exactly one location
- **THEN** the monitor form renders a non-interactive region chip showing the location name and a helper text indicating single-region preview
- **AND** the create and update monitor server actions submit the location from the catalog data, not from a constant

#### Scenario: Catalog contains multiple enabled locations

- **WHEN** the enabled subset of the probe-location catalog contains more than one location
- **THEN** the monitor form renders a real selection control bound to the enabled locations
- **AND** the dashboard does not impose a single-selection default that is not present in the catalog

### Requirement: System confirms destructive actions with an in-app dialog

The dashboard SHALL confirm permanent deletion of services, monitors, notification channels, and escalation policies using an in-app confirmation dialog rather than `window.confirm`.

#### Scenario: Operator triggers destructive delete

- **WHEN** operator activates a destructive delete control for a service, monitor, notification channel, or escalation policy
- **THEN** the dashboard opens an in-app confirmation dialog whose description matches the existing per-resource confirm message
- **AND** the cancel control receives focus by default
- **AND** the dialog is dismissable by keyboard, by clicking outside, and by the cancel control

#### Scenario: Operator confirms destructive delete

- **WHEN** operator activates the confirm control inside the dialog
- **THEN** the dashboard calls the existing delete API and navigates as before
- **AND** focus moves to a sensible next target such as the next list item, the parent list, or a Create CTA — not back to `<body>`

### Requirement: System degrades gracefully when dashboard APIs are unavailable

The dashboard SHALL render an operator-readable unavailable state when a backing API request fails, preserving the shared dashboard shell.

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

### Requirement: System keeps policy surfaces on the dark monitoring design language

The dashboard SHALL apply the dark monitoring design tokens and the shared AppShell to every route under the policies module, including policy edit.

#### Scenario: Operator opens policy edit

- **WHEN** operator opens an existing escalation policy detail page
- **THEN** the page is wrapped in the shared AppShell with the `Policies` module marked active
- **AND** status banners use the standard status tokens rather than light-mode Tailwind colors
- **AND** escalation-state UI uses the primary token for attention-queue indicators rather than raw sky/rose colors

### Requirement: System removes internal scaffolding from operator chrome

The dashboard SHALL NOT expose internal scaffolding language (for example "Bootstrap assumptions" or "built-in catalog assumption") in the operator-facing sidebar, header, or top-level home page.

#### Scenario: Operator scans sidebar and home page

- **WHEN** operator opens any dashboard route
- **THEN** the sidebar and header contain product navigation only and do not list internal scaffolding panels
- **AND** the home page heading and the sidebar header do not disagree on the product name

#### Scenario: Audit trail route

- **WHEN** operator navigates to `/audit-trail`
- **THEN** the route either redirects to a real surface (for example `/incidents`) or renders a real empty state that points operators to the surface where audit information lives
- **AND** the route does not render developer-only notes as primary operator content

### Requirement: System provides accessibility baseline across dashboard surfaces

The dashboard SHALL meet an accessibility baseline: per-page primary heading, skip-to-main link, announced empty states, labeled status indicators, and accessible tab semantics.

#### Scenario: Operator scans a dashboard page with a screen reader

- **WHEN** operator opens any dashboard route
- **THEN** the page exposes a primary `<h1>` inside the main content describing the page purpose
- **AND** a skip-to-main-content link is the first focusable element of the page
- **AND** status chips convey state by both text and `aria-label`, not by color alone

#### Scenario: Operator encounters an empty state

- **WHEN** any surface renders its empty state
- **THEN** the empty state is announced via `aria-live="polite"`
- **AND** the empty state explains the next operator action

#### Scenario: Operator uses tab navigation on a dashboard page

- **WHEN** operator navigates a tabbed interface (for example monitor detail or incident detail)
- **THEN** the tablist uses `role="tablist"` semantics
- **AND** tabs are reachable by keyboard and announce selection state

### Requirement: System uses shared table primitives on the channels surface

The dashboard SHALL render notification channels using the shared `<Table>` primitive.

#### Scenario: Operator scans channels table

- **WHEN** operator opens `/integrations/channels`
- **THEN** the channels list renders using the shared Table primitive
- **AND** row dividers, padding, and typography match the incidents and runs tables

### Requirement: System uses human-readable attention-queue labels

The dashboard SHALL translate internal attention-queue tones into operator-readable labels on the home page.

#### Scenario: Operator opens home page

- **WHEN** operator opens the home page
- **THEN** attention-queue items display human-readable labels such as "Action needed", "At risk", or "Heads-up" instead of raw tone identifiers (`down`, `warn`, `info`)
- **AND** the all-clear empty state copy describes the absence of operator attention without exposing internal terminology

## MODIFIED Requirements

### Requirement: System supports monitor management through dashboard

The dashboard SHALL allow operators to manage monitor configuration from the dashboard.

#### Scenario: Operator creates monitor from dashboard

- **WHEN** operator submits a valid create-monitor form
- **THEN** system creates the monitor through the existing monitor create API and reflects the new monitor in dashboard views
- **AND** the submitted probe-location identifiers are taken from the server-side probe-location catalog data rather than from a hard-coded client constant

#### Scenario: Operator updates monitor from dashboard

- **WHEN** operator submits valid monitor changes
- **THEN** system updates the monitor through the existing monitor update API and reflects the saved state in dashboard views
- **AND** the submitted probe-location identifiers are taken from the server-side probe-location catalog data rather than from a hard-coded client constant

#### Scenario: Operator enables or disables monitor from dashboard

- **WHEN** operator triggers enable or disable control for a monitor
- **THEN** system calls the existing action endpoint and reflects the resulting enabled state in dashboard views
- **AND** if the change is destructive (for example disabling a monitor that is part of the rollup), the dashboard requires an in-app confirmation before issuing the action