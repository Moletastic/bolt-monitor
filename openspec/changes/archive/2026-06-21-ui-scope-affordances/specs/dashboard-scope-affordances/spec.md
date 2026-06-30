## Purpose

Define scope-aware affordances on service and notification channel surfaces so operators can see at a glance which monitors belong to a service and which notification routes reference a channel.

## ADDED Requirements

### Requirement: System shows per-monitor traffic-light dots on service surfaces

The dashboard SHALL render a row of status dots, one per child monitor, on each service card and on each row of the home service health matrix.

#### Scenario: Operator scans a service card with monitors

- **WHEN** a service has one or more child monitors
- **THEN** the service card shows one status dot per child monitor
- **AND** each dot's color matches the monitor's current status using existing status tokens
- **AND** hovering or focusing a dot reveals the monitor name

#### Scenario: Operator scans a service card without monitors

- **WHEN** a service has zero child monitors
- **THEN** the service card shows no dot row
- **AND** the existing "no monitors" status banner remains the only signal

#### Scenario: Operator scans the home service health matrix

- **WHEN** the home page renders the service health matrix
- **THEN** each row includes the same per-monitor dot row as on the services list card
- **AND** the dots do not replace the existing rollup status chip

### Requirement: System shows notification channel usage scope on the channels list

The dashboard SHALL render, on each row of the notification channels list, a usage indicator showing how many notification routes reference the channel.

#### Scenario: Operator scans a channels list row

- **WHEN** the operator opens the notification channels list
- **THEN** each row shows a "Used by N routes" disclosure where N is the count of escalation policies whose business-hours or off-hours path references the channel

#### Scenario: Operator expands a channel's usage disclosure

- **WHEN** the operator expands the usage disclosure for a channel
- **THEN** the disclosure lists the referencing routes by name
- **AND** each route name links to the corresponding policy detail page

#### Scenario: Channel is unreferenced

- **WHEN** no notification route references a channel
- **THEN** the channel row shows an "Unused" indicator instead of a count
- **AND** the indicator does not link to any disclosure

### Requirement: Scope affordances remain informational

The per-monitor traffic-light dots SHALL be informational; the only interactive target on a service card remains the card itself. The channel usage scope SHALL expose a single disclosure per row with no other new interactive elements.

#### Scenario: Operator clicks a service card

- **WHEN** the operator activates a service card
- **THEN** the navigation target is the service detail page
- **AND** no individual dot intercepts the click

#### Scenario: Operator clicks a channel usage disclosure

- **WHEN** the operator expands a channel's usage disclosure
- **THEN** the only new interactive elements are the disclosure toggle and the route-name links inside the expanded list
