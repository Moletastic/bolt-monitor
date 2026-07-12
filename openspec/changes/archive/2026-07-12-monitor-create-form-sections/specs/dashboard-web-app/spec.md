## ADDED Requirements

### Requirement: System provides sectioned monitor create and edit form

The dashboard SHALL provide a monitor form organized around identity, protocol-specific request configuration, and validation expectations while preserving existing HTTP monitor submission behavior.

#### Scenario: Operator opens monitor creation page
- **WHEN** the operator navigates to `/services/{serviceId}/monitors/new`
- **THEN** system shows the monitor creation form using the available dashboard content width
- **AND** system does not show a separate “Create flow notes” side card
- **AND** system preserves existing service lookup, breadcrumb, and unavailable-state behavior

#### Scenario: Operator reviews monitor identity
- **WHEN** the operator views the monitor form
- **THEN** system shows an `Identity` section
- **AND** the section includes monitor name
- **AND** the section includes check frequency

#### Scenario: Operator reviews protocol choices
- **WHEN** the operator views the monitor form
- **THEN** system shows protocol choices between the `Identity` section and the `Request` section
- **AND** `HTTP` is selected by default
- **AND** `TCP` is disabled
- **AND** `gRPC` is disabled
- **AND** disabled protocol choices communicate that they are coming soon
- **AND** form submission continues to create or update HTTP monitors only

#### Scenario: Operator configures HTTP request
- **WHEN** the operator views the HTTP `Request` section
- **THEN** system shows method, target URL, and timeout in milliseconds controls
- **AND** system shows headers as editable key/value rows
- **AND** each header row includes a delete control
- **AND** system provides a full-width add-header control
- **AND** new monitors default headers to `Content-Type: application/json`
- **AND** submitted headers preserve the existing HTTP configuration payload shape
- **AND** system does not require or submit deprecated probe-location configuration from the dashboard form

#### Scenario: Operator configures HTTP validation
- **WHEN** the operator views the HTTP `Validation` section
- **THEN** system shows expected status codes as removable badge-style selected values
- **AND** new monitors default expected status codes to `200`
- **AND** system offers common HTTP status code options
- **AND** system does not expose arbitrary status code freeform entry in this phase
- **AND** system hides the expected-body-contains control from the active form
- **AND** submitted expected status codes preserve the existing HTTP configuration payload shape

#### Scenario: Operator edits a monitor through the shared form
- **WHEN** a monitor edit flow renders the shared monitor form
- **THEN** system uses the same sectioned layout as monitor creation
- **AND** system preserves existing monitor update submission behavior

#### Scenario: Dashboard creates a monitor without deprecated probe locations
- **WHEN** the dashboard submits a monitor creation request without `probeLocations`
- **THEN** system accepts the request using service-owned execution defaults
- **AND** system does not require dashboard operators to choose execution locations

### Requirement: System removes temporary monitor detail edit form

The dashboard SHALL stop rendering the current embedded monitor edit form on the monitor detail page until the monitor detail edit experience is refactored separately.

#### Scenario: Operator opens monitor detail page
- **WHEN** the operator navigates to `/services/{serviceId}/monitors/{monitorId}`
- **THEN** system does not show the embedded monitor edit form
- **AND** system preserves the rest of the monitor detail status and operational information
