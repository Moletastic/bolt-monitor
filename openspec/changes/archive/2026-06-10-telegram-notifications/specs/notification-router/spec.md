## ADDED Requirements

### Requirement: Router receives notification events from SQS queue
The system SHALL have a `notify-runtime` Lambda that receives notification events from the `notification-queue` SQS queue.

#### Scenario: SQS event triggers router
- **WHEN** a message appears in `notification-queue`
- **THEN** `notify-runtime` Lambda is invoked with the event payload

### Requirement: Router fans out to channel senders based on event type
The system SHALL route notification events to the appropriate `NotificationSender` implementation based on channel type.

#### Scenario: Route to Telegram sender
- **WHEN** a notification event's channel type is `telegram`
- **THEN** router invokes `TelegramSender.Send()` with the notification payload

### Requirement: Router reads channel config from DynamoDB
The system SHALL read channel configuration from DynamoDB to determine sender and delivery details.

#### Scenario: Lookup channel for monitor
- **WHEN** router receives a notification event for a monitor
- **THEN** it queries DynamoDB for `MonitorNotificationLink` entries for that monitor
- **AND** for each linked channel, retrieves the full `NotificationChannel` entity

### Requirement: Router applies tenant default if no explicit links
The system SHALL use the tenant's default notification channel when a monitor has no explicit channel links.

#### Scenario: No explicit links, default exists
- **WHEN** a monitor has no `MonitorNotificationLink` entries
- **AND** the tenant has a default notification channel (`isDefault: true`)
- **THEN** router routes notification to the default channel

### Requirement: Router skips disabled channels
The system SHALL NOT send notifications to channels that are disabled.

#### Scenario: Channel is disabled
- **WHEN** a linked or default channel has `enabled: false`
- **THEN** router skips that channel and logs the skip

### Requirement: Router sends pre-formatted message to sender
The system SHALL pass a pre-formatted human-readable message to the channel sender.

#### Scenario: Send message
- **WHEN** router has resolved the target channel and notification details
- **THEN** it calls `sender.Send(ctx, notification)` where `notification.Message` is the formatted text
