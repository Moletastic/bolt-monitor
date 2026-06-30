## 1. Infrastructure

- [x] 1.1 Add `notification-queue` SQS queue to SST bootstrap stack
- [x] 1.2 Add `notify-runtime` Lambda to SST bootstrap stack (SQS-triggered)
- [x] 1.3 Add DynamoDB entity types for `NotificationChannel` and `MonitorNotificationLink` to schema

## 2. Shared Notification Interfaces

- [x] 2.1 Create `shared/notifications/notification.go` with `Notification` struct
- [x] 2.2 Create `shared/notifications/sender.go` with `NotificationSender` interface
- [x] 2.3 Create `shared/notifications/router.go` with `NotificationRouter` that fans out to senders

## 3. DynamoDB Repository

- [x] 3.1 Create `shared/notifications/repository.go` with `NotificationChannelRepository` interface
- [x] 3.2 Implement `DynamoNotificationChannelRepository` with Get, List, Create, Update, Delete
- [x] 3.3 Implement `MonitorNotificationLinkRepository` with GetByMonitor, CreateLink, DeleteLink, UpdateLink

## 4. API Endpoints

- [x] 4.1 Add `POST /api/v1/notification-channels` endpoint
- [x] 4.2 Add `GET /api/v1/notification-channels` endpoint
- [x] 4.3 Add `PATCH /api/v1/notification-channels/{channelId}` endpoint
- [x] 4.4 Add `DELETE /api/v1/notification-channels/{channelId}` endpoint
- [x] 4.5 Add `POST /api/v1/notification-channels/{channelId}/test` endpoint
- [x] 4.6 Add `PUT /api/v1/services/{serviceId}/monitors/{monitorId}/notification-channels` endpoint
- [x] 4.7 Add `GET /api/v1/services/{serviceId}/monitors/{monitorId}/notification-channels` endpoint
- [x] 4.8 Add `POST /api/v1/notification-channels/{channelId}/detect-chat-id` endpoint (Telegram auto-detect)

## 5. Telegram Sender

- [x] 5.1 Create `shared/notifications/telegram.go` implementing `NotificationSender`
- [x] 5.2 Implement `Send()` using Telegram Bot API `sendMessage` endpoint
- [x] 5.3 Implement `ValidateConfig()` for bot token validation
- [x] 5.4 Implement `ChannelType()` returning `"telegram"`
- [x] 5.5 Add MarkdownV2 escaping for message text
- [x] 5.6 Support `disable_notification` flag from channel config

## 6. Notify Runtime Lambda

- [x] 6.1 Implement SQS event handler in `services/notify-runtime/`
- [x] 6.2 Implement router that reads `MonitorNotificationLink` entries for target monitor
- [x] 6.3 Implement router fallback to tenant default channel when no explicit links
- [x] 6.4 Implement router skip of disabled channels
- [x] 6.5 Wire Telegram sender into router's sender map

## 7. Check Runtime Integration

- [x] 7.1 Modify `RecordExecutionResult()` to detect state transitions
- [x] 7.2 Add SQS enqueue call after writing incident when state transition occurs
- [x] 7.3 Enqueue `incident.opened` event when UP→DOWN transition
- [x] 7.4 Enqueue `incident.resolved` event when DOWN→UP transition
- [x] 7.5 Skip enqueue when state is unchanged or monitor is in maintenance

## 8. Dashboard UI

- [x] 8.1 Add notification channels list page
- [x] 8.2 Add Telegram channel creation form (bot token input, name)
- [x] 8.3 Add "Detect Chat ID" button and flow
- [x] 8.4 Add channel enable/disable toggle
- [x] 8.5 Add channel deletion with link check
- [x] 8.6 Add monitor notification channel linking UI
- [x] 8.7 Add "Test Notification" button per channel
- [x] 8.8 Add default channel designation toggle