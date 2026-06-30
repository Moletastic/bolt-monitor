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
- **THEN** system shows a consistently sized technology icon that is visually distinguishable in service overview, service cards, and service detail summary
- **AND** unknown or missing technology uses a consistent fallback icon with the same visual footprint

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

### Requirement: System exposes monitor overview with protocol context
System SHALL show protocol/type context when listing monitors.

#### Scenario: Operator scans monitor overview on desktop
- **WHEN** operator views a monitor overview table
- **THEN** system includes a dedicated protocol/type column separate from monitor name
- **AND** rows show monitor name, protocol/type, current status, enabled state, last check, duration, probe location, and available action
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
- **THEN** system shows a current-status summary with monitor name, protocol/type, target, enabled state, current status, last outcome, last check time, duration, probe location, and cadence
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
- **AND** the overview includes scheduler recurring execution state, probe location catalog summary when available, and safe setup/environment context

#### Scenario: Settings source data is unavailable
- **WHEN** scheduler configuration or probe location data cannot be loaded
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
