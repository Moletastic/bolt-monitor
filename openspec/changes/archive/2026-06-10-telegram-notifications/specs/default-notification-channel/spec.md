## ADDED Requirements

### Requirement: Tenant can designate one channel as the default
The system SHALL allow a tenant to mark one notification channel as the default for their tenant.

#### Scenario: Set channel as default
- **WHEN** a tenant marks a channel as default (`isDefault: true`)
- **THEN** any previously default channel for that tenant is unmarked
- **AND** only one channel per tenant is default at a time

#### Scenario: Default channel applies to monitors without explicit links
- **WHEN** a monitor has no `MonitorNotificationLink` entries
- **AND** the tenant has a default notification channel
- **THEN** notifications for that monitor route to the default channel

### Requirement: Only enabled default channels apply
The system SHALL NOT route to a default channel if it is disabled.

#### Scenario: Default channel is disabled
- **WHEN** a tenant's default channel is disabled
- **THEN** monitors without explicit links receive no notifications (unless another default is set)

### Requirement: Default channel can be changed
The system SHALL allow a tenant to change their default channel at any time.

#### Scenario: Change default channel
- **WHEN** a tenant marks a different channel as default
- **THEN** the previous default remains as a non-default channel
- **AND** new notifications route to the new default channel
