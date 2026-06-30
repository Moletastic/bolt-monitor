## REMOVED Requirements

### Requirement: System routes notifications through channel configuration
**Reason**: The notification channel model (NotificationChannel records, MonitorNotificationLink, NotificationRouter, SenderRegistry) is replaced entirely by escalation policy-driven notification execution. All notification routing is now determined by escalation policy steps, not by per-monitor channel links.

**Migration**: Operators must create escalation policies and assign them to services before incidents can generate notifications. The notify-runtime Lambda is repurposed to receive escalation step events rather than direct incident events.

### Requirement: System supports per-monitor notification channel links
**Reason**: Per-monitor notification links are removed. Notification targeting is now determined by the escalation policy assigned to the parent service and the escalation path (business-hours or off-hours) selected at runtime.

**Migration**: Configure escalation policies and service bindings before creating monitors. Notification routing is service-scoped, not monitor-scoped.

### Requirement: System emits incident.opened and incident.resolved notification events directly
**Reason**: check-runtime no longer emits incident.opened or incident.resolved notification events. These events are replaced by escalation-driven notifications. check-runtime emits incident.down (triggers escalation start) and incident.up (triggers escalation suppression) events instead.

**Migration**: Ensure escalation policies are configured for all services before creating monitors. The notify-runtime receives escalation step events and fires the appropriate channel notifications at each step.

### Requirement: System supports default tenant notification channel
**Reason**: Default channel fallback is removed. If a service has no escalation policy, no notification is sent.

**Migration**: Assign escalation policies to all services to ensure notifications are sent.

## ADDED Requirements

### Requirement: System supports multi-channel notification sending
System SHALL support sending notifications through multiple channel types: telegram, email, sms, webhook, pagerduty.

#### Scenario: Telegram notification is sent
- **WHEN** an escalation step fires with channel type telegram
- **THEN** system sends a Telegram message to the configured target using the notify-runtime Lambda

#### Scenario: Email notification is sent
- **WHEN** an escalation step fires with channel type email
- **THEN** system sends an email to the configured email address using the notify-runtime Lambda

#### Scenario: SMS notification is sent
- **WHEN** an escalation step fires with channel type sms
- **THEN** system sends an SMS to the configured phone number using the notify-runtime Lambda

#### Scenario: Webhook notification is sent
- **WHEN** an escalation step fires with channel type webhook
- **THEN** system POSTs incident data to the configured webhook URL using the notify-runtime Lambda

#### Scenario: PagerDuty notification is sent
- **WHEN** an escalation step fires with channel type pagerduty
- **THEN** system creates a PagerDuty incident via their REST API integration using the notify-runtime Lambda
