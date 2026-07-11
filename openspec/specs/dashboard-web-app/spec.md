## Purpose

Define the operator dashboard web application used to inspect monitoring health, triage operational attention, and manage service-first monitor workflows.
## Requirements
### Requirement: System provides operator dashboard web application
System SHALL provide a web application for operators to inspect monitoring health, triage operational attention, and manage monitors through a module-oriented console layout.

#### Scenario: Operator opens dashboard home
- **WHEN** operator navigates to the dashboard application
- **THEN** system shows an operational overview framed inside the shared dashboard sidebar shell
- **AND** the overview summarizes service health, incident state, scheduler state, and setup gaps using available dashboard APIs

#### Scenario: Operator sees prioritized attention
- **WHEN** operator opens the dashboard home and there are down services, open incidents, disabled scheduler state, services without monitors, disabled monitor coverage, or draft services
- **THEN** system shows a prioritized attention area that identifies the items needing operator review
- **AND** each actionable item links to the existing module route where the operator can inspect or manage it

#### Scenario: Operator reviews service health matrix
- **WHEN** operator opens the dashboard home and services exist
- **THEN** system shows a compact service health matrix with service identity, rollup status, lifecycle state, monitor coverage, and recent update context
- **AND** each service row links to the existing service detail route

#### Scenario: Operator opens dashboard with no configured services
- **WHEN** operator navigates to the dashboard application and no services exist
- **THEN** system shows an empty-state path to create the first service
- **AND** system does not show misleading zero-health summaries as if monitoring coverage exists

#### Scenario: Dashboard overview API context is unavailable
- **WHEN** one or more dashboard overview API requests fail
- **THEN** system shows an actionable unavailable or partial-state message for the affected overview content
- **AND** system preserves the shared dashboard shell and navigation

### Requirement: System exposes monitor overview under services module

System SHALL provide the current monitor-oriented overview inside the `Services` module rather than on the root dashboard route.

#### Scenario: Operator opens services module

- **WHEN** operator navigates to the services landing route
- **THEN** system shows the current monitor overview backed by real monitor API data

#### Scenario: Operator opens service list with services

- **WHEN** services exist in the Services module
- **THEN** system shows each service as an actionable card that navigates to service detail from the full non-interactive card area
- **AND** each card emphasizes service name, description, rollup status, lifecycle, technology, monitor coverage, and update context
- **AND** when the service has child monitors, each card shows a per-monitor traffic-light dot row reflecting each monitor's current status
- **AND** raw service identifiers are not shown as primary card content

#### Scenario: Operator scans service technology

- **WHEN** service technology is available
- **THEN** system shows a consistently sized technology icon that is visually distinguishable in service overview, service cards, and service detail summary
- **AND** unknown or missing technology uses a consistent fallback icon with the same visual footprint

### Requirement: System provides service-list layout controls

The dashboard SHALL present the Services landing page as a scannable service list with health indicators, list controls, and viewport-appropriate service creation affordances.

#### Scenario: Operator opens service list on desktop

- **WHEN** services exist and the operator opens the Services landing page on a wide viewport
- **THEN** system shows a visible `Services` page title and concise service-list description
- **AND** system shows equal-size summary indicators for active services, draft services, and services that are down now to the right of the title area
- **AND** system does not show a duplicated operations card containing lower-fidelity service previews above the service card list

#### Scenario: Operator controls the service list on desktop

- **WHEN** services exist and the operator views the Services landing page on a wide viewport
- **THEN** system shows a controls row above the service cards
- **AND** the controls row includes a search field for narrowing the visible service cards
- **AND** the controls row includes filter affordances for service-list refinement
- **AND** the controls row includes a `Create service` action that navigates to the service creation flow

#### Scenario: Operator opens service list on mobile

- **WHEN** services exist and the operator opens the Services landing page on a narrow viewport
- **THEN** system shows active, draft, and down-now indicators in a single row at the top of the visible content
- **AND** system does not show the desktop page title and description in the visible content flow
- **AND** system preserves an accessible page heading for assistive technology
- **AND** system shows search and filter controls in the next row
- **AND** system renders service cards one per row

#### Scenario: Operator creates a service on mobile

- **WHEN** services exist and the operator views the Services landing page on a narrow viewport
- **THEN** system shows the `Create service` action as a fixed floating action button at the bottom-right of the viewport
- **AND** the floating action button has an accessible name describing that it creates a service
- **AND** the floating action button remains reachable without covering primary service card content

#### Scenario: Service list layout is loading

- **WHEN** the Services landing page is loading
- **THEN** system shows loading placeholders that match the final service-list layout structure for the current viewport

### Requirement: System exposes service detail as operational summary
System SHALL provide service detail content that summarizes monitoring health and coverage for the service.

#### Scenario: Operator opens service detail
- **WHEN** operator opens an individual service
- **THEN** system shows a monitoring-oriented service summary with service name, description, rollup status, lifecycle state, technology, monitor count, enabled monitor coverage, and last update context
- **AND** raw service identifiers are not shown as primary header content
- **AND** the create-monitor path remains visible from the summary area

#### Scenario: Service has setup or health gaps
- **WHEN** the service is draft, has no monitors, has disabled monitor coverage, or has a down rollup status
- **THEN** system shows an operator-readable setup or health signal in the service summary

### Requirement: System renders expanded service technology icons

The dashboard SHALL render and select a generic technology icon for every supported service category.

#### Scenario: Operator scans service category icons
- **WHEN** dashboard surfaces render a service with category `web`, `api`, `worker`, `scheduler`, `storage`, `search`, `auth`, `payments`, `analytics`, `observability`, `ai`, or `integration`
- **THEN** system shows a distinct generic technology icon for that category
- **AND** system shows an understandable category label where text labels are part of the surface

#### Scenario: Operator scans service purpose icons
- **WHEN** dashboard surfaces render a service with category `media`, `content`, `finance`, `learning`, `gaming`, `commerce`, `messaging`, `support`, `marketing`, `admin`, `security`, `location`, or `social`
- **THEN** system shows a distinct generic purpose icon for that category
- **AND** system shows an understandable category label where text labels are part of the surface

#### Scenario: Operator selects service category
- **WHEN** the operator creates or edits a service in the dashboard
- **THEN** system includes every supported service category in the service icon/category selector
- **AND** each selectable category is represented with an icon and understandable label

#### Scenario: Dashboard renders existing category
- **WHEN** dashboard surfaces render a service with existing category `server`, `database`, `cache`, `http`, `queue`, `container`, or `function`
- **THEN** system continues to show the existing category with a technology icon and label

#### Scenario: Dashboard renders missing or unknown category
- **WHEN** dashboard surfaces render a service with no category or a category not recognized by the dashboard catalog
- **THEN** system uses a safe fallback service icon
- **AND** system does not fail rendering the service surface

#### Scenario: Dashboard category catalog matches backend support
- **WHEN** the expanded service category catalog is implemented
- **THEN** dashboard TypeScript category definitions include every backend-supported category value
- **AND** dashboard icon rendering provides a mapping for every dashboard category value

### Requirement: System exposes monitor overview with protocol context
System SHALL show protocol/type context when listing monitors.

#### Scenario: Operator scans monitor overview on desktop
- **WHEN** operator views a monitor overview table
- **THEN** system includes a dedicated protocol/type column separate from monitor name
- **AND** rows show monitor name, protocol/type, current status, enabled state, last check, duration, and available action
- **AND** raw monitor identifiers are not shown as primary row content

#### Scenario: Operator scans monitor overview on mobile
- **WHEN** operator views monitor overview cards on a narrow viewport
- **THEN** system keeps protocol/type visible on each monitor card
- **AND** raw monitor identifiers are not shown as primary card content

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

### Requirement: System provides useful integrations module state

System SHALL make notification channel management usable immediately when operators open the Integrations module.

#### Scenario: Operator opens integrations module

- **WHEN** operator navigates to `/integrations`
- **THEN** system starts loading notification channels without requiring a manual refresh action
- **AND** system shows loading feedback while the channel list is pending

#### Scenario: Notification channels are loaded

- **WHEN** notification channels are returned by the API
- **THEN** system shows configured channels with channel name, type, enabled state, default state, and available actions
- **AND** each row shows a usage-scope indicator describing how many notification routes reference the channel, with an expandable disclosure listing the referencing routes

#### Scenario: No notification channels exist

- **WHEN** notification channel API returns an empty collection
- **THEN** system shows an empty state that explains no alert channels are configured and provides a path to add one

#### Scenario: Notification channels cannot be loaded

- **WHEN** the notification channel API request fails
- **THEN** system shows an actionable unavailable state and preserves the manual retry path

### Requirement: System exposes incidents module as operational incident center
System SHALL provide an incidents module that is useful for both populated and empty incident states.

#### Scenario: Operator opens incidents module
- **WHEN** operator navigates to `/incidents`
- **THEN** system lists incidents from the incident API with opened time, summary, status, origin, and available drill-down action
- **AND** raw monitor identifiers are not shown as primary incident labels

#### Scenario: Incident collection is empty
- **WHEN** the selected incident filter has no matching incidents
- **THEN** system shows an empty state that distinguishes healthy no-open-incident state from no-history state
- **AND** system explains that incidents are created by monitor execution rather than manual UI creation

#### Scenario: Incident collection cannot be loaded
- **WHEN** the incident API request fails
- **THEN** system shows an unavailable state inside the shared dashboard shell

### Requirement: System exposes settings module overview
System SHALL provide a settings module overview for dashboard control-plane context.

#### Scenario: Operator opens settings module
- **WHEN** operator navigates to `/config`
- **THEN** system shows a settings overview instead of placeholder content
- **AND** the overview includes scheduler recurring execution state and safe setup/environment context

#### Scenario: Settings source data is unavailable
- **WHEN** scheduler configuration data cannot be loaded
- **THEN** system shows an unavailable state for the affected settings section while preserving the rest of the settings page

### Requirement: System preserves operator-focused identifiers in dashboard UI
System SHALL prioritize human-readable operational identity over storage identifiers in dashboard views.

#### Scenario: Operator scans dashboard surfaces
- **WHEN** system renders service, monitor, incident, integration, or settings surfaces
- **THEN** primary visible labels use human-readable names, summaries, statuses, protocols, targets, timestamps, or actions
- **AND** raw service IDs, monitor IDs, channel IDs, and incident IDs are not shown as primary content unless no human-readable value exists

#### Scenario: Debug identifier is needed
- **WHEN** an identifier is useful for support or debugging
- **THEN** system MAY expose it through a low-emphasis metadata or copy affordance rather than as the main label

### Requirement: System supports monitor management through dashboard

The dashboard SHALL allow operators to manage monitor configuration from the dashboard.

#### Scenario: Operator creates monitor from dashboard

- **WHEN** operator submits a valid create-monitor form
- **THEN** system creates the monitor through the existing monitor create API and reflects the new monitor in dashboard views
- **AND** the submitted payload does not include probe-location or region selection

#### Scenario: Operator updates monitor from dashboard

- **WHEN** operator submits valid monitor changes
- **THEN** system updates the monitor through the existing monitor update API and reflects the saved state in dashboard views
- **AND** the submitted payload does not include probe-location or region selection

#### Scenario: Operator enables or disables monitor from dashboard

- **WHEN** operator triggers enable or disable control for a monitor
- **THEN** system calls the existing action endpoint and reflects the resulting enabled state in dashboard views
- **AND** if the change is destructive (for example disabling a monitor that is part of the rollup), the dashboard requires an in-app confirmation before issuing the action

### Requirement: System preserves monitoring design language in dashboard

System SHALL implement dashboard UI using the repository's monitoring design system.

#### Scenario: Dashboard renders status-oriented UI

- **WHEN** system renders dashboard surfaces
- **THEN** typography, colors, spacing, density, and status emphasis follow `DESIGN.md` and remain consistent with the intended monitoring-console visual language
- **AND** timestamps rendered anywhere on a dashboard surface (table cells, summary cards, status banners, settings cards, service detail panels) use the mono font token so they read as data rather than prose

### Requirement: System supports service deletion through dashboard
System SHALL allow operators to permanently delete eligible services from the dashboard.

#### Scenario: Operator deletes service from service detail
- **WHEN** operator confirms deletion for an eligible service from the service detail view
- **THEN** system deletes the service through the service delete API
- **AND** system redirects the operator to the service list
- **AND** system refreshes dashboard service data so the deleted service is no longer shown

#### Scenario: Operator attempts to delete active service
- **WHEN** operator attempts to delete an active service from the dashboard
- **THEN** system explains that active services must be archived or otherwise made inactive before deletion
- **AND** system preserves the current service detail view

#### Scenario: Service deletion fails
- **WHEN** the service delete API rejects or fails a dashboard delete request
- **THEN** system shows an actionable error message
- **AND** system does not navigate as if deletion succeeded

### Requirement: System supports monitor deletion through dashboard
System SHALL allow operators to permanently delete eligible monitors from the dashboard.

#### Scenario: Operator deletes monitor from monitor detail
- **WHEN** operator confirms deletion for an eligible monitor from the monitor detail view
- **THEN** system deletes the monitor through the nested monitor delete API
- **AND** system redirects the operator to the parent service detail view
- **AND** system refreshes dashboard service and monitor data so the deleted monitor is no longer shown

#### Scenario: Operator attempts to delete last active-service monitor
- **WHEN** operator attempts to delete the only monitor under an active service
- **THEN** system explains why the monitor cannot be deleted in the current service state
- **AND** system preserves the current monitor detail view

#### Scenario: Monitor deletion fails
- **WHEN** the monitor delete API rejects or fails a dashboard delete request
- **THEN** system shows an actionable error message
- **AND** system does not navigate as if deletion succeeded

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

- **WHEN** a page awaits a single API call that fails (for example `/admin/scheduler`)
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
- **THEN** the sidebar contains product navigation only and does not list internal scaffolding panels
- **AND** the dashboard does not render a sticky top header above the main content carrying only an eyebrow label, tagline, or duplicate create CTA
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

### Requirement: System shows per-monitor traffic-light dots on the home service health matrix

The home page service health matrix SHALL render the same per-monitor traffic-light dot row on each matrix row that the services list card renders.

#### Scenario: Operator scans the home service health matrix

- **WHEN** the home page renders the service health matrix with services that have child monitors
- **THEN** each matrix row includes a per-monitor dot row matching the current status of each child monitor
- **AND** the dot row sits alongside the existing rollup status chip without replacing it

### Requirement: System marks unreferenced notification channels

The channels list SHALL distinguish referenced channels from unreferenced ones so operators can spot orphan configuration.

#### Scenario: Channel is unreferenced

- **WHEN** no notification route references a channel
- **THEN** the channel row shows an "Unused" indicator in place of the usage count
- **AND** the indicator does not link to any disclosure

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

### Requirement: System uses chevron breadcrumb separators

Dashboard breadcrumbs SHALL use decorative right-chevron separators between breadcrumb items.

#### Scenario: Breadcrumbs render multiple items
- **WHEN** breadcrumbs contain more than one item
- **THEN** system visually separates items with muted right-chevron separators
- **AND** separators are hidden from assistive technology
- **AND** parent links and current-page semantics remain unchanged

### Requirement: System uses sharper dashboard border radius

The dashboard SHALL use a reduced border-radius scale for general operational surfaces.

#### Scenario: General dashboard surfaces render
- **WHEN** system renders cards, panels, dialogs, menus, inputs, selects, buttons, or feedback banners
- **THEN** system uses the reduced dashboard radius scale
- **AND** surfaces appear less curved than the previous dashboard style

#### Scenario: Semantic round shapes render
- **WHEN** system renders status chips, traffic-light dots, circular icons, progress bars, or floating action buttons
- **THEN** system preserves intentional pill or circle shapes

#### Scenario: Radius scale is configured
- **WHEN** dashboard styling is configured
- **THEN** the Tailwind border-radius tokens are reduced from the previous scale
- **AND** components use shared tokens instead of one-off radius values where practical

### Requirement: System provides global top-bar search

The dashboard SHALL provide a top-bar search control for finding and navigating to services, monitors, notification routes, and notification channels.

#### Scenario: Operator views dashboard shell
- **WHEN** an operator opens a dashboard page inside the shared shell
- **THEN** system shows a top-bar search input with a search icon
- **AND** the input communicates that it can search services, monitors, routes, and channels

#### Scenario: Operator types search query
- **WHEN** the operator enters at least the minimum number of search characters
- **THEN** system waits for a debounce interval before requesting search results
- **AND** system shows user feedback while results are loading
- **AND** system avoids issuing a request for every keystroke

#### Scenario: Operator sees search results
- **WHEN** matching search results are returned
- **THEN** system displays a clickable result list below or near the search input
- **AND** each result shows a resource-specific icon
- **AND** each result shows primary text and secondary text appropriate to its resource type
- **AND** each result navigates to the returned resource link

#### Scenario: Operator searches with no matches
- **WHEN** search completes with no matching results
- **THEN** system shows an empty search feedback message without navigating away

#### Scenario: Search request fails
- **WHEN** the search API request fails
- **THEN** system shows an actionable error state near the search input
- **AND** system preserves the current page and input value

#### Scenario: Operator uses keyboard search interaction
- **WHEN** the search input or results are focused
- **THEN** system preserves keyboard accessibility for entering a query, reaching result links, dismissing the result list, and navigating via result activation

#### Scenario: Operator uses search on narrow viewport
- **WHEN** the dashboard is viewed on a narrow viewport
- **THEN** system keeps the search input usable without hiding primary navigation or page content
- **AND** result links remain tappable and readable
