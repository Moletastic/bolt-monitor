## Why

The system currently creates incidents internally when checks fail but provides no way to alert operators. Telegram is the first notification channel to be implemented, establishing a multi-channel notification architecture that future channels (email, Slack, etc.) will also use.

## What Changes

- New `NotificationChannel` entity stored in DynamoDB (per tenant)
- New `MonitorNotificationLink` join table (channels linked to monitors, with event filter)
- New `notify-runtime` Lambda triggered by SQS `notification-queue`
- New `NotificationSender` interface in `shared/` for multi-channel support
- Telegram implementation of the interface
- API endpoints for channel CRUD and monitor-channel linking
- Dashboard UI for configuring Telegram channels and linking to monitors
- Test notification button that sends a test message via Telegram
- Default/global notification channel per tenant
- Auto-detection of Telegram chat ID when user messages the bot

## Capabilities

### New Capabilities

- `notification-channel`: Configure and manage notification channels (e.g., Telegram bots) per tenant. Supports enable/disable, CRUD operations, and per-channel config encrypted at rest.
- `notification-router`: Route notification events from the check runtime to appropriate channel senders. Reads channel config from DynamoDB, fans out to channel-specific senders.
- `telegram-sender`: Telegram-specific notification sender using bot token and chat ID. Implements `NotificationSender` interface.
- `monitor-notification-link`: Associate notification channels with monitors, with per-channel event filters (`incident.opened`, `incident.resolved`).
- `notification-test`: Send a test notification through a configured channel to verify setup.
- `default-notification-channel`: Tenant-level default channel applied to all monitors unless overridden.

### Modified Capabilities

- `monitor-check-execution`: The check runtime will enqueue a notification event to SQS when a check run results in a state transition (UP→DOWN or DOWN→UP), rather than writing an incident silently.

## Impact

- **New Lambda**: `notify-runtime` (SQS-triggered)
- **New DynamoDB entities**: `NotificationChannel`, `MonitorNotificationLink`
- **New SQS queue**: `notification-queue`
- **New API endpoints**:
  - `POST /api/v1/notification-channels`
  - `GET /api/v1/notification-channels`
  - `DELETE /api/v1/notification-channels/{channelId}`
  - `PATCH /api/v1/notification-channels/{channelId}`
  - `POST /api/v1/notification-channels/{channelId}/test`
  - `PUT /api/v1/services/{serviceId}/monitors/{monitorId}/notification-channels`
  - `GET /api/v1/services/{serviceId}/monitors/{monitorId}/notification-channels`
- **New shared module**: `shared/notifications/` with interface definitions
- **Dashboard**: Integrations page extended with Telegram channel management UI
