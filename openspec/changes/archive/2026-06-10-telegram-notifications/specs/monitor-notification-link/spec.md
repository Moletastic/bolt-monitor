## ADDED Requirements

### Requirement: Monitors can be linked to multiple notification channels
The system SHALL allow a monitor to be linked to zero or more notification channels.

#### Scenario: Monitor with no links
- **WHEN** a monitor has no `MonitorNotificationLink` entries
- **THEN** only the tenant default channel applies (if any)

#### Scenario: Monitor with multiple links
- **WHEN** a monitor is linked to channels A and B
- **THEN** notifications are sent to both channels

### Requirement: Link stores event filter per channel
The system SHALL store which event types trigger notifications per monitor-channel link.

#### Scenario: Link with open and resolved events
- **WHEN** a link is created with events `["incident.opened", "incident.resolved"]`
- **THEN** notifications are sent for both incident open and resolve events

#### Scenario: Link with opened only
- **WHEN** a link is created with events `["incident.opened"]`
- **THEN** notifications are sent only when an incident opens

### Requirement: Link can be enabled or disabled per monitor
The system SHALL allow toggling a link's enabled state without removing the link.

#### Scenario: Disable link
- **WHEN** a tenant disables a monitor's link to a channel
- **THEN** that link does not receive notifications until re-enabled

### Requirement: Link deletion removes channel association
The system SHALL remove a monitor's channel association when the link is deleted.

#### Scenario: Delete link
- **WHEN** a tenant removes a monitor's link to a channel
- **THEN** the `MonitorNotificationLink` entry is deleted from DynamoDB
- **AND** the monitor no longer receives notifications on that channel

### Requirement: Link query returns all channels for a monitor
The system SHALL provide a way to query all notification channel links for a given monitor.

#### Scenario: List links for monitor
- **WHEN** a tenant requests notification channels for a monitor
- **THEN** system returns all `MonitorNotificationLink` entries where PK matches the monitor
