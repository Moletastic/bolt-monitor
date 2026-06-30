## ADDED Requirements

### Requirement: System reads enabled probe locations through the dashboard catalog API

The dashboard SHALL read the enabled subset of the probe-location catalog through the probe-location read API and SHALL NOT hard-code probe-location identifiers in dashboard actions or forms.

#### Scenario: Dashboard renders monitor location field

- **WHEN** the dashboard renders the monitor create or edit form
- **THEN** the available locations are derived from the canonical probe-location catalog read at request time
- **AND** the catalog call uses the existing probe-location read API rather than a constant or fixture

#### Scenario: Single enabled location is shown honestly

- **WHEN** the enabled subset of the catalog contains exactly one location
- **THEN** the dashboard renders a non-interactive region chip with helper copy indicating single-region preview
- **AND** the dashboard does not present the picker as if it were a multi-option selector

#### Scenario: Multiple enabled locations are presented as a real selector

- **WHEN** the enabled subset of the catalog contains more than one location
- **THEN** the dashboard renders a real selection control bound to the enabled locations
- **AND** submitted monitor payloads carry the operator's selection rather than a hard-coded default

### Requirement: System removes hard-coded probe-location defaults from dashboard actions

The dashboard SHALL NOT ship a hard-coded default probe-location identifier in its create or update monitor server actions.

#### Scenario: Operator creates monitor

- **WHEN** operator submits the create monitor form
- **THEN** the dashboard server action derives the submitted probe location from the server-side probe-location catalog data
- **AND** the action source code does not contain a hard-coded location constant

#### Scenario: Operator updates monitor

- **WHEN** operator submits the edit monitor form
- **THEN** the dashboard server action derives the submitted probe location from the server-side probe-location catalog data
- **AND** the action source code does not contain a hard-coded location constant