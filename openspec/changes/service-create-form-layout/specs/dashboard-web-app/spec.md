## ADDED Requirements

### Requirement: System provides focused service creation form layout

The dashboard SHALL provide a dedicated service creation page that uses full-width form content organized around service identity and notification setup.

#### Scenario: Operator opens service creation page
- **WHEN** the operator navigates to `/services/new`
- **THEN** system shows a page title and concise description for creating a service
- **AND** system uses the full dashboard content width for the creation form
- **AND** system does not show a separate “Create flow notes” side card

#### Scenario: Operator enters service identity
- **WHEN** the operator views the service creation form
- **THEN** system shows a `Service identity` section with an icon and section name
- **AND** the section groups service icon/category selection and service name in the same row when viewport width allows
- **AND** the service icon/category selector shows actual service category icons alongside understandable labels
- **AND** the description field appears below the icon/category and service name controls

#### Scenario: Operator configures notifications
- **WHEN** the operator views the service creation form
- **THEN** system shows a `Notifications` section with an icon and section name
- **AND** the section includes notification route selection
- **AND** the section includes a business-hours switch below notification route selection
- **AND** the business-hours switch includes explanatory text describing its notification-routing purpose
- **AND** when business hours are enabled, system preserves controls for timezone, time window, and days of week

#### Scenario: Operator submits service creation
- **WHEN** the operator submits a valid service creation form
- **THEN** system creates a draft service through the existing service creation behavior
- **AND** system redirects to the created service detail view
- **AND** system does not attempt to create monitors as part of service creation

#### Scenario: Operator edits an existing service
- **WHEN** the operator edits an existing service
- **THEN** system preserves existing update behavior
- **AND** system does not show create-only page framing that would imply a new service is being created
