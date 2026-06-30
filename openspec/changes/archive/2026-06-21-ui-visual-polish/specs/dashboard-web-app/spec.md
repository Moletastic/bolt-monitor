## MODIFIED Requirements

### Requirement: System exposes monitor overview under services module

System SHALL provide the current monitor-oriented overview inside the `Services` module rather than on the root dashboard route.

#### Scenario: Operator opens services module

- **WHEN** operator navigates to the services landing route
- **THEN** system shows the current monitor overview backed by real monitor API data

#### Scenario: Operator opens service list with services

- **WHEN** services exist in the Services module
- **THEN** system shows each service as an actionable card that navigates to service detail from the full non-interactive card area
- **AND** each card emphasizes service name, description, rollup status, lifecycle, technology, monitor coverage, and update context
- **AND** raw service identifiers are not shown as primary card content

#### Scenario: Operator scans service technology

- **WHEN** service technology is available
- **THEN** system shows a consistently sized technology icon that visually fills its frame in service overview, service cards, and service detail summary
- **AND** the icon footprint matches the `sm`, `md`, or `lg` size requested at each call site
- **AND** unknown or missing technology uses a consistent fallback icon with the same visual footprint

### Requirement: System preserves monitoring design language in dashboard

System SHALL implement dashboard UI using the repository's monitoring design system.

#### Scenario: Dashboard renders status-oriented UI

- **WHEN** system renders dashboard surfaces
- **THEN** typography, colors, spacing, density, and status emphasis follow `DESIGN.md` and remain consistent with the intended monitoring-console visual language
- **AND** timestamps rendered anywhere on a dashboard surface (table cells, summary cards, status banners, settings cards, service detail panels) use the mono font token so they read as data rather than prose

### Requirement: System removes internal scaffolding from operator chrome

The dashboard SHALL NOT expose internal scaffolding language (for example "Bootstrap assumptions" or "built-in catalog assumption") in the operator-facing sidebar, header, or top-level home page.

#### Scenario: Operator scans sidebar and home page

- **WHEN** operator opens any dashboard route
- **THEN** the sidebar contains product navigation only and does not list internal scaffolding panels
- **AND** the dashboard does not render a sticky top header above the main content carrying only an eyebrow label, tagline, or duplicate create CTA
- **AND** the home page heading and the sidebar header do not disagree on the product name

#### Scenario: Audit trail route

- **WHEN** operator navigates to `/audit-trail`
- **THEN** the route either redirects to a real surface (for example `/incidents`) or renders a real empty state that points operators to the surface where audit information lives
- **AND** the route does not render developer-only notes as primary operator content

## ADDED Requirements

### Requirement: System renders service create and edit forms without read-only lifecycle field

The dashboard SHALL NOT render the lifecycle state as an input or read-only field inside the service create or edit form.

#### Scenario: Operator opens service create form

- **WHEN** operator opens the new service form
- **THEN** the form does not contain a Lifecycle label, value, or helper text
- **AND** lifecycle state remains visible on the service detail summary, the services list card, and the home service health matrix

#### Scenario: Operator opens service edit form

- **WHEN** operator opens the edit service form
- **THEN** the form does not contain a Lifecycle label, value, or helper text
- **AND** the form collects only inputs the operator can change
