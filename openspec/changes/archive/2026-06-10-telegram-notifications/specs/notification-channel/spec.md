## ADDED Requirements

### Requirement: Notification channels are tenant-scoped entities
The system SHALL store notification channels as DynamoDB entities scoped to a tenant.

#### Scenario: Channel creation
- **WHEN** a tenant creates a notification channel with type, name, and config
- **THEN** system stores the channel under `TENANT#<tenantId>/CHANNEL#<channelId>`

#### Scenario: Channel retrieval
- **WHEN** a tenant lists their notification channels
- **THEN** system returns all channels where PK begins with `TENANT#<tenantId>`

### Requirement: Channel config contains type-specific settings
The system SHALL store channel-type-specific configuration as encrypted JSON in DynamoDB.

#### Scenario: Telegram channel config
- **WHEN** a channel of type `telegram` is created
- **THEN** config includes `botToken` (encrypted) and `chatId` (empty initially for auto-detect flow)

### Requirement: Channels can be enabled or disabled
The system SHALL allow toggling a channel's enabled state without deleting the channel.

#### Scenario: Disable channel
- **WHEN** a tenant disables a notification channel
- **THEN** the channel's `enabled` flag is set to false and no notifications route to it

#### Scenario: Enable channel
- **WHEN** a tenant enables a notification channel
- **THEN** the channel's `enabled` flag is set to true and notifications resume

### Requirement: Channel deletion requires no active monitor links
The system SHALL reject deletion of a notification channel that has monitor links unless force-delete is specified.

#### Scenario: Delete channel with links
- **WHEN** a tenant attempts to delete a channel that has monitor links
- **THEN** system returns error indicating channels are linked

#### Scenario: Delete channel without links
- **WHEN** a tenant deletes a channel with no monitor links
- **THEN** channel is removed from DynamoDB

### Requirement: Channel type is validated on creation
The system SHALL reject channel creation for unsupported notification types.

#### Scenario: Create unsupported channel type
- **WHEN** a tenant creates a channel with type `unsupported`
- **THEN** system returns validation error

#### Scenario: Create telegram channel
- **WHEN** a tenant creates a channel with type `telegram`
- **THEN** system accepts and stores the channel
