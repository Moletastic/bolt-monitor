## ADDED Requirements

### Requirement: System provides dashboard breadcrumbs on nested pages

The dashboard SHALL show breadcrumb navigation on pages that have meaningful parent context below a module landing page.

#### Scenario: Operator opens dashboard root or module landing page
- **WHEN** the operator opens `/`, `/services`, `/policies`, `/integrations/channels`, `/incidents`, or `/config`
- **THEN** system does not show breadcrumb navigation
- **AND** system relies on the page heading and sidebar active state for orientation

#### Scenario: Operator opens service pages
- **WHEN** the operator opens `/services/new`
- **THEN** system shows breadcrumbs `Services / Create service`
- **WHEN** the operator opens `/services/{serviceId}`
- **THEN** system shows breadcrumbs `Services / {service name}`
- **WHEN** the operator opens `/services/{serviceId}/monitors/new`
- **THEN** system shows breadcrumbs `Services / {service name} / Create monitor`
- **WHEN** the operator opens `/services/{serviceId}/monitors/{monitorId}`
- **THEN** system shows breadcrumbs `Services / {service name} / {monitor name}`

#### Scenario: Operator opens notification route pages
- **WHEN** the operator opens `/policies/new`
- **THEN** system shows breadcrumbs `Notification routes / Create route`
- **WHEN** the operator opens `/policies/{policyId}`
- **THEN** system shows breadcrumbs `Notification routes / {policy name}`

#### Scenario: Operator opens channel pages
- **WHEN** the operator opens `/integrations/channels/new`
- **THEN** system shows breadcrumbs `Channels / Create channel`
- **WHEN** the operator opens `/integrations/channels/{channelId}`
- **THEN** system shows breadcrumbs `Channels / {channel name}`

#### Scenario: Operator opens incident or settings subpages
- **WHEN** the operator opens `/incidents/{incidentId}`
- **THEN** system shows breadcrumbs `Incidents / {incident summary or incident ID}`
- **WHEN** the operator opens `/admin/scheduler`
- **THEN** system shows breadcrumbs `Settings / Scheduler`
- **WHEN** the operator opens `/locations`
- **THEN** system shows breadcrumbs `Settings / Probe locations`
- **WHEN** the operator opens `/audit-trail`
- **THEN** system shows breadcrumbs `Incidents / Audit trail`

### Requirement: System renders breadcrumbs accessibly

Dashboard breadcrumb navigation SHALL be accessible and link-based.

#### Scenario: Breadcrumbs render
- **WHEN** breadcrumbs are shown
- **THEN** system renders them inside navigation semantics labelled for breadcrumbs
- **AND** every parent breadcrumb is a link
- **AND** the current page breadcrumb is not a link
- **AND** the current page breadcrumb indicates current page semantics

#### Scenario: Breadcrumb label is unavailable
- **WHEN** a dynamic resource name is unavailable but the page still renders
- **THEN** system uses a stable fallback label for the current breadcrumb
- **AND** system does not expose raw opaque IDs as primary breadcrumb text unless no human-readable fallback exists

#### Scenario: Breadcrumbs render on narrow viewport
- **WHEN** breadcrumbs are shown on a narrow viewport
- **THEN** system keeps breadcrumb links tappable
- **AND** system prevents long labels from breaking page layout
