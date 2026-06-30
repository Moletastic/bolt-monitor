## ADDED Requirements

### Requirement: Test notification sends a test message through a channel
The system SHALL provide a test endpoint that sends a pre-defined test message through a configured channel.

#### Scenario: Test notification succeeds
- **WHEN** a tenant calls `POST /api/v1/notification-channels/{channelId}/test`
- **AND** the channel has valid config (bot token and chat ID for Telegram)
- **THEN** a test message is sent through that channel
- **AND** returns success

#### Scenario: Test notification fails due to invalid config
- **WHEN** a tenant calls `POST /api/v1/notification-channels/{channelId}/test`
- **AND** the channel has invalid or missing config
- **THEN** an error is returned describing the configuration issue

### Requirement: Test notification does not create an incident
The system SHALL ensure test notifications are sent directly without creating an incident record.

#### Scenario: Test notification bypasses incident creation
- **WHEN** a test notification is sent
- **THEN** no `Incident` record is created or modified in DynamoDB

### Requirement: Test message is clearly identifiable as a test
The system SHALL format the test notification message to indicate it is a test and not a real incident.

#### Scenario: Test message format
- **WHEN** a test notification is sent
- **THEN** the message text includes indication that this is a test (e.g., "[Test]")
