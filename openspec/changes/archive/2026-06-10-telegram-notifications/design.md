## Context

The system currently creates incidents in DynamoDB when check runs fail, but no external notification is sent to operators. The check runtime (`check-runtime`) writes results and incidents synchronously with no async notification path.

Notification requirements established during exploration:
- Event-driven: notify only on state *changes* (UP→DOWN, DOWN→UP), not every check run
- Loose coupling: check runtime should not block on notification delivery
- Multi-channel ready: architecture must support future Email, Slack, etc. via interface
- Normalized channel model: channels shared across monitors (not embedded in Monitor)
- Telegram bot token + auto-detected chat ID setup flow
- Test notification button for verification
- Default/global notification channel per tenant

## Goals / Non-Goals

**Goals:**
- Telegram as the first notification channel
- Async, queue-based notification delivery via SQS
- Interface-driven channel sender abstraction for multi-channel future
- Normalized DynamoDB model for channels and monitor-channel links
- Per-monitor event filtering (`incident.opened`, `incident.resolved`)
- Tenant-level default notification channel
- Telegram chat ID auto-detection (user messages bot)
- Test notification capability

**Non-Goals:**
- Notification templating (v1 uses simple fixed messages)
- Push notification channels (mobile push, etc.)
- Notification aggregation or rate limiting (deferred)
- Per-channel retry policies (Lambda default retry is sufficient for v1)
- Notification preferences per user (channels are tenant-level for v1)

## Decisions

### Decision 1: Notification delivery via SQS-decoupled Lambda

**Choice:** `notify-runtime` Lambda triggered by `notification-queue` SQS queue.

**Rationale:** This follows the same pattern as `check-runtime` (scheduler/worker split). The check runtime enqueues a notification event and returns immediately — notification delivery is fire-and-forget. This keeps check latency unaffected by Telegram API latency.

**Alternatives considered:**
- Inline in `check-runtime` worker: rejected — synchrony blocks check loop, adds latency
- Direct SQS→Lambda without custom Lambda: rejected — router needs to read channel config from DynamoDB to route to correct sender

### Decision 2: Router reads channel config from DynamoDB (not worker)

**Choice:** Worker enqueues one event per monitor with state transition. Router Lambda reads DynamoDB to find which channels are linked to that monitor.

**Rationale:** This provides loose coupling — monitor-channel topology changes don't require re-enqueueing. The router is the fan-out point that owns "which channels for this monitor."

**Alternatives considered:**
- Worker enqueues N events (one per channel): rejected — doubles SQS messages, topology changes require re-enqueue
- Worker writes notification event with channel IDs embedded: rejected — makes worker aware of routing logic

### Decision 3: Normalized DynamoDB model for channels

**Choice:** `NotificationChannel` entity (per tenant) and `MonitorNotificationLink` join table (per monitor per channel).

**Rationale:** Channels are shared infrastructure — one Telegram bot can notify for many monitors. The join table allows per-monitor event filtering without duplicating channel config.

**DynamoDB key patterns:**
```
NotificationChannel:
 PK: TENANT#<tenantId>
  SK: CHANNEL#<channelId>

MonitorNotificationLink:
  PK: MONITOR#<tenantId>#<serviceId>#<monitorId>
  SK: CHANNEL#<channelId>
```

### Decision 4: `NotificationSender` interface in `shared/notifications/`

**Choice:** Define interface in `shared/` so both `notify-runtime` and future channel senders live there.

```go
type NotificationSender interface {
    Send(ctx context.Context, notification Notification) error
    ChannelType() string
    ValidateConfig(config JSON) error
}

type Notification struct {
    EventType   string    // "incident.opened" | "incident.resolved"
    MonitorID   string
    ServiceID   string
    TenantID    string
    MonitorName string
    ServiceName string
    Timestamp  time.Time
    Message    string    // pre-formatted, simple text
}
```

**Rationale:** Go interfaces define contracts without inheritance complexity. Telegram sender, Email sender, Slack sender all implement `Send()`. Router holds a map `channelType → sender` and dispatches by type.

### Decision 5: Simple messages for v1 (no templating)

**Choice:** Notifications are fixed-format strings, not templates.

**Message format:**
```
🚨 Incident Opened: <monitorName> is DOWN
Service: <serviceName>
URL: <monitorURL>
Error: <errorMessage>
Time: <timestamp>
```

```
✅ Incident Resolved: <monitorName> is UP
Service: <serviceName>
URL: <monitorURL>
Duration: <downtime>
Time: <timestamp>
```

**Rationale:** Simplicity for v1. Template rendering (like Uptime-Kuma's Liquid engine) adds complexity. Future capability spec will address templating.

### Decision 6: Telegram chat ID auto-detection

**Choice:** When user creates a Telegram channel, they provide bot token. System discovers chat ID via Telegram API `getUpdates` after user sends a message to the bot.

**Flow:**
1. User creates bot via BotFather, gets bot token
2. User pastes token in dashboard, saves channel (chat_id empty initially)
3. User messages the bot directly in Telegram
4. User clicks "Detect Chat ID" button in dashboard → backend calls Telegram `getUpdates`, extracts chat ID, updates channel
5. Channel is now ready to send notifications

**Rationale:** Follows the exact UX described in exploration. No polling infrastructure needed — discovery is on-demand.

### Decision 7: Default/global notification channel per tenant

**Choice:** `NotificationChannel` entity has an `isDefault` boolean. The router applies the tenant default to any monitor without explicit links.

**Rationale:** Reduces friction — user sets up one Telegram channel, all monitors notify through it by default unless overridden.

## Risks / Trade-offs

- **[Risk] Telegram API rate limits** → Lambda retry handles transient failures. For v1, no explicit rate limit handling.
- **[Risk] Chat ID not detected before first notification** → API validates `chatId` is non-empty before sending. Test button helps verify setup.
- **[Risk] DynamoDB read on every notification event** → Router reads monitor's channel links from DynamoDB. Acceptable for v1; consider caching in future.
- **[Trade-off] SQS adds eventual consistency** → Notification may arrive seconds after incident opens. Acceptable for alerting use case.
- **[Trade-off] Go interface per channel** → Adding a new channel requires implementing `Send()` in a new file. No runtime discovery — follows existing Go patterns in this repo.

## Migration Plan

1. **Deploy infrastructure first**: Add `notification-queue` SQS, `notify-runtime` Lambda (stub), new DynamoDB entity types. No code changes to existing runtime.
2. **Implement in order**: shared interfaces → DynamoDB repository → API endpoints → notify-runtime → dashboard UI.
3. **Verify with test button**: Before routing real incidents, test notification confirms bot token + chat ID work.
4. **Toggle via feature flag or monitor-level setting**: Existing monitors unaffected until explicitly linked to a channel.

## Open Questions

1. **Encryption of bot token**: Should bot token be encrypted at rest in DynamoDB? (Strongly recommended — tokens are secrets)
2. **Channel deletion with links**: What happens when a user deletes a channel that has monitor links? Cascade delete links or reject deletion?
3. **Notification ordering**: If a monitor flaps UP→DOWN→UP quickly, could notifications arrive out of order? Should the router handle deduplication?
4. **Incident acknowledged state**: Should `incident.acknowledged` trigger a notification? (User explicitly took ownership — may want to suppress subsequent DOWN→UP for same incident)
